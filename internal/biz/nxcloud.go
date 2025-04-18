package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/util"
	"fission-basic/kit/sqlx"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"fission-basic/internal/pkg/nxcloud"
	"fission-basic/internal/pkg/queue"
)

type NXCloudUsecase struct {
	nxCloudRepo       NXCloudRepo
	l                 *log.Helper
	business          *conf.Business
	officalQ          *queue.Official
	unOfficalQ        *queue.UnOfficial
	reNewQ            *queue.RenewMsg
	callbackQ         *queue.CallMsg
	repeatHelpQ       *queue.RepeatHelp
	redisService      *redis.RedisService
	msgRepo           MsgRepo
	waMsg             *WaMsgService
	waMsgSendRepo     WaMsgSendRepo
	officialRallyRepo OfficialRallyRepo
}

func NewNxCLoudUsecase(
	nxCloudRepo NXCloudRepo,
	msgRepo MsgRepo,
	l log.Logger,
	officialQueue *queue.Official,
	unofficalQ *queue.UnOfficial,
	reNewQ *queue.RenewMsg,
	callbackQ *queue.CallMsg,
	repeatHelpQ *queue.RepeatHelp,
	waMsg *WaMsgService,
	waMsgSendRepo WaMsgSendRepo,
	officialRallyRepo OfficialRallyRepo,
	business *conf.Business,
	redisService *redis.RedisService,
) *NXCloudUsecase {
	return &NXCloudUsecase{
		msgRepo:           msgRepo,
		nxCloudRepo:       nxCloudRepo,
		l:                 log.NewHelper(l),
		officalQ:          officialQueue,
		unOfficalQ:        unofficalQ,
		reNewQ:            reNewQ,
		callbackQ:         callbackQ,
		repeatHelpQ:       repeatHelpQ,
		waMsg:             waMsg,
		waMsgSendRepo:     waMsgSendRepo,
		officialRallyRepo: officialRallyRepo,
		business:          business,
		redisService:      redisService,
	}
}

type CreateMsgRequest struct {
	Content string
}

// CreateMsg 参团消息
func (nx *NXCloudUsecase) CreateMsg(ctx context.Context, info *nxcloud.NXCloudInfo) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.CreateMsg")
	defer cost()

	msgType := info.MsgType
	if msgType == nxcloud.MsgTypeRallyCode {
		return nx.createUnOfficialMsg(ctx, info)
	}

	if msgType == nxcloud.MsgTypeAttend {
		return nx.createOfficialMsg(ctx, info)
	}

	return fmt.Errorf("unknown msg type=%s", msgType)
}

func (nx *NXCloudUsecase) createOfficialMsg(
	ctx context.Context,
	cloudInfo *nxcloud.NXCloudInfo,
) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.createOfficialMsg")
	defer cost()

	generation, _ := strconv.ParseInt(cloudInfo.Generation, 10, 64)
	officialMsg := &OfficialMsgRecord{
		WaID:       cloudInfo.WaID,
		RallyCode:  cloudInfo.IdentificationCode,
		Channel:    cloudInfo.Channel,
		Language:   cloudInfo.Language,
		Generation: int(generation),
		NickName:   cloudInfo.UserNickName,
		SendTime:   cloudInfo.SendTime,
		State:      MsgStateDoing,
	}

	sendQ := func() error {
		// send official queue
		return nx.officalQ.SendBack(&queue.RallyData{
			WaID:       cloudInfo.WaID,
			RallyCode:  cloudInfo.IdentificationCode,
			NickName:   cloudInfo.UserNickName,
			Channel:    cloudInfo.Channel,
			Language:   cloudInfo.Language,
			SendTime:   cloudInfo.SendTime,
			Generation: cloudInfo.Generation,
		})
	}

	err := nx.nxCloudRepo.SaveOfficialMsg(ctx, officialMsg, cloudInfo.MsgID, cloudInfo.Content)
	if err == nil {
		return sendQ()
	}

	if !sqlx.IsDuplicateError(err) {
		nx.l.WithContext(ctx).Errorf("SaveOfficialMsg failed, err=%v, info=%+v", err, cloudInfo)
		return err
	}

	// 唯一键冲突，则更新语言
	msg, err := nx.officialRallyRepo.FindMsg(ctx, cloudInfo.WaID, cloudInfo.IdentificationCode)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			nx.l.WithContext(ctx).Errorf("FindMsg failed, err=%v, info=%+v", err, cloudInfo)
			return err
		}

		nx.l.WithContext(ctx).Infof("FindMsg failed, err=%v, info=%+v", err, cloudInfo)
		return err
	}

	if msg.Language == cloudInfo.Language {
		// 同一种语言，不需要更新
		return nil
	}

	msg.SendTime = cloudInfo.SendTime
	msg.Language = cloudInfo.Language

	err = nx.nxCloudRepo.UpdateOfficialMsgLaguage(ctx, msg, cloudInfo.MsgID, cloudInfo.Content)
	if err != nil {
		nx.l.WithContext(ctx).Errorf("UpdateOfficialMsgLaguage failed, err=%v, info=%+v", err, cloudInfo)
		return err
	}

	// send official queue
	return sendQ()
}

