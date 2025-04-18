package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/queue"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/util"
	"fission-basic/kit/sqlx"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"
)

type UnOfficialRallyUsecase struct {
	l            *log.Helper
	attendEnable bool
	rallyRepo    UnOfficialRallyRepo
	userInfoRepo UserInfoRepo
	maxJoinNum   int

	waMsg          *WaMsgService
	cdkUsecase     *CdkUsecase
	activityInfoUC *ActivityInfoUsecase
	business       *conf.Business

	q                  *queue.UnOfficial
	redisClusterClient *redis.ClusterClient
}

func NewUnOfficialRallyUsecase(
	d *conf.Data,
	rallyRepo UnOfficialRallyRepo,
	waMsg *WaMsgService,
	cdkUsecase *CdkUsecase,
	l log.Logger,
	userInfoRepo UserInfoRepo,
	activityInfoUC *ActivityInfoUsecase,
	business *conf.Business,
	redisClusterClient *redis.ClusterClient,
	q *queue.UnOfficial,
) *UnOfficialRallyUsecase {
	return &UnOfficialRallyUsecase{
		attendEnable:       d.AttendEnable,
		rallyRepo:          rallyRepo,
		userInfoRepo:       userInfoRepo,
		l:                  log.NewHelper(l),
		waMsg:              waMsg,
		maxJoinNum:         int(d.JoinGroup.MaxNum),
		cdkUsecase:         cdkUsecase,
		activityInfoUC:     activityInfoUC,
		business:           business,
		redisClusterClient: redisClusterClient,
		q:                  q,
	}
}

func (ur *UnOfficialRallyUsecase) Handler(
	ctx context.Context,
	rallyData *queue.RallyData,
	helpCodeFact HelpCodeInterface,
) error {
	cost := util.MethodCost(ctx, ur.l, "UnOfficialRallyUsecase.Handler")
	defer cost()

	activityInfo, err := ur.activityInfoUC.GetActivityInfo(ctx)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetActivityInfo failed, err=%v", err)
		return err
	}

	// 这里不能重推，一定不能
	locked, unlock, err := redis.ConsumerLock(ctx, ur.redisClusterClient, rallyData.RallyCode, 60*time.Second)
	if err != nil {
		ur.l.WithContext(ctx).Warnf("unofficialRally Consumer get Lock failed, err=%v", err)
		return err
	}
	if !locked {
		ur.l.WithContext(ctx).Warnf("unofficialRally Consumer lock failed")
		return err
	}
	defer unlock()

	// 活动期&缓冲期
	if activityInfo.ActivityStatus == constants.ATStatusStarted ||
		activityInfo.ActivityStatus == constants.ATStatusBuffer {
		return ur.attendHandle(ctx, activityInfo, rallyData, helpCodeFact)
	}

	// 结束期
	return ur.unattendHandle(ctx, rallyData, true)
}

func (ur *UnOfficialRallyUsecase) RepeatHelpHandler(
	ctx context.Context,
	rallyData *queue.RallyData,
	helpCodeFact HelpCodeInterface,
) error {
	activityInfo, err := ur.activityInfoUC.GetActivityInfo(ctx)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetActivityInfo failed, err=%v", err)
		return err
	}

	if activityInfo.ActivityStatus == constants.ATStatusStarted ||
		activityInfo.ActivityStatus == constants.ATStatusBuffer {
		return ur.repeatHelp(ctx, rallyData.WaID, rallyData, false)
	}

	// 结束期
	return ur.unattendHandle(ctx, rallyData, false)
}

