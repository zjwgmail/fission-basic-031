package biz

import (
	"context"
	"time"
)

type ActivityInfoRepo interface {
	GetActivityInfo(ctx context.Context, id string) (*ActivityInfoDto, error)
	UpdateActivityInfo(ctx context.Context, stu *UpdateActivityInfoDto) error
}

type ActivityInfoDto struct {
	Id             string
	ActivityName   string
	ActivityStatus string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	StartAt        time.Time
	EndAt          time.Time
	EndBufferDay   int
	EndBufferAt    time.Time
	ReallyEndAt    time.Time
	CostMax        float64
}

type UpdateActivityInfoDto struct {
	Id             string
	ActivityStatus string
}
