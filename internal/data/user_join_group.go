package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.UserJoinGroupRepo = (*UserJoinGroup)(nil)

type UserJoinGroup struct {
	data *Data
	l    *log.Helper
}

func NewUserJoinGroup(
	d *Data,
	l log.Logger,
) biz.UserJoinGroupRepo {
	return &UserJoinGroup{
		data: d,
		l:    log.NewHelper(l),
	}
}

func (u *UserJoinGroup) ListGtIdGtJoinGroupTime(ctx context.Context, id int64,
	startTimestamp int64, endTimestamp int64, limit int) ([]*biz.UserJoinGroup, error) {
	userJoinGroups, err := model.UserJoinGroupListGtIdBetweenJoinGroupTime(ctx, u.data.db, id, startTimestamp, endTimestamp, uint(limit))
	if err != nil {
		return nil, err
	}
	var bizUserJoinGroups []*biz.UserJoinGroup
	for _, userJoinGroup := range userJoinGroups {
		bizUserJoinGroups = append(bizUserJoinGroups, ConvertUserJoinGroup2Biz(userJoinGroup))
	}
	return bizUserJoinGroups, nil
}

func (u *UserJoinGroup) GetFirstLeJoinGroupTime(ctx context.Context, timestamp int64) (*biz.UserJoinGroup, error) {
	userJoinGroup, err := model.UserJoinGroupGetFirstGtJoinGroupTime(ctx, u.data.db, timestamp)
	if err != nil {
		return nil, err
	}
	return ConvertUserJoinGroup2Biz(userJoinGroup), nil
}
