package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"fission-basic/api/constants"
	"fission-basic/internal/biz"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/util"
)

type UserRemindService struct {
	l                 *log.Helper
	userRemindUsecase *biz.UserRemindUsecase

	// FIXME: 将来某一天
	helpCodeService    *HelpCodeService
	activityUsecase    *biz.ActivityInfoUsecase
	redisClusterClient *redis.ClusterClient
}

func NewUserRemindService(
	l log.Logger,
	userRemindUsecase *biz.UserRemindUsecase,
	helpCodeService *HelpCodeService,
	activityUsecase *biz.ActivityInfoUsecase,
	redisClusterClient *redis.ClusterClient,
) *UserRemindService {
	return &UserRemindService{
		l:                  log.NewHelper(l),
		userRemindUsecase:  userRemindUsecase,
		helpCodeService:    helpCodeService,
		activityUsecase:    activityUsecase,
		redisClusterClient: redisClusterClient,
	}
}

func (u *UserRemindService) RemindJoinGroupV3(ctx context.Context) error {
	cost := util.MethodCost(ctx, u.l, "UserRemindService.RemindJoinGroupV3")
	defer cost()

	locked, unlock, err := redis.JobLock(ctx, u.redisClusterClient, "RemindJoinGroupV3", 12*time.Minute)
	if err != nil || !locked {
		u.l.WithContext(ctx).Infof("RemindJoinGroupV3 lock failed, err=%v, locked=%v", err, locked)
		return err
	}
	defer unlock()

	activity, err := u.activityUsecase.GetActivityInfo(ctx)
	if err != nil {
		u.l.WithContext(ctx).Errorf("get activity info failed, err=%v", err)
		// 未知活动状态时不打扰用户，不再处理
		return err
	}

	if activity.ActivityStatus == constants.ATStatusEnd {
		u.l.WithContext(ctx).Infof("activity status is end, no need to renew")
		return nil
	}

	err = u.userRemindUsecase.RemindJoinGroupV3(ctx, u.helpCodeService)
	if err != nil {
		u.l.WithContext(ctx).Errorf("user remind v3 failed, err=%v", err)
		return err
	}

	return nil
}

func (u *UserRemindService) RenewV22(ctx context.Context) error {
	cost := util.MethodCost(ctx, u.l, "UserRemindService.RenewV22")
	defer cost()

	locked, unlock, err := redis.JobLock(ctx, u.redisClusterClient, "RenewV22", 16*time.Minute)
	if err != nil || !locked {
		u.l.WithContext(ctx).Infof("RenewV22 lock failed, err=%v, locked=%v", err, locked)
		return err
	}
	defer unlock()

	activity, err := u.activityUsecase.GetActivityInfo(ctx)
	if err != nil {
		u.l.WithContext(ctx).Errorf("get activity info failed, err=%v", err)
		// 未知活动状态时不打扰用户，不再处理
		return err
	}
	if activity.ActivityStatus == constants.ATStatusEnd {
		u.l.WithContext(ctx).Infof("activity status is end, no need to renew")
		return nil
	}

	err = u.userRemindUsecase.FreeDurationRenewV22(ctx)
	if err != nil {
		u.l.WithContext(ctx).Errorf("user remind v22 failed, err=%v", err)
		return err
	}

	return nil
}

func (u *UserRemindService) CDKV0(ctx context.Context) error {
	cost := util.MethodCost(ctx, u.l, "UserRemindService.CDKV0")
	defer cost()

	locked, unlock, err := redis.JobLock(ctx, u.redisClusterClient, "CDKV0", 8*time.Minute)
	if err != nil || !locked {
		u.l.WithContext(ctx).Infof("CDKV0 lock failed, err=%v, locked=%v", err, locked)
		return err
	}
	defer unlock()

	err = u.userRemindUsecase.CDKV0(ctx)
	if err != nil {
		u.l.WithContext(ctx).Errorf("user remind v0 failed, err=%v", err)
		return err
	}

	return nil
}
