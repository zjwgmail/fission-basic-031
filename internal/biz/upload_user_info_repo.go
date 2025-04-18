package biz

import (
	"context"
	"time"
)

type UploadUserInfoRepo interface {
	InsertBatch(ctx context.Context, list []*UploadUserInfoDTO) error
	UpdateState(ctx context.Context, phoneNumber string, state int) error
	ListInNumber(ctx context.Context, phoneNumberList []string) ([]*UploadUserInfoDTO, error)
	ListGtIdWithState(ctx context.Context, id int, state int, limit uint) ([]*UploadUserInfoDTO, error)
}

type UploadUserInfoDTO struct {
	Id           int
	PhoneNumber  string
	LastSendTime time.Time
	State        int
}
