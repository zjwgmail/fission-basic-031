package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

var _ biz.HelpCodeRepo = (*HelpCode)(nil)

type HelpCode struct {
	data *Data
	l    *log.Helper
}

func NewHelpCode(d *Data, logger log.Logger) biz.HelpCodeRepo {
	return &HelpCode{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (hc *HelpCode) CreateEmptyHelpCode(ctx context.Context) (int64, error) {
	hcEntity := model.HelpCodeEntity{
		CreateTime: time.Now(),
	}

	id, err := model.HelpCodeInsert(ctx, hc.data.db, &hcEntity)

	return id, err
}

func (hc *HelpCode) UpdateHelpCode(ctx context.Context, hcParam *biz.HelpCodeParam) error {
	id := hcParam.Id

	hcEntity := model.HelpCodeEntity{
		Id:       id,
		HelpCode: hcParam.HelpCode,
	}
	return model.HelpCodeUpdateById(ctx, hc.data.db, &hcEntity)
}

func (hc *HelpCode) UpdateShortLink(ctx context.Context, hcParam *biz.HelpCodeParam) error {
	return model.UpdateShortLinkByHelpCode(ctx, hc.data.db, hcParam)
}

func (hc *HelpCode) ListShortLinkByHelpCode(ctx context.Context, hcParam *biz.HelpCodeParam) ([]string, error) {
	return model.ListShortLinkByHelpCode(ctx, hc.data.db, hcParam)
}

func (hc *HelpCode) GetDataById(ctx context.Context, id int64) (string, map[int]string, error) {
	return model.HelpCodeGetById(ctx, hc.data.db, id)
}

func (hc *HelpCode) DeleteById(ctx context.Context, id int64) error {
	return model.HelpCodeDeleteById(ctx, hc.data.db, id)
}

func (hc *HelpCode) GetMaxId(ctx context.Context) (int64, error) {
	return model.HelpCodeGetMaxId(ctx, hc.data.db)
}

func (hc *HelpCode) ListGtId(ctx context.Context, id int64, limit uint) ([]*biz.HelpCode, error) {
	modes, err := model.HelpCodeListGtId(ctx, hc.data.db, id, limit)
	if err != nil {
		return nil, err
	}
	var bizModes []*biz.HelpCode
	for _, mode := range modes {
		code2Biz := ConvertHelpCode2Biz(mode)
		bizModes = append(bizModes, code2Biz)
	}
	return bizModes, nil
}
