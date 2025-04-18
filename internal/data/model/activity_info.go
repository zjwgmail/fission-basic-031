package model

import (
	"context"
	"database/sql"
	"errors"
	"fission-basic/api/constants"
	"time"

	"fission-basic/kit/sqlx"
)

const tableActivityInfo = `activity_info`

// ActivityInfo 对应 activity_info 表的结构体
type ActivityInfo struct {
	Id             string        `db:"id"`
	ActivityName   string        `db:"activity_name"`
	ActivityStatus string        `db:"activity_status"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
	StartAt        sql.NullTime  `db:"start_at"`
	EndAt          sql.NullTime  `db:"end_at"`
	EndBufferDay   sql.NullInt64 `db:"end_buffer_day"`
	EndBufferAt    sql.NullTime  `db:"end_buffer_at"`
	ReallyEndAt    sql.NullTime  `db:"really_end_at"`

	CostMax float64 `db:"cost_max"`
}

func GetActivityInfo(ctx context.Context, db sqlx.DB, id string) (*ActivityInfo, error) {

	var wms ActivityInfo
	err := sqlx.GetContext(ctx, db, &wms, tableActivityInfo, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return nil, err
	}

	return &wms, nil
}

func UpdateActivityStatus(ctx context.Context, db sqlx.DB, msg *ActivityInfo) error {
	if "" == msg.Id || "" == msg.ActivityStatus {
		return errors.New("activity id or status should not be empty")
	}

	activityInfo, err := GetActivityInfo(ctx, db, msg.Id)
	if err != nil {
		return err
	}
	if "" == activityInfo.Id {
		return errors.New("activity id don't exists")
	}

	where := map[string]interface{}{
		"id": msg.Id,
	}

	update := map[string]interface{}{
		"activity_status": msg.ActivityStatus,
	}

	if constants.ATStatusBuffer == msg.ActivityStatus {
		// 缓冲期的时候，需要计算最终截止时间
		now := time.Now()
		reallyEndTime := now.Add(time.Duration(activityInfo.EndBufferDay.Int64*24) * time.Hour)

		update["end_buffer_at"] = now
		update["really_end_at"] = reallyEndTime
	}

	return sqlx.UpdateContext(ctx, db, tableActivityInfo, where, update)
}
