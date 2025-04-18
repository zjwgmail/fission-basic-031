package biz

import (
	"context"
	"fission-basic/internal/util"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type HelpCodeUsecase struct {
	repo HelpCodeRepo
	l    *log.Helper
}

func NewHelpCodeUsecase(repo HelpCodeRepo, l log.Logger) *HelpCodeUsecase {
	return &HelpCodeUsecase{
		repo: repo,
		l:    log.NewHelper(l),
	}
}

func (hc *HelpCodeUsecase) Add(ctx context.Context) (string, int64, error) {
	id, err := hc.repo.CreateEmptyHelpCode(ctx)
	if err != nil {
		return "", 0, err
	}
	hcParam := HelpCodeParam{
		Id:       id,
		HelpCode: util.ToBase32(id),
	}
	err = hc.repo.UpdateHelpCode(ctx, &hcParam)
	if err != nil {
		return "", id, err
	}

	return hcParam.HelpCode, id, nil
}

func (hc *HelpCodeUsecase) UpdateShortLinkByHelpCode(ctx context.Context, version int, shortLink string, helpCode string) error {
	hcParam := HelpCodeParam{
		ShortLinkVersion: "short_link_v" + strconv.Itoa(version),
		ShortLink:        shortLink,
		HelpCode:         helpCode,
	}
	err := hc.repo.UpdateShortLink(ctx, &hcParam)
	if err != nil {
		return err
	}
	return nil
}

func (hc *HelpCodeUsecase) GetShortLinkByHelpCodeAndVersion(ctx context.Context, helpCode string, shortLinkVersion int) (string, error) {
	hcParam := HelpCodeParam{
		HelpCode: helpCode,
	}
	shortLinks, err := hc.repo.ListShortLinkByHelpCode(ctx, &hcParam)
	if err != nil {
		return "", err
	}
	return shortLinks[shortLinkVersion], err
}

func (hc *HelpCodeUsecase) GetDataById(ctx context.Context, id int64) (string, map[int]string, error) {
	return hc.repo.GetDataById(ctx, id)
}

func (hc *HelpCodeUsecase) DeleteById(ctx context.Context, id int64) error {
	return hc.repo.DeleteById(ctx, id)
}

func (hc *HelpCodeUsecase) GetMaxId(ctx context.Context) (int64, error) {
	return hc.repo.GetMaxId(ctx)
}

func (hc *HelpCodeUsecase) ListGtId(ctx context.Context, id int64, limit uint) ([]*HelpCode, error) {
	return hc.repo.ListGtId(ctx, id, limit)
}

type HelpCode struct {
	Id          int64
	Del         string
	CreateTime  time.Time
	UpdateTime  time.Time
	HelpCode    string
	ShortLinkV0 string
	ShortLinkV1 string
	ShortLinkV2 string
	ShortLinkV3 string
	ShortLinkV4 string
	ShortLinkV5 string
}
