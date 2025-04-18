package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.WaUserScoreRepo = (*WaUserScore)(nil)

type WaUserScore struct {
	data *Data
	l    *log.Helper
}

func NewWaUserScore(d *Data, logger log.Logger) biz.WaUserScoreRepo {
	return &WaUserScore{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (w *WaUserScore) UpdateState(ctx context.Context, waId string, state int) error {
	return model.WaUserScoreUpdateState(ctx, w.data.db, waId, state)
}

func (w *WaUserScore) PageBySocialScore(ctx context.Context, limit uint, length uint) ([]*biz.WaUserScoreDTO, error) {
	list, err := model.WaUserScorePageBySocialScore(ctx, w.data.db, limit, length)
	if err != nil {
		return nil, err
	}
	return ConvertWaUserScore2BizList(list), nil
}

func (w *WaUserScore) PageByRecurringProb(ctx context.Context, limit uint, length uint) ([]*biz.WaUserScoreDTO, error) {
	list, err := model.WaUserScorePageByRecurringProb(ctx, w.data.db, limit, length)
	if err != nil {
		return nil, err
	}
	return ConvertWaUserScore2BizList(list), nil
}
