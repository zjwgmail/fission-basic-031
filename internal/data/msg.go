package data

import (
	"context"
	"fission-basic/api/constants"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/kit/sqlx"
	"fission-basic/util"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"
)

var _ biz.MsgRepo = (*Msg)(nil)

type Msg struct {
	data *Data
	l    *log.Helper
}

func NewMsg(d *Data, logger log.Logger) biz.MsgRepo {
	return &Msg{
		data: d,
		l:    log.NewHelper(logger),
	}
}

// CompleteReceiptAndAddMsgRetry implements biz.MsgRepo.
func (m *Msg) CompleteReceiptAndAddMsgRetry(ctx context.Context, msgState int, msgID string, waID string, content string, msgType, buildMsgParam, table string) error {

	msgRetry := &model.WaMsgRetry{
		WaID:          waID,
		MsgType:       msgType,
		State:         msgState,
		Content:       content,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		BuildMsgParam: buildMsgParam,
	}
	// 修改回执表的状态
	return sqlx.TxContext(ctx, m.data.db, func(ctx context.Context, db sqlx.DB) error {
		if table == "retry" {
			err := model.UpdateWaMsgRetryState(ctx, db, msgID, msgState)
			if err != nil {
				m.l.Error(fmt.Sprintf("UpdateWaMsgRetryState msgID:%v,error: %v", msgID, err))
				return err
			}
		} else {
			err := model.UpdateWaMsgSendState(ctx, db, msgID, msgState)
			if err != nil {
				m.l.Error(fmt.Sprintf("UpdateWaMsgSendState msgID:%v,error: %v", msgID, err))
				return err
			}

			_, err = model.InsertWaMsgRetry(ctx, db, msgRetry)
			if err != nil {
				m.l.Error(fmt.Sprintf("InsertWaMsgRetry msgID:%v,error: %v", msgID, err))
				return err
			}
		}

		err := model.UpdateReceiptMsgRecordState(ctx, db, msgID, biz.MsgStateDoing, biz.MsgStateComplete)
		if err != nil {
			m.l.Error(fmt.Sprintf("UpdateReceiptMsgRecordState msgID:%v,error: %v", msgID, err))
			return err
		}
		return nil
	})
}

// CompleteReceiptAndDeleteRetry implements biz.MsgRepo.
func (m *Msg) CompleteReceiptAndDeleteRetry(ctx context.Context, msgID string) error {

	return sqlx.TxContext(ctx, m.data.db, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateWaMsgRetryState(ctx, db, msgID, constants.MsgSendStateNxSuccess)
		if err != nil {
			m.l.Error(fmt.Sprintf("UpdateWaMsgRetryState msgID:%v,error: %v", msgID, err))
			return err
		}
		err = model.UpdateReceiptMsgRecordState(ctx, db, msgID, biz.MsgStateDoing, biz.MsgStateComplete)
		if err != nil {
			m.l.Error(fmt.Sprintf("UpdateReceiptMsgRecordState msgID:%v,error: %v", msgID, err))
			return err
		}
		return nil
	})
}

// CompleteReceiptMsg implements biz.MsgRepo.
func (m *Msg) CompleteReceiptMsg(ctx context.Context, msgID string) error {
	return sqlx.TxContext(ctx, m.data.db, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateWaMsgSendState(ctx, db, msgID, constants.MsgSendStateNxSuccess)
		if err != nil {
			m.l.Error(fmt.Sprintf("UpdateWaMsgSendState msgID:%v,error: %v", msgID, err))
			return err
		}
		err = model.UpdateReceiptMsgRecordState(ctx, db, msgID, biz.MsgStateDoing, biz.MsgStateComplete)
		if err != nil {
			m.l.Error(fmt.Sprintf("UpdateReceiptMsgRecordState msgID:%v,error: %v", msgID, err))
			return err
		}
		return nil
	})
}

// FindMsgReceipt implements biz.MsgRepo.
func (m *Msg) FindMsgReceipt(ctx context.Context, msgID string) (*biz.ReceiptMsgRecord, error) {
	record, err := model.GetReceiptMsgRecord(ctx, m.data.db, msgID)
	if err != nil {
		return nil, err
	}
	res := &biz.ReceiptMsgRecord{}
	err = util.CopyFieldsByJson(*record, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Msg) ListDoingReceiptMsgRecords(ctx context.Context,
	minID int, offset, length uint, maxTime time.Time) ([]*biz.ReceiptMsgRecord, error) {
	records, err := model.SelectReceiptMsgRecords(ctx, m.data.db, biz.MsgStateDoing, minID, offset, length, maxTime)
	if err != nil {
		m.l.WithContext(ctx).Errorf("list doing msg failed, err=%v, minID=%d, offset=%d, length=%d, maxTime=%v",
			err, minID, offset, length, maxTime)
		return nil, err
	}

	return lo.Map(
		records,
		func(msg *model.ReceiptMsgRecord, _ int) *biz.ReceiptMsgRecord {
			return convertReceiptMsgRecord2Biz(msg)
		},
	), nil
}

// FindWaMsgRetry implements biz.MsgRepo.
func (m *Msg) FindWaMsgRetry(ctx context.Context, msgID string) (*biz.WaMsgRetry, error) {
	record, err := model.GetWaMsgRetry(ctx, m.data.db, msgID)
	if err != nil {
		return nil, err
	}
	res := &biz.WaMsgRetry{}
	err = util.CopyFieldsByJson(*record, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// FindWaMsgSend implements biz.MsgRepo.
func (m *Msg) FindWaMsgSend(ctx context.Context, msgID string) (*biz.WaMsgSend, error) {
	record, err := model.GetWaMsgSend(ctx, m.data.db, msgID)
	if err != nil {
		return nil, err
	}
	res := &biz.WaMsgSend{}
	err = util.CopyFieldsByJson(*record, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
