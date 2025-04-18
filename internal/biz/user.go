package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type UserInfoUsecase struct {
	repo UserInfoRepo
	l    *log.Helper
}

func NewUserInfoUsecase(
	repo UserInfoRepo,
	l log.Logger,
) *UserInfoUsecase {
	return &UserInfoUsecase{
		l:    log.NewHelper(l),
		repo: repo,
	}
}

func (u *UserInfoUsecase) GetUserInfoByHelpCode(ctx context.Context, rallyCode string) (*UserInfo, error) {
	userInfo, err := u.repo.GetUserInfoByHelpCode(ctx, rallyCode)
	if err != nil {
		u.l.WithContext(ctx).Errorf("GetUserInfoByHelpCode failed, err=%v, rallyCode=%s", err, rallyCode)
		return nil, err
	}

	return userInfo, nil
}
