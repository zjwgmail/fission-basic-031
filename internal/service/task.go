package service

import (
	"context"
	"encoding/json"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/pkg/nxcloud"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/util"
	"fmt"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"

	"fission-basic/contrib/task"
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/feishu"
	"fission-basic/internal/pkg/queue"
	taskq "fission-basic/kit/task"
)

type TaskService struct {
	d           *conf.Data
	officialQ   *queue.Official
	unOfficialQ *queue.UnOfficial
	renewMsgQ   *queue.RenewMsg
	callbackQ   *queue.CallMsg
	repeatHelpQ *queue.RepeatHelp
	gwQ         *queue.GW
	gwRecall    *queue.GWRecall
	gwUnknown   *queue.GWUnknown

	report *feishu.Develop

	l *log.Helper

	officialUsecase   *biz.OfficialRallyUsecase
	unOfficialUsecase *biz.UnOfficialRallyUsecase
	msgUsecase        *biz.MsgUsecase

	// 先这么写，后续再改，service里面不应该有交叉
	*HelpCodeService
	waMsgService       *biz.WaMsgService
	userInfoRepo       biz.UserInfoRepo
	resendJob          *biz.ResendJob
	resendRetryJob     *biz.ResendRetryJob
	waMsgSendRepo      biz.WaMsgSendRepo
	redisClusterClient *redis.ClusterClient
	nxCloudService     *NxCloudService
}

func NewTaskService(
	d *conf.Data,
	officialQ *queue.Official,
	unOfficialQ *queue.UnOfficial,
	renewMsgQ *queue.RenewMsg,
	report *feishu.Develop,
	officialUsecase *biz.OfficialRallyUsecase,
	unOfficialUsecase *biz.UnOfficialRallyUsecase,
	msgUsecase *biz.MsgUsecase,
	callbackQ *queue.CallMsg,
	repeatHelpQ *queue.RepeatHelp,
	gwQ *queue.GW,
	gwRecall *queue.GWRecall,
	gwUnknown *queue.GWUnknown,
	helpCodeService *HelpCodeService,
	l log.Logger,
	waMsgService *biz.WaMsgService,
	userInfoRepo biz.UserInfoRepo,
	redisClusterClient *redis.ClusterClient,
	resendJob *biz.ResendJob,
	resendRetryJob *biz.ResendRetryJob,
	waMsgSendRepo biz.WaMsgSendRepo,
	nxCloudService *NxCloudService,
) *TaskService {
	return &TaskService{
		d:                  d,
		officialQ:          officialQ,
		unOfficialQ:        unOfficialQ,
		renewMsgQ:          renewMsgQ,
		callbackQ:          callbackQ,
		repeatHelpQ:        repeatHelpQ,
		report:             report,
		officialUsecase:    officialUsecase,
		unOfficialUsecase:  unOfficialUsecase,
		msgUsecase:         msgUsecase,
		l:                  log.NewHelper(l),
		waMsgService:       waMsgService,
		HelpCodeService:    helpCodeService,
		userInfoRepo:       userInfoRepo,
		resendJob:          resendJob,
		resendRetryJob:     resendRetryJob,
		waMsgSendRepo:      waMsgSendRepo,
		redisClusterClient: redisClusterClient,
		gwQ:                gwQ,
		gwRecall:           gwRecall,
		gwUnknown:          gwUnknown,
		nxCloudService:     nxCloudService,
	}
}

