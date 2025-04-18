package biz

import (
	"context"
	"time"

	"fission-basic/internal/pojo/dto"
)

type OfficialRallyRepo interface {
	// 完成助力消息处理
	CompleteRally(ctx context.Context, waID, rallyCode string,
		waMsgSends []*dto.WaMsgSend,
	) error

	// 开团: 新建用户开团，消息处理完成
	CreateUserGroup(ctx context.Context,
		userInfo *UserInfo, lastSendTime int64,
		waID, rallyCode, helpCode string,
		waMsgSends []*dto.WaMsgSend,
	) error

	// 更新用户语言，消息处理完成
	UpdateUserInfoLanguageByWaID(ctx context.Context,
		waID, rallyCode, language string,
		waMsgSends []*dto.WaMsgSend,
	) error

	FindMsg(ctx context.Context,
		waID, rallyCode string) (*OfficialMsgRecord, error)

	// 查找尚未处理完成的消息
	ListDoingMsg(ctx context.Context, minID int, offset, length uint, maxTime time.Time) ([]*OfficialMsgRecord, error)

	// 查找用户开团数据
	FindUserCreateGroup(ctx context.Context, waID string) (*UserCreateGroup, error)
}

type UserInfo struct {
	ID         int64     // 自增主键，唯一标识每条记录
	Del        int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
	CreateTime time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time // 记录更新时间，更新记录时自动更新为当前时间
	WaID       string    // 用户的唯一标识
	HelpCode   string    // 助力码
	Channel    string    // 用户来源的渠道
	Language   string    // 用户使用的语言
	Generation int       // 用户参与活动的代数
	JoinCount  int       // 用户的助力人数
	CDKv0      string    // 类型为 v0 的 CDK 码
	CDKv3      string    // 类型为 v3 的 CDK 码
	CDKv6      string    // 类型为 v6 的 CDK 码
	CDKv9      string    // 类型为 v9 的 CDK 码
	CDKv12     string    // 类型为 v12 的 CDK 码
	CDKv15     string    // 类型为 v15 的 CDK 码
	Nickname   string    // 用户的昵称
}

type UserCreateGroup struct {
	ID              int    // 自增主键，唯一标识每条记录
	CreateWAID      string // 开团人ID
	HelpCode        string // 助力码
	Generation      int    // 代次
	CreateGroupTime int64
	CreateTime      time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime      time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del             int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
}
