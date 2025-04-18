package queue

import (
	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"

	"github.com/go-redis/redis"
)

// 回执
type GWRecall struct {
	*taskq.Queue
}

func NewGWRecal(
	cli *redis.ClusterClient,
	d *conf.Data,
) *GWRecall {
	return &GWRecall{
		Queue: taskq.NewQueue(cli, d.Queue.GwRecall),
	}
}
