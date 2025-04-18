package queue

import (
	taskq "fission-basic/kit/task"

	"github.com/go-redis/redis"

	"fission-basic/internal/conf"
)

type CallMsg struct {
	*taskq.Queue
}

// 回执消息队列
func NewCallMsg(cli *redis.ClusterClient, d *conf.Data) *CallMsg {
	return &CallMsg{
		Queue: taskq.NewQueue(cli, d.Queue.CallMsg),
	}
}
