package queue

import (
	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"

	"github.com/go-redis/redis"
)

type GW struct {
	*taskq.Queue
}

func NewGW(
	cli *redis.ClusterClient,
	d *conf.Data,
) *GW {
	return &GW{
		Queue: taskq.NewQueue(cli, d.Queue.Gw),
	}
}