func (ur *UnOfficialRallyUsecase) attendHandle(
	ctx context.Context,
	activityInfo *ActivityInfoDto,
	rallyData *queue.RallyData,
	helpCodeFact HelpCodeInterface,
) error {
	cost := util.MethodCost(ctx, ur.l, "UnOfficialRallyUsecase.attendHandle")
	defer cost()
	start := time.Now()

	waID := rallyData.WaID
	helpCode := rallyData.RallyCode

	msg, err := ur.rallyRepo.FindMsg(ctx, waID, helpCode)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			ur.l.WithContext(ctx).
				Errorf("FindMsg failed, err=%v, waID=%s, helpCode=%s", err, waID, helpCode)
			return nil
		}

		// 未找到当做尚未处理
		ur.l.WithContext(ctx).
			Warnf("FindMsg failed, err=%v, waID=%s, rallyCode=%s", err, waID, helpCode)
	}
	// 已处理过
	if msg != nil && msg.State == MsgStateComplete {
		// TODO: 实际是重复处理了，按说不用发消息？@jianwu确认
		return nil
	}

	// 是否助力过
	_, err = ur.rallyRepo.FindUserJoinGroupByWaID(ctx, waID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			ur.l.WithContext(ctx).
				Errorf("FindUserJoinGroupByWaID failed, err=%v, waID=%s", err, waID)
			return err
		}
		// 尚未助力过
		err = nil
	} else {
		// 已助力过
		return ur.repeatHelp(ctx, waID, rallyData, true)
	}

	// 查询助力人信息
	ugs, err := ur.rallyRepo.ListUserJoinGroups(ctx, helpCode)
	if err != nil {
		ur.l.WithContext(ctx).
			Errorf("ListUserJoinGroups failed, err=%v, rallyCode=%s", err, helpCode)
		return err
	}

	ur.l.WithContext(ctx).Infof("query database cost: %s", time.Since(start).String())

	rallyInfo := BaseInfo{
		WaID: waID,
		// RallyCode: helpCode, // 暂时没有
		NickName: rallyData.NickName,
		CDKType:  -1,
		SendTime: rallyData.SendTime,

		// 新开团信息
		Generation: int(rallyData.GenerationInt64()),
		Channel:    rallyData.Channel,
		Language:   rallyData.Language,
	}

	// 查询开团人信息
	helpUserGroup, err := ur.rallyRepo.FindUserCreateGroupByHelpCode(ctx, helpCode)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			// 尚未开团，正常不会，除非延迟严重
			ur.l.WithContext(ctx).
				Warnf("GetUserGroupByHelpCode failed, err=%v, helpCode=%s", err, helpCode)
			return err
		}

		ur.l.WithContext(ctx).
			Errorf("GetUserGroupByHelpCode failed, err=%v, helpCode=%s", err, helpCode)
		return err
	}

	if helpUserGroup.CreateWAID == waID {
		// 自己不能给自己助力
		return ur.canNotHelpSelf(ctx, waID, rallyData, helpCode)
	}

	if len(ugs) >= ur.maxJoinNum {
		// 助力时：
		// 		活动期: 若被助力人已满15次，则1. 助力人开团；2. 消耗助力人助力次数；3. 客态开团消息。4. 不给被助力人发送任何消息。
		// 		缓冲期: 只新增用户

		// 活动期
		if activityInfo.ActivityStatus == constants.ATStatusStarted {
			return ur.startedMaxJoinGroup(ctx, &rallyInfo, helpCode, helpUserGroup.CreateWAID, len(ugs)+1, helpCodeFact)
		}

		// 缓冲期
		return ur.bufferMaxJoinGroup(ctx, &rallyInfo, helpCode, helpUserGroup.CreateWAID, len(ugs)+1, helpCodeFact)
	}

	helpInfo := BaseInfo{
		WaID: helpUserGroup.CreateWAID,
		// UserNickname: helpUserGroup.UserNickname, // 暂时没有
		RallyCode:  helpCode,
		CDKType:    -1,
		Generation: helpUserGroup.Generation,
	}

	joinWaIDs := lo.Map(ugs, func(ug *UserJoinGroup, _ int) string { return ug.JoinWaID })

	// 活动期
	if activityInfo.ActivityStatus == constants.ATStatusStarted {
		err = ur.joinGroup(ctx, &rallyInfo, &helpInfo, joinWaIDs, len(ugs), len(ugs)+1, helpCodeFact)
		if err != nil {
			return err
		}
		return nil
	}

	// 缓冲期
	return ur.bufferJoinGroup(ctx, &rallyInfo, &helpInfo, joinWaIDs, len(ugs), len(ugs)+1, helpCodeFact)
}

