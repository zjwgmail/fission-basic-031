package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"

	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/queue"
	"fission-basic/internal/pojo/dto"
	"fission-basic/kit/sqlx"
)

type OfficialRallyUsecase struct {
	attendEnable bool
	d            *conf.Data
	b            *conf.Business
	rallyRepo    OfficialRallyRepo
	l            *log.Helper

	q *queue.Official

	// TODO: 后续改掉，不应该这么调用
	waMsg          *WaMsgService
	activityInfoUC *ActivityInfoUsecase
}

func NewOfficialRallyUsecase(
	d *conf.Data,
	b *conf.Business,
	rallyRepo OfficialRallyRepo,
	q *queue.Official,
	l log.Logger,
	waMsg *WaMsgService,
	activityInfoUC *ActivityInfoUsecase,
) *OfficialRallyUsecase {
	return &OfficialRallyUsecase{
		attendEnable:   d.AttendEnable,
		rallyRepo:      rallyRepo,
		l:              log.NewHelper(l),
		d:              d,
		b:              b,
		q:              q,
		waMsg:          waMsg,
		activityInfoUC: activityInfoUC,
	}
}

type HelpCodeInterface interface {
	GetHelpCode(ctx context.Context) (string, error)
	GetShortLinkByHelpCode(ctx context.Context, helpCode string, shortLinkVersion int) (string, error)
}

func (fu *OfficialRallyUsecase) Handler(
	ctx context.Context,
	data *queue.RallyData,
	helpCodeFact HelpCodeInterface,
) error {
	activityInfo, err := fu.activityInfoUC.GetActivityInfo(ctx)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("GetActivityInfo failed, err=%v", err)
		return err
	}

	// 活动进行中
	if activityInfo.ActivityStatus == constants.ATStatusStarted {
		return fu.attendHandle(ctx, data, helpCodeFact)
	}

	//结束期和缓冲期
	return fu.unattendHandle(ctx, activityInfo, data)
}

func (fu *OfficialRallyUsecase) unattendHandle(
	ctx context.Context,
	activityInfo *ActivityInfoDto,
	data *queue.RallyData,
) error {
	msgType := constants.FounderCanNotStartGroupMsg
	if activityInfo.ActivityStatus == constants.ATStatusEnd {
		msgType = constants.EndCanNotStartGroupMsg
	}

	// 活动缓冲期
	messages, err := fu.waMsg.FounderCanNotStartGroupMsg(ctx, &dto.BuildMsgInfo{
		WaId:       data.WaID,
		MsgType:    msgType,
		Channel:    data.Channel,
		Language:   data.Language,
		Generation: data.Generation,
		RallyCode:  data.RallyCode,
	})
	if err != nil {
		fu.l.Errorf("FounderCanNotStartGroupMsg failed, err=%v, data=%v", err, data)
		return err
	}

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	waID, rallyCode := data.WaID, data.RallyCode
	err = fu.rallyRepo.CompleteRally(ctx, waID, rallyCode, msgSends)
	if err != nil {
		if errors.Is(err, sqlx.ErrRowsAffected) {
			// 已经完成或错误数据
			fu.l.WithContext(ctx).
				Warnf(`completeRally failed, err=%v, waID=%s, rallyCode=%s`, err, waID, rallyCode)
			return nil
		}
		fu.l.WithContext(ctx).
			Errorf(`completeRally failed, err=%v, waID=%s, rallyCode=%s`, err, waID, rallyCode)
		return err
	}

	r, err := fu.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, r=%s", err, waID, rallyCode, r)
		return nil
	}

	return nil
}

func (fu *OfficialRallyUsecase) attendHandle(
	ctx context.Context,
	data *queue.RallyData,
	helpCodeFact HelpCodeInterface,
) error {
	// fu.l.WithContext(ctx).Debugw("data", data)

	waID := data.WaID
	rallyCode := data.RallyCode
	officialMsg, err := fu.rallyRepo.FindMsg(ctx, waID, rallyCode)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			fu.l.WithContext(ctx).
				Errorf("find official msg failed, err=%v, waID=%s, rallyCode=%s", err, waID, rallyCode)
			return err
		}

		// 没找到走未处理逻辑，只记录一条日志，不做其它处理
		fu.l.WithContext(ctx).
			Warnf("find official msg failed, err=%v, waID=%s, rallyCode=%s", err, waID, rallyCode)
	}

	// 已完成
	if officialMsg != nil && officialMsg.State == MsgStateComplete {
		// TODO: 实际这是重复处理了，直接舍弃消息即可，待@jianwu确认
		return nil
	}

	_, err = fu.rallyRepo.FindUserCreateGroup(ctx, waID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			fu.l.WithContext(ctx).Errorf("find user create failed, err=%v, waID=%s", err, waID)
			return err
		}

		// 开团
		return fu.newUserGroup(ctx, data, helpCodeFact)
	}

	// 已经开团
	// fu.l.Debugw("ucg", userCreateGroup)

	return fu.switchLanguage(ctx, data)
}

