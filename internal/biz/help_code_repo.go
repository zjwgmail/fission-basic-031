package biz

import (
	"context"
)

type HelpCodeRepo interface {
	CreateEmptyHelpCode(ctx context.Context) (int64, error)
	UpdateHelpCode(ctx context.Context, hcParam *HelpCodeParam) error
	UpdateShortLink(ctx context.Context, hcParam *HelpCodeParam) error
	ListShortLinkByHelpCode(ctx context.Context, hcParam *HelpCodeParam) ([]string, error)
	GetDataById(ctx context.Context, id int64) (string, map[int]string, error)
	DeleteById(ctx context.Context, id int64) error
	GetMaxId(ctx context.Context) (int64, error)
	ListGtId(ctx context.Context, id int64, limit uint) ([]*HelpCode, error)
}

type HelpCodeParam struct {
	Id               int64
	HelpCode         string
	ShortLinkVersion string
	ShortLink        string
}
