package biz

import (
	"context"
	"encoding/json"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/util/goroutine_pool"
	"fission-basic/util"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// spec: "*/9 * * * *" todo参考这个设置
const RESEND_JOB_LOCK_TIME = time.Minute * 10

type ResendRetryJob struct {
	repo             WaMsgRetryRepo
	userRemindRepo   UserRemindRepo
	activityInfoRepo ActivityInfoRepo
	l                *log.Helper
	bootstrap        *conf.Bootstrap
	redisService     *redis.RedisService
	waMsgService     *WaMsgService
}

var resendRetryGoroutinePool = goroutine_pool.NewGoroutinePool(1)

func NewResendRetryJob(repo WaMsgRetryRepo, userRemindRepo UserRemindRepo, activityInfoRepo ActivityInfoRepo, l log.Logger, bootstrap *conf.Bootstrap, redisService *redis.RedisService, waMsgService *WaMsgService) *ResendRetryJob {
	return &ResendRetryJob{
		repo:             repo,
		userRemindRepo:   userRemindRepo,
		activityInfoRepo: activityInfoRepo,
		l:                log.NewHelper(l),
		bootstrap:        bootstrap,
		redisService:     redisService,
		waMsgService:     waMsgService,
	}
}

func (w *ResendRetryJob) ResendRetryJobHandle(ctx context.Context, methodName string) {

	template := w.redisService
	taskLockKey := constants.GetTaskLockKey(w.bootstrap.Business.Activity.Id, methodName)

	getLock, err := template.SetNX(methodName, taskLockKey, "1", 11*time.Minute)
	if err != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s],call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s],get reids lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := template.Del(taskLockKey)
		if !del {
			w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s]，del redis lock fail", methodName))
		}
	}()

	w.l.WithContext(ctx).Info(fmt.Sprintf("method[%s],running", methodName))
	// 查询活动信息，若为unstart、end就不执行
	activityInfo, err := w.activityInfoRepo.GetActivityInfo(ctx, w.bootstrap.Business.Activity.Id)
	if err != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s],query activity fail. activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		return
	}
	if activityInfo.ActivityStatus == constants.ATStatusUnStart || activityInfo.ActivityStatus == constants.ATStatusEnd {
		w.l.WithContext(ctx).Warn(fmt.Sprintf("method[%s],activity is not running，activity's id:%v", methodName, w.bootstrap.Business.Activity.Id))
		return
	}

	// 查询未发送的消息
	// 分页参数 todo 调大
	limit := uint(1000)
	minId := "0"
	stateList := []int{constants.MsgSendStateUnSend, constants.MsgSendStateFail, constants.MsgSendStateNxFail, constants.MsgSendStateNxTimeout}

	// 循环直到所有消息都被处理
	for {
		// 调用支持分页的方法获取未发送消息的waId列表
		waIdList, err := w.repo.ListRetryWaIdByState(ctx, minId, limit, stateList)
		if err != nil {
			w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,Failed to query the list of unsent Waids. Procedure，err:%v", methodName, err))
			break // 如果有错误，跳出循环
		}

		if len(waIdList) == 0 {
			break // 如果没有更多的消息，跳出循环
		}

		// 处理当前页的消息
		for i := range waIdList {
			waId := waIdList[i]
			ctx2 := context.Background()
			// 查询是否在免打扰时间
			isDisturb, err := util.IsNotDisturbTime(ctx2, w.bootstrap.Business.Activity.IsDebug, waId)
			if err != nil {
				w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,get IsNotDisturbTime fial waId:%v,err:%v", methodName, waId, err))
			}
			if !isDisturb {

				ctx3 := context.Background()
				resendRetryGoroutinePool.Execute(func(param interface{}) {
					u, ok := param.(string) // 断言u是User类型
					if !ok {
						w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,Assertion error occurred，waId:%v", methodName, u))
					}
					w.l.WithContext(ctx).Infof(fmt.Sprintf("mthod:%s,resendRetryGoroutinePool The pool execution task starts，waId:%v", methodName, u))
					w.ReSendMsgByWaId(ctx3, waId, true)
					w.l.WithContext(ctx).Infof(fmt.Sprintf("mthod:%s,resendRetryGoroutinePool The pool execution task end，waId:%v", methodName, u))
				}, waId)
			}

		}

		minId = waIdList[len(waIdList)-1]

		// 等待当前页的消息处理完成
		resendGoroutinePool.Wait()

		// 准备下一页
		//page++
	}

}

