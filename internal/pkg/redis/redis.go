package redis

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/go-redis/redis"
	"github.com/google/wire"

	"fission-basic/internal/conf"
)

var ProviderSet = wire.NewSet(
	NewRedisClient,
)

var ConsumerProviderSet = wire.NewSet(
	NewRedisClient,
	NewRedisService,
)

var JobProviderSet = wire.NewSet(
	NewRedisClient,
	NewRedisService,
)

type ClusterClient = redis.ClusterClient

func NewRedisClient(d *conf.Data) *ClusterClient {
	log.Infof("redis.Addr=%v", d.Redis.Addr)
	log.Infof("redis.Pwd=%v", d.Redis.Password)
	log.Infof("redis.PoolSize=%v", d.Redis.PoolSize)
	log.Infof("redis.MinIdleConns=%v", d.Redis.MinIdleConns)
	//todo zsj 优化参数
	return redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:              []string{d.Redis.Addr},
			Password:           d.Redis.Password,
			MaxRedirects:       3,
			MaxConnAge:         60 * time.Second,
			WriteTimeout:       3 * time.Second,
			DialTimeout:        5 * time.Second,
			ReadTimeout:        3 * time.Second,
			PoolSize:           int(d.Redis.PoolSize),
			MinIdleConns:       int(d.Redis.MinIdleConns),
			PoolTimeout:        5 * time.Second,
			IdleTimeout:        300 * time.Second,
			IdleCheckFrequency: 60 * time.Second,
		},
		// &redis.Options{
		// 	Addr:         d.Redis.Addr,
		// 	Network:      d.Redis.Network,
		// 	Password:     d.Redis.Password,
		// 	DB:           6, // 这个有点坑
		// 	MaxConnAge:   30 * time.Second,
		// 	DialTimeout:  2 * time.Second,
		// 	ReadTimeout:  10 * time.Second,
		// 	PoolSize:     200,
		// 	MinIdleConns: 60,
		// 	IdleTimeout:  10 * time.Second,
		// },
	)
}

func JobLock(ctx context.Context, redisClient *redis.ClusterClient, key string, expiration time.Duration) (bool, func() bool, error) {
	return lock(ctx, redisClient, "job:lock:"+key, expiration)
}

func ConsumerLock(ctx context.Context, redisClient *redis.ClusterClient, key string, expiration time.Duration) (bool, func() bool, error) {
	return lock(ctx, redisClient, "consumer:lock:"+key, expiration)
}

func lock(ctx context.Context, redisClient *redis.ClusterClient, key string, expiration time.Duration) (bool, func() bool, error) {
	res, err := redisClient.SetNX(key, "1", expiration).Result()
	if err != nil {
		return false, nil, err
	}

	if !res {
		return res, nil, nil
	}

	return res, func() bool {
		return Del(ctx, redisClient, key)
	}, nil
}

func Del(ctx context.Context, redisClient *redis.ClusterClient, key string) bool {
	if "" == key {
		return false
	}

	_, err := redisClient.Del(key).Result()
	if err != nil {
		return false
	}

	return true
}
