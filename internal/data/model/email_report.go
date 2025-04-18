package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tableEmailReport = `email_report`

type EmailReportEntity struct {
	Id              int64  `db:"id"`
	Date            string `db:"date"`
	Utc             int    `db:"utc"`
	Language        string `db:"language"`
	Channel         string `db:"channel"`
	CountryCode     string `db:"country_code"`
	GenerationCount string `db:"generation_count"`
	DailyJoinCount  string `db:"daily_join_count"`
	TotalJoinCount  string `db:"total_join_count"`
	CountV3         int    `db:"count_v3"`
	CountV22        int    `db:"count_v22"`
	CountV36        int    `db:"count_v36"`
	SuccessCount    int    `db:"success_count"`
	FailedCount     int    `db:"failed_count"`
	TimeoutCount    int    `db:"timeout_count"`
	InterceptCount  int    `db:"intercept_count"`
}

func EmailReportSelect(ctx context.Context, db sqlx.DB, erEntity *EmailReportEntity) (*EmailReportEntity, error) {
	result := &EmailReportEntity{}
	err := sqlx.GetContext(ctx, db, result, tableEmailReport, map[string]interface{}{
		"date":         erEntity.Date,
		"utc":          erEntity.Utc,
		"channel":      erEntity.Channel,
		"country_code": erEntity.CountryCode,
		"language":     erEntity.Language,
	})
	return result, err
}

func EmailReportInsert(ctx context.Context, db sqlx.DB, erEntity *EmailReportEntity) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableEmailReport, erEntity)
}

func EmailReportUpdate(ctx context.Context, db sqlx.DB, erEntity *EmailReportEntity) error {
	where := map[string]interface{}{
		"date":         erEntity.Date,
		"utc":          erEntity.Utc,
		"channel":      erEntity.Channel,
		"country_code": erEntity.CountryCode,
		"language":     erEntity.Language,
	}
	update := map[string]interface{}{
		"generation_count": erEntity.GenerationCount,
		"daily_join_count": erEntity.DailyJoinCount,
		"total_join_count": erEntity.TotalJoinCount,
		"count_v3":         erEntity.CountV3,
		"count_v22":        erEntity.CountV22,
		"count_v36":        erEntity.CountV36,
		"success_count":    erEntity.SuccessCount,
		"failed_count":     erEntity.FailedCount,
	}
	return sqlx.UpdateContext(ctx, db, tableEmailReport, where, update)
}

func EmailReportList(ctx context.Context, db sqlx.DB, utc int) ([]*EmailReportEntity, error) {
	var list []*EmailReportEntity
	err := sqlx.SelectContext(ctx, db, &list, tableEmailReport, map[string]interface{}{
		"utc": utc,
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}
