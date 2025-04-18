package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"

	"github.com/didi/gendry/builder"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNoRows       = sql.ErrNoRows
	ErrTxDone       = sql.ErrTxDone
	ErrRowsAffected = errors.New("RowsAffected Error") // 当数据库操作没改变数据时，返回此错误
	ErrDuplicate    = &mysql.MySQLError{
		Number: 1062,
		// SQLState: [5]byte{'2', '7', '0', '0', '0'},
		// Message: "Duplicate entry '(.*)' for key '(.*)'",
	}
)

func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrDuplicate)
}

type (
	Rows   = sqlx.Rows
	Stmt   = sql.Stmt
	Result = sql.Result
)

type DB interface {
	Rebind(query string) string
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (Result, error)
	NamedExec(query string, arg interface{}) (Result, error)
	Queryx(query string, args ...interface{}) (*Rows, error)
	Prepare(query string) (*Stmt, error)

	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (Result, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	PrepareContext(ctx context.Context, query string) (*Stmt, error)
}

type Config struct {
	DriverName      string
	Server          string
	MaxIdle         int64
	MaxOpen         int64
	ConnMaxLifetime time.Duration
}

func Open(cfg *Config) (DB, error) {
	if cfg.DriverName == "" {
		// 目前只能用mysql
		cfg.DriverName = "mysql"
	}

	if cfg.ConnMaxLifetime < 30*time.Second {
		cfg.ConnMaxLifetime = 30 * time.Second
	}

	log.Infof("open db success,driverName:%v,server:%v,maxIdle:%v,maxOpen:%v,connMaxLifetime:%v", cfg.DriverName, cfg.Server, cfg.MaxIdle, cfg.MaxOpen, cfg.ConnMaxLifetime)
	d, err := sqlx.Connect(cfg.DriverName, cfg.Server)
	if err != nil {
		log.Errorf("open db error,err:%v", err)
		return nil, err
	}

	d.SetMaxIdleConns(int(cfg.MaxIdle))
	d.SetConnMaxIdleTime(cfg.ConnMaxLifetime)
	d.SetMaxOpenConns(int(cfg.MaxOpen))
	d.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return d, nil
}

func ExeSql(ctx context.Context, db DB, sql string, args ...interface{}) (Result, error) {
	exec, err := db.Exec(sql, args...)
	if err != nil {
		return nil, err
	}
	return exec, nil
}

func InsertIgnoreContext(ctx context.Context, db DB, table string, i interface{}) (int64, error) {
	r, err := db.NamedExecContext(ctx, BuildInsertIgnoreSQL(table, GetFields(i)...), i)
	if err != nil {
		return 0, err
	}

	return r.LastInsertId()
}

func InsertContext(ctx context.Context, db DB, table string, i interface{}) (int64, error) {
	r, err := db.NamedExecContext(ctx, BuildInsertSQL(table, GetFields(i)...), i)
	if err != nil {
		return 0, err
	}

	return r.LastInsertId()
}

func UpdateContext(ctx context.Context, db DB, table string, where, update map[string]interface{}) error {
	query, args, err := builder.BuildUpdate(table, where, update)
	if err != nil && !errors.Is(err, ErrRowsAffected) {
		return err
	}

	return ExecContext(ctx, db, query, args...)
}

func UpdateContextIgnore(ctx context.Context, db DB, table string, where, update map[string]interface{}) error {
	query, args, err := builder.BuildUpdate(table, where, update)
	if err != nil {
		return err
	}

	return ExecContextIgnore(ctx, db, query, args...)
}

func BulkInsertContext(ctx context.Context, db DB, table string, args ...interface{}) (err error) {
	names, tags := getNameAndTags(args[0])
	fieldStr := strings.Join(tags, ",")
	valueStr := "(?" + strings.Repeat(",?", len(tags)-1) + ")"
	valueStr += strings.Repeat(","+valueStr, len(args)-1)
	values := make([]interface{}, 0, len(args)*len(tags))
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for _, name := range names {
			values = append(values, v.FieldByName(name).Interface())
		}
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("INSERT INTO `%s`(%s) VALUES%s", table, fieldStr, valueStr), values...)
	return
}

func BulkInsertIgnoreContext(ctx context.Context, db DB, table string, args []interface{}) (err error) {
	names, tags := getNameAndTags(args[0])
	fieldStr := strings.Join(tags, ",")
	valueStr := "(?" + strings.Repeat(",?", len(tags)-1) + ")"
	valueStr += strings.Repeat(","+valueStr, len(args)-1)
	values := make([]interface{}, 0, len(args)*len(tags))
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for _, name := range names {
			values = append(values, v.FieldByName(name).Interface())
		}
	}

	_, err = db.ExecContext(ctx, fmt.Sprintf("INSERT IGNORE INTO `%s`(%s) VALUES%s", table, fieldStr, valueStr), values...)
	return
}

func GetContext(ctx context.Context, db DB, dest interface{}, table string, where map[string]interface{}, fields ...string) error {
	if len(fields) == 0 {
		fields = GetFields(dest)
	}

	query, args, err := builder.BuildSelect(table, where, fields)
	if err != nil {
		return err
	}

	return db.GetContext(ctx, dest, query, args...)
}

func SelectContext(ctx context.Context, db DB, dest interface{}, table string, where map[string]interface{}, fields ...string) error {
	if len(fields) == 0 {
		fields = GetFields(dest)
	}

	query, args, err := builder.BuildSelect(table, where, fields)
	if err != nil {
		return err
	}

	return db.SelectContext(ctx, dest, query, args...)
}

func DeleteContext(ctx context.Context, db DB, table string, where map[string]interface{}) error {
	query, args, err := builder.BuildDelete(table, where)
	if err != nil {
		return err
	}

	return ExecContext(ctx, db, query, args...)
}

func DeleteContextIgnore(ctx context.Context, db DB, table string, where map[string]interface{}) error {
	query, args, err := builder.BuildDelete(table, where)
	if err != nil {
		return err
	}

	return ExecContextIgnore(ctx, db, query, args...)
}

func ExecContextIgnore(ctx context.Context, db DB, query string, args ...interface{}) error {
	// 执行SQL
	_, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func ExecContext(ctx context.Context, db DB, query string, args ...interface{}) error {
	// 执行SQL
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// 检查影响的结果
	num, err := res.RowsAffected()
	if err == nil && num == 0 {
		err = ErrRowsAffected
		return err
	}

	return nil
}

func TxContext(ctx context.Context, db DB, fn func(context.Context, DB) error) error {
	if x, ok := db.(interface{ Beginx() (*sqlx.Tx, error) }); ok {
		tx, err := x.Beginx()
		if err != nil {
			return err
		}
		if err = fn(ctx, tx); err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				err = fmt.Errorf("%w: %v", err, err2)
			}
		} else {
			err = tx.Commit()
		}
		return err
	}

	return fn(ctx, db)
}
