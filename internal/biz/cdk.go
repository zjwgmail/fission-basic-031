package biz

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/kit/sqlx"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
)

type CdkUsecase struct {
	userInfoRepo     UserInfoRepo
	SystemConfigRepo SystemConfigRepo
	redisService     *redis.RedisService
	bizConf          *conf.Business
	l                *log.Helper
}

func NewCdkUsecase(userInfoRepo UserInfoRepo,
	SystemConfigRepo SystemConfigRepo,
	redis *redis.RedisService,
	bizConf *conf.Business,
	l log.Logger) *CdkUsecase {
	return &CdkUsecase{
		userInfoRepo:     userInfoRepo,
		SystemConfigRepo: SystemConfigRepo,
		redisService:     redis,
		bizConf:          bizConf,
		l:                log.NewHelper(l),
	}
}

func (c *CdkUsecase) GetCDK(ctx context.Context, waId string, cdkType int) (string, bool, error) {
	// db校验是否已分配cdk
	userInfo, err := c.userInfoRepo.GetUserInfo(ctx, waId)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			return "", false, err
		}
	} else {
		cdk := getCDKByUserInfo(cdkType, userInfo)
		if cdk != "" {
			return cdk, false, nil
		}
	}
	// redis获取cdk
	cdkQueueName := constants.CdkQueueKeyPrefix + strconv.Itoa(cdkType)
	cdk, err := c.redisService.PopQueueData(ctx, cdkQueueName)
	if err != nil {
		return "", false, err
	}
	// FIXME: 待讨论
	// err = c.userInfoRepo.UpdateUserInfoCdkByWaId(ctx, waId, cdk, cdkType)
	// if err != nil {
	// 	if !errors.Is(err, sqlx.ErrRowsAffected) {
	// 		return "", false, err
	// 	}
	// 	// 有些场景可以忽略
	// }
	currentCount, err := c.redisService.GetQueueSize(ctx, cdkQueueName)
	if err != nil {
		return "", false, err
	}
	value := c.redisService.Get(cdkQueueName + constants.CdkTotalCountKeySuffix)
	totalCount, err := strconv.Atoi(value)
	if err != nil {
		return "", false, err
	}
	needAlarm := float64(currentCount) < float64(totalCount)*c.bizConf.Cdk.AlarmThreshold
	if needAlarm {
		c.l.Infof("cdkQueueName:%s,currentCount:%d,totalCount:%d", cdkQueueName, currentCount, totalCount)
	}
	return cdk, needAlarm, nil
}

func getCDKByUserInfo(cdkType int, userInfo *UserInfo) string {
	switch cdkType {
	case 0:
		return userInfo.CDKv0
	case 3:
		return userInfo.CDKv3
	case 6:
		return userInfo.CDKv6
	case 9:
		return userInfo.CDKv9
	case 12:
		return userInfo.CDKv12
	case 15:
		return userInfo.CDKv15
	}
	return ""
}
