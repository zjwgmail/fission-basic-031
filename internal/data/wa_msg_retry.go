package data

import (
	"context"
	"errors"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pojo/dto"
	"fission-basic/util"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.WaMsgRetryRepo = (*WaMsgRetry)(nil)

type WaMsgRetry struct {
	data *Data
	l    *log.Helper
}

func NewWaMsgRetry(d *Data, logger log.Logger) biz.WaMsgRetryRepo {
	return &WaMsgRetry{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (s *WaMsgRetry) ListRetryWaIdByState(ctx context.Context, minWaId string, limit uint, state []int) ([]string, error) {
	waIdList, err := model.ListWaIdOfRetryByState(ctx, s.data.db, minWaId, limit, state)
	if err != nil {
		return nil, err
	}
	return waIdList, nil
}

func (s *WaMsgRetry) ListMsgRetryByWaIdAndState(ctx context.Context, state []int, waId string) ([]*dto.WaMsgRetryDto, error) {
	msgSendList, err := model.ListMsgRetryByWaIdAndState(ctx, s.data.db, state, waId)
	if err != nil {
		return nil, err
	}
	var bizMsgRetryList []*dto.WaMsgRetryDto
	for _, msgSend := range msgSendList {
		send := &dto.WaMsgRetryDto{}
		util.CopyFieldsByJson(*msgSend, send)
		bizMsgRetryList = append(bizMsgRetryList, send)
	}
	return bizMsgRetryList, nil
}

func (s *WaMsgRetry) UpdateWaRetryMsg(ctx context.Context, stu *dto.WaMsgRetryDto) error {
	if stu.ID <= 0 {
		s.l.WithContext(ctx).Error(fmt.Sprintf("msg's id is null, stu:%+v", stu))
		return errors.New(fmt.Sprintf("msg's id is null,stu:%v", stu))
	}

	req := &model.WaMsgRetry{}
	err := util.CopyFieldsByJson(*stu, req)
	if err != nil {
		s.l.WithContext(ctx).Error(fmt.Sprintf("copyFieldsByJson error: %v", err))
		return err
	}

	err = model.UpdateWaMsgRetry(ctx, s.data.db, req)
	return err
}