func (s *TaskService) ListTasks() []*task.Task {
	var tasks []*task.Task

	// 官方助力消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.officialQ.Queue,
			Func:   s.officialRallyCodeConumser,
			Number: 50,
			D:      time.Millisecond * 500, // 50ms
		},
	)...)

	// 非官方助理码消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.unOfficialQ.Queue,
			Func:   s.unOfficialRallyCodeConumser,
			Number: 50,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)

	// 免续费消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.renewMsgQ.Queue,
			Func:   s.renewMsgConsumer,
			Number: 10,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)

	// 回执消息消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.callbackQ.Queue,
			Func:   s.recallConsumer,
			Number: 100,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)

	// 重复助力消费者
	tasks = append(tasks, lo.Repeat(
		1,
		&task.Task{
			Queue:  s.repeatHelpQ.Queue,
			Func:   s.repeatHelpConsumer,
			Number: 30,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)

	// 流程承接消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.gwQ.Queue,
			Func:   s.GwConsumer,
			Number: 50,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)
	// 流程承接消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.gwRecall.Queue,
			Func:   s.GwConsumer,
			Number: 50,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)
	// 流程承接消费者
	tasks = append(tasks, lo.Repeat(
		2,
		&task.Task{
			Queue:  s.gwUnknown.Queue,
			Func:   s.GwConsumer,
			Number: 50,
			D:      time.Millisecond * 50, // 50ms
		},
	)...)

	return tasks
	// return []*task.Task{
	// 	{
	// 		Queue:  s.officialQ.Queue,
	// 		Func:   s.officialRallyCodeConumser,
	// 		Number: 5,
	// 		D:      time.Millisecond * 50, // 50ms
	// 	},
	// 	{
	// 		Queue:  s.unOfficialQ.Queue,
	// 		Func:   s.unOfficialRallyCodeConumser,
	// 		Number: 5,
	// 		D:      time.Millisecond * 50, // 50ms
	// 	},
	// 	{
	// 		Queue:  s.renewMsgQ.Queue,
	// 		Func:   s.renewMsgConsumer,
	// 		Number: 1,
	// 		D:      time.Millisecond * 50, // 50ms
	// 	},
	// 	{
	// 		Queue:  s.callbackQ.Queue,
	// 		Func:   s.recallConsumer,
	// 		Number: 1,
	// 		D:      time.Millisecond * 50, // 50ms
	// 	},
	// 	{
	// 		Queue:  s.repeatHelpQ.Queue,
	// 		Func:   s.repeatHelpConsumer,
	// 		Number: 1,
	// 		D:      time.Millisecond * 50, // 50ms
	// 	},
	// }
}

func (s *TaskService) GwConsumer(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	for i := range ids {
		err := s.GwHandler(ctx, ids[i])
		if err != nil {
			// 这里不能return err，会导致消费者停止；利用重试机制重新处理
			continue
		}
	}

	return nil
}

func (s *TaskService) GwHandler(ctx context.Context, id string) error {
	var req v1.UserAttendInfoRequest
	err := jsoniter.Unmarshal([]byte(id), &req)
	if err != nil {
		s.l.WithContext(ctx).Errorf("gwHandler-unmarshal-failed, err=%v, id=%s", err, id)
		return err
	}

	_, err = s.nxCloudService.userAttendInfo(ctx, &req)
	if err != nil {
		s.l.WithContext(ctx).Errorf("gwHandler-UserAttendInfo-failed, err=%v, id=%+v", err, req)
		return err
	}

	s.l.WithContext(ctx).Info("gwHandler-UserAttendInfo-success ")

	return nil
}

func (s *TaskService) OfficialTaskMonitor(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisClusterClient, "official_task_monitor", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	msgs := append(
		[]string{},
		s.buildMonitorMsg(ctx, "官方助力码", s.officialQ.Queue),
		s.buildMonitorMsg(ctx, "非官方助力码", s.unOfficialQ.Queue),
		s.buildMonitorMsg(ctx, "续免费消息回复延长", s.renewMsgQ.Queue),
		s.buildMonitorMsg(ctx, "回调消息", s.callbackQ.Queue),
		s.buildMonitorMsg(ctx, "重复助力", s.repeatHelpQ.Queue),
		s.buildMonitorMsg(ctx, "网关消费队列", s.gwQ.Queue),
		s.buildMonitorMsg(ctx, "网关未知消费队列", s.gwUnknown.Queue),
	)

	err = s.report.SendTextMsg(
		ctx,
		strings.Join(msgs, "\n"),
	)
	if err != nil {
		s.l.WithContext(ctx).Errorf("OfficialTaskMonitor-send-failed, err=%v", err)
		return nil
	}

	return nil
}

func (s *TaskService) UnOfficialTaskMonitor(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisClusterClient, "unofficial_task_monitor", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	return s.monitor(ctx, "非官方助力码队列", s.officialQ.Queue)
}