func (fu *OfficialRallyUsecase) switchLanguage(ctx context.Context, data *queue.RallyData) error {
	messages, err := fu.waMsg.SwitchLangMsg(ctx, &dto.BuildMsgInfo{
		WaId:       data.WaID,
		MsgType:    constants.SwitchLangMsg,
		Channel:    data.Channel,
		Language:   data.Language,
		Generation: data.Generation,
		RallyCode:  data.RallyCode,
	})
	if err != nil {
		fu.l.WithContext(ctx).Errorf("SwitchLangMsg failed, err=%v, data=%v",
			err, data)
		return err
	}

	msgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	err = fu.rallyRepo.UpdateUserInfoLanguageByWaID(ctx, data.WaID, data.RallyCode, data.Language, msgSends)
	if err != nil {
		if errors.Is(err, sqlx.ErrRowsAffected) {
			//TODO: 未实际更新，不发送消息了，@jianwu确认
			return nil
		}

		fu.l.WithContext(ctx).
			Errorf(`UpdateUserInfoLanguageByWaID failed, err=%v, data=%s`,
				err, data)
		return err
	}

	r, err := fu.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, data=%v, ret=%s",
			err, data, r)
		return nil
	}

	return nil
}

// newUserGroup 新开团
func (fu *OfficialRallyUsecase) newUserGroup(
	ctx context.Context,
	data *queue.RallyData,
	helpCodeFact HelpCodeInterface,
) error {
	helpCode, err := helpCodeFact.GetHelpCode(ctx)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("GetHelpCode failed, err=%v, data=%v", err, data)
		return err
	}

	shortLink, err := helpCodeFact.GetShortLinkByHelpCode(ctx, helpCode, 0)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v, data=%v", err, data)
		return err
	}

	messages, err := fu.waMsg.StartGroupMsg2NX(ctx, &dto.BuildMsgInfo{
		WaId:       data.WaID,
		MsgType:    constants.StartGroupMsg,
		Channel:    data.Channel,
		Language:   data.Language,
		Generation: fmt.Sprint(data.GenerationInt64() + 1),
		RallyCode:  helpCode,
	}, shortLink, nil, false)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("StartGroupMsg2NX failed, err=%v, data=%v", err, data)
		return err
	}

	waMsgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	userInfo := &UserInfo{
		WaID:       data.WaID,
		HelpCode:   helpCode,
		Channel:    data.Channel,
		Language:   data.Language,
		Generation: 1,
		Nickname:   data.NickName,
		JoinCount:  0,
	}

	err = fu.rallyRepo.CreateUserGroup(ctx, userInfo, data.SendTime,
		data.WaID, data.RallyCode, helpCode, waMsgSends)
	if err != nil {
		fu.l.WithContext(ctx).
			Errorf("CrateUserGroup failed, err=%v, waID=%s, rallyCode=%s", err, data.WaID, helpCode)
		return err
	}

	r, err := fu.waMsg.SendMsgList2NX(ctx, messages)
	if err != nil {
		fu.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, waID=%s, rallyCode=%s, r=%s",
			err, data.WaID, helpCode, r)
		return nil
	}

	return nil
}

func (fu *OfficialRallyUsecase) RetryMsg(ctx context.Context) error {
	fu.l.WithContext(ctx).Infof("official RetryMsg start")
	defer fu.l.WithContext(ctx).Infof("official RetryMsg end")

	var (
		minID       = 0
		maxSendTime = time.Now().Add(-2 * time.Minute) // 暂定2分钟
		offset      = uint(0)
		length      = uint(100)
		// 每20个推一次
		batchSize = 20
		rallys    = make([]*queue.RallyData, 0, batchSize)
		sendQ     = func() {
			err := fu.q.SendBacks(rallys)
			if err != nil {
				// 忽略这个错误
				fu.l.WithContext(ctx).Errorf("SendBackForce failed, err=%v, rallys=%+v", err, rallys)
				err = nil
			}
			rallys = rallys[:0]
		}
	)

	for {
		msgs, err := fu.rallyRepo.ListDoingMsg(ctx, minID, offset, length, maxSendTime)
		if err != nil {
			fu.l.WithContext(ctx).Errorf("list doing msg failed, err=%v", err)
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
			rallys = append(rallys, rally)

			// 够一个批次
			if (i+1)%batchSize == 0 {
				sendQ()
			}
		}

		if len(rallys) > 0 {
			sendQ()
		}

		if len(msgs) < int(length) {
			break
		}

		minID = msgs[len(msgs)-1].ID
	}

	return nil
}
