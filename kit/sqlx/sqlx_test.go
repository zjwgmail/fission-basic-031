package sqlx

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/didi/gendry/scanner"
	_ "github.com/go-sql-driver/mysql"
)

/*
CREATE TABLE `student` (
	`id` int NOT NULL AUTO_INCREMENT,
	`name` varchar(255),
	`created_at` datetime DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`)
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='学生';
*/

const tableStudent = `student`

type Student struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

var db DB

func init() {
	b, err := Open(&Config{
		DriverName: "mysql",
		// Server:     `root:@tcp(127.0.0.1:3306)/fission?charset=utf8mb4&parseTime=True&loc=Local`,
		Server: ``,
	})
	if err != nil {
		panic(err)
	}

	db = b
}

func TestQueryXContext(t *testing.T) {
	ctx := context.Background()
	rows, err := db.QueryxContext(ctx, "select * from user_info limit 1;")
	if err != nil {
		t.Error(err)
		return
	}

	defer rows.Close()

	row, err := scanner.ScanMapDecode(rows)
	if err != nil {
		t.Errorf("Failed to scan row: %v", err)
		return
	}

	b, err := json.Marshal(row)
	if err != nil {
		t.Errorf("Failed to marshal results to JSON: %v", err)
		return
	}

	t.Log(string(b))
}

func TestInsert(t *testing.T) {
	id, err := InsertContext(context.Background(), db, tableStudent, &Student{
		// ID:        1,
		Name:      "gee",
		CreatedAt: time.Now(),
	})

	if err != nil {
		if IsDuplicateError(err) {
			fmt.Println("唯一键冲突错误:", err)
		}
		panic(err)
	}

	t.Log(id)
}

func TestGet(t *testing.T) {
	var stu Student
	err := GetContext(context.Background(), db, &stu, tableStudent,
		map[string]interface{}{
			"id": 2,
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(stu)
}

func TestSelect(t *testing.T) {
	var stus []Student
	err := SelectContext(context.Background(), db, &stus, tableStudent,
		map[string]interface{}{
			"_limit": []uint{0, 10},
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(stus)
}

func BulkHelper(s ...*Student) []interface{} {
	var objects []interface{}
	for i := range s {
		objects = append(objects, s[i])
	}
	return objects
}

func TestBulkInsertContext(t *testing.T) {
	stus := []*Student{
		{
			Name:      "b3",
			CreatedAt: time.Now(),
		},
		{
			Name:      "b4",
			CreatedAt: time.Now(),
		},
	}

	BulkHelper(stus...)

	err := BulkInsertContext(context.Background(), db, tableStudent, BulkHelper(stus...)...)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("success")
}

func TestUpdateUserInfo(t *testing.T) {
	ctx := context.Background()
	err := UpdateContext(ctx, db, "user_info",
		map[string]interface{}{
			"id":         1,
			"join_count": 1,
		},
		map[string]interface{}{
			"join_count": 1,
		},
	)

	if err != nil {
		panic(err)
	}

	t.Log("success")
}

func TestUpdate(t *testing.T) {
	err := UpdateContext(context.Background(), db, tableStudent,
		map[string]interface{}{
			"id": 2,
		},
		map[string]interface{}{
			"name": "san.zhang",
		},
	)

	if err != nil {
		t.Error(err)
		return
	}
}

func TestTx(t *testing.T) {
	err := TxContext(context.Background(), db, func(ctx context.Context, tx DB) error {
		_, err := InsertContext(ctx, tx, tableStudent, &Student{
			Name:      "tx1",
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}

		_, err = InsertContext(ctx, tx, tableStudent, &Student{
			Name:      "tx2",
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestCreateTable(t *testing.T) {
	sql := `
CREATE TABLE student02 (
	id int NOT NULL AUTO_INCREMENT,
	name varchar(255),
	created_at datetime DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='学生';
`
	err := ExecContext(context.Background(), db, sql)
	if err != nil {
		t.Error(err)
	}
}
