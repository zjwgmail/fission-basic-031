package data

import (
	"context"
	"encoding/json"
	v1 "fission-basic/api/fission/v1"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"

	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/kit/sqlx"
	"fission-basic/util"
)

var _ biz.NXCloudRepo = (*NXCloud)(nil)

type NXCloud struct {
	data *Data
	l    *log.Helper
}

func NewNXCloud(d *Data, logger log.Logger) biz.NXCloudRepo {
	return &NXCloud{
		data: d,
		l:    log.NewHelper(logger),
	}
}

// SaveOfficialMsg implements biz.NXCloudRepo.
func (n *NXCloud) SaveOfficialMsg(
	ctx context.Context,
	msg *biz.OfficialMsgRecord,
	msgID string,
	content string,
) error {
	now := time.Now()
	officialMsgRecord := model.OfficialMsgRecord{
		WaID:       msg.WaID,
		RallyCode:  msg.RallyCode,
		State:      biz.MsgStateDoing,
		Channel:    msg.Channel,
		Generation: msg.Generation,
		Language:   msg.Language,
		Nickname:   msg.NickName,
		SendTime:   msg.SendTime,
		CreateTime: now,
		UpdateTime: now,
		Del:        biz.NotDeleted,
	}
	n.l.WithContext(ctx).Infof("insert officialMsgRecord:%+v", officialMsgRecord)

	f := func(ctx context.Context, db sqlx.DB) error {
		// save official msg
		_, err := model.InsertOfficialMsgRecord(ctx, db, &officialMsgRecord)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "officialMsgRecord", officialMsgRecord)
			return err
		}

		return nil
	}

	err := n.saveMsg(ctx, msg.WaID, msgID, content, msg.SendTime, now, msg.Generation, f)
	if err != nil {
		n.l.WithContext(ctx).
			Errorf(`save official msg failed, err=%v, msg=%+v`, err, msg)
		return err
	}

	return nil
}

// SaveUnOfficialMsg implements biz.NXCloudRepo.
func (n *NXCloud) SaveUnOfficialMsg(ctx context.Context,
	unOfficialMsg *biz.UnOfficialMsgRecord,
	msgID, conent string) error {
	now := time.Now()

	unOfficialMsgRecord := model.UnOfficialMsgRecord{
		WaID:       unOfficialMsg.WaID,
		RallyCode:  unOfficialMsg.RallyCode,
		State:      biz.MsgStateDoing,
		Channel:    unOfficialMsg.Channel,
		Language:   unOfficialMsg.Language,
		Generation: unOfficialMsg.Generation,
		Nickname:   unOfficialMsg.NickName,
		SendTime:   unOfficialMsg.SendTime,
		CreateTime: now,
		UpdateTime: now,
		Del:        biz.NotDeleted,
	}

	f := func(ctx context.Context, db sqlx.DB) error {
		// save unofficial msg
		_, err := model.InsertUnOfficialMsgRecord(ctx, db, &unOfficialMsgRecord)
		if err != nil {
			n.l.WithContext(ctx).
				Errorf("save unofficial msg failed, err=%v, unOfficialMsgRecord=%v", err, unOfficialMsgRecord)
			return err
		}
		return nil
	}

	err := n.saveMsg(ctx, unOfficialMsg.WaID, msgID, conent, unOfficialMsg.SendTime, now, unOfficialMsg.Generation, f)
	if err != nil {
		n.l.WithContext(ctx).
			Errorf(`save unofficial msg failed, err=%v, unOfficialMsg=%+v, msgID=%s`, err, unOfficialMsg, msgID)
		return err
	}

	return nil
}

