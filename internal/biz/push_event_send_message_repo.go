package biz

import "context"

type PushEventSendMessageRepo interface {
	InsertIgnore(ctx context.Context, dto *PushEventSendMessageDTO) error
	UpdateCostByMsgId(ctx context.Context, msgId string, cost int) error
	GetByMsgId(ctx context.Context, msgId string) (*PushEventSendMessageDTO, error)
}

type PushEventSendMessageDTO struct {
	MessageId string
	Cost      int
	Version   int
}
