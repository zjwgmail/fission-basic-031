package taskq

import (
	"fission-basic/internal/conf"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

var (
	d = conf.Data{
		Redis: &conf.Data_Redis{
			Addr:     "r-2zes4wcldf135nfcbipd.redis.rds.aliyuncs.com:6379",
			Password: "Redis123",
		},
	}
	cli = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{d.Redis.Addr},
		Password:     d.Redis.Password,
		MaxConnAge:   30 * time.Second,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  10 * time.Second,
		PoolSize:     200,
		MinIdleConns: 60,
		IdleTimeout:  10 * time.Second,
	})
	q = NewQueue(cli, "test")
)

func TestSend(t *testing.T) {
	err := q.SendFront([]string{"10", "20", "30", "40"}, false)

	if err != nil {
		panic(err)
	}
}

func TestRelase(t *testing.T) {
	err := q.Release([]string{"1", "2", "3", "4"})
	if err != nil {
		panic(err)
	}
}

func TestReceive(t *testing.T) {
	ids, err := q.Receive(3, time.Millisecond*10)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("ids:", ids)
}

func TestDelete(t *testing.T) {
	err := q.Delete([]string{"10", "20", "30", "40"})
	if err != nil {
		panic(err)
	}
}

func TestLen(t *testing.T) {
	l, err := q.Len()
	if err != nil {
		panic(err)
	}
	t.Log("len:", l)
}