func (ur *UnOfficialRallyUsecase) canNotHelpSelf(ctx context.Context, waID string, rallyData *queue.RallyData, helpCode string) error {
	messages, err := ur.waMsg.CanNotHelpOneselfMsg2NX(ctx, &dto.BuildMsgInfo{
		WaId:       waID,
		MsgType:    constants.CanNotHelpOneselfMsg,
		Channel:    rallyData.Channel,
		Language:   rallyData.Language,
		Generation: rallyData.Generation,
		RallyCode:  rallyData.RallyCode,
	})
	if err != nil {
		ur.l.WithContext(ctx).Errorf("CanNotHelpOneselfMsg2NX failed, err=%v", err)
		return err
	}

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	err = ur.rallyRepo.CompleteRally(ctx, waID, helpCode, msgSends, true)
	if err != nil {
		ur.l.WithContext(ctx).
			Errorf("CompleteRally failed, err=%v, waID=%s, helpCode=%s", err, waID, helpCode)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, msgs=%+v, ret=%s",
			err, messages, r)
		return nil
	}

	return nil
}

// 活动期: 若被助力人已满15次，则1. 助力人开团；2. 消耗助力人助力次数；3. 客态开团消息。4. 不给被助力人发送任何消息。
func (ur *UnOfficialRallyUsecase) startedMaxJoinGroup(ctx context.Context,
	rallyInfo *BaseInfo,
	helpCode, helpWaID string, newHelpNum int,
	helpCodeFact HelpCodeInterface,
) error {
	err := ur.buildRallyInfo(ctx, rallyInfo, helpCodeFact)
	if err != nil {
		ur.l.Errorf("buildRallyInfo failed, err=%v, waID=%s, helpCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	rallyMsgs, err := ur.buildRallyMsgInfo(ctx, rallyInfo)
	if err != nil {
		return err
	}
	msgSends := lo.Map(rallyMsgs, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	err = ur.rallyRepo.CreateStartedMaxJoinGroup(ctx, rallyInfo, helpCode, helpWaID, newHelpNum, msgSends)
	if err != nil {
		ur.l.Errorf("CompleteRally failed, err=%v, waID=%s, helpCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, rallyMsgs)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, ret=%s",
			err, rallyInfo.WaID, rallyInfo.RallyCode, r)
		return nil
	}

	return nil
}

func (ur *UnOfficialRallyUsecase) bufferMaxJoinGroup(
	ctx context.Context,
	rallyInfo *BaseInfo,
	helpCode, helpWaID string, newHelpNum int,
	helpCodeFact HelpCodeInterface,
) error {
	rallyMessages, err := ur.waMsg.CanNotStartGroupMsg(ctx, &dto.BuildMsgInfo{
		WaId:       rallyInfo.WaID,
		MsgType:    constants.CanNotStartGroupMsg,
		Channel:    rallyInfo.Channel,
		Language:   rallyInfo.Language,
		Generation: fmt.Sprint(rallyInfo.Generation),
		RallyCode:  rallyInfo.RallyCode,
	})
	if err != nil {
		ur.l.Errorf("CanNotStartGroupMsg failed, err=%v, waID=%s, helpCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	msgSends := lo.Map(rallyMessages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	err = ur.rallyRepo.CreateBufferMaxJoinGroup(ctx, rallyInfo, helpCode, helpWaID, newHelpNum, msgSends)
	if err != nil {
		ur.l.Errorf("CompleteRally failed, err=%v, waID=%s, helpCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, rallyMessages)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, ret=%s",
			err, rallyInfo.WaID, rallyInfo.RallyCode, r)
		return nil
	}

	return nil
}

func (ur *UnOfficialRallyUsecase) repeatHelp(ctx context.Context, waID string,
	rallyData *queue.RallyData, withMsgDB bool) error {
	messages, err := ur.waMsg.RepeatHelpMsg2NX(ctx, &dto.BuildMsgInfo{
		WaId:       waID,
		MsgType:    constants.RepeatHelpMsg,
		Channel:    rallyData.Channel,
		Language:   rallyData.Language,
		Generation: rallyData.Generation,
		RallyCode:  rallyData.RallyCode,
	}, rallyData.NickName)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("RepeatHelpMsg2NX failed, err=%v", err)
		return err
	}

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	err = ur.rallyRepo.CompleteRally(ctx, waID, rallyData.RallyCode, msgSends, withMsgDB)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("CompleteRally failed, err=%v", err)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, ret=%s",
			err, waID, rallyData.RallyCode, r)
		return nil
	}

	return nil
}

// joinGroup 助力
func (ur *UnOfficialRallyUsecase) bufferJoinGroup(
	ctx context.Context,
	rallyInfo, helpInfo *BaseInfo,
	joinWaIDs []string,
	oldJoinNum, newJoinNum int, // 参团人数
	helpCodeFact HelpCodeInterface,
) error {
	ur.l.WithContext(ctx).Infof("bufferJoinGroup, rallyInfo=%+v, helpInfo=%+v, oldJoinNum=%d, newJoinNum=%d",
		rallyInfo, helpInfo, oldJoinNum, newJoinNum)

	err := ur.buildHelpBaseInfo(ctx, helpInfo, oldJoinNum, helpCodeFact)
	if err != nil {
		ur.l.Errorf("buildHelpBaseInfo failed, err=%v, waID=%s, helpCode=%s", err, helpInfo.WaID, helpInfo.RallyCode)
		return err
	}

	messages, err := ur.buildHelpMsgInfo(ctx, joinWaIDs, rallyInfo, helpInfo, oldJoinNum)
	if err != nil {
		return err
	}

	rallyMessages, err := ur.waMsg.CanNotStartGroupMsg(ctx, &dto.BuildMsgInfo{
		WaId:       rallyInfo.WaID,
		MsgType:    constants.CanNotStartGroupMsg,
		Language:   rallyInfo.Language,
		Channel:    rallyInfo.Channel,
		Generation: fmt.Sprint(rallyInfo.Generation),
	})
	if err != nil {
		ur.l.WithContext(ctx).Errorf("CanNotStartGroupMsg failed, err=%v, waID=%s, rallyCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}
	messages = append(messages, rallyMessages...)

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	err = ur.rallyRepo.
		CreateBufferJoinGroup(ctx, rallyInfo, helpInfo, msgSends, newJoinNum)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("CreateJoinGroup2 failed, err=%v, waID=%s, rallyCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, ret=%s",
			err, rallyInfo.WaID, rallyInfo.RallyCode, r)
		return nil
	}

	return nil
}

// 传递需要的信息，不知道叫啥名字
type BaseInfo struct {
	WaID               string
	RallyCode          string
	NickName           string
	RallyCodeShortLink string

	CDK          string
	CdkShortLink string
	CDKType      int // 初始化值成-1，因为0有业务含义

	NeedCreateGroup bool  // 需要开团
	SendTime        int64 // 助力时间

	Language   string
	Channel    string
	Generation int
}

func (ur *UnOfficialRallyUsecase) buildHelpBaseInfo(
	ctx context.Context,
	helpInfo *BaseInfo,
	joinCount int,
	helpCodeFact HelpCodeInterface,
) error {
	cost := util.MethodCost(ctx, ur.l, "UnOfficialRallyUsecase.buildHelpBaseInfo")
	defer cost()

	version := (joinCount + 1) / 3 // 基于3、6、9、12、15获奖
	helpShortLink, err := helpCodeFact.GetShortLinkByHelpCode(ctx, helpInfo.RallyCode, 0)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v, helpCode=%s, version=%d", err, helpInfo.RallyCode, version)
		return err
	}

	helpInfo.RallyCodeShortLink = helpShortLink
	if joinCount < 15 {
		if (joinCount+1)%3 == 0 {
			cdkType := joinCount + 1
			cdk, ok, err := ur.cdkUsecase.GetCDK(ctx, helpInfo.WaID, cdkType)
			if err != nil {
				ur.l.WithContext(ctx).Errorf("GetCDK failed, err=%v, waID=%s, cdkType=%d", err, helpInfo.WaID, cdkType)
				return err
			}
			if ok {
				err := ur.activityInfoUC.UpdateActivityInfo(ctx, &UpdateActivityInfoDto{
					Id:             ur.business.Activity.Id,
					ActivityStatus: constants.ATStatusBuffer,
				})
				if err != nil && errors.Is(err, sqlx.ErrRowsAffected) {
					ur.l.WithContext(ctx).Errorf("UpdateActivityInfo failed, err=%v, waID=%s, cdkType=%d", err, helpInfo.WaID, cdkType)
					// 忽略error
					err = nil
				}
			}
			helpInfo.CDK = cdk
			//todo 世杰验证
			helpInfo.CdkShortLink, err = helpCodeFact.GetShortLinkByHelpCode(ctx, helpInfo.RallyCode, cdkType/3)
			if err != nil {
				ur.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v, waID=%s, cdkType=%d", err, helpInfo.WaID, cdkType)
				return err
			}
			helpInfo.CDKType = cdkType
		}
	}

	userInfo, err := ur.userInfoRepo.GetUserInfo(ctx, helpInfo.WaID)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetUserInfo failed, err=%v, waID=%s", err, helpInfo.WaID)
		return err
	}

	helpInfo.Language = userInfo.Language
	helpInfo.Channel = userInfo.Channel
	helpInfo.Generation = userInfo.Generation
	helpInfo.NickName = userInfo.Nickname

	return nil
}

func (ur *UnOfficialRallyUsecase) buildRallyInfo(
	ctx context.Context,
	rallyInfo *BaseInfo,
	helpCodeFact HelpCodeInterface,
) error {
	cost := util.MethodCost(ctx, ur.l, "UnOfficialRallyUsecase.buildRallyInfo")
	defer cost()

	// 是否已经开团
	ucg, err := ur.rallyRepo.FindUserCreateGroup(ctx, rallyInfo.WaID)
	if err == nil {
		// 已开团不需要处理
		rallyInfo.RallyCode = ucg.HelpCode
		shortLink, err := helpCodeFact.GetShortLinkByHelpCode(ctx, rallyInfo.RallyCode, 0)
		if err != nil {
			ur.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v, waID=%s", err, rallyInfo.WaID)
			return err
		}
		rallyInfo.RallyCodeShortLink = shortLink
		return nil
	}

	if !errors.Is(err, sqlx.ErrNoRows) {
		ur.l.WithContext(ctx).
			Errorf("FindUserCreateGroup failed, err=%v, waID=%s", err, rallyInfo.WaID)
		return err
	}

	// 未开团
	start := time.Now()
	rallyInfo.NeedCreateGroup = true
	rallyCode, err := helpCodeFact.GetHelpCode(ctx)
	ur.l.WithContext(ctx).Infof("GetHelpCode cost=%v ", time.Since(start))
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetHelpCode failed, err=%v, waID=%s", err, rallyInfo.WaID)
		return err
	}
	rallyInfo.RallyCode = rallyCode

	start = time.Now()
	shortLink, err := helpCodeFact.GetShortLinkByHelpCode(ctx, rallyCode, 0)
	ur.l.WithContext(ctx).Infof("GetShortLinkByHelpCode cost=%v ", time.Since(start))
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v, waID=%s", err, rallyInfo.WaID)
		return err
	}
	rallyInfo.RallyCodeShortLink = shortLink

	// free cdk
	cdk, ok, err := ur.cdkUsecase.GetCDK(ctx, rallyInfo.WaID, 0)
	ur.l.WithContext(ctx).Infof("get cdk cost=%v ", time.Since(start))
	if err != nil {
		ur.l.WithContext(ctx).Errorf("GetCDK failed, err=%v, waID=%s", err, rallyInfo.WaID)
		return err
	}
	if ok {
		err := ur.activityInfoUC.UpdateActivityInfo(ctx, &UpdateActivityInfoDto{
			Id:             ur.business.Activity.Id,
			ActivityStatus: constants.ATStatusBuffer,
		})
		if err != nil && errors.Is(err, sqlx.ErrRowsAffected) {
			ur.l.WithContext(ctx).Errorf("UpdateActivityInfo failed, err=%v, waID=%s", err, rallyInfo.WaID)
			// 忽略error
		}
	}

	rallyInfo.CDK = cdk
	rallyInfo.CDKType = 0

	return nil
}

