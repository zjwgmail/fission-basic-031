package data

import (
	"context"
	"errors"
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto"
	"fission-basic/util"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.WaMsgSendRepo = (*WaMsgSend)(nil)

type WaMsgSend struct {
	data         *Data
	l            *log.Helper
	redisService *redis.RedisService
	business     *conf.Business
}

func NewWaMsgSend(d *Data, redisService *redis.RedisService, business *conf.Business, logger log.Logger) biz.WaMsgSendRepo {
	return &WaMsgSend{
		data:         d,
		redisService: redisService,
		business:     business,
		l:            log.NewHelper(logger),
	}
}

func (s *WaMsgSend) AddWaMsgSend(ctx context.Context, stu *dto.WaMsgSend) (int64, error) {

	req := &model.WaMsgSend{}
	err := util.CopyFieldsByJson(*stu, req)
	if err != nil {
		s.l.WithContext(ctx).Error(fmt.Sprintf("copyFieldsByJson error: %v", err))
		return 0, err
	}

	id, err := model.InsertWaMsgSend(ctx, s.data.db, req)

	return id, err
}

func (s *WaMsgSend) UpdateWaMsg(ctx context.Context, stu *dto.WaMsgSend) error {
	if stu.ID <= 0 {
		s.l.WithContext(ctx).Error(fmt.Sprintf("msg's id is null, stu:%+v", stu))
		return errors.New(fmt.Sprintf("msg's id is null,stu:%v", stu))
	}

	req := &model.WaMsgSend{}
	err := util.CopyFieldsByJson(*stu, req)
	if err != nil {
		s.l.WithContext(ctx).Error(fmt.Sprintf("copyFieldsByJson error: %v", err))
		return err
	}

	err = model.UpdateWaMsgSend(ctx, s.data.db, req)
	return err
}

func (s *WaMsgSend) UpdateWaMsgSendStateByWaMsgId(ctx context.Context, waMsgID string, state int) error {
	if waMsgID == "" {
		s.l.WithContext(ctx).Error(fmt.Sprintf("msg's waMsgID is null,waMsgID:%v", waMsgID))
		return errors.New(fmt.Sprintf("msg's waMsgID is null,waMsgID:%v", waMsgID))
	}
	err := model.UpdateWaMsgSendState(ctx, s.data.db, waMsgID, state)

	return err
}

func (s *WaMsgSend) ListGtId(ctx context.Context, id int64, limit int) ([]*biz.WaMsgSend, error) {
	msgSendList, err := model.ListMsgSendGtId(ctx, s.data.db, id, uint(limit))
	if err != nil {
		return nil, err
	}
	var bizMsgSendList []*biz.WaMsgSend
	for _, msgSend := range msgSendList {
		bizMsgSendList = append(bizMsgSendList, convertWaMsgSend2SimpleBiz(msgSend))
	}
	return bizMsgSendList, nil
}

func (s *WaMsgSend) ListGtIdInPts(ctx context.Context, id int64, pts []string, limit int) ([]*biz.WaMsgSend, error) {
	msgSendList, err := model.ListMsgSendGtIdInPts(ctx, s.data.db, id, pts, uint(limit))
	if err != nil {
		return nil, err
	}
	var bizMsgSendList []*biz.WaMsgSend
	for _, msgSend := range msgSendList {
		bizMsgSendList = append(bizMsgSendList, convertWaMsgSend2SimpleBiz(msgSend))
	}
	return bizMsgSendList, nil
}

func (s *WaMsgSend) ListWaIdByState(ctx context.Context, minWaId string, limit uint, state []int) ([]string, error) {
	waIdList, err := model.ListWaIdByState(ctx, s.data.db, minWaId, limit, state)
	if err != nil {
		return nil, err
	}
	return waIdList, nil
}

func (s *WaMsgSend) ListMsgSendByWaIdAndState(ctx context.Context, state []int, waId string, ptList []string) ([]*biz.WaMsgSend, error) {
	msgSendList, err := model.ListMsgSendByWaIdAndState(ctx, s.data.db, state, waId, ptList)
	if err != nil {
		return nil, err
	}
	var bizMsgSendList []*biz.WaMsgSend
	for _, msgSend := range msgSendList {
		send := &biz.WaMsgSend{}
		util.CopyFieldsByJson(*msgSend, send)
		bizMsgSendList = append(bizMsgSendList, send)
	}
	return bizMsgSendList, nil
}
