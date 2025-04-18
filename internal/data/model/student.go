package model

import (
	"context"
	"time"

	"fission-basic/kit/sqlx"
)

const tableStudent = `student`

type Student struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func CreateStudentTable(ctx context.Context, db sqlx.DB) error {
	sql := `
CREATE TABLE IF NOT EXISTS student (
	id int NOT NULL AUTO_INCREMENT,
	name varchar(255),
	created_at datetime DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='学生';
`
	err := sqlx.ExecContext(ctx, db, sql, nil)
	return err
}

func DroptableStudent(ctx context.Context, db sqlx.DB) error {
	sql := `
DROP TABLE IF EXISTS student;
`
	err := sqlx.ExecContext(ctx, db, sql, nil)
	return err
}

func InsertStudent(ctx context.Context, db sqlx.DB, stu *Student) (int64, error) {
	return sqlx.InsertIgnoreContext(ctx, db, tableStudent, stu)
}

func GetStudentByName(ctx context.Context, db sqlx.DB, name string) (*Student, error) {
	where := map[string]interface{}{
		"name": name,
	}

	var stu Student
	err := sqlx.GetContext(ctx, db, &stu, tableStudent, where)
	if err != nil {
		return nil, err
	}

	return &stu, nil
}

func SelectStudents(ctx context.Context, db sqlx.DB, offset, length uint) ([]Student, error) {
	where := map[string]interface{}{
		"_limit": []uint{offset, length},
	}

	var stus []Student
	err := sqlx.SelectContext(ctx, db, &stus, tableStudent, where)
	if err != nil {
		return nil, err
	}

	return stus, nil
}

func DeleteStudent(ctx context.Context, db sqlx.DB, id int64) error {
	where := map[string]interface{}{
		"id": id,
	}

	return sqlx.DeleteContext(ctx, db, tableStudent, where)
}

func CountStudent(ctx context.Context, db sqlx.DB) (int64, error) {
	var total int64
	err := sqlx.GetContext(ctx, db, &total, tableStudent, map[string]interface{}{}, "count(*)")
	if err != nil {
		return 0, err
	}
	return total, nil
}
