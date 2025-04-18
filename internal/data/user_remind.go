package data

import (
	"context"
	"errors"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pojo/dto"
	"fission-basic/kit/sqlx"
	"fission-basic/util"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"
)

var _ biz.UserRemindRepo = (*UserRemind)(nil)

type UserRemind struct {
	data *Data
	l    *log.Helper
}

func NewUserRemind(
	d *Data,
	l log.Logger,
) biz.UserRemindRepo {
	return &UserRemind{
		data: d,
		l:    log.NewHelper(l),
	}
}

// CompleteUserRemindV3Status implements biz.UserRemindRepo.
func (u *UserRemind) CompleteUserRemindV3Status(ctx context.Context,
	waID string, oldStatus, status int,
	waMsgs []*dto.WaMsgSend) error {
	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateUserRemindV3Status(ctx, u.data.db, waID, oldStatus, status)
		if err != nil {
			u.l.WithContext(ctx).Errorf("update user remind v3 status failed, err=%v, waID=%s, oldStatus=%d, status=%d", err, waID, oldStatus, status)
			return err
		}

		err = u.addWaMsgs(ctx, time.Now(), db, waMsgs)
		if err != nil {
			return err
		}

		return nil
	})
}

// ListUserRemindTODOV3 implements biz.UserRemindRepo.
func (u *UserRemind) ListUserRemindTODOV3(ctx context.Context, offset, length uint, minID int64, minSendTime int64) ([]*biz.UserRemind, error) {
	userReminds, err := model.SelectUserRemindsTODOV3(ctx, u.data.db, offset, length, minID, minSendTime)
	if err != nil {
		u.l.WithContext(ctx).Errorf("select user remind failed, err=%v", err)
		return nil, err
	}

	return lo.Map(userReminds, func(userRemind *model.UserRemind, _ int) *biz.UserRemind {
		return convertUserRemind2Biz(userRemind)
	}), nil
}

// ListUserRemindTODOV0 implements biz.UserRemindRepo.
func (u *UserRemind) ListUserRemindTODOV0(ctx context.Context, offset, length uint, minID, minSendTime int64) ([]*biz.UserRemind, error) {
	userReminds, err := model.SelectUserRemindsTODOV0(ctx, u.data.db, offset, length, minID, minSendTime)
	if err != nil {
		u.l.WithContext(ctx).Errorf("select user remind failed, err=%v", err)
		return nil, err
	}

	return lo.Map(userReminds, func(userRemind *model.UserRemind, _ int) *biz.UserRemind {
		return convertUserRemind2Biz(userRemind)
	}), nil
}

// ListUserRemindTODOV22 implements biz.UserRemindRepo.
func (u *UserRemind) ListUserRemindTODOV22(ctx context.Context, offset, length uint, minID, minSendTime int64) ([]*biz.UserRemind, error) {
	userReminds, err := model.SelectUserRemindsTODOV22(ctx, u.data.db, offset, length, minID, minSendTime)
	if err != nil {
		u.l.WithContext(ctx).Errorf("select user remind failed, err=%v", err)
		return nil, err
	}

	return lo.Map(userReminds, func(userRemind *model.UserRemind, _ int) *biz.UserRemind {
		return convertUserRemind2Biz(userRemind)
	}), nil
}

// CompleteUserRemindV0Status implements biz.UserRemindRepo.
func (u *UserRemind) CompleteUserRemindV0Status(ctx context.Context,
	waID string, oldStatus, status int,
	waMsgSends []*dto.WaMsgSend,
) error {
	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateUserRemindV0Status(ctx, u.data.db, waID, oldStatus, status)
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("update user remind freeCDK status failed, err=%v, waID=%s, oldStatus=%d, status=%d", err, waID, oldStatus, status)
			return err
		}

		err = u.addWaMsgs(ctx, time.Now(), db, waMsgSends)
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserRemind) addWaMsgs(ctx context.Context, t time.Time, db sqlx.DB, waMsgSends []*dto.WaMsgSend) error {
	for i := range waMsgSends {
		id, err := model.InsertWaMsgSend(ctx, db,
			&model.WaMsgSend{
				WaMsgID:       waMsgSends[i].WaMsgID,
				WaID:          waMsgSends[i].WaID,
				MsgType:       waMsgSends[i].MsgType,
				State:         waMsgSends[i].State,
				Content:       waMsgSends[i].Content,
				BuildMsgParam: waMsgSends[i].BuildMsgParam,
				SendRes:       waMsgSends[i].SendRes,
				CreateTime:    t,
				UpdateTime:    t,
				Del:           biz.NotDeleted,
			},
		)
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("insert wa msg send failed, err=%v", err)
			return err
		}
		waMsgSends[i].ID = id
	}
	return nil
}

// CompleteUserRemindV22Status implements biz.UserRemindRepo.
func (u *UserRemind) CompleteUserRemindV22Status(ctx context.Context,
	waID string, oldStatus, status int,
	waMsgSends []*dto.WaMsgSend,
) error {
	now := time.Now()
	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateUserRemindV22Status(ctx, u.data.db, waID, oldStatus, status)
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("update user remind v22 status failed, err=%v, waID=%s, oldStatus=%d, status=%d", err, waID, oldStatus, status)
			return err
		}

		err = u.addWaMsgs(ctx, now, db, waMsgSends)
		if err != nil {
			return err
		}

		return nil
	})
}

// GetUserInfo implements biz.UserRemindRepo.
func (u *UserRemind) GetUserInfo(ctx context.Context, waID string) (*biz.UserInfo, error) {
	userInfo, err := model.GetUserInfo(ctx, u.data.db, waID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).Errorf("get user info failed, err=%v, waID=%s", err, waID)
		}
		return nil, err
	}

	return convertUserInfo2Biz(userInfo), nil
}

// GetUserRemindInfo implements biz.UserRemindRepo.
func (u *UserRemind) GetUserRemindInfo(ctx context.Context, waID string) (*dto.UserRemindDto, error) {
	userInfo, err := model.GetUserRemindByWaID(ctx, u.data.db, waID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).Errorf("get UserRemind info failed, err=%v, waID=%s", err, waID)
		}
		return nil, err
	}

	res := &dto.UserRemindDto{}
	err = util.CopyFieldsByJson(*userInfo, res)
	if err != nil {
		u.l.WithContext(ctx).Errorf("copy struct failed, err=%v, waID=%s", err, waID)
		return nil, err
	}
	return res, nil
}
