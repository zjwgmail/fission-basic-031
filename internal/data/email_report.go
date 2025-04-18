package data

import (
	"context"
	"errors"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/kit/sqlx"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.EmailReportRepo = (*EmailReport)(nil)

type EmailReport struct {
	data *Data
	l    *log.Helper
}

func NewEmailReport(d *Data, logger log.Logger) biz.EmailReportRepo {
	return &EmailReport{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (er *EmailReport) AddBatchEmailReport(ctx context.Context, list []*biz.EmailReportDTO, utc int) (int, error) {
	for _, dto := range list {
		entity := ConvertEmailReport2Entity(dto, utc)
		report, err := model.EmailReportSelect(ctx, er.data.db, entity)
		if err != nil && !errors.Is(err, sqlx.ErrNoRows) {
			return 0, err
		}
		if report.Id == 0 {
			_, err := model.EmailReportInsert(ctx, er.data.db, entity)
			if err != nil {
				return 0, err
			}
		}
		if err := model.EmailReportUpdate(ctx, er.data.db, entity); err != nil {
			if errors.Is(err, sqlx.ErrRowsAffected) {
				continue
			}
			return 0, err
		}
	}
	return len(list), nil
}

func (er *EmailReport) ListAllEmailReport(ctx context.Context, utc int) ([]*biz.EmailReportDTO, error) {
	list, err := model.EmailReportList(ctx, er.data.db, utc)
	if err != nil {
		return nil, err
	}
	result := make([]*biz.EmailReportDTO, 0)
	for _, entity := range list {
		dto := ConvertEmailReport2DTO(entity)
		result = append(result, dto)
	}
	return result, nil
}
