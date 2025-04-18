package biz

import (
	"context"
)

type FeishuReportRepo interface {
	AddFeishuReport(ctx context.Context, frParam *FeishuReportParam) (int64, error)
}

type FeishuReportParam struct {
	Date           string
	Time           string
	FirstCount     int
	FissionCount   int
	CoverCount     int
	CdkCount       string
	FailedCount    int
	TimeoutCount   int
	InterceptCount int
}
