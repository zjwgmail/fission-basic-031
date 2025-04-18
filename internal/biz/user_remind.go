package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"

	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/util/goroutine_pool"
)

type UserRemindUsecase struct {
	d      *conf.Data
	l      *log.Helper
	urRepo UserRemindRepo

	waMsgUsease *WaMsgService
}

func NewUserRemindUsecase(
	d *conf.Data,
	l log.Logger,
	urRepo UserRemindRepo,
	waMsgUsease *WaMsgService,
) *UserRemindUsecase {
	return &UserRemindUsecase{
		l:           log.NewHelper(l),
		urRepo:      urRepo,
		d:           d,
		waMsgUsease: waMsgUsease,
	}
}

// RemindJoinGroupV3 催团
func (u *UserRemindUsecase) RemindJoinGroupV3(
	ctx context.Context,
	helpCode HelpCodeInterface,
) error {
	var (
		minID       = int64(0)
		minSendTime = time.Now().Unix()
		offset      = uint(0)
		length      = uint(2000)
	)

	v3Func := func(ctx context.Context, userRemind *UserRemind) error {
		userInfo, err := u.urRepo.GetUserInfo(ctx, userRemind.WaID)
		if err != nil {
			u.l.WithContext(ctx).Errorf("get user info failed, err=%v, waID=%s", err, userRemind.WaID)
			return nil
		}

		var (
			v3NewStatus int
			sendMsg     bool
		)

		if userInfo.JoinCount >= int(u.d.JoinGroup.MaxNum) {
			sendMsg = false
			v3NewStatus = UserRemindVXStatusMsgDone
		} else {
			sendMsg = true
			v3NewStatus = UserRemindVXStatusMsgSend
		}

		msgs, callback, err := u.buildRemindV3Messages(ctx, sendMsg, userRemind, userInfo, helpCode)
		if err != nil {
			u.l.WithContext(ctx).Errorf("buildRemindV3Messages failed, err=%v", err)
			return nil
		}

		err = u.urRepo.CompleteUserRemindV3Status(ctx, userRemind.WaID, userRemind.StatusV3, v3NewStatus, msgs)
		if err != nil {
			u.l.WithContext(ctx).Errorf("complete user remind v3 status failed, err=%v", err)
			return nil
		}

		if callback != nil {
			callback()
		}

		return nil
	}

	u.l.WithContext(ctx).Infof("start remind join group v3")
	for {
		userReminds, err := u.urRepo.ListUserRemindTODOV3(ctx, offset, length, minID, minSendTime)
		if err != nil {
			u.l.WithContext(ctx).Errorf("ListUserRemindTODOV3 failed, err=%v", err)
			return err
		}
		if len(userReminds) == 0 {
			u.l.WithContext(ctx).Infof("all done no user remind v3")
			break
		}

		funcs := make([]func(ctx context.Context) error, 0, len(userReminds))
		for i := range userReminds {
			userRemind := userReminds[i]
			funcs = append(funcs, func(ctx context.Context) error {
				return v3Func(ctx, userRemind)
			})
		}

		err = goroutine_pool.ParallN2(ctx, 30, funcs)
		if err != nil {
			u.l.WithContext(ctx).Errorf("ParallN2 failed, err=%v", err)
		}

		minID = userReminds[len(userReminds)-1].ID
	}

	return nil
}

