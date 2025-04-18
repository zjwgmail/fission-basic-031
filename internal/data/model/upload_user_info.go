package model

import (
	"context"
	"fission-basic/kit/sqlx"
	"time"
)

const tableUploadUserInfo = `upload_user_info`

type UploadUserInfoEntity struct {
	Id           int       `db:"id"`
	PhoneNumber  string    `db:"phone_number"`
	LastSendTime time.Time `db:"last_send_time"`
	State        int       `db:"state"`
}

func UploadUserInfoInsertBatch(ctx context.Context, db sqlx.DB, uuis []*UploadUserInfoEntity) error {
	var objects []interface{}
	for i := range uuis {
		objects = append(objects, *uuis[i])
	}
	return sqlx.BulkInsertIgnoreContext(ctx, db, tableUploadUserInfo, objects)
}

func UploadUserInfoUpdateState(ctx context.Context, db sqlx.DB, phoneNumber string, state int) error {
	where := map[string]interface{}{
		"phone_number": phoneNumber,
	}
	return sqlx.UpdateContext(ctx, db, tableUploadUserInfo, where, map[string]interface{}{"state": state})
}

func UploadUserInfoListInNumber(ctx context.Context, db sqlx.DB, phoneNumberList []string) ([]*UploadUserInfoEntity, error) {
	var list []*UploadUserInfoEntity
	err := sqlx.SelectContext(ctx, db, &list, tableUploadUserInfo, map[string]interface{}{
		"phone_number in": phoneNumberList,
	})
	return list, err
}

func UploadUserInfoListGtIdWithState(ctx context.Context, db sqlx.DB, id int, state int, limit uint) ([]*UploadUserInfoEntity, error) {
	var list []*UploadUserInfoEntity
	err := sqlx.SelectContext(ctx, db, &list, tableUploadUserInfo, map[string]interface{}{
		"id > ":    id,
		"state":    state,
		"_orderby": "id",
		"_limit":   []uint{limit},
	})
	return list, err
}