func (ur *UnOfficialRallyUsecase) buildHelpMsgInfo(
	ctx context.Context,
	joinWaIDs []string,
	rallyInfo, helpInfo *BaseInfo,
	joinCount int,
) ([]*dto.SendNxListParamsDto, error) {
	cost := util.MethodCost(ctx, ur.l, "buildHelpMsgInfo")
	defer cost()

	userInfos, err := ur.userInfoRepo.FindUserInfos(ctx, joinWaIDs)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("FindUserInfos faileds, err=%v, joinWaIds=%v", err, joinWaIDs)
		return nil, err
	}

	helpNickNames := lo.Map(userInfos, func(userInfo *UserInfo, _ int) *dto.HelpNickNameInfo {
		return &dto.HelpNickNameInfo{
			UserNickname: userInfo.Nickname,
		}
	})
	helpNickNames = append(helpNickNames, &dto.HelpNickNameInfo{
		UserNickname: rallyInfo.NickName,
	})

	// 有奖励
	if helpInfo.CDK != "" {
		channel, language, generation := helpInfo.Channel, helpInfo.Language, helpInfo.Generation
		helpMessages, err := ur.waMsg.HelpOverMsg2NX(ctx, &dto.BuildMsgInfo{
			WaId:       helpInfo.WaID,
			MsgType:    fmt.Sprintf("%s%d", constants.HelpOverMsgPrefix, (joinCount+1)%3),
			Channel:    channel,
			Language:   language,
			Generation: fmt.Sprint(generation),
			RallyCode:  helpInfo.RallyCode,
		}, helpInfo.CdkShortLink, helpNickNames, constants.BizTypeInteractive, helpInfo.RallyCodeShortLink)
		if err != nil {
			ur.l.WithContext(ctx).Errorf("HelpOverMsg2NX failed, err=%v, helpInfo=%+v", err, helpInfo)
			return nil, err
		}

		return helpMessages, nil
	}

	// 无奖励
	channel, language, generation := helpInfo.Channel, helpInfo.Language, helpInfo.Generation
	ur.l.WithContext(ctx).Infof("HelpTaskSingleSuccessMsg2NX param %v", helpInfo)
	helpMessages, err := ur.waMsg.HelpTaskSingleSuccessMsg2NX(ctx, &dto.BuildMsgInfo{
		WaId:       helpInfo.WaID,
		MsgType:    fmt.Sprintf("%s%d", constants.HelpOverMsgPrefix, (joinCount+1)%3),
		Channel:    channel,
		Language:   language,
		Generation: fmt.Sprint(generation),
		RallyCode:  helpInfo.RallyCode,
	}, helpInfo.RallyCodeShortLink, helpNickNames)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("HelpOverMsg2NX failed, err=%v, helpInfo=%+v", err, helpInfo)
		return nil, err
	}

	return helpMessages, nil
}

