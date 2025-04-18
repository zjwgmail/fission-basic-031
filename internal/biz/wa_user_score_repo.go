package biz

import (
	"context"
	"time"
)

type WaUserScoreRepo interface {
	UpdateState(ctx context.Context, waId string, state int) error
	PageBySocialScore(ctx context.Context, limit uint, length uint) ([]*WaUserScoreDTO, error)
	PageByRecurringProb(ctx context.Context, limit uint, length uint) ([]*WaUserScoreDTO, error)
}

type WaUserScoreDTO struct {
	Id            int
	WaId          string
	LastLoginTime time.Time
	State         int
	SocialScore   int
	RecurringProb float64
}