func (u *UserRemindUsecase) buildRemindV3Messages(ctx context.Context,
	sendMsg bool,
	userRemind *UserRemind, userInfo *UserInfo,
	helpCode HelpCodeInterface,
) ([]*dto.WaMsgSend, func(), error) {
	if !sendMsg {
		return nil, nil, nil
	}

	shorLink, err := helpCode.GetShortLinkByHelpCode(ctx, userInfo.HelpCode, 0)
	if err != nil {
		u.l.WithContext(ctx).Errorf("GetShortLinkByHelpCode failed, err=%v", err)
		return nil, nil, err
	}

	messages, err := u.waMsgUsease.PromoteClusteringMsg2NX(
		ctx,
		&dto.BuildMsgInfo{
			MsgType:    constants.PromoteClusteringMsg,
			WaId:       userRemind.WaID,
			RallyCode:  userInfo.HelpCode,
			Language:   userInfo.Language,
			Channel:    userInfo.Channel,
			Generation: fmt.Sprint(userInfo.Generation),
		},
		shorLink,
		constants.BizTypeInteractive,
		nil,
	)
	if err != nil {
		u.l.WithContext(ctx).Errorf("renew free msg failed, err=%v", err)
		return nil, nil, err
	}
	waMsgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	return waMsgSends, func() {
		r, err := u.waMsgUsease.SendMsgList2NX(ctx, messages)
		if err != nil {
			u.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, data=%v, ret=%s", err, messages, r)
			return
		}
	}, nil
}

// FreeDurationRenewV22 免费24小时续时
func (u *UserRemindUsecase) FreeDurationRenewV22(ctx context.Context) error {
	var (
		minID       = int64(0)
		minSendTime = time.Now().Unix()
		offset      = uint(0)
		length      = uint(2000)
	)

	v22Func := func(ctx context.Context, userRemind *UserRemind) error {
		userInfo, err := u.urRepo.GetUserInfo(ctx, userRemind.WaID)
		if err != nil {
			u.l.WithContext(ctx).Errorf("get user info failed, err=%v, waID=%s", err, userRemind.WaID)
			return nil
		}

		var (
			v22NewStatus int
			sendMsg      bool
		)

		if userInfo.JoinCount >= int(u.d.JoinGroup.MaxNum) {
			sendMsg = false
			v22NewStatus = UserRemindVXStatusMsgDone
		} else {
			sendMsg = true
			v22NewStatus = UserRemindVXStatusMsgSend
		}

		waMsgSends, callback, err := u.buildFree22Messages(ctx, sendMsg, userRemind, userInfo)
		if err != nil {
			u.l.WithContext(ctx).Errorf("build free22 messages failed, err=%v", err)
			return nil
		}

		err = u.urRepo.CompleteUserRemindV22Status(ctx, userRemind.WaID, userRemind.StatusV22, v22NewStatus, waMsgSends)
		if err != nil {
			u.l.WithContext(ctx).Errorf("complete user remind v22 status failed, err=%v", err)
			return nil
		}

		if callback != nil {
			callback()
		}

		return nil
	}

	for {
		userReminds, err := u.urRepo.ListUserRemindTODOV22(ctx, offset, length, minID, minSendTime)
		if err != nil {
			u.l.WithContext(ctx).Errorf("list user remind failed, err=%v", err)
			return err
		}

		funcs := make([]func(context.Context) error, 0, len(userReminds))
		for i := range userReminds {
			userRemind := userReminds[i]
			funcs = append(funcs, func(ctx context.Context) error {
				return v22Func(ctx, userRemind)
			})
		}

		err = goroutine_pool.ParallN2(ctx, 40, funcs)
		if err != nil {
			u.l.WithContext(ctx).Errorf("ParallN2 failed,err=%v", err)
		}

		if len(userReminds) < int(length) {
			break
		}

		minID = userReminds[len(userReminds)-1].ID
	}

	return nil
}

func (u *UserRemindUsecase) buildFree22Messages(ctx context.Context,
	sendMsg bool,
	userRemind *UserRemind, userInfo *UserInfo,
) ([]*dto.WaMsgSend, func(), error) {
	if !sendMsg {
		return nil, nil, nil
	}

	messages, err := u.waMsgUsease.RenewFreeMsg(
		ctx,
		&dto.BuildMsgInfo{
			MsgType:    constants.PromoteClusteringMsg,
			WaId:       userRemind.WaID,
			RallyCode:  userInfo.HelpCode,
			Language:   userInfo.Language,
			Channel:    userInfo.Channel,
			Generation: fmt.Sprint(userInfo.Generation),
		},
		constants.BizTypeInteractive,
	)
	if err != nil {
		u.l.WithContext(ctx).Errorf("renew free msg failed, err=%v", err)
		return nil, nil, err
	}
	waMsgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	return waMsgSends, func() {
		r, err := u.waMsgUsease.SendMsgList2NX(ctx, messages)
		if err != nil {
			u.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, data=%v, ret=%s", err, messages, r)
			return
		}
	}, nil
}

