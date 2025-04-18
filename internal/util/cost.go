package util

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func MethodCost(ctx context.Context, l *log.Helper, method string) func() {
	start := time.Now()
	return func() {
		l.WithContext(ctx).Infof("methodCost, method:%s cost: %s", method, time.Since(start).String())
	}
}
