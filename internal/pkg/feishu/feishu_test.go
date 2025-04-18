package feishu

import (
	"context"
	"fission-basic/internal/conf"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

func TestSendTextMsg(t *testing.T) {
	feishu := NewFeishu(&conf.Data{
		Feishu: &conf.Data_Feishu{
			Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/f5d6f895-dbba-4f4c-87dc-d9542d64bf9c",
		},
	}, log.NewStdLogger(os.Stdout))

	feishu.SendTextMsg(context.TODO(), "test msg")
}