func (s *TaskService) RenewMsgMonitor(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisClusterClient, "renew_msg_monitor", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	return s.monitor(ctx, "续免费消息回复延长队列", s.renewMsgQ.Queue)
}

func (s *TaskService) CallMsgMonitor(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisClusterClient, "call_msg_monitor", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	return s.monitor(ctx, "消息回执延长队列", s.callbackQ.Queue)
}

func (s *TaskService) GwMsgMonitor(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisClusterClient, "gw_msg_monitor", 60*time.Second)
	if err != nil || !locked {
		return err
	}

	defer unlock()

	return s.monitor(ctx, "网关消费队列", s.gwQ.Queue)
}

func (s *TaskService) buildMonitorMsg(ctx context.Context, taskName string, t *taskq.Queue) string {
	len, err := t.Len()
	if err != nil {
		s.l.WithContext(ctx).Errorf("%s queue len failed, err=%v", taskName, err)
		// 这里不返回错误
		return fmt.Sprintf("%s现存数据量: %v", taskName, err)
	}

	return fmt.Sprintf("> %s现存数据量: %d", taskName, len)
}

func (s *TaskService) monitor(ctx context.Context, taskName string, t *taskq.Queue) error {
	len, err := t.Len()
	if err != nil {
		s.l.WithContext(ctx).Errorf("%s queue len failed, err=%v", taskName, err)
		// 这里不返回错误
		return nil
	}

	err = s.report.SendTextMsg(
		ctx,
		fmt.Sprintf("%s现存数据量: %d", taskName, len),
	)
	if err != nil {
		s.l.WithContext(ctx).Errorf("%s send msg failed, err=%v", taskName, err)
		return nil
	}

	return err
}

func (s *TaskService) officialRallyCodeConumser(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// s.l.WithContext(ctx).Debugf("official ids=%+v", ids)

	for i := range ids {
		err := s.officialRallyCodeHandler(ctx, ids[i])
		if err != nil {
			s.l.WithContext(ctx).Errorf("officialRallyCodeHandler failed, err=%v, id=%s", err, ids[i])
			// 这里不能return err，会导致消费者停止；利用重试机制重新处理
			continue
		}
	}

	return nil
}

func (s *TaskService) renewMsgConsumer(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	for i := range ids {
		err := s.renewMsgHandler(ctx, ids[i])
		if err != nil {
			s.l.WithContext(ctx).Errorf("renewMsgHandler failed, err=%v, id=%d", err, ids[i])
			continue
		}
	}

	return nil
}

// renewMsgHandler
func (s *TaskService) renewMsgHandler(ctx context.Context, waID string) error {
	//idStrs := strings.Split(id, "|")
	//if len(idStrs) != 2 {
	//	s.l.WithContext(ctx).Errorf("wrong rally msg id=%s", id)
	//	return nil
	//}
	//
	//waID, rallyCode := idStrs[0], idStrs[1]

	userInfo, err := s.userInfoRepo.GetUserInfo(ctx, waID)
	if err != nil {
		s.l.WithContext(ctx).Errorf("GetUserInfo failed, err=%v, waID=%s", err, waID)
		return err
	}

	// s.l.WithContext(ctx).Debugf("waID=%s, rallyCode=%s", waID, userInfo.HelpCode)

	shortLink, err := s.HelpCodeService.GetShortLinkByHelpCode(ctx, userInfo.HelpCode, 0)
	if err != nil {
		s.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v, waID=%s", err, waID)
		return err
	}
	msgInfoEntity := &dto.BuildMsgInfo{
		WaId:       waID,
		MsgType:    constants.RenewFreeReplyMsg,
		Channel:    userInfo.Channel,
		Language:   userInfo.Language,
		Generation: strconv.Itoa(userInfo.Generation),
		RallyCode:  userInfo.HelpCode,
	}

	// 发送回复消息
	msg, err := s.waMsgService.RenewFreeReplyMsg(ctx, msgInfoEntity, shortLink, constants.BizTypeInteractive)
	if err != nil {
		s.l.WithContext(ctx).Errorf("build RenewFreeReplyMsg failed, err=%v, waID=%s", err, waID)
		return err
	}

	// 新增消息表
	for _, paramsDto := range msg {
		id, err := s.waMsgSendRepo.AddWaMsgSend(ctx, paramsDto.WaMsgSend)
		if err != nil {
			return err
		}
		paramsDto.WaMsgSend.ID = id
	}

	_, err = s.waMsgService.SendMsgList2NX(ctx, msg)
	if err != nil {
		s.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s",
			err, waID, userInfo.HelpCode)
		return err
	}

	// TODO  重试发送所有未成功发送的消息
	s.resendJob.ReSendMsgByWaId(ctx, waID, false)
	s.resendRetryJob.ReSendMsgByWaId(ctx, waID, false)
	return nil
}