func (ur *UnOfficialRallyUsecase) buildRallyMsgInfo(
	ctx context.Context,
	rallyInfo *BaseInfo,
) ([]*dto.SendNxListParamsDto, error) {
	channel, language, generation := rallyInfo.Channel, rallyInfo.Language, rallyInfo.Generation

	// 助力人处理
	rallyMsgs, err := ur.waMsg.StartGroupMsg2NX(ctx, &dto.BuildMsgInfo{
		WaId:       rallyInfo.WaID,
		MsgType:    constants.StartGroupMsg,
		Channel:    channel,
		Language:   language,
		Generation: fmt.Sprint(generation),
		RallyCode:  rallyInfo.RallyCode,
	}, rallyInfo.RallyCodeShortLink, nil, true)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("StartGroupMsg2NX failed, err=%v, waID=%s, rallyCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return nil, err
	}

	return rallyMsgs, nil
}

// joinGroup 助力
func (ur *UnOfficialRallyUsecase) joinGroup(
	ctx context.Context,
	rallyInfo, helpInfo *BaseInfo,
	joinWaIDs []string,
	oldJoinNum, newJoinNum int, // 参团人数
	helpCodeFact HelpCodeInterface,
) error {
	cost := util.MethodCost(ctx, ur.l, "unofficialRallyUsecase.joinGroup")
	defer cost()

	err := ur.buildHelpBaseInfo(ctx, helpInfo, oldJoinNum, helpCodeFact)
	if err != nil {
		ur.l.Errorf("buildHelpBaseInfo failed, err=%v, waID=%s, helpCode=%s", err, helpInfo.WaID, helpInfo.RallyCode)
		return err
	}

	err = ur.buildRallyInfo(ctx, rallyInfo, helpCodeFact)
	if err != nil {
		ur.l.Errorf("buildRallyInfo failed, err=%v, waID=%s, helpCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	start := time.Now()
	messages, err := ur.buildHelpMsgInfo(ctx, joinWaIDs, rallyInfo, helpInfo, oldJoinNum)
	if err != nil {
		return err
	}

	rallyMsgs, err := ur.buildRallyMsgInfo(ctx, rallyInfo)
	if err != nil {
		return err
	}
	ur.l.WithContext(ctx).Infof("buildRallyAndHelpMsgInfo cost: %s", time.Since(start).String())

	messages = append(messages, rallyMsgs...)

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	// 需要处理的事项：
	// 1. 助力人
	//     case1(新增助力人): newJoinGroup、newCreateGroup、newUserRemind、newUserInfo
	//     case2(不新增助力人): newJoinGroup、updateUserRemind
	// 2. 被助力人
	// 	   case1(有奖励): updateUserInfo.JoinNum+对应CDK
	//     case2(无奖励): updateUserInfo.JoinNum
	// 3. 发送消息
	//	   msgSends表新增数据
	// 4. 更新消息完成
	//     非官方消息处理完成
	err = ur.rallyRepo.
		CreateJoinGroup2(ctx, rallyInfo, helpInfo, msgSends, newJoinNum)
	ur.l.WithContext(ctx).Infof("CreateJoinGroup2 cost: %s", time.Since(start).String())
	if err != nil {
		ur.l.WithContext(ctx).Errorf("CreateJoinGroup2 failed, err=%v, waID=%s, rallyCode=%s", err, rallyInfo.WaID, rallyInfo.RallyCode)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, messages)
	ur.l.WithContext(ctx).Infof("SendMsgList2NX cost: %s", time.Since(start).String())
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, ret=%s",
			err, rallyInfo.WaID, rallyInfo.RallyCode, r)
		return nil
	}

	return nil
}

