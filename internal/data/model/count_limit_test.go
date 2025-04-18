package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"testing"
)

// var db sqlx.DB

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

func TestCountAddOne(t *testing.T) {
	err := CountLimitAddOne(context.Background(), db, "test")
	if err != nil {
		panic(err)
	}
}

func TestInsertCountLimit(t *testing.T) {
	err := CountLimitInsert(context.Background(), db, "test")
	if err != nil {
		panic(err)
	}
}

func TestCountLimitGet(t *testing.T) {
	count, err := CountLimitGet(context.Background(), db, "test")
	if err != nil {
		panic(err)
	}
	t.Log(count)
}
