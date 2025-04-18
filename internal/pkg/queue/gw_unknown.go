package queue

import (
	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"

	"github.com/go-redis/redis"
)

type GWUnknown struct {
	*taskq.Queue
}

func NewGWUnknown(
	cli *redis.ClusterClient,
	d *conf.Data,
) *GWUnknown {
	return &GWUnknown{
		Queue: taskq.NewQueue(cli, d.Queue.GwUnknown),
	}
}