// unattendHandle: 不可以助力流程
func (ur *UnOfficialRallyUsecase) unattendHandle(
	ctx context.Context,
	rallyData *queue.RallyData,
	withMsgDB bool,
) error {
	messages, err := ur.waMsg.EndCanNotHelpMsg(ctx, &dto.BuildMsgInfo{
		WaId:       rallyData.WaID,
		MsgType:    constants.EndCanNotHelpMsg,
		Channel:    rallyData.Channel,
		Language:   rallyData.Language,
		Generation: rallyData.Generation,
	})
	if err != nil {
		ur.l.WithContext(ctx).Errorf("EndCanNotHelpMsg failed, err=%v", err)
	}

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	waID, rallyCode := rallyData.WaID, rallyData.RallyCode
	err = ur.rallyRepo.CompleteRally(ctx, waID, rallyCode, msgSends, withMsgDB)
	if err != nil {
		if errors.Is(err, sqlx.ErrRowsAffected) {
			// 已经完成或错误数据
			ur.l.WithContext(ctx).
				Warnf(`completeRally failed, err=%v, waID=%s, helpCode=%s`, err, waID, rallyCode)
			return nil
		}

		ur.l.WithContext(ctx).
			Errorf(`completeRally failed, err=%v, waID=%s, helpCode=%s`, err, waID, rallyCode)
		return err
	}

	r, err := ur.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		ur.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, ret=%s", err, waID, rallyCode, r)
		return nil
	}

	return nil
}

