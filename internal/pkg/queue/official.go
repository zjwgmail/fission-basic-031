package queue

import (
	"fmt"

	"github.com/go-redis/redis"

	"fission-basic/internal/conf"
	taskq "fission-basic/kit/task"
)

type Official struct {
	*rally
}

func NewOfficialQueue(cli *redis.ClusterClient, d *conf.Data) *Official {
	fmt.Println("NewOfficialQueue=", d.Queue.OfficialKey)
	return &Official{
		rally: &rally{
			Queue: taskq.NewQueue(cli, d.Queue.OfficialKey),
		},
	}
}
