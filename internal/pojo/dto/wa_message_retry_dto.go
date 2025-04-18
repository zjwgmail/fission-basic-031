package dto

import "time"

type WaMsgRetryDto struct {
	ID            int64     // 自增主键，唯一标识记录
	WaID          string    // wa用户ID
	WaMsgID       string    // wa消息id
	MsgType       string    // 消息类型
	State         int       // 发送状态 0:初始化 1:发送成功 2:发送失败
	Content       string    // 发送消息内容
	CreateTime    time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime    time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del           int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
	BuildMsgParam string    // 发送消息的具体内容
	SendRes       string    // 发送返回结果
}
