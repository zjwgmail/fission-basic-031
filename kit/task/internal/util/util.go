package util

import (
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

func ImportLib(script string, lib ...string) string {
	return strings.Join(append(lib, script), "\r\n")
}

func ImportKitLib(script string) string {
	return ImportLib(script, KitLib)
}

func StringArgs(ids []string) []interface{} {
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return args
}

func Receive(cli *redis.ClusterClient, script *redis.Script, key string, limit int, d time.Duration) ([]string, error) {
	rkeys := []string{ /*fmt.Sprint(timenowUS()),*/ key}
	rargs := []interface{}{timeoutUS(d), limit}

	r, err := script.Run(cli, rkeys, rargs...).Result()
	if err != nil {
		return nil, err
	}

	vals, ok := r.([]interface{})
	if !ok {
		return nil, errors.New("invalid strings")
	}

	if len(vals) <= 2 {
		return nil, nil
	}

	ret := make([]string, 0, len(vals)-2)
	for _, v := range vals[2:] {
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("invalid string")
		}
		ret = append(ret, s)
	}

	return ret, nil
}

func timenowUS() int64 {
	return TimestampUS(time.Now())
}

func TimestampUS(t time.Time) int64 {
	return t.UnixNano() / int64(time.Microsecond)
}

func timeoutUS(d time.Duration) int64 {
	return int64(d / time.Microsecond)
}
