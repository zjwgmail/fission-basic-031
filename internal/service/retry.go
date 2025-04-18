package service

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/util"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type RetryService struct {
	l                      *log.Helper
	officialRallyUsecase   *biz.OfficialRallyUsecase
	unOfficialRallyUsecase *biz.UnOfficialRallyUsecase
	msgUsecase             *biz.MsgUsecase
	redisCusterClient      *redis.ClusterClient
}

func NewRetryService(
	l log.Logger,
	officialRallyUsecase *biz.OfficialRallyUsecase,
	unOfficialRallyUsecase *biz.UnOfficialRallyUsecase,
	msgUsecase *biz.MsgUsecase,
	redisCusterClient *redis.ClusterClient,
) *RetryService {
	return &RetryService{
		l:                      log.NewHelper(l),
		officialRallyUsecase:   officialRallyUsecase,
		unOfficialRallyUsecase: unOfficialRallyUsecase,
		msgUsecase:             msgUsecase,
		redisCusterClient:      redisCusterClient,
	}
}

func (s *RetryService) RetryOfficialMsgRecord(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisCusterClient, "retry_official_msg_record", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	err = s.officialRallyUsecase.RetryMsg(ctx)
	if err != nil {
		s.l.WithContext(ctx).Errorf("RetryOfficialMsgRecord failed, err=%v", err)
		return err
	}

	return nil
}

func (s *RetryService) RetryUnOfficialMsgRecord(ctx context.Context) error {
	cost := util.MethodCost(ctx, s.l, "RetryService.RetryUnOfficialMsgRecord")
	defer cost()

	locked, unlock, err := redis.JobLock(ctx, s.redisCusterClient, "retry_unofficial_msg_record", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	err = s.unOfficialRallyUsecase.RetryMsg(ctx)
	if err != nil {
		s.l.WithContext(ctx).Errorf("RetryUnOfficialMsgRecord failed, err=%v", err)
		return err
	}

	return nil
}

func (s *RetryService) ReceiptMsgRecord(ctx context.Context) error {
	locked, unlock, err := redis.JobLock(ctx, s.redisCusterClient, "retry_receipt_msg_record", 60*time.Second)
	if err != nil || !locked {
		return err
	}
	defer unlock()

	err = s.msgUsecase.RetryReceiptMsgRecord(ctx)
	if err != nil {
		s.l.WithContext(ctx).Errorf("ReceiptMsgRecord failed, err=%v", err)
		return err
	}

	return nil
}
