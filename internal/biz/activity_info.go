package biz

import (
	"context"
	"fission-basic/internal/conf"
	"fission-basic/internal/util"

	"github.com/go-kratos/kratos/v2/log"
)

type ActivityInfoUsecase struct {
	repo      ActivityInfoRepo
	l         *log.Helper
	bootstrap *conf.Bootstrap
}

func NewActivityInfoUsecase(repo ActivityInfoRepo, l log.Logger, bootstrap *conf.Bootstrap) *ActivityInfoUsecase {
	return &ActivityInfoUsecase{
		repo:      repo,
		l:         log.NewHelper(l),
		bootstrap: bootstrap,
	}
}

// UpdateActivityInfo 更新
//
//	 ActivityStatus 枚举
//	 ATStatusUnStart = "unstart"
//		ATStatusStarted = "started"
//		ATStatusBuffer  = "buffer"
//		ATStatusEnd     = "end"
func (hc *ActivityInfoUsecase) UpdateActivityInfo(ctx context.Context, update *UpdateActivityInfoDto) error {

	err := hc.repo.UpdateActivityInfo(ctx, update)

	if err != nil {
		return err
	}

	return nil
}

func (hc *ActivityInfoUsecase) GetActivityInfo(ctx context.Context) (*ActivityInfoDto, error) {
	cost := util.MethodCost(ctx, hc.l, "ActivityInfoUsecase.GetActivityInfo")
	defer cost()

	activityInfo, err := hc.repo.GetActivityInfo(ctx, hc.bootstrap.Business.Activity.Id)

	if err != nil {
		return nil, err
	}

	return activityInfo, nil
}
