package service

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/util"
	"fmt"
	"runtime/debug"

	"github.com/go-kratos/kratos/v2/log"
)

type CronService struct {
	resendRetryJob  *biz.ResendRetryJob
	resendJob       *biz.ResendJob
	activityJob     *biz.ActivityJob
	feishuReportJob *biz.FeishuReportJob
	emailReportJob  *biz.EmailReportJob
	drainageJob     *biz.DrainageJob
	l               *log.Helper
}

func NewCronService(
	activityJob *biz.ActivityJob,
	resendRetryJob *biz.ResendRetryJob,
	resendJob *biz.ResendJob,
	feishuReportJob *biz.FeishuReportJob,
	emailReportJob *biz.EmailReportJob,
	drainageJob *biz.DrainageJob,
	l log.Logger,
) *CronService {
	return &CronService{
		activityJob:     activityJob,
		resendRetryJob:  resendRetryJob,
		resendJob:       resendJob,
		feishuReportJob: feishuReportJob,
		emailReportJob:  emailReportJob,
		drainageJob:     drainageJob,
		l:               log.NewHelper(l),
	}
}

func (c *CronService) ResendRetryMsg(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[ResendRetryMsg]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("ResendRetryMsg start"))
	c.resendRetryJob.ResendRetryJobHandle(context.Background(), "ResendRetryMsg")

	c.l.Info(fmt.Sprintf("ResendRetryMsg end"))
	return nil
}

func (c *CronService) ResendMsg(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[ResendMsg]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("ResendMsg start"))
	c.resendJob.ResendJobHandle(context.Background(), "ResendMsg")

	c.l.Info(fmt.Sprintf("ResendMsg end"))
	return nil
}

func (c *CronService) ActivityJob(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[ActivityJob]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("ActivityJob start"))
	c.activityJob.ActivityJobHandle(context.Background(), "ActivityJob")

	c.l.Info(fmt.Sprintf("ActivityJob end"))
	return nil
}

func (c *CronService) FeishuReportTask(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[FeishuReportTask]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("FeishuReportTask start"))
	c.feishuReportJob.SendReport(ctx)
	_ = fmt.Sprintf("FeishuReportTask start")
	return nil
}

func (c *CronService) EmailReportUtc8Task(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[EmailReportUtc8Task]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("EmailReportUtc8Task start"))
	c.emailReportJob.SendReport(ctx, 8)
	_ = fmt.Sprintf("EmailReportUtc8Task start")
	return nil
}

func (c *CronService) EmailReportUtc0Task(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[EmailReportUtc0Task]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("EmailReportUtc0Task start"))
	c.emailReportJob.SendReport(ctx, 0)
	_ = fmt.Sprintf("EmailReportUtc0Task start")
	return nil
}

func (c *CronService) EmailReportUtcMinus8Task(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[EmailReportUtcMinus8Task]，panic occurs,err: %v", e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("EmailReportUtcMinus8Task start"))
	c.emailReportJob.SendReport(ctx, -8)
	_ = fmt.Sprintf("EmailReportUtcMinus8Task start")
	return nil
}

func (c *CronService) PushEvent1Send(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[%s]，panic occurs,err: %v", util.GetCurrentFuncName(), e))
			return
		}
	}()
	//c.l.Info(fmt.Sprintf("uploadUserSend start"))
	c.drainageJob.SendPushEvent1Msg(ctx)
	//_ = fmt.Sprintf("uploadUserSend start")
	return nil
}

func (c *CronService) PushEvent2Send(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[%s]，panic occurs,err: %v", util.GetCurrentFuncName(), string(debug.Stack())))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("uploadUserSend start"))
	c.drainageJob.SendPushEvent2Msg(ctx)
	_ = fmt.Sprintf("uploadUserSend end")
	return nil
}

func (c *CronService) PushEvent3Send(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[%s]，panic occurs,err: %v", util.GetCurrentFuncName(), e))
			return
		}
	}()
	//c.l.Info(fmt.Sprintf("uploadUserSend start"))
	c.drainageJob.SendPushEvent3Msg(ctx)
	//_ = fmt.Sprintf("uploadUserSend start")
	return nil
}

func (c *CronService) PushEvent4Send(ctx context.Context) error {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			c.l.Error(fmt.Sprintf("method[%s]，panic occurs,err: %v", util.GetCurrentFuncName(), e))
			return
		}
	}()
	c.l.Info(fmt.Sprintf("uploadUserSend start"))
	c.drainageJob.SendPushEvent4Msg(ctx)
	_ = fmt.Sprintf("uploadUserSend start")
	return nil
}
