package data

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pkg/redis"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.ActivityInfoRepo = (*ActivityInfo)(nil)

type ActivityInfo struct {
	data         *Data
	l            *log.Helper
	redisService *redis.RedisService
}

func NewActivityInfo(d *Data, logger log.Logger, redisService *redis.RedisService) biz.ActivityInfoRepo {
	return &ActivityInfo{
		data:         d,
		l:            log.NewHelper(logger),
		redisService: redisService,
	}
}

func (s *ActivityInfo) GetActivityInfo(ctx context.Context, id string) (*biz.ActivityInfoDto, error) {

	// getInfo := s.redisService.Get(constants.ActivityInfoKey)
	// if getInfo != "" {
	// 	res := &biz.ActivityInfoDto{}
	// 	err := json.Unmarshal([]byte(getInfo), res)
	// 	if err != nil {
	// 		s.l.WithContext(ctx).Errorf("Failed to unmarshal json: %v", err)
	// 		return nil, err
	// 	}
	// 	return res, nil
	// }

	info, err := model.GetActivityInfo(ctx, s.data.db, id)
	if err != nil {
		s.l.WithContext(ctx).Error(fmt.Sprintf("get activityInfo from database. error: %v", err))
		return nil, err
	}

	req := &biz.ActivityInfoDto{
		Id:             info.Id,
		ActivityStatus: info.ActivityStatus,
		CreatedAt:      info.CreatedAt,
		UpdatedAt:      info.UpdatedAt,
		StartAt:        info.StartAt.Time,
		EndAt:          info.EndAt.Time,
		EndBufferDay:   int(info.EndBufferDay.Int64),
		EndBufferAt:    info.EndBufferAt.Time,
		ReallyEndAt:    info.ReallyEndAt.Time,
		CostMax:        info.CostMax,
	}

	// marshal, err := json.Marshal(req)
	// if err != nil {
	// 	s.l.WithContext(ctx).Errorf("Failed to marshal json: %v", err)
	// 	return nil, err
	// }
	// set := s.redisService.Set(constants.ActivityInfoKey, string(marshal))
	// if !set {
	// 	return nil, errors.New("failed to set activityInfo to redis")
	// }
	return req, nil
}

func (s *ActivityInfo) UpdateActivityInfo(ctx context.Context, stu *biz.UpdateActivityInfoDto) error {
	req := &model.ActivityInfo{
		Id:             stu.Id,
		ActivityStatus: stu.ActivityStatus,
	}
	err := model.UpdateActivityStatus(ctx, s.data.db, req)
	if err != nil {
		s.l.WithContext(ctx).Error(fmt.Sprintf("UpdateActivityInfo error: %v", err))
		return err
	}

	del := s.redisService.Del(constants.ActivityInfoKey)
	if !del {
		return errors.New("delete activityInfo redis error")
	}
	return nil
}
