package biz

import (
	"context"
	"fission-basic/internal/pojo/dto"
)

type UserRemindRepo interface {
	// ListUserRemindTODOV3 获取未发送v3消息的用户提醒
	ListUserRemindTODOV3(ctx context.Context, offset, length uint, minID, minSendTime int64) ([]*UserRemind, error)
	// ListUserRemindTODOV22 获取未发送v22消息的用户提醒
	ListUserRemindTODOV22(ctx context.Context, offset, length uint, minID, minSendTime int64) ([]*UserRemind, error)
	// ListUserRemindTODOV0 获取未发送v0消息的用户提醒
	ListUserRemindTODOV0(ctx context.Context, offset, length uint, minID, minSendTime int64) ([]*UserRemind, error)

	// V3 消息处理完成
	CompleteUserRemindV3Status(ctx context.Context, waID string, oldStatus, status int, waMsgSends []*dto.WaMsgSend) error
	// V22 消息处理完成
	CompleteUserRemindV22Status(ctx context.Context, waID string, oldStatus, status int, waMsgSends []*dto.WaMsgSend) error
	// FreeSDK 消息处理完成
	CompleteUserRemindV0Status(ctx context.Context, waID string, oldStatus, status int, waMsgSends []*dto.WaMsgSend) error

	// 获取用户信息
	GetUserInfo(ctx context.Context, waID string) (*UserInfo, error)

	// GetUserRemindInfo 获取用户提醒信息
	GetUserRemindInfo(ctx context.Context, waID string) (*dto.UserRemindDto, error)
}
