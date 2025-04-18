package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.WaMsgReceivedRepo = (*WaMsgReceived)(nil)

type WaMsgReceived struct {
	data *Data
	l    *log.Helper
}

func NewWaMsgReceived(d *Data, logger log.Logger) biz.WaMsgReceivedRepo {
	return &WaMsgReceived{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (w WaMsgReceived) ListGtIdReceivedTime(ctx context.Context, suffix string, startTimeStamp int64, endTimeStamp int64, minId int, limit uint) ([]*biz.WaMsgReceivedDTO, error) {
	entityList, err := model.WaMsgReceivedListGtIdReceivedTime(ctx, w.data.db, suffix, startTimeStamp, endTimeStamp, minId, limit)
	if err != nil {
		return nil, err
	}
	return waMsgReceivedList2DTO(entityList), nil
}
