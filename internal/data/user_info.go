package data

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"

	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/kit/sqlx"
)

var _ biz.UserInfoRepo = (*UserInfo)(nil)

type UserInfo struct {
	data *Data
	l    *log.Helper
}

func NewUserInfo(
	d *Data,
	l log.Logger,
) biz.UserInfoRepo {
	return &UserInfo{
		data: d,
		l:    log.NewHelper(l),
	}
}

// GetUserCreateGroupByHellpCode implements biz.UserInfoRepo.
func (u *UserInfo) GetUserCreateGroupByHellpCode(ctx context.Context, helpCode string) (*biz.UserCreateGroup, error) {
	userCreateGroup, err := model.GetUserCreateGroupByHelpCode(ctx, u.data.db, helpCode)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).
				Errorf("get user create group failed, err=%v, helpCode=%s", err, helpCode)
		}
		return nil, err
	}

	return ConvertUserCreateGroup2Biz(userCreateGroup), nil
}

// GetUserInfo implements biz.UserInfoRepo.
func (u *UserInfo) GetUserInfo(ctx context.Context, waID string) (*biz.UserInfo, error) {
	userInfo, err := model.GetUserInfo(ctx, u.data.db, waID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).Infof("get user info failed, err=%v, waID=%s", err, waID)
		}
		return nil, err
	}
	return convertUserInfo2Biz(userInfo), nil
}

func (u *UserInfo) GetUserInfoByHelpCode(ctx context.Context, helpCode string) (*biz.UserInfo, error) {
	userInfo, err := model.GetUserInfoByHelpCode(ctx, u.data.db, helpCode)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).Errorf("GetUserInfoByHelpCode failed, err=%v, helpCode=%s", err, helpCode)
		}
		return nil, err
	}

	return convertUserInfo2Biz(userInfo), nil
}

func (u *UserInfo) ListUserInfoByHelpCodes(ctx context.Context, helpCode []string) ([]*biz.UserInfo, error) {
	userInfos, err := model.ListUserInfoByHelpCodes(ctx, u.data.db, helpCode)
	if err != nil {
		return nil, err
	}

	return lo.Map(userInfos, func(userInfo *model.UserInfo, _ int) *biz.UserInfo {
		return convertUserInfo2Biz(userInfo)
	}), nil
}

func (u *UserInfo) UpdateUserInfoCdkByWaId(ctx context.Context, waID string, cdk string, cdkType int) error {
	return model.UpdateUserInfoCdkByWaId(ctx, u.data.db, waID, cdk, cdkType)
}

func (u *UserInfo) FindUserInfos(ctx context.Context, waIDs []string) ([]*biz.UserInfo, error) {
	userInfos, err := model.SelectUserInfosByWaIDs(ctx, u.data.db, waIDs)
	if err != nil {
		return nil, err
	}

	return lo.Map(userInfos, func(userInfo *model.UserInfo, _ int) *biz.UserInfo {
		return convertUserInfo2Biz(userInfo)
	}), nil
}

func (u *UserInfo) ListGtIdLtEndTime(ctx context.Context, minId int64, endTime time.Time, limit int) ([]*biz.UserInfo, error) {
	userInfos, err := model.ListUserInfoGtIdLtEndTime(ctx, u.data.db, minId, endTime, uint(limit))
	if err != nil {
		return nil, err
	}

	var bizUserInfos []*biz.UserInfo
	for _, userInfo := range userInfos {
		bizUserInfos = append(bizUserInfos, convertUserInfo2Biz(userInfo))
	}
	return bizUserInfos, nil
}
