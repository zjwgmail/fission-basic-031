package queue

import (
	"github.com/go-redis/redis"

	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"
)

type UnOfficial struct {
	*rally
}

func NewUnOfficialQueue(cli *redis.ClusterClient, d *conf.Data) *UnOfficial {
	return &UnOfficial{
		rally: &rally{
			Queue: taskq.NewQueue(cli, d.Queue.UnofficialKey),
		},
	}
}
