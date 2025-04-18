package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/pkg/nxcloud"
	"fission-basic/internal/pkg/queue"
	"fission-basic/internal/pkg/redis"
	"time"

	"fission-basic/internal/conf"
	"fission-basic/kit/sqlx"

	"github.com/go-kratos/kratos/v2/log"
)

type MsgUsecase struct {
	d                        *conf.Data
	l                        *log.Helper
	repo                     MsgRepo
	waUserScoreRepo          WaUserScoreRepo
	pushEventSendMessageRepo PushEventSendMessageRepo
	redisService             *redis.RedisService
	q                        *queue.CallMsg
}

func NewMsgUsecase(
	d *conf.Data,
	l log.Logger,
	repo MsgRepo,
	waUserScoreRepo WaUserScoreRepo,
	pushEventSendMessageRepo PushEventSendMessageRepo,
	redisService *redis.RedisService,
	q *queue.CallMsg,
) *MsgUsecase {
	return &MsgUsecase{
		d:                        d,
		l:                        log.NewHelper(l),
		repo:                     repo,
		pushEventSendMessageRepo: pushEventSendMessageRepo,
		waUserScoreRepo:          waUserScoreRepo,
		redisService:             redisService,
		q:                        q,
	}
}

// CallbackHandle 回执消息处理
func (msg *MsgUsecase) RecallHandle(ctx context.Context, msgInfo *nxcloud.ReceiptMsgQueueDTO) error {
	msgReceipt, err := msg.repo.FindMsgReceipt(ctx, msgInfo.MsgID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			msg.l.WithContext(ctx).Warnf("find msg receipt failed, err=%v, waId:%v,msgID=%s", err, msgInfo.WaID, msgInfo.MsgID)
			return nil
		}
		msg.l.WithContext(ctx).Errorf("find msg receipt failed, err=%v, waId:%v,msgID=%s", err, msgInfo.WaID, msgInfo.MsgID)
		return err
	}

	if msgReceipt.State != MsgStateDoing {
		msg.l.WithContext(ctx).Infof("msg state is done, waId:%v,msgID=%s", msgInfo.WaID, msgInfo.MsgID)
		return nil
	}

	table := "retry"
	waMsgSend := &WaMsgSend{}
	_, err = msg.repo.FindWaMsgRetry(ctx, msgInfo.MsgID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			msg.l.WithContext(ctx).Errorf("FindWaMsgRetry failed, err=%v, waId:%v,msgID=%s", err, msgInfo.WaID, msgInfo.MsgID)
			return err
		} else {
			waMsgSend, err = msg.repo.FindWaMsgSend(ctx, msgInfo.MsgID)
			if err != nil {
				if !errors.Is(err, sqlx.ErrNoRows) {
					msg.l.WithContext(ctx).Errorf("FindWaMsgSend failed, err=%v, waId:%v,msgID=%s", err, msgInfo.WaID, msgInfo.MsgID)
					return nil
				} else {
					msg.l.WithContext(ctx).Errorf("waMsgId is not in the database and is not processed, err=%v, waId:%v,msgID=%s", err, msgInfo.WaID, msgInfo.MsgID)
					return nil
				}
			} else {
				table = "send"
			}
		}
	}

	// 成功、失败
	if msgReceipt.MsgState == constants.MsgSendStateNxSuccess {
		completeFunc := msg.repo.CompleteReceiptMsg
		if table == "retry" {
			completeFunc = msg.repo.CompleteReceiptAndDeleteRetry
		}

		err = completeFunc(ctx, msgInfo.MsgID)
		if err != nil {
			msg.l.WithContext(ctx).Errorf("complete msgRecall failed, err=%v, waId:%v,msgID=%s", err, msgInfo.WaID, msgInfo.MsgID)
			return err
		}

		var currency string
		var price float64
		var foreignPrice float64
		for _, cost := range msgInfo.Costs {
			currency = cost.Currency
			price = price + cost.Price
			foreignPrice = foreignPrice + cost.ForeignPrice
		}
		msg.l.WithContext(ctx).Infof("msg costInfo. currency=%v,price=%v,foreignPrice=%v", currency, price, foreignPrice)
		if foreignPrice > 0 {
			// 计费 wanjiaju
			err = msg.pushEventSendMessageRepo.UpdateCostByMsgId(ctx, msgInfo.MsgID, int(foreignPrice*10000))
			if err != nil {
				msg.l.WithContext(ctx).Errorf("UpdateCostByMsgId failed, err=%v, waId:%v,msgID=%s", err, waMsgSend.WaID, msgInfo.MsgID)
			}
		}
		return nil
	}

	err = msg.repo.CompleteReceiptAndAddMsgRetry(ctx, msgReceipt.MsgState, msgInfo.MsgID, waMsgSend.WaID, waMsgSend.Content, waMsgSend.MsgType, waMsgSend.BuildMsgParam, table)
	if err != nil {
		msg.l.WithContext(ctx).Errorf("CompleteReceiptAndAddMsgRetry failed, err=%v,waId:%v msgID=%s", err, waMsgSend.WaID, msgInfo.MsgID)
		return err
	}

	return nil
}

func (msg *MsgUsecase) RetryReceiptMsgRecord(ctx context.Context) error {
	msg.l.WithContext(ctx).Info("retry receipt msg record start")
	defer msg.l.WithContext(ctx).Info("retry receipt msg record end")

	var (
		minID       = 0
		maxSendTime = time.Now().Add(-2 * time.Minute) // 暂定2分钟
		offset      = uint(0)
		length      = uint(100)
		batchSize   = 20
		qDatas      = make([]string, 0, batchSize)
		sendQ       = func() {
			err := msg.q.SendBack(qDatas, true)
			if err != nil {
				// 忽略这个错误
				msg.l.WithContext(ctx).Errorf("SendBack failed, err=%v, msg=%+v", err, qDatas)
			}
			qDatas = qDatas[:0]
		}
	)

	for {
		msgs, err := msg.repo.ListDoingReceiptMsgRecords(ctx, minID, offset, length, maxSendTime)
		if err != nil {
			msg.l.WithContext(ctx).Errorf("list doing msg failed, err=%v, minID=%d, offset=%d, length=%d, maxTime=%v",
				err, minID, offset, length, maxSendTime)
			return err
		}

		for i := range msgs {
			info := msgs[i]
			var Costs []*v1.Cost
			if info.CostInfo != "" {
				err := json.Unmarshal([]byte(info.CostInfo), &Costs)
				if err != nil {
					msg.l.WithContext(ctx).Errorf("json unmarshal failed, err=%v, info=%v", err, info)
					continue
				}
			}

			// 发送队列的消息
			queueDTO := &nxcloud.ReceiptMsgQueueDTO{
				WaID:     info.WaID,
				MsgID:    info.MsgID,
				MsgType:  nxcloud.MsgTypeCallback,
				MsgState: info.MsgState,
				Costs:    Costs,
			}

			marshal, err := json.Marshal(queueDTO)
			if err != nil {
				msg.l.Errorf("queueDTO convert to json failed, err=%v, info=%v", err, info)
				return err
			}

			qDatas = append(qDatas, string(marshal))
			if (i+1)%batchSize == 0 {
				sendQ()
			}
		}

		if len(qDatas) > 0 {
			sendQ()
		}

		if len(msgs) < int(length) {
			break
		}

		minID = msgs[len(msgs)-1].ID
	}

	return nil
}
