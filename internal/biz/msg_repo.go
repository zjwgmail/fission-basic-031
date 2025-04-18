package biz

import (
	"context"
	"time"
)

type MsgRepo interface {
	FindWaMsgSend(ctx context.Context, msgID string) (*WaMsgSend, error)
	FindWaMsgRetry(ctx context.Context, msgID string) (*WaMsgRetry, error)
	FindMsgReceipt(ctx context.Context, msgID string) (*ReceiptMsgRecord, error)

	// CompleteReceiptMsg 完成回执消息处理
	CompleteReceiptMsg(ctx context.Context, msgID string) error

	// DeleteWaMsgSend 完成回执消息处理+删除重试记录
	CompleteReceiptAndDeleteRetry(ctx context.Context, msgID string) error

	CompleteReceiptAndAddMsgRetry(ctx context.Context, msgState int, msgID, waID, content string, msgType, buildMsgParam, table string) error

	ListDoingReceiptMsgRecords(ctx context.Context, minID int, offset, length uint, maxTime time.Time) ([]*ReceiptMsgRecord, error)
}

type ReceiptMsgRecord struct {
	ID         int       // 自增主键，唯一标识每条记录
	MsgID      string    // wa消息id
	MsgState   int       // wa消息状态
	State      int       // 状态表 1:未完成 2:已完成
	CreateTime time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
	CostInfo   string    // 花费实体json数据
	WaID       string    // wa用户ID
}

type WaMsgSend struct {
	ID            int64     // 自增主键一部分，唯一标识记录
	WaMsgID       string    // wa消息id
	WaID          string    // wa用户ID
	State         int       // 发送状态 1:发送成功 2:发送失败
	Content       string    // 发送消息内容
	MsgType       string    // 消息类型
	Pt            string    // 分区字段
	CreateTime    time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime    time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del           int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
	BuildMsgParam string    `db:"build_msg_param"` // 发送消息的具体内容
	SendRes       string    `db:"send_res"`        // 发送返回结果
}

type WaMsgRetry struct {
	ID         int64     // 自增主键，唯一标识记录
	WaID       string    // wa用户ID
	WaMsgID    string    // wa消息id
	MsgType    string    // 消息类型
	State      int       // 发送状态 0:初始化 1:发送成功 2:发送失败
	Content    string    // 发送消息内容
	CreateTime time.Time // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      // 标记记录是否删除，0 表示未删除，1 表示已删除
}