// 免费CDK消息
func (u *UserRemindUsecase) builV0dMessages(ctx context.Context,
	sendMsg bool,
	userRemind *UserRemind, userInfo *UserInfo,
) ([]*dto.WaMsgSend, func(), error) {
	if !sendMsg {
		return nil, nil, nil
	}

	messages, err := u.waMsgUsease.FreeCdkMsg2NX(
		ctx,
		&dto.BuildMsgInfo{
			MsgType:    constants.FreeCdkMsg,
			WaId:       userRemind.WaID,
			RallyCode:  userInfo.HelpCode,
			Language:   userInfo.Language,
			Channel:    userInfo.Channel,
			Generation: fmt.Sprint(userInfo.Generation),
		},
		userInfo.CDKv0,
		constants.BizTypeInteractive,
	)
	if err != nil {
		u.l.WithContext(ctx).Errorf("FreeCdkMsg2NX failed, err=%v", err)
		return nil, nil, err
	}
	waMsgSends := lo.Map(messages, func(m *dto.SendNxListParamsDto, _ int) *dto.WaMsgSend {
		return m.WaMsgSend
	})

	return waMsgSends, func() {
		r, err := u.waMsgUsease.SendMsgList2NX(ctx, messages)
		if err != nil {
			u.l.WithContext(ctx).Errorf("SendMsgList2NX failed, err=%v, data=%v, ret=%s", err, messages, r)
			return
		}
	}, nil
}

// CDKV0 免费SDK消息（红包）
func (u *UserRemindUsecase) CDKV0(ctx context.Context) error {
	var (
		minID       = int64(0)
		minSendTime = time.Now().Unix()
		offset      = uint(0)
		length      = uint(2000)
	)

	v0Func := func(ctx context.Context, userRemind *UserRemind) error {

		userInfo, err := u.urRepo.GetUserInfo(ctx, userRemind.WaID)
		if err != nil {
			u.l.WithContext(ctx).Errorf("get user info failed, err=%v, waID=%s", err, userRemind.WaID)
			return nil
		}

		var (
			v0NewStatus = UserRemindVXStatusMsgSend
			sendMsg     = true
		)
		waMsgSends, callback, err := u.builV0dMessages(ctx, sendMsg, userRemind, userInfo)
		if err != nil {
			u.l.WithContext(ctx).Errorf("builV0dMessages failed, err=%v", err)
			return nil
		}

		err = u.urRepo.CompleteUserRemindV0Status(ctx, userRemind.WaID, userRemind.StatusV0, v0NewStatus, waMsgSends)
		if err != nil {
			u.l.WithContext(ctx).Errorf("complete user remind v0 status failed, err=%v", err)
			return nil
		}

		if callback != nil {
			callback()
		}

		return nil
	}

	for {
		userReminds, err := u.urRepo.ListUserRemindTODOV0(ctx, offset, length, minID, minSendTime)
		if err != nil {
			u.l.WithContext(ctx).Errorf("list user remind failed, err=%v", err)
			return err
		}

		funcs := make([]func(context.Context) error, 0, len(userReminds))
		for i := range userReminds {
			userRemind := userReminds[i]
			funcs = append(funcs, func(ctx context.Context) error {
				return v0Func(ctx, userRemind)
			})
		}

		err = goroutine_pool.ParallN2(ctx, 15, funcs)
		if err != nil {
			u.l.WithContext(ctx).Errorf("ParallN2 failed, err=%v", err)
		}

		if len(userReminds) < int(length) {
			break
		}

		minID = userReminds[len(userReminds)-1].ID
	}

	return nil
}
