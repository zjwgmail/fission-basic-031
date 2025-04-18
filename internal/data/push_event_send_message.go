package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.PushEventSendMessageRepo = (*PushEventSendMessage)(nil)

type PushEventSendMessage struct {
	data *Data
	l    *log.Helper
}

func NewPushEventSendMessage(d *Data, logger log.Logger) biz.PushEventSendMessageRepo {
	return &PushEventSendMessage{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (p *PushEventSendMessage) InsertIgnore(ctx context.Context, dto *biz.PushEventSendMessageDTO) error {
	_, err := model.PushEventSendMessageInsertIgnore(ctx, p.data.db, dto.MessageId, dto.Cost, dto.Version)
	return err
}

func (p *PushEventSendMessage) UpdateCostByMsgId(ctx context.Context, msgId string, cost int) error {
	return model.PushEventSendMessageUpdateCostByMessageId(ctx, p.data.db, msgId, cost)
}

func (p *PushEventSendMessage) GetByMsgId(ctx context.Context, msgId string) (*biz.PushEventSendMessageDTO, error) {
	entity, err := model.PushEventSendMessageGetByMessageId(ctx, p.data.db, msgId)
	if err != nil {
		return nil, err
	}
	return &biz.PushEventSendMessageDTO{
		Cost:      entity.Cost,
		MessageId: entity.MessageId,
		Version:   entity.Version,
	}, nil
}
