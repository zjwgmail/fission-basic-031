package biz

import (
	"context"
	v1 "fission-basic/api/fission/v1"
	"time"
)

type NXCloudRepo interface {
	// 保存官方助力消息
	SaveOfficialMsg(ctx context.Context,
		officialMsgRecord *OfficialMsgRecord,
		msgID, content string,
	) error

	// 更新官方助力消息语言
	UpdateOfficialMsgLaguage(ctx context.Context,
		officialMsgRecord *OfficialMsgRecord,
		msgID, content string,
	) error

	// 保存非官方助力消息
	SaveUnOfficialMsg(ctx context.Context,
		unOfficialMsgRecord *UnOfficialMsgRecord,
		msgID, content string,
	) error

	// 保存消息
	SaveReceiveMsg(ctx context.Context,
		waID string,
		msgID, content string,
		sendTime int64,
		generation int,
	) error

	SaveReceiptMsg(ctx context.Context, waID, msgID, content string, msgState int, sendTime int64, cost []*v1.Cost) error

	OnlySaveReceiveMsg(ctx context.Context, waID, msgID, content string, sendTime int64) error

	// 保存续免费回复消息
	SaveRenewMsg(ctx context.Context,
		waID string,
		msgID, content string,
		sendTime int64,
	) error
}

type WaMsgReceived struct {
	ID              int64
	Del             int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
	CreateTime      time.Time // 记录创建的时间，插入记录时自动记录当前时间
	UpdateTime      time.Time // 记录更新的时间，插入或更新记录时自动更新为当前时间
	WaMsgID         string    // wa 消息的唯一标识
	WaID            string    // 可能是用户或会话的 WhatsApp ID
	Content         string    // 发送消息的具体内容
	MsgReceivedTime int64     // 消息发送的时间，以大整数表示
}

type UserRemind struct {
	ID           int64
	WaID         string    // 用户的 WhatsApp ID
	LastSendTime int64     // 最后消息发送时间
	SendTimeV0   int64     // 免费CDK消息发送时间
	StatusV0     int       // 提醒消息发送状态，1:未发送 2:已发送
	SendTimeV22  int64     // v22 信息发送时间
	StatusV22    int       // 提醒消息发送状态，1:未发送 2:已发送
	SendTimeV3   int64     // v3 信息发送时间
	StatusV3     int       // 提醒消息发送状态，1:未发送 2:已发送
	SendTimeV36  int64     // v36 信息发送时间
	StatusV36    int       // 提醒消息发送状态，1:未发送 2:已发送 3:不发送
	CreateTime   time.Time // 记录创建的时间，插入记录时自动记录当前时间
	UpdateTime   time.Time // 记录更新的时间，插入或更新记录时自动更新为当前时间
	Del          int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
}

type OfficialMsgRecord struct {
	ID         int       // 自增主键，唯一标识每条记录
	WaID       string    // 开团人标识
	RallyCode  string    // 助力码，用于识别助力活动
	State      int       // 状态表 1:未完成 2:已完成
	Channel    string    // 用户来源的渠道
	Language   string    // 用户使用的语言
	Generation int       // 用户参与活动的代数
	NickName   string    // 用户的昵称
	SendTime   int64     // 消息发送时间
	CreateTime time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
}

type UnOfficialMsgRecord struct {
	ID         int       // 自增主键，唯一标识每条记录
	WaID       string    // 开团人标识
	RallyCode  string    // 助力码，用于识别助力活动
	State      int       // 状态表 1:未完成 2:已完成
	Channel    string    // 用户来源的渠道
	Language   string    // 用户使用的语言
	Generation int       // 用户参与活动的代数
	NickName   string    // 用户的昵称
	SendTime   int64     // 消息发送时间
	CreateTime time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
}