func (ur *UnOfficialRallyUsecase) RetryMsg(ctx context.Context) error {
	ur.l.WithContext(ctx).Info("unofficial rally retry msg")
	defer ur.l.WithContext(ctx).Info("unofficial rally retry msg done")

	var (
		minID       = 0
		maxSendTime = time.Now().Add(-2 * time.Minute) // 暂定2分钟
		offset      = uint(0)
		length      = uint(100)
		batchSize   = 20
		rallyDatas  = make([]*queue.RallyData, 0, batchSize)
		sendQ       = func() {
			err := ur.q.SendBacks(rallyDatas)
			if err != nil {
				// 忽略这个错误
				ur.l.WithContext(ctx).Errorf("SendBackForce failed, err=%v, rallys=%+v", err, rallyDatas)
			}
			rallyDatas = rallyDatas[:0]
		}
	)

	for {
		msgs, err := ur.rallyRepo.ListDoingMsgs(ctx, minID, offset, length, maxSendTime)
		if err != nil {
			ur.l.WithContext(ctx).Errorf("list doing msg failed, err=%v", err)
			return err
		}

		for i := range msgs {
			rally := &queue.RallyData{
				WaID:       msgs[i].WaID,
				RallyCode:  msgs[i].RallyCode,
				Channel:    msgs[i].Channel,
				Language:   msgs[i].Language,
				Generation: fmt.Sprint(msgs[i].Generation),
				NickName:   msgs[i].NickName,
				SendTime:   msgs[i].SendTime,
			}

			rallyDatas = append(rallyDatas, rally)
			if (i+1)%20 == 0 {
				sendQ()
			}
		}

		if len(rallyDatas) > 0 {
			sendQ()
		}

		if len(msgs) < int(length) {
			break
		}

		minID = msgs[len(msgs)-1].ID
	}

	return nil
}
