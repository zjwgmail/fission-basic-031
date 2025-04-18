package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tablePushEventSendMessage = "push_event_send_message"

type PushEventSendMessageEntity struct {
	ID        int    `db:"id"`
	MessageId string `db:"message_id"`
	Cost      int    `db:"cost"`
	Version   int    `db:"version"`
}

func PushEventSendMessageInsertIgnore(ctx context.Context, db sqlx.DB, messageId string, cost int, version int) (int64, error) {
	entity := &PushEventSendMessageEntity{
		MessageId: messageId,
		Cost:      cost,
		Version:   version,
	}
	return sqlx.InsertIgnoreContext(ctx, db, tablePushEventSendMessage, entity)
}

func PushEventSendMessageUpdateCostByMessageId(ctx context.Context, db sqlx.DB, messageId string, cost int) error {
	where := map[string]interface{}{
		"message_id": messageId,
	}
	update := map[string]interface{}{
		"cost": cost,
	}
	return sqlx.UpdateContext(ctx, db, tablePushEventSendMessage, where, update)
}

func PushEventSendMessageGetByMessageId(ctx context.Context, db sqlx.DB, messageId string) (*PushEventSendMessageEntity, error) {
	where := map[string]interface{}{
		"message_id": messageId,
	}
	entity := &PushEventSendMessageEntity{}
	err := sqlx.GetContext(ctx, db, entity, tablePushEventSendMessage, where)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