func (nx *NXCloudUsecase) createUnOfficialMsg(
	ctx context.Context,
	cloudInfo *nxcloud.NXCloudInfo,
) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.createUnOfficialMsg")
	defer cost()

	generation, _ := strconv.ParseInt(cloudInfo.Generation, 10, 64)
	unOfficialMsg := &UnOfficialMsgRecord{
		WaID:       cloudInfo.WaID,
		RallyCode:  cloudInfo.IdentificationCode,
		Channel:    cloudInfo.Channel,
		Language:   cloudInfo.Language,
		Generation: int(generation),
		NickName:   cloudInfo.UserNickName,
		SendTime:   cloudInfo.SendTime,
		State:      MsgStateDoing,
	}

	err := nx.nxCloudRepo.SaveUnOfficialMsg(ctx, unOfficialMsg, cloudInfo.MsgID, cloudInfo.Content)
	if err == nil {
		return nx.unOfficalQ.SendBack(&queue.RallyData{
			WaID:       cloudInfo.WaID,
			RallyCode:  cloudInfo.IdentificationCode,
			NickName:   cloudInfo.UserNickName,
			Channel:    cloudInfo.Channel,
			Language:   cloudInfo.Language,
			SendTime:   cloudInfo.SendTime,
			Generation: cloudInfo.Generation,
		})
	}

	if !sqlx.IsDuplicateError(err) {
		nx.l.WithContext(ctx).Errorf("SaveUnOfficialMsg failed, err=%v, info=%+v", err, cloudInfo)
		return err
	}

	// 唯一键冲突，则放到另外一个队列，只发消息就行
	return nx.repeatHelpQ.SendBack(&queue.RallyData{
		WaID:       cloudInfo.WaID,
		RallyCode:  cloudInfo.IdentificationCode,
		NickName:   cloudInfo.UserNickName,
		Channel:    cloudInfo.Channel,
		Language:   cloudInfo.Language,
		SendTime:   cloudInfo.SendTime,
		Generation: cloudInfo.Generation,
	})
}

func (nx *NXCloudUsecase) Recall(ctx context.Context, info *nxcloud.NXCloudInfo) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.Recall")
	defer cost()

	// todo 发送消息表、重发消息表，判断是否存在
	// 放 Redis里面
	start := time.Now()
	findCost := func() {
		nx.l.WithContext(ctx).Infof("find msg send cost=%s", time.Since(start).String())
	}

	// redis 判断
	redisKey := fmt.Sprintf(constants.MsgSignKey, nx.business.Activity.Id, info.MsgID)
	isExists := nx.redisService.Exits(redisKey)
	if !isExists {
		if info.Costs == nil || len(info.Costs) <= 0 || info.Costs[0].ForeignPrice <= 0 {
			nx.l.WithContext(ctx).Error(fmt.Sprintf("this msg is not in redis and msg is free. not doing this msg. redisKey: %v", redisKey))
			return errors.New("this msg is not in redis and msg is free. not doing this msg.")
		}
		nx.l.WithContext(ctx).Error(fmt.Sprintf("this msg is not in redis and msg is not free. doing this msg. redisKey: %v", redisKey))
	}

	//_, err := nx.msgRepo.FindWaMsgSend(ctx, info.MsgID)
	//if err != nil {
	//	if !errors.Is(err, sqlx.ErrNoRows) {
	//		nx.l.Errorf("query waMsgSend failed, err=%v, info=%v", err, info)
	//		return err
	//	}
	//	_, err = nx.msgRepo.FindWaMsgRetry(ctx, info.MsgID)
	//	if err != nil {
	//		if !errors.Is(err, sqlx.ErrNoRows) {
	//			nx.l.Errorf("query WaMsgRetry failed, err=%v, info=%v", err, info)
	//			return err
	//		}
	//		nx.l.Infof("This message is not in the database, so it is not processed, err=%v, info=%v", err, info)
	//		findCost()
	//		return err
	//	}
	//}
	findCost()

	msgState := constants.NxStatusMsgStateMap[info.Status]

	// 这个需要保存消息接受表，消息回执表。
	err := nx.nxCloudRepo.SaveReceiptMsg(ctx, info.WaID, info.MsgID, info.Content, msgState, info.SendTime, info.Costs)
	if err != nil {
		nx.l.Errorf("SaveReceiveMsg failed, err=%v, info=%v", err, info)
		return err
	}
	// 发送队列的消息
	queueDTO := &nxcloud.ReceiptMsgQueueDTO{
		WaID:     info.WaID,
		MsgID:    info.MsgID,
		MsgType:  info.MsgType,
		MsgState: msgState,
		Costs:    info.Costs,
	}

	marshal, err := json.Marshal(queueDTO)
	if err != nil {
		nx.l.Errorf("queueDTO convert to json failed, err=%v, info=%v", err, info)
		return err
	}

	nx.l.Infof("Send a receipt queue message for redis, info=%v", queueDTO)
	return nx.callbackQ.SendBack([]string{string(marshal)}, true)
}

