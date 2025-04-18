package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tablePushEvent4User = "push_event4_user"

type PushEvent4User struct {
	WaId string `db:"wa_id"`
}

func PushEvent4UserInsertIgnore(ctx context.Context, db sqlx.DB, waId string) (int64, error) {
	if waId == "" {
		return 0, nil
	}
	entity := &PushEvent4User{
		WaId: waId,
	}
	return sqlx.InsertIgnoreContext(ctx, db, tablePushEvent4User, entity)
}
