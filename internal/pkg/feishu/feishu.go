package feishu

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-resty/resty/v2"
	"github.com/google/wire"

	"fission-basic/internal/conf"
)

var ProviderSet = wire.NewSet(NewFeishu, NewDevelop)

type Feishu struct {
	webHook string
	l       *log.Helper
}

func NewFeishu(d *conf.Data,
	l log.Logger,
) *Feishu {
	return &Feishu{
		webHook: d.Feishu.Webhook,
		l:       log.NewHelper(l),
	}
}

func (f *Feishu) SendTextMsg(ctx context.Context, msg string) error {
	content := map[string]any{
		"text": "有手机 \n" + msg,
	}
	body := map[string]any{
		"msg_type": "text",
		"content":  content,
	}

	return f.sendMsg(ctx, body)
}

func (f *Feishu) sendMsg(ctx context.Context, msg map[string]any) error {
	client := resty.New()

	_, err := client.R().SetBody(msg).Post(f.webHook)
	if err != nil {
		f.l.WithContext(ctx).Errorf("send feishu msg failed, err=%v, msg=%s", err, msg)
		return err
	}

	// f.l.WithContext(ctx).Debugf("resp=%v", resp)
	return nil
}
