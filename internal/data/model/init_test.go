package model

import (
	"context"
	"testing"

	"fission-basic/kit/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

var db sqlx.DB

func init() {
	b, err := sqlx.Open(&sqlx.Config{
		DriverName: "mysql",
		Server:     ``,
	})
	if err != nil {
		panic(err)
	}

	db = b
}

func TestInit(t *testing.T) {
	ctx := context.Background()
	err := InitDB(ctx, db)
	if err != nil {
		t.Error(err)
	}
}
