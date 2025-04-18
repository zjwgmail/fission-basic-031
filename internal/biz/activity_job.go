package biz

import (
	"context"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

const lockTimeout = time.Second * 120

type ActivityJob struct {
	repo         ActivityInfoRepo
	l            *log.Helper
	bootstrap    *conf.Bootstrap
	redisService *redis.RedisService
}

func NewActivityJob(repo ActivityInfoRepo, l log.Logger, bootstrap *conf.Bootstrap, redisService *redis.RedisService) *ActivityJob {
	return &ActivityJob{
		repo:         repo,
		l:            log.NewHelper(l),
		bootstrap:    bootstrap,
		redisService: redisService,
	}
}

func (w *ActivityJob) ActivityJobHandle(ctx context.Context, methodName string) {

	template := w.redisService
	taskLockKey := constants.GetTaskLockKey(w.bootstrap.Business.Activity.Id, methodName)

	getLock, err := template.SetNX(methodName, taskLockKey, "1", lockTimeout)
	if err != nil {
		w.l.Error(fmt.Sprintf("method[%s],call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		w.l.Error(fmt.Sprintf("method[%s],get reids lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := template.Del(taskLockKey)
		if !del {
			w.l.Error(fmt.Sprintf("method[%s]，del redis lock fail", methodName))
		}
	}()

	// 查询活动信息
	activityInfo, err := w.repo.GetActivityInfo(ctx, w.bootstrap.Business.Activity.Id)
	if err != nil {
		w.l.Error(fmt.Sprintf("method[%s],query activity fail. activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		return
	}

	switch activityInfo.ActivityStatus {
	case constants.ATStatusUnStart:
		// 未开始
		startAt := activityInfo.StartAt
		nowTime := time.Now()
		if nowTime.After(startAt) || nowTime.Equal(startAt) {
			// 到达开始时间
			update := &UpdateActivityInfoDto{
				Id:             w.bootstrap.Business.Activity.Id,
				ActivityStatus: constants.ATStatusStarted,
			}
			w.l.Infof(fmt.Sprintf("method[%s],ActivityJobHandle activity to ATStatusStarted activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
			err = w.repo.UpdateActivityInfo(ctx, update)
			if err != nil {
				w.l.Error(fmt.Sprintf("method[%s],update activity info fail，activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
				return
			}
			w.l.Info(fmt.Sprintf("method[%s],activity is running;activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		} else {
			w.l.Info(fmt.Sprintf("method[%s],The active start time is not reached;activity‘s id:%v", methodName, w.bootstrap.Business.Activity.Id))
		}
	case constants.ATStatusStarted:
		endAt := activityInfo.EndAt
		nowTime := time.Now()
		if !w.bootstrap.Business.Activity.IsDebug && (nowTime.After(endAt) || nowTime.Equal(endAt)) {
			// 到达结束时间
			update := &UpdateActivityInfoDto{
				Id:             w.bootstrap.Business.Activity.Id,
				ActivityStatus: constants.ATStatusBuffer,
			}
			w.l.Infof(fmt.Sprintf("method[%s],ActivityJobHandle activity to ATStatusBuffer activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
			err = w.repo.UpdateActivityInfo(ctx, update)
			if err != nil {
				w.l.Error(fmt.Sprintf("method[%s],Failed to update the activity information，activity‘s id:%v", methodName, w.bootstrap.Business.Activity.Id))
				return
			}
			w.l.Info(fmt.Sprintf("method[%s],Activity has ended;activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		} else {
			w.l.Info(fmt.Sprintf("method[%s],The activity has not reached its end time;activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		}
	case constants.ATStatusBuffer:
		// 结束
		reallyEndAt := activityInfo.ReallyEndAt
		nowTime := time.Now()
		if nowTime.After(reallyEndAt) || nowTime.Equal(reallyEndAt) {
			// 到达开始时间
			update := &UpdateActivityInfoDto{
				Id:             w.bootstrap.Business.Activity.Id,
				ActivityStatus: constants.ATStatusEnd,
			}
			w.l.Infof(fmt.Sprintf("method[%s],ActivityJobHandle activity to ATStatusEnd activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
			err = w.repo.UpdateActivityInfo(ctx, update)
			if err != nil {
				w.l.Error(fmt.Sprintf("method[%s],Failed to update the activity information，activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
				return
			}
			w.l.Info(fmt.Sprintf("method[%s],activity end;activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		} else {
			w.l.Info(fmt.Sprintf("method[%s],The activity has not reached its end time;activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		}
	case constants.ATStatusEnd:
		w.l.Info(fmt.Sprintf("method[%s],Activity has ended;activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
	}

}