func (nx *NXCloudUsecase) OnlySaveMsg(ctx context.Context, info *nxcloud.NXCloudInfo) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.OnlySaveMsg")
	defer cost()

	gen, _ := strconv.ParseInt(info.Generation, 10, 64)
	return nx.nxCloudRepo.SaveReceiveMsg(
		ctx, info.WaID, info.MsgID, info.Content, info.SendTime, int(gen))
}

func (nx *NXCloudUsecase) RenewMsg(ctx context.Context, info *nxcloud.NXCloudInfo) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.RenewMsg")
	defer cost()

	err := nx.nxCloudRepo.SaveRenewMsg(ctx, info.WaID, info.MsgID, info.Content, info.SendTime)
	if err != nil {
		nx.l.Errorf("SaveReceiveMsg failed, err=%v, info=%v", err, info)
		return err
	}

	// 发送消息
	//key := fmt.Sprintf(`%s|%s`, info.WaID, info.RallyCode)
	return nx.reNewQ.SendBack([]string{info.WaID}, false)
}

func (nx *NXCloudUsecase) NotWhiteMsg(ctx context.Context, info *nxcloud.NXCloudInfo) error {
	cost := util.MethodCost(ctx, nx.l, "NXCloudUsecase.NotWhiteMsg")
	defer cost()

	// 这个需要保存消息接受表，消息回执表。
	err := nx.nxCloudRepo.OnlySaveReceiveMsg(ctx, info.WaID, info.MsgID, info.Content, info.SendTime)
	if err != nil {
		nx.l.Errorf("SaveReceiveMsg failed, err=%v, info=%v", err, info)
		return err
	}

	msgInfoEntity := &dto.BuildMsgInfo{
		WaId:       info.WaID,
		MsgType:    constants.CannotAttendActivityMsg,
		Channel:    info.Channel,
		Language:   info.Language,
		Generation: info.Generation,
		RallyCode:  info.RallyCode,
	}
	// 发送非白消息
	activity2NX, err := nx.waMsg.CannotAttendActivity2NX(ctx, msgInfoEntity)
	if err != nil {
		nx.l.Errorf("queueDTO convert to json failed, err=%v, info=%v", err, info)
		return err
	}

	// 新增消息表
	for _, paramsDto := range activity2NX {
		id, err := nx.waMsgSendRepo.AddWaMsgSend(ctx, paramsDto.WaMsgSend)
		if err != nil {
			nx.l.Errorf("build msg failed, err=%v, info=%v", err, info)
			return err
		}
		paramsDto.WaMsgSend.ID = id
	}

	// 后续会更新消息表，需要消息id
	_, err = nx.waMsg.SendMsgList2NX(ctx, activity2NX)
	if err != nil {
		nx.l.Errorf("send msg to nx failed, err=%v, info=%v", err, info)
		return err
	}

	nx.l.Infof("Send a not white success")
	return nil
}
