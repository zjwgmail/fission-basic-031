package biz

import (
	"context"
	"fission-basic/internal/pojo/dto"
)

type WaMsgSendRepo interface {
	AddWaMsgSend(ctx context.Context, stu *dto.WaMsgSend) (int64, error)
	UpdateWaMsg(ctx context.Context, stu *dto.WaMsgSend) error
	UpdateWaMsgSendStateByWaMsgId(ctx context.Context, waMsgID string, state int) error
	ListGtId(ctx context.Context, id int64, limit int) ([]*WaMsgSend, error)
	ListGtIdInPts(ctx context.Context, id int64, pts []string, limit int) ([]*WaMsgSend, error)

	ListWaIdByState(ctx context.Context, minWaId string, limit uint, state []int) ([]string, error)
	ListMsgSendByWaIdAndState(ctx context.Context, state []int, waId string, ptList []string) ([]*WaMsgSend, error)
}
