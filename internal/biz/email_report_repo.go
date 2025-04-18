package biz

import "context"

type EmailReportRepo interface {
	AddBatchEmailReport(ctx context.Context, list []*EmailReportDTO, utc int) (int, error)
	ListAllEmailReport(ctx context.Context, utc int) ([]*EmailReportDTO, error)
}

type EmailReportDTO struct {
	Date            string
	Utc             string
	Language        string
	Channel         string
	CountryCode     string
	GenerationCount [7]int
	DailyJoinCount  [16]int
	TotalJoinCount  [16]int
	CountV3         int
	CountV22        int
	CountV36        int
	SuccessCount    int
	FailedCount     int
	TimeoutCount    int
	InterceptCount  int
}
