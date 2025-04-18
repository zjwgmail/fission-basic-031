package redis

import (
	"context"
	"fission-basic/internal/conf"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func TestRedisLock(t *testing.T) {

	d := conf.Data{
		Redis: &conf.Data_Redis{
			Addr:     "r-2zes4wcldf135nfcbipd.redis.rds.aliyuncs.com:6379",
			Password: "Redis123",
		},
	}

	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{d.Redis.Addr},
		Password:     d.Redis.Password,
		MaxConnAge:   30 * time.Second,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  10 * time.Second,
		PoolSize:     200,
		MinIdleConns: 60,
		IdleTimeout:  10 * time.Second,
	})

	locked, unlock, err := ConsumerLock(context.Background(), cli, "test", 60*time.Second)
	if err != nil {
		t.Errorf("get lock failed, err=%v", err)
		return
	}
	if !locked {
		t.Error("lock failed")
		return
	}
	defer unlock()
}
