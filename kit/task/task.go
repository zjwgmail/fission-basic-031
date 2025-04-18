package taskq

import (
	"strings"
	"time"

	"fission-basic/kit/task/internal/queue"

	"github.com/go-redis/redis"
)

type Queue struct {
	key    string
	client *redis.ClusterClient
}

func NewQueue(cli *redis.ClusterClient, key string) *Queue {
	return &Queue{
		client: cli,
		key:    niceKey(key),
	}
}

func (q *Queue) Name() string {
	return rawKey(q.key)
}

func (q *Queue) SendBack(ids []string, force bool) error {
	_, err := queue.Send(q.client, q.key, ids, false, force)
	return err
}

func (q *Queue) SendFront(ids []string, force bool) error {
	_, err := queue.Send(q.client, q.key, ids, true, force)
	return err
}

func (q *Queue) Receive(n uint, d time.Duration) ([]string, error) {
	return queue.Receive(q.client, q.key, int(n), d)
}

func (q *Queue) Release(ids []string) error {
	return queue.Release(q.client, q.key, ids)
}

func (q *Queue) Delete(ids []string) error {
	return queue.Delete(q.client, q.key, ids)
}

func (q *Queue) Len() (int, error) {
	return queue.Len(q.client, q.key)
}

const prefix = "taskq:"

func niceKey(key string) string {
	return prefix + key
}

func rawKey(key string) string {
	return strings.TrimPrefix(key, prefix)
}