func (s *TaskService) officialRallyCodeHandler(ctx context.Context, id string) error {
	data, err := s.officialQ.UnWrap(id)
	if err != nil {
		s.l.WithContext(ctx).Errorf("unWrap failed, err=%v, id=%d", err, id)
		return err
	}

	return s.officialUsecase.Handler(ctx, data, s.HelpCodeService)
}

func (s *TaskService) unOfficialRallyCodeConumser(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	cost := util.MethodCost(ctx, s.l, "taskService.unOfficialRallyCodeConumser")
	defer cost()

	start := time.Now()
	defer func() {
		cost := time.Since(start)
		s.l.WithContext(ctx).Infof("unOfficialRallyCodeConumser cost=%v", cost.Seconds())
	}()

	for i := range ids {
		err := s.unOfficialRallyCodeHandler(ctx, ids[i])
		if err != nil {
			s.l.WithContext(ctx).Errorf("unofficialRallyCodeHandler failed, err=%v, id=%d", err, ids[i])
			// 这里不能return err，会导致消费者停止；利用重试机制重新处理
			continue
		}
	}

	return nil
}

func (s *TaskService) unOfficialRallyCodeHandler(ctx context.Context, id string) error {
	rallyData, err := s.unOfficialQ.UnWrap(id)
	if err != nil {
		s.l.WithContext(ctx).Errorf("UnWrap failed, err=%v, id=%d", err, id)
		return err
	}

	return s.unOfficialUsecase.Handler(ctx, rallyData, s.HelpCodeService)
}

func (s *TaskService) repeatHelpConsumer(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	for i := range ids {
		err := s.repeatHelpHandler(ctx, ids[i])
		if err != nil {
			s.l.WithContext(ctx).Errorf("unofficialRallyCodeHandler failed, err=%v, id=%d", err, ids[i])
			// 这里不能return err，会导致消费者停止；利用重试机制重新处理
			continue
		}
	}

	return nil
}

func (s *TaskService) repeatHelpHandler(ctx context.Context, id string) error {
	rallyData, err := s.unOfficialQ.UnWrap(id)
	if err != nil {
		s.l.WithContext(ctx).Errorf("UnWrap failed, err=%v, id=%d", err, id)
		return err
	}

	// s.l.WithContext(ctx).Debugf("rallyData=%+v", rallyData)

	return s.unOfficialUsecase.RepeatHelpHandler(ctx, rallyData, s.HelpCodeService)
}

func (s *TaskService) recallConsumer(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	for i := range ids {
		err := s.recallHandler(ctx, ids[i])
		if err != nil {
			s.l.WithContext(ctx).Errorf("callbackHandler failed, err=%v, id=%d", err, ids[i])
			continue
		}
	}

	return nil
}

// recallHandler 回执
func (s *TaskService) recallHandler(ctx context.Context, msgInfoStr string) error {
	queueDTO := &nxcloud.ReceiptMsgQueueDTO{}

	err := json.Unmarshal([]byte(msgInfoStr), queueDTO)
	if err != nil {
		s.l.WithContext(ctx).Errorf("redis receipt queue message convert to queueDTO failed, err=%v, msgInfoStr=%v", err, msgInfoStr)
		return err
	}
	return s.msgUsecase.RecallHandle(ctx, queueDTO)
}
