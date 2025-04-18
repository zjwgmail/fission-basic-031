package dto

import (
	"fission-basic/internal/pojo/dto/nx"
)

type HelpParam struct {
	WaId      string `json:"wa_id"`
	RallyCode string `json:"rally_code"`
	IsHelp    bool   `json:"is_help"`
}

type BuildMsgInfo struct {
	WaId       string `json:"wa_id"`
	MsgType    string `json:"msg_type"`
	Channel    string `json:"channel"`
	Language   string `json:"language"`
	Generation string `json:"generation"`
	RallyCode  string `json:"rally_code"`
}

type HelpNickNameInfo struct {
	Id           int64  `json:"id"` // userInfo 表的id
	UserNickname string `json:"user_nick_name"`
	// 助力码
	RallyCode string `json:"rally_code"`
}

type SendNxListParamsDto struct {
	SendMsg       nx.NxReq
	MsgInfoEntity *BuildMsgInfo
	WaMsgSend     *WaMsgSend
}

// WaMsgSend 对应 wa_msg_send 表的结构体
type WaMsgSend struct {
	ID            int64  // 自增主键，唯一标识每条记录
	WaMsgID       string // wa 消息的唯一标识
	WaID          string // 可能是用户或会话的 WhatsApp ID
	MsgType       string // 消息类型
	State         int    // 发送状态 1:发送成功 2:发送失败
	Content       string // 发送消息的具体内容
	BuildMsgParam string // 发送消息的具体内容
	SendRes       string // 发送返回结果
	Pt            string // 分区字段
}
