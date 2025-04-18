package biz

import (
	"context"
	"fission-basic/internal/pojo/dto"
)

type WaMsgRetryRepo interface {
	UpdateWaRetryMsg(ctx context.Context, stu *dto.WaMsgRetryDto) error
	ListRetryWaIdByState(ctx context.Context, minWaId string, limit uint, state []int) ([]string, error)
	ListMsgRetryByWaIdAndState(ctx context.Context, state []int, waId string) ([]*dto.WaMsgRetryDto, error)
}
