package biz

import "context"

type UserJoinGroupRepo interface {
	ListGtIdGtJoinGroupTime(ctx context.Context, id int64, startTimestamp int64, endTimestamp int64, limit int) ([]*UserJoinGroup, error)
	GetFirstLeJoinGroupTime(ctx context.Context, timestamp int64) (*UserJoinGroup, error)
}
