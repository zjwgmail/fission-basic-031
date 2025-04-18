package constants

import (
	"fmt"
	"strings"
)

var (
	MsgSendStateUnSend    = 0 // 未发送
	MsgSendStateSuccess   = 1 // 发送成功
	MsgSendStateFail      = 2 // 发送失败
	MsgSendStateNxSuccess = 3 // nx发送wa成功
	MsgSendStateNxFail    = 4 // nx发送wa失败
	MsgSendStateNxTimeout = 5 // nx发送wa超时

	UserRemindVXStatusMsgNotSend = 1 // 未发送
	UserRemindVXStatusMsgSend    = 2 // 已发送
	UserRemindVXStatusMsgDone    = 9 // 未发送，也不需要再发送的特殊状态

	NxStatusSent    = "sent"
	NxStatusFailed  = "failed"
	NxStatusTimeout = "timeout"

	NxStatusMsgStateMap = map[string]int{
		NxStatusSent:    MsgSendStateNxSuccess,
		NxStatusFailed:  MsgSendStateNxFail,
		NxStatusTimeout: MsgSendStateNxTimeout,
	}

	ReceiveMsg = "receiveMsg" // 接收的消息

	ActivityTaskMsg         = "activityTaskMsg"         // 参与活动消息
	CannotAttendActivityMsg = "cannotAttendActivityMsg" // 非白，不能参与活动消息
	CanNotHelpOneselfMsg    = "CanNotHelpOneselfMsg"    // 不能助力自己
	RepeatHelpMsg           = "repeatHelpMsg"           // 重复助力消息
	StartGroupMsg           = "startGroupMsg"           // 开团消息

	SwitchLangMsg                  = "switchLangMsg"              // 切换语言消息
	HelpStartGroupMsg              = "helpStartGroupMsg"          // 受邀人开团消息
	FounderCanNotStartGroupMsg     = "founderCanNotStartGroupMsg" // 缓冲期主态不能开团消息
	CanNotStartGroupMsg            = "canNotStartGroupMsg"        // 缓冲期客态不能开团消息
	HelpTaskSingleStartMsg         = "helpTaskSingleStartMsg"     // 助力人参与活动信息
	HelpTaskSingleSuccessMsgPrefix = "helpTaskSingleSuccessMsg"   // 被人助力成功信息前缀

	HelpOverMsgPrefix = "helpOverMsg"  // 助力完成信息
	HelpOverMsg1      = "helpOverMsg1" // 3人助力完成信息
	HelpOverMsg2      = "helpOverMsg2" // 6人助力完成信息
	HelpOverMsg3      = "helpOverMsg3" // 9人助力完成信息
	HelpOverMsg4      = "helpOverMsg4" // 12人助力完成信息
	HelpOverMsg5      = "helpOverMsg5" // 15人助力完成信息

	FreeCdkMsg             = "freeCdkMsg"             // 免费CDK信息
	RedPacketReadyMsg      = "redPacketReadyMsg"      // 红包预发信息
	RedPacketSendMsg       = "redPacketSendMsg"       // 红包发放信息
	RenewFreeMsg           = "renewFreeMsg"           // 续免费信息
	PayRenewFreeMsg        = "payRenewFreeMsg"        // 付费-续免费信息
	PromoteClusteringMsg   = "promoteClusteringMsg"   // 催促成团消息
	EndCanNotStartGroupMsg = "endCanNotStartGroupMsg" // 结束期-不能开团消息
	EndCanNotHelpMsg       = "endCanNotHelpMsg"       // 结束期-不能助力消息
	RenewFreeReplyMsg      = "renewFreeReplyMsg"      // 续订回复信息

	WaRedirectListPrefix = "https://wa.me/?text="

	BizTypeInteractive = 1 // 互动消息类型
	BizTypeTemplate    = 2 // 模板消息类型

	Generation01            = "01"    // 初代
	FirstIdentificationCode = "00000" // 初代识别码

	ATStatusUnStart = "unstart" // 活动未开始
	ATStatusStarted = "started" // 活动进行中
	ATStatusBuffer  = "buffer"  // 活动缓冲期
	ATStatusEnd     = "end"     // 活动已结束

	//all redis key
	MsgSignKey        = "activity:mlbb25031:%v:msg:sign:%v"
	NxMsgIdKey        = "activity:mlbb25031:%v:nxMsgId:%v"
	ActivityInfoKey   = "activity:mlbb25031:activityInfo"
	HelpTextLockKey   = "activity:mlbb25031:%v:helpText:lock"
	HelpTextWeightKey = "activity:mlbb25031:%v:helpText:weight"
	NotWhiteSetKey    = "activity:mlbb25031:%v:notWhite:phoneSet:"
	NotWhiteCountKey  = "activity:mlbb25031:%v:notWhite:count:%v:%v:%v"
	TaskLock          = "activity:mlbb25031:%v:lock:task:%v"
	ReSendMsgLock     = "activity:mlbb25031:%v:lock:reSendMsg:%v"

	HelpCodeKey            = "help_codes_mlbb25031"      // 助力码队列key
	ImageDowngrade         = "image_downgrade_mlbb25031" // 图片降级判断
	CdkQueueKeyPrefix      = "activity_mlbb25031_cdk_v"  // cdk队列key前缀
	CdkTotalCountKeySuffix = "_count"                    // cdk总数key后缀

	//activity_mlbb25031_cdk_v0
	CdkV0        = "activity_mlbb25031_cdk_v0"
	CdkV0_COUNT  = CdkV0 + CdkTotalCountKeySuffix
	CdkV3        = "activity_mlbb25031_cdk_v3"
	CdkV3_COUNT  = CdkV3 + CdkTotalCountKeySuffix
	CdkV6        = "activity_mlbb25031_cdk_v6"
	CdkV6_COUNT  = CdkV6 + CdkTotalCountKeySuffix
	CdkV9        = "activity_mlbb25031_cdk_v9"
	CdkV9_COUNT  = CdkV9 + CdkTotalCountKeySuffix
	CdkV12       = "activity_mlbb25031_cdk_v12"
	CdkV12_COUNT = CdkV12 + CdkTotalCountKeySuffix
	CdkV15       = "activity_mlbb25031_cdk_v15"
	CdkV15_COUNT = CdkV15 + CdkTotalCountKeySuffix

	PushEvent2CountKey     = "activity:mlbb25031:pushEvent2Count2"
	PushEvent3CountKey     = "activity:mlbb25031:pushEvent3Count"
	PushEvent3OnceCountKey = "activity:mlbb25031:pushEvent3OnceCount:0307"
	PushEvent4WaIdsKey     = "activity:mlbb25031:pushEvent4WaIds"

	PushEvent2CountLimit     = 0
	PushEvent3CountLimit     = 700000
	PushEvent3OnceCountLimit = 500000

	EmailReportJobTaskLockPrefix = "email_report_job_lock_mlbb25031"
	PushEventJobTaskLockPrefix   = "push_event_job_2_lock_v7_mlbb25031"
	PushEventJob3TaskLockPrefix  = "push_event_job_3_lock_v8_mlbb25031"
)

