package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tableFeishuReport = `feishu_report`

type FeishuReportEntity struct {
	Id             int64  `db:"id"`
	Date           string `db:"date"`
	Time           string `db:"time"`
	FirstCount     int    `db:"first_count"`
	FissionCount   int    `db:"fission_count"`
	CoverCount     int    `db:"cover_count"`
	CdkCount       string `db:"cdk_count"`
	FailedCount    int    `db:"failed_count"`
	TimeoutCount   int    `db:"timeout_count"`
	InterceptCount int    `db:"intercept_count"`
}

func InsertFeishuReport(ctx context.Context, db sqlx.DB, frEntity *FeishuReportEntity) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableFeishuReport, frEntity)
}
