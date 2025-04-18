package dto

import "time"

// UserRemindDto 对应 user_remind 表的结构体
type UserRemindDto struct {
	ID           int64     // 自增主键，唯一标识每条记录
	WaID         string    // 用户的 WhatsApp ID
	LastSendTime int64     // 最后消息发送时间
	SendTimeV0   int64     // V0 信息发送时间
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
