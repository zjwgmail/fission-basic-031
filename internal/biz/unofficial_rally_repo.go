package biz

import (
	"context"
	"time"

	"fission-basic/internal/pojo/dto"
)

type UnOfficialRallyRepo interface {
	// 完成助力消息处理
	CompleteRally(ctx context.Context,
		waID, rallyCode string,
		msgSends []*dto.WaMsgSend, withMsgDB bool) error

	CreateJoinGroup2(
		ctx context.Context,
		rallyInfo, helpInfo *BaseInfo,
		msgSends []*dto.WaMsgSend,
		newJoinNum int,
	) error

	// 已经够最大助力次数时，助力人需要客态开团
	CreateStartedMaxJoinGroup(ctx context.Context,
		rallyInfo *BaseInfo,
		helpCode, helpWaID string, newHelpNum int,
		msgSends []*dto.WaMsgSend,
	) error

	CreateBufferMaxJoinGroup(ctx context.Context,
		rallyInfo *BaseInfo,
		helpCode, helpWaID string, newHelpNum int,
		msgSends []*dto.WaMsgSend,
	) error

	CreateBufferJoinGroup(
		ctx context.Context,
		rallyInfo, helpInfo *BaseInfo,
		msgSends []*dto.WaMsgSend,
		newJoinNum int,
	) error

	FindMsg(ctx context.Context, waID, rallyCode string) (*UnOfficialMsgRecord, error)

	ListDoingMsgs(ctx context.Context, minID int, offset, length uint, maxTime time.Time) ([]*UnOfficialMsgRecord, error)

	// 查找用户开团数据
	FindUserCreateGroup(ctx context.Context, waID string) (*UserCreateGroup, error)
	// 查询开团数据
	FindUserCreateGroupByHelpCode(ctx context.Context, helpCode string) (*UserCreateGroup, error)

	// 查询参团数据
	ListUserJoinGroups(ctx context.Context, helpCode string) ([]*UserJoinGroup, error)

	FindUserJoinGroupByWaID(ctx context.Context, waID string) (*UserJoinGroup, error)
}

type UserJoinGroup struct {
	ID            int64     // 自增主键，唯一标识每条记录
	JoinWaID      string    // 助力人ID
	HelpCode      string    // 被助力码
	JoinGroupTime int64     // 助力时间
	CreateTime    time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime    time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del           int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
}