func (n *NXCloud) saveMsg(ctx context.Context,
	waID string, msgID string, content string, sendTime int64,
	now time.Time, generation int,
	extendFunc func(ctx context.Context, db sqlx.DB) error,
) error {
	msgReceived := model.WaMsgReceived{
		WaMsgID:         msgID,
		WaID:            waID,
		Content:         content,
		MsgReceivedTime: sendTime,
		CreateTime:      now,
		UpdateTime:      now,
		Del:             biz.NotDeleted,
	}

	newUserRemind := false
	_, err := model.GetUserRemindByWaID(ctx, n.data.db, waID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			newUserRemind = true
		} else {
			n.l.WithContext(ctx).Errorf("GetUserRemindByWaID failed, err=%v, waID=%s", err, waID)
			return err
		}
	}

	var userRemindDBFunc dbFunc

	if newUserRemind {
		userRemindDBFunc = n.newUserRemindDBFunc(ctx, waID, sendTime, now, generation)
	} else {
		// 续免费
		sendV22, err := util.GetSendRenewMsgTime(ctx, waID, 22, sendTime)
		if err != nil {
			n.l.WithContext(ctx).
				Errorf("GetSendRenewMsgTime failed, err=%v, waID=%s", err, waID)
			// 不返回
			err = nil
		}

		userRemindDBFunc = func(ctx context.Context, db sqlx.DB) error {
			err := model.UpdateUserRemindV22SendTime(ctx, db, waID, sendTime, sendV22)
			if err != nil {
				if errors.Is(err, sqlx.ErrRowsAffected) {
					n.l.WithContext(ctx).Infof("UpdateUserRemindV22SendTime failed, err=%v, waID=%s", err, waID)
					// fixme： 测试的时候时间经常一直，不做报错处理
					return nil
				}
				n.l.WithContext(ctx).Errorw("err", err, "waID", waID, "sendTime", sendTime)
				return err
			}

			return err
		}
	}

	return sqlx.TxContext(ctx, n.data.db, func(ctx context.Context, db sqlx.DB) error {
		err := userRemindDBFunc(ctx, db)
		if err != nil {
			return err
		}

		// save receive msg
		_, err = model.InsertWaMsgReceived(ctx, db, &msgReceived)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "msgReceived", msgReceived)
			return err
		}

		if extendFunc != nil {
			err := extendFunc(ctx, db)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (n *NXCloud) newUserRemindDBFunc(ctx context.Context, waID string, sendTime int64, now time.Time, generation int) dbFunc {
	// 催促成团
	clusterTime, err := util.GetSendClusteringTime(ctx, waID, 3, sendTime)
	if err != nil {
		n.l.WithContext(ctx).
			Errorf("GetSendClusteringTime failed, err=%v, waID=%s", err, waID)
		// 不返回
		err = nil
	}

	// 续免费
	sendV22, err := util.GetSendRenewMsgTime(ctx, waID, 22, sendTime)
	if err != nil {
		n.l.WithContext(ctx).
			Errorf("GetSendRenewMsgTime failed, err=%v, waID=%s", err, waID)
		// 不返回
		err = nil
	}

	unSendStatus := 1
	userRemind := model.UserRemind{
		// ID:           0,
		WaID:         waID,
		LastSendTime: sendTime,
		// SendTimeV0:   freeCDK,
		SendTimeV3:  clusterTime,
		SendTimeV22: sendV22,
		StatusV0:    unSendStatus,
		StatusV22:   unSendStatus,
		StatusV3:    unSendStatus,
		StatusV36:   unSendStatus,
		CreateTime:  now,
		UpdateTime:  now,
		Del:         biz.NotDeleted,
	}

	// 初代v0为空
	if generation == 1 {
		userRemind.StatusV0 = 9
	} else {
		// 免费CDK时间
		freeCDK, err := util.GetSendClusteringTime(ctx, waID, 8, sendTime)
		if err != nil {
			n.l.WithContext(ctx).
				Errorf("GetSendClusteringTime failed, err=%v, waID=%s", err, waID)
			// 不返回
			err = nil
		}

		userRemind.SendTimeV0 = freeCDK
		userRemind.StatusV0 = unSendStatus
	}

	return func(ctx context.Context, db sqlx.DB) error {
		_, err := model.InsertUserRemind(ctx, db, &userRemind)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "userRemind", userRemind)
			return err
		}

		return nil
	}
}

func (n *NXCloud) SaveRenewMsg(ctx context.Context, waID string, msgID, content string, sendTime int64) error {

	// 续免费
	sendV22, err := util.GetSendRenewMsgTime(ctx, waID, 22, sendTime)
	if err != nil {
		n.l.WithContext(ctx).
			Errorf("GetSendRenewMsgTime failed, err=%v, waID=%s", err, waID)
		// 不返回
		err = nil
	}

	now := time.Now()
	msgReceived := model.WaMsgReceived{
		WaMsgID:         msgID,
		WaID:            waID,
		Content:         content,
		MsgReceivedTime: sendTime,
		CreateTime:      now,
		UpdateTime:      now,
		Del:             biz.NotDeleted,
	}

	return sqlx.TxContext(ctx, n.data.db, func(ctx context.Context, db sqlx.DB) error {

		err := model.UpdateUserRemindV22SendTime(ctx, db, waID, sendTime, sendV22)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "waID", waID, "sendTime", sendTime)
			return err
		}

		// save receive msg
		_, err = model.InsertWaMsgReceived(ctx, db, &msgReceived)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "msgReceived", msgReceived)
			return err
		}

		return nil
	})
}

