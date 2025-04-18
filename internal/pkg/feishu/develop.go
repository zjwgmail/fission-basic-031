package feishu

import (
	"context"
	"fission-basic/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-resty/resty/v2"
)

type Develop struct {
	webHook string
	l       *log.Helper
}

func NewDevelop(d *conf.Data,
	l log.Logger,
) *Develop {
	return &Develop{
		webHook: d.Feishu.DevelopWebhook,
		l:       log.NewHelper(l),
	}
}

func (f *Develop) SendTextMsg(ctx context.Context, msg string) error {
	content := map[string]any{
		"text": "有手机 \n" + msg,
	}
	body := map[string]any{
		"msg_type": "text",
		"content":  content,
	}

	return f.sendMsg(ctx, body)
}

func (f *Develop) sendMsg(ctx context.Context, msg map[string]any) error {
	client := resty.New()

	_, err := client.R().SetBody(msg).Post(f.webHook)
	if err != nil {
		f.l.WithContext(ctx).Errorf("send feishu msg failed, err=%v, msg=%s", err, msg)
		return err
	}

	return nil
}