func (w *ResendRetryJob) ReSendMsgByWaId(ctx context.Context, waId string, isPt bool) {
	methodName := "ReSendMsgByWaId"

	// 查询是否是免费期
	userRemind, err := w.userRemindRepo.GetUserRemindInfo(ctx, waId)
	if err != nil {
		w.l.WithContext(ctx).Warn(fmt.Sprintf("mthod:%s,Failed to query userRemindRepo. waId:%v，err:%v", methodName, waId, err))
		return
	}
	timeLastSend := userRemind.LastSendTime
	nowUnix := time.Now().Unix()

	if nowUnix-timeLastSend >= 86400 {
		w.l.WithContext(ctx).Warn(fmt.Sprintf("mthod:%s,This user is not free period does not send messages. waId:%v，err:%v", methodName, waId, err))
		return
	}

	template := w.redisService
	taskLockKey := constants.GetReSendMsgLockKey(w.bootstrap.Business.Activity.Id, waId)

	getLock, err := template.SetNX(methodName, taskLockKey, "1", lockTimeout)
	if err != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s],call redis nx fail，this waId not reSend. waId:%v", methodName, waId))
		return
	}
	if !getLock {
		w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s],get reids lock fail，this waId not reSend. waId:%v", methodName, waId))
		return
	}
	defer func() {
		del := template.Del(taskLockKey)
		if !del {
			w.l.WithContext(ctx).Error(fmt.Sprintf("method[%s]，del redis lock fail. waId:%v", methodName, waId))
		}
	}()

	w.l.WithContext(ctx).Info(fmt.Sprintf("mthod:%s,This user is free and unlock period send messages. waId:%v", methodName, waId))

	// 查询用户信息
	userInfo, err := w.userRemindRepo.GetUserInfo(ctx, waId)
	if err != nil {
		w.l.WithContext(ctx).Warn(fmt.Sprintf("mthod:%s,Failed to query userInfoRepo. waId:%v，err:%v", methodName, waId, err))
		return
	}

	//ptList := make([]string, 0)
	//if isPt {
	//	ptList = append(ptList, util.GetPtTimeList()...)
	//}

	stateList := []int{constants.MsgSendStateUnSend, constants.MsgSendStateFail, constants.MsgSendStateNxFail, constants.MsgSendStateNxTimeout}
	msgList, err := w.repo.ListMsgRetryByWaIdAndState(ctx, stateList, waId)
	if err != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,Failed to query msgSendRepo. waId:%v，err:%v", methodName, waId, err))
		return
	}

	w.l.WithContext(ctx).Info(fmt.Sprintf("mthod:%s,waId:%v，msgList'len is:%v", methodName, waId, len(msgList)))

	if len(msgList) > 0 {
		for _, msg := range msgList {
			w.l.WithContext(ctx).Info(fmt.Sprintf("mthod:%s,Execution of the resending message begins. waId:%v，msgId:%v", methodName, waId, msg.ID))
			w.reSendMsg(ctx, methodName, msg, userInfo)
			w.l.WithContext(ctx).Info(fmt.Sprintf("mthod:%s,Execution of the resending message end. waId:%v，msgId:%v", methodName, waId, msg.ID))
		}
	}
	w.l.WithContext(ctx).Info(fmt.Sprintf("mthod:%s,Execution of the all message end. waId:%v", methodName, waId))
}

func (w *ResendRetryJob) reSendMsg(ctx context.Context, methodName string, msgInfoEntity *dto.WaMsgRetryDto, userInfo *UserInfo) {
	ctx = context.Background()

	buildMsgParams := &conf.MsgInfo{}
	err := json.Unmarshal([]byte(msgInfoEntity.BuildMsgParam), buildMsgParams)
	if err != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,msg buildMsgParams convert to json fail. waId:%v，err:%v", methodName, msgInfoEntity.WaID, err))
		return
	}

	buildMsgInfo := []*dto.BuildMsgInfo{
		{
			WaId:       msgInfoEntity.WaID,
			MsgType:    msgInfoEntity.MsgType,
			Channel:    userInfo.Channel,
			Language:   userInfo.Language,
			Generation: strconv.Itoa(userInfo.Generation),
			RallyCode:  userInfo.HelpCode,
		},
	}
	// 免费期，互动
	sendNxListParamsDto, nxErr := w.waMsgService.BuildInteractionMessage2NX(ctx, buildMsgInfo, []*conf.MsgInfo{buildMsgParams})
	if nxErr != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,BuildInteractionMessage2NX  fail. waId:%v，err:%v", methodName, msgInfoEntity.WaID, err))
		return
	}

	for _, paramsDto := range sendNxListParamsDto {
		paramsDto.WaMsgSend.ID = msgInfoEntity.ID
	}

	w.l.WithContext(ctx).Info(fmt.Sprintf("mthod:%s,send nx messages. waId:%v", methodName, msgInfoEntity.WaID))

	_, nxErr = w.waMsgService.SendMsgList2NXOfRetry(ctx, sendNxListParamsDto)
	if nxErr != nil {
		w.l.WithContext(ctx).Error(fmt.Sprintf("mthod:%s,SendMsgList2NXOfRetry fail. waId:%v，err:%v", methodName, msgInfoEntity.WaID, err))
		return
	}
}
