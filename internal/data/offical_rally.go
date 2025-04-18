package data

import (
	"context"
	"errors"
	"time"

	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pojo/dto"
	"fission-basic/kit/sqlx"

	"github.com/samber/lo"
)

var _ biz.OfficialRallyRepo = (*OfficialRally)(nil)

type OfficialRally struct {
	*Rally
}

func NewOfficialRally(
	rally *Rally,
) biz.OfficialRallyRepo {
	return &OfficialRally{
		Rally: rally,
	}
}

func (o *OfficialRally) completeMsgRecordState(ctx context.Context, db sqlx.DB, waID, rallyCode string) error {
	err := model.UpdateOfficialMsgRecordState(ctx, db, waID, rallyCode,
		biz.MsgStateDoing, biz.MsgStateComplete)
	if err != nil {
		o.l.WithContext(ctx).
			Errorf("update msg record state failed, err=%v", err)
		return err
	}

	return nil
}

// UpdateUserInfoLanguageByWaID implements biz.OfficialRallyRepo.
func (o *OfficialRally) UpdateUserInfoLanguageByWaID(
	ctx context.Context, waID, rallyCode, language string,
	msgSends []*dto.WaMsgSend) error {
	err := o.Rally.updateUserInfoLanguageByWaID(
		ctx, waID, rallyCode, language, msgSends, time.Now(),
		func(ctx context.Context, db sqlx.DB) error {
			return o.completeMsgRecordState(ctx, db, waID, rallyCode)
		})
	if err != nil {
		if !errors.Is(err, sqlx.ErrRowsAffected) {
			o.l.WithContext(ctx).
				Errorf("update user info language failed, err=%v, waID=%s, rallyCode=%s, language=%s",
					err, waID, rallyCode, language)
			return err
		}

		err1 := o.completeMsgRecordState(ctx, o.data.db, waID, rallyCode)
		if err1 != nil {
			o.l.WithContext(ctx).Errorf("complete msg record state failed, err=%v, waID=%s, rallyCode=%s", err1, waID, rallyCode)
		}

		return err
	}

	return nil
}

// CompleteRally implements biz.OfficialRallyRepo.
// 完成官方消息处理
func (o *OfficialRally) CompleteRally(ctx context.Context, waID string, rallyCode string,
	waMsgSends []*dto.WaMsgSend,
) error {
	now := time.Now()
	return sqlx.TxContext(ctx, o.data.db, func(ctx context.Context, tx sqlx.DB) error {
		err := o.completeMsgRecordState(ctx, tx, waID, rallyCode)
		if err != nil {
			return err
		}

		err = o.saveMsgSend(ctx, now, tx, waMsgSends)
		if err != nil {
			return nil
		}

		return nil
	})
}

// CreateUserGroup implements biz.OfficialRallyRepo.
// 开团
func (o *OfficialRally) CreateUserGroup(ctx context.Context,
	userInfo *biz.UserInfo, lastSendTime int64,
	waID, rallyCode, helpCode string,
	waMsgSends []*dto.WaMsgSend) error {
	f := func(ctx context.Context, tx sqlx.DB) error {
		err := model.UpdateOfficialMsgRecordState(ctx, tx, waID, rallyCode,
			biz.MsgStateDoing, biz.MsgStateComplete)
		if err != nil {
			o.l.WithContext(ctx).
				Errorf("update official msg record state failed, err=%v", err)
			return err
		}

		return nil
	}

	return o.Rally.createGroup(ctx, userInfo, lastSendTime, waID, helpCode, waMsgSends, f)
}

// FindMsg implements biz.OfficialRallyRepo.
func (o *OfficialRally) FindMsg(ctx context.Context, waID, rallyCode string) (
	*biz.OfficialMsgRecord, error) {
	msg, err := model.GetOfficialMsgRecord(ctx, o.data.db, waID, rallyCode)
	if err != nil {
		return nil, err
	}

	return convertOfficialMsgRecord2Biz(msg), nil
}

func (o *OfficialRally) ListDoingMsg(ctx context.Context, minID int, offset, length uint,
	maxTime time.Time) ([]*biz.OfficialMsgRecord, error) {
	msgs, err := model.SelectOfficialMsgRecords(ctx, o.data.db, minID, offset, length, biz.MsgStateDoing, maxTime)
	if err != nil {
		o.l.WithContext(ctx).Errorf("list doing msg failed, err=%v, minID=%d, offset=%d, length=%d, maxTime=%v",
			err, minID, offset, length, maxTime)
		return nil, err
	}

	return lo.Map(
		msgs,
		func(msg *model.OfficialMsgRecord, _ int) *biz.OfficialMsgRecord {
			return convertOfficialMsgRecord2Biz(msg)
		},
	), nil
}
