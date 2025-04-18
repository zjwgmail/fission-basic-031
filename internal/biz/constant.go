package biz

const (
	MsgStateDoing    int = 1 // 处理中
	MsgStateComplete int = 2 // 已完成
)

const (
	Deleted    int8 = 1 // 已删除
	NotDeleted int8 = 2 // 未删除
)

const (
	UserRemindVXStatusMsgNotSend = 1 // 未发送
	UserRemindVXStatusMsgSend    = 2 // 已发送
	UserRemindVXStatusMsgDone    = 9 // 未发送，也不需要再发送的特殊状态
)

const (
	RecallStateSuccess = 1 // 消息发送成功
	RecallStateFaild   = 2 // 消息发送失败
	RecallStateTimeout = 3 // 消息发送超时
)
