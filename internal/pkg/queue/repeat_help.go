package queue

import (
	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"

	"github.com/go-redis/redis"
)

type RepeatHelp struct {
	*rally
}

func NewRepeatHelp(cli *redis.ClusterClient, d *conf.Data) *RepeatHelp {
	return &RepeatHelp{
		rally: &rally{
			Queue: taskq.NewQueue(cli, d.Queue.RepeatHelpKey),
		},
	}
}
