package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.FeishuReportRepo = (*FeishuReport)(nil)

type FeishuReport struct {
	data *Data
	l    *log.Helper
}

func NewFeishuReport(d *Data, logger log.Logger) biz.FeishuReportRepo {
	return &FeishuReport{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (fr *FeishuReport) AddFeishuReport(ctx context.Context, frParam *biz.FeishuReportParam) (int64, error) {
	frEntity := ConvertFeishuReportParam2Entity(frParam)
	return model.InsertFeishuReport(ctx, fr.data.db, frEntity)
}
