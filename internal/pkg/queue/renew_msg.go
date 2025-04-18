package queue

import (
	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"

	"github.com/go-redis/redis"
)

type RenewMsg struct {
	*taskq.Queue
}

func NewRenewMsg(cli *redis.ClusterClient, d *conf.Data) *RenewMsg {
	return &RenewMsg{
		Queue: taskq.NewQueue(cli, d.Queue.RenewMsg),
	}
}
