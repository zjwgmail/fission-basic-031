package biz

import (
	"context"
	"time"
)

type UserInfoRepo interface {
	// 查询开团信息
	GetUserCreateGroupByHellpCode(ctx context.Context, helpCode string) (*UserCreateGroup, error)

	// 查询用户信息
	GetUserInfo(ctx context.Context, waID string) (*UserInfo, error)

	FindUserInfos(ctx context.Context, waIDs []string) ([]*UserInfo, error)

	// 查询用户信息通过助力码
	GetUserInfoByHelpCode(ctx context.Context, helpCode string) (*UserInfo, error)

	// 查询用户信息通过助力码
	ListUserInfoByHelpCodes(ctx context.Context, helpCode []string) ([]*UserInfo, error)

	// 更新用户cdk
	UpdateUserInfoCdkByWaId(ctx context.Context, waID string, cdk string, cdkType int) error

	// 查询用户信息列表
	ListGtIdLtEndTime(ctx context.Context, minId int64, endTime time.Time, limit int) ([]*UserInfo, error)
}