func (n *NXCloud) SaveReceiveMsg(ctx context.Context, waID, msgID, content string, sendTime int64, generation int) error {
	now := time.Now()
	msgReceived := model.WaMsgReceived{
		WaMsgID:         msgID,
		WaID:            waID,
		Content:         content,
		MsgReceivedTime: sendTime,
		CreateTime:      now,
		UpdateTime:      now,
		Del:             biz.NotDeleted,
	}

	newUserRemind := false
	_, err := model.GetUserRemindByWaID(ctx, n.data.db, waID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			newUserRemind = true
		} else {
			n.l.WithContext(ctx).Errorf("GetUserRemindByWaID failed, err=%v, waID=%s", err, waID)
			return err
		}
	}

	var userRemindDBFunc dbFunc

	if !newUserRemind {
		// 续免费
		sendV22, err := util.GetSendRenewMsgTime(ctx, waID, 22, sendTime)
		if err != nil {
			n.l.WithContext(ctx).
				Errorf("GetSendRenewMsgTime failed, err=%v, waID=%s", err, waID)
			// 不返回
			err = nil
		}

		userRemindDBFunc = func(ctx context.Context, db sqlx.DB) error {
			err := model.UpdateUserRemindV22SendTime(ctx, db, waID, sendTime, sendV22)
			if err != nil {
				if errors.Is(err, sqlx.ErrRowsAffected) {
					n.l.WithContext(ctx).Infof("UpdateUserRemindV22SendTime failed, err=%v, waID=%s", err, waID)
					// fixme： 测试的时候时间经常一直，不做报错处理
					return nil
				}
				n.l.WithContext(ctx).Errorw("err", err, "waID", waID, "sendTime", sendTime)
				return err
			}

			return err
		}
	}

	return sqlx.TxContext(ctx, n.data.db, func(ctx context.Context, db sqlx.DB) error {
		if !newUserRemind {
			err := userRemindDBFunc(ctx, db)
			if err != nil {
				return err
			}
		}

		// save receive msg
		_, err := model.InsertWaMsgReceived(ctx, db, &msgReceived)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "msgReceived", msgReceived)
			return err
		}

		return nil
	})
	// return n.saveMsg(ctx, waID, msgID, content, sendTime, time.Now(), generation, nil)
}

func (n *NXCloud) SaveReceiptMsg(ctx context.Context, waID, msgID, content string, msgState int, sendTime int64, cost []*v1.Cost) error {
	t := time.Now()

	msgReceived := model.WaMsgReceived{
		WaMsgID:         msgID,
		WaID:            waID,
		Content:         content,
		MsgReceivedTime: sendTime,
		CreateTime:      t,
		UpdateTime:      t,
		Del:             biz.NotDeleted,
	}

	receiptMsgRecord := model.ReceiptMsgRecord{
		MsgID:      msgID,
		MsgState:   msgState,
		State:      biz.MsgStateDoing,
		CreateTime: t,
		UpdateTime: t,
		Del:        biz.NotDeleted,
		WaId:       waID,
	}

	// 格式化时间为 "20060102" 格式
	formattedDate := t.Format("20060102")
	receiptMsgRecord.Pt = formattedDate

	if cost != nil && len(cost) >= 0 {
		marshal, err := json.Marshal(cost)
		if err != nil {
			n.l.Errorf("queueDTO convert to json failed, err=%v, cost=%v", err, cost)
			return err
		}
		receiptMsgRecord.CostInfo = string(marshal)
	}

	return sqlx.TxContext(ctx, n.data.db, func(ctx context.Context, db sqlx.DB) error {

		// save receive msg
		_, err := model.InsertReceiptMsgRecord(ctx, db, &receiptMsgRecord)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "receiptMsgRecord", receiptMsgRecord)
			return err
		}

		// save receive msg
		_, err = model.InsertWaMsgReceived(ctx, db, &msgReceived)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "msgReceived", msgReceived)
			return err
		}

		return nil
	})
}

func (n *NXCloud) OnlySaveReceiveMsg(ctx context.Context, waID, msgID, content string, sendTime int64) error {
	t := time.Now()

	msgReceived := model.WaMsgReceived{
		WaMsgID:         msgID,
		WaID:            waID,
		Content:         content,
		MsgReceivedTime: sendTime,
		CreateTime:      t,
		UpdateTime:      t,
		Del:             biz.NotDeleted,
	}

	return sqlx.TxContext(ctx, n.data.db, func(ctx context.Context, db sqlx.DB) error {

		// save receive msg
		_, err := model.InsertWaMsgReceived(ctx, db, &msgReceived)
		if err != nil {
			n.l.WithContext(ctx).Errorw("err", err, "msgReceived", msgReceived)
			return err
		}

		return nil
	})
}

// 更新官方助力消息语言
func (n *NXCloud) UpdateOfficialMsgLaguage(ctx context.Context,
	msg *biz.OfficialMsgRecord,
	msgID, content string,
) error {
	now := time.Now()

	f := func(ctx context.Context, db sqlx.DB) error {
		// save official msg
		err := model.UpdateOfficialMsgRecordLanguageAndState(ctx, db, msg.WaID, msg.RallyCode, msg.Language, biz.MsgStateDoing)
		if err != nil {
			n.l.WithContext(ctx).Errorf("UpdateOfficialMsgRecordLanguageAndState failed, err=%v, msg=%+v",
				err, msg)
			return err
		}

		return nil
	}

	err := n.saveMsg(ctx, msg.WaID, msgID, content, msg.SendTime, now, msg.Generation, f)
	if err != nil {
		n.l.WithContext(ctx).
			Errorf(`save official msg failed, err=%v, msg=%+v`, err, msg)
		return err
	}

	return nil
}
