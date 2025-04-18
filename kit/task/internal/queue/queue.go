package queue

import (
	"time"

	"fission-basic/kit/task/internal/util"

	"github.com/go-redis/redis"
)

func Receive(cli *redis.ClusterClient, key string, n int, d time.Duration) ([]string, error) {
	return util.Receive(cli, receiveScript, key, n, d)
}

func Send(cli *redis.ClusterClient, key string, ids []string, front, force bool) (int, error) {
	rkeys := []string{key}
	rargs := make([]interface{}, 0, len(ids)+1)

	rargs = append(rargs, force)
	rargs = append(rargs, util.StringArgs(ids)...)

	if front {
		return lsendScript.Run(cli, rkeys, rargs...).Int()
	}

	return rsendScript.Run(cli, rkeys, rargs...).Int()
}

func Release(cli *redis.ClusterClient, key string, ids []string) error {
	return releaseScript.Run(cli, []string{key}, util.StringArgs(ids)...).Err()
}

func Len(cli *redis.ClusterClient, key string) (int, error) {
	return lenScript.Run(cli, []string{key}).Int()
}

func Delete(cli *redis.ClusterClient, key string, ids []string) error {
	return delScript.Run(cli, []string{key}, util.StringArgs(ids)...).Err()
}