func GetHelpTextLockKey(activityId string) string {
	return fmt.Sprintf(HelpTextLockKey, activityId)
}

func GetHelpTextWeightKey(activityId string) string {
	return fmt.Sprintf(HelpTextWeightKey, activityId)
}

func GetNotWhiteSetKey(activity string) string {
	return fmt.Sprintf(NotWhiteSetKey, activity)
}

func GetNotWhiteCountKey(activity string, date, channel, language string) string {
	date = ReplaceChineseMonthDay(date)
	return fmt.Sprintf(NotWhiteCountKey, activity, date, channel, language)
}

func GetTaskLockKey(activity, taskName string) string {
	return fmt.Sprintf(TaskLock, activity, taskName)
}

func NickJobLock(key string) string {
	return fmt.Sprintf("job:lock:%s", key)
}

func GetReSendMsgLockKey(activity, waId string) string {
	return fmt.Sprintf(ReSendMsgLock, activity, waId)
}

func ReplaceChineseMonthDay(s string) string {
	// 将字符串中的"月"替换为"M"
	s = strings.ReplaceAll(s, "月", "M")
	// 将字符串中的"日"或"号"替换为"D"
	s = strings.ReplaceAll(s, "日", "D")
	s = strings.ReplaceAll(s, "号", "D")
	return s
}
