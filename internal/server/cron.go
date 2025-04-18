package server

import (
	"fission-basic/contrib/cron"
	"fission-basic/internal/conf"
	"fission-basic/internal/service"

	"github.com/go-kratos/kratos/v2/log"
)

func NewCronServer(d *conf.Data,
	c *service.CronService,
	t *service.TaskService,
	ur *service.UserRemindService,
	retry *service.RetryService,
	logger log.Logger,
) *cron.Server {
	cr := cron.NewServer()
	if d.CronTask.ResendRetryMsg.Enable {
		_, _ = cr.AddFunc(d.CronTask.ResendRetryMsg.Spec, c.ResendRetryMsg)
	}

	if d.CronTask.ResendMsg.Enable {
		_, _ = cr.AddFunc(d.CronTask.ResendMsg.Spec, c.ResendMsg)
	}

	//////////////////////////////////////////////////
	// 队列监控
	if d.CronTask.OfficialQueueMonitor.Enable {
		_, _ = cr.AddFunc(d.CronTask.OfficialQueueMonitor.Spec, t.OfficialTaskMonitor)
	}
	// if d.CronTask.UnofficialQueueMonitor.Enable {
	// 	_, _ = cr.AddFunc(d.CronTask.UnofficialQueueMonitor.Spec, t.UnOfficialTaskMonitor)
	// }
	// if d.CronTask.RenewQueueMonitor.Enable {
	// 	_, _ = cr.AddFunc(d.CronTask.RenewQueueMonitor.Spec, t.RenewMsgMonitor)
	// }
	// if d.CronTask.CallMsgQueueMonitor.Enable {
	// 	_, _ = cr.AddFunc(d.CronTask.CallMsgQueueMonitor.Spec, t.CallMsgMonitor)
	// }
	// if d.CronTask.GwQueueMonitor.Enable {
	// 	_, _ = cr.AddFunc(d.CronTask.GwQueueMonitor.Spec, t.GwMsgMonitor)
	// }
	//////////////////////////////////////////////////

	if d.CronTask.ActivityTask.Enable {
		_, _ = cr.AddFunc(d.CronTask.ActivityTask.Spec, c.ActivityJob)
	}
	if d.CronTask.FeishuReportTask.Enable {
		_, _ = cr.AddFunc(d.CronTask.FeishuReportTask.Spec, c.FeishuReportTask)
	}
	if d.CronTask.EmailReportUtc8Task.Enable {
		_, _ = cr.AddFunc(d.CronTask.EmailReportUtc8Task.Spec, c.EmailReportUtc8Task)
	}
	if d.CronTask.EmailReportUtc0Task.Enable {
		_, _ = cr.AddFunc(d.CronTask.EmailReportUtc0Task.Spec, c.EmailReportUtc0Task)
	}
	if d.CronTask.EmailReportUtcMinus8Task.Enable {
		_, _ = cr.AddFunc(d.CronTask.EmailReportUtcMinus8Task.Spec, c.EmailReportUtcMinus8Task)
	}

	//////////////////////////////////////////////////
	// 重试
	if d.CronTask.RetryOfficialMsg.Enable {
		_, _ = cr.AddFunc(d.CronTask.RetryOfficialMsg.Spec, retry.RetryOfficialMsgRecord)
	}
	if d.CronTask.RetryUnofficialMsg.Enable {
		_, _ = cr.AddFunc(d.CronTask.RetryUnofficialMsg.Spec, retry.RetryUnOfficialMsgRecord)
	}
	if d.CronTask.RetryReceiptMsgRecord.Enable {
		_, _ = cr.AddFunc(d.CronTask.RetryReceiptMsgRecord.Spec, retry.ReceiptMsgRecord)
	}
	//////////////////////////////////////////////////

	//////////////////////////////////////////////////
	// 用户提醒
	if d.CronTask.UserRemindFreeCdk.Enable {
		_, _ = cr.AddFunc(d.CronTask.UserRemindFreeCdk.Spec, ur.CDKV0)
	}
	if d.CronTask.UserRemindV22.Enable {
		_, _ = cr.AddFunc(d.CronTask.UserRemindV22.Spec, ur.RenewV22)
	}
	if d.CronTask.UserRemindV3.Enable {
		_, _ = cr.AddFunc(d.CronTask.UserRemindV3.Spec, ur.RemindJoinGroupV3)
	}
	//////////////////////////////////////////////////

	//////////////////////////////////////////////////
	// 引流
	if d.CronTask.PushEvent1Send.Enable {
		_, _ = cr.AddFunc(d.CronTask.PushEvent1Send.Spec, c.PushEvent1Send)
	}
	if d.CronTask.PushEvent2Send.Enable {
		_, _ = cr.AddFunc(d.CronTask.PushEvent2Send.Spec, c.PushEvent2Send)
	}
	if d.CronTask.PushEvent3Send.Enable {
		_, _ = cr.AddFunc(d.CronTask.PushEvent3Send.Spec, c.PushEvent3Send)
	}
	if d.CronTask.PushEvent4Send.Enable {
		_, _ = cr.AddFunc(d.CronTask.PushEvent4Send.Spec, c.PushEvent4Send)
	}
	//////////////////////////////////////////////////
	return cr
}
