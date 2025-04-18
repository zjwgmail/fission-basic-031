package biz

import "context"

type WaMsgReceivedRepo interface {
	ListGtIdReceivedTime(ctx context.Context, suffix string, startTimeStamp int64, endTimeStamp int64, minId int, limit uint) ([]*WaMsgReceivedDTO, error)
}

type WaMsgReceivedDTO struct {
	Id              int
	WaId            string
	MsgReceivedTime int64
}
