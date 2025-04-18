package redis

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/util/encoder/json"
	"fission-basic/util"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
)

type RedisService struct {
	redisClient *redis.ClusterClient
	l           *log.Helper
	confData    *conf.Data
	business    *conf.Business
}

func NewRedisService(
	redisClient *redis.ClusterClient,
	l log.Logger,
	confData *conf.Data,
	business *conf.Business,
) *RedisService {
	service := &RedisService{
		redisClient: redisClient,
		l:           log.NewHelper(l),
		confData:    confData,
		business:    business,
	}
	service.InitHelpWeight()
	return service
}

// CheckMsgSignIsExists 校验消息幂等
func (n *RedisService) CheckMsgSignIsExists(ctx context.Context, signNiuxin string) (bool, error) {
	template := n.redisClient
	redisKey := fmt.Sprintf(constants.MsgSignKey, n.business.Activity.Id, signNiuxin)
	exists, err := template.Exists(redisKey).Result()
	if err != nil {
		n.l.WithContext(ctx).Errorf("redis校验key是否存在报错,传过来的sign:%v；err:%v", signNiuxin, err)
		return true, err
	}
	if exists != 0 {
		n.l.WithContext(ctx).Warnf("消息sign已存在不做处理,传过来的sign:%v ", signNiuxin)
		return true, nil
	}

	n.l.WithContext(ctx).Infof("消息不重复，处理此消息，sign：%v ", signNiuxin)
	_, err = template.Set(redisKey, "1", 2*time.Minute).Result()
	if err != nil {
		n.l.WithContext(ctx).Errorf("消息重复缓存新增失败,传过来的sign:%v ", signNiuxin)
		return true, err
	}
	return false, nil
}

func (n *RedisService) CreateQueueAndPut(ctx context.Context, queueName string, queueData string) error {
	template := n.redisClient
	_, err := template.LPush(queueName, queueData).Result()
	if err != nil {
		return fmt.Errorf("redis create queue error,queueName:%v；err:%v", queueName, err)
	}
	return nil
}

func (n *RedisService) CreateQueueAndBatchPut(ctx context.Context, queueName string, queueData ...string) error {
	template := n.redisClient
	_, err := template.LPush(queueName, queueData).Result()
	if err != nil {
		return fmt.Errorf("redis create queue error,queueName:%v；err:%v", queueName, err)
	}
	return nil
}

func (n *RedisService) PopQueueData(ctx context.Context, queueName string) (string, error) {
	template := n.redisClient
	data, err := template.RPop(queueName).Result()
	if err != nil {
		return "", fmt.Errorf("redis get queue error,queueName:%v；err:%v", queueName, err)
	}
	return data, nil
}

func (n *RedisService) GetQueueSize(ctx context.Context, queueName string) (int, error) {
	template := n.redisClient
	count, err := template.LLen(queueName).Result()
	if err != nil {
		return 0, fmt.Errorf("redis get queue size error,queueName:%v；err:%v", queueName, err)
	}
	return int(count), nil
}

func (n *RedisService) LIndex(ctx context.Context, queueName string, index int64) (string, error) {
	template := n.redisClient
	data, err := template.LIndex(queueName, index).Result()
	if err != nil {
		return "", fmt.Errorf("redis LIndex error, queueName:%v, index:%v, err:%v", queueName, index, err)
	}
	return data, nil
}

func (n *RedisService) InitHelpWeight() {
	methodName := util.GetCurrentFuncName()
	template := n.redisClient

	lockKey := constants.GetHelpTextLockKey(n.business.Activity.Id)

	getLock, err := template.SetNX(lockKey, "1", time.Second*60).Result()
	if err != nil {
		n.l.Errorf(fmt.Sprintf("method[%s],call redis nx fail，this server not run this task err:%v", methodName, err))
		return
	}
	if !getLock {
		n.l.Errorf(fmt.Sprintf("method[%s],get lock fail from redis，this server not run this task err:%v", methodName, err))
		return
	}

	defer func() {
		_, err = template.Del(lockKey).Result()
		if err != nil {
			n.l.Error(fmt.Sprintf("method[%s],delete lock fail from redis，this server not run this task err:%v", methodName, err))
		}
	}()

	helpTextWeightKey := constants.GetHelpTextWeightKey(n.business.Activity.Id)
	helpTextList := n.business.Activity.HelpTextList
	paramsBytes, err := json.NewEncoder().Encode(helpTextList)
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s]，HelpTextList convert json fail,HelpTextList:%v,err:%v", methodName, helpTextList, err))
		return
	}
	paramsStr := string(paramsBytes)
	_, err = template.Set(helpTextWeightKey, paramsStr, -1).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s]，query %v init help weight fail,err：%v", methodName, helpTextWeightKey, err))
		return
	}
	return
}

// GetHelpTextWeight 获取权重的
func (n *RedisService) GetHelpTextWeight(ctx context.Context) (*conf.Business_HelpText, error) {
	template := n.redisClient
	methodName := util.GetCurrentFuncName()

	helpTextWeightKey := constants.GetHelpTextWeightKey(n.business.Activity.Id)

	code, err := template.Exists(helpTextWeightKey).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s],Check whether an error occurs on the redis key %v,err：%v", methodName, helpTextWeightKey, err))
		return nil, err
	}
	if code == 0 {
		n.l.Error(fmt.Sprintf("method[%s],helpText data is'nt exists on the redis key %v", methodName, helpTextWeightKey))
		return nil, err
	}

	str, err := template.Get(helpTextWeightKey).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s],query %v from the redis fail", methodName, helpTextWeightKey))
		return nil, err
	}

	helpTextList := make([]*conf.Business_HelpText, 0)
	err = json.NewEncoder().Decode([]byte(str), &helpTextList)
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s],HelpTextList convert json fail,HelpTextList:%v,err:%v", methodName, helpTextList, err))
		return nil, err
	}
	if len(helpTextList) <= 0 {
		n.l.Error(fmt.Sprintf("method[%s],HelpTextList's length is zero. HelpTextList:%v,err:%v", methodName, helpTextList, err))
		return nil, err
	}

	// 计算权重的累积和
	var totalWeight int
	for _, helpText := range helpTextList {
		totalWeight += int(helpText.Weight)
	}

	// 生成一个随机数，范围从 0 到 totalWeight
	rand.Seed(time.Now().UnixNano()) // 随机种子
	randomWeight := rand.Intn(totalWeight)

	// 根据随机数选择对应的值
	for _, helpText := range helpTextList {
		randomWeight -= int(helpText.Weight)
		if randomWeight <= 0 {
			return helpText, nil
		}
	}

	return nil, errors.New("no value")
}

func (n *RedisService) SAddKey(methodName string, key string, value string) (int64, error) {
	// 给非白拦截redis增加phone
	template := n.redisClient
	newCount, err := template.SAdd(key, value).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s],add redis key:%v fail,err:%v", methodName, key, err))
		return 0, err
	}
	n.l.Info(fmt.Sprintf("method[%s],add redis key:%v success.addCount：%v", methodName, key, newCount))
	return newCount, nil
}

func (n *RedisService) AddIncrKey(methodName string, key string) (int64, error) {
	// 给非白拦截redis增加次数
	template := n.redisClient
	newCount, err := template.Incr(key).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s],add redis key:%v fail,err:%v", methodName, key, err))
		return 0, err
	}
	n.l.Info(fmt.Sprintf("method[%s],add redis key:%v success", methodName, key))
	return newCount, nil
}

// SetNX 分布式锁
func (n *RedisService) SetNX(methodName string, key, value string, timeout time.Duration) (bool, error) {
	template := n.redisClient
	res, err := template.SetNX(key, value, timeout).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("method[%s],setnx redis key:%v fail,err:%v", methodName, key, err))
		return false, err
	}
	n.l.Info(fmt.Sprintf("method[%s],setnx redis key:%v success", methodName, key))
	return res, nil
}

func (n *RedisService) Del(key string) bool {
	if "" == key {
		n.l.Error(fmt.Sprintf("key is empty"))
		return false
	}
	_, err := n.redisClient.Del(key).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("redis del %s fail!, err=%v", key, err))
		return false
	}
	return true
}

func (n *RedisService) Set(key string, value string) bool {
	if "" == key {
		n.l.Error(fmt.Sprintf("key is empty"))
		return false
	}
	_, err := n.redisClient.Set(key, value, -1).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("redis set [%s]:[%s] fail!, err=%v", key, value, err))
		return false
	}
	return true
}

func (n *RedisService) Get(key string) string {
	if "" == key {
		n.l.Error(fmt.Sprintf("key is empty"))
		return ""
	}
	result, err := n.redisClient.Get(key).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("redis get [%s] fail!, err=%v", key, err))
		return ""
	}
	return result
}

func (n *RedisService) Exits(key string) bool {
	code, err := n.redisClient.Exists(key).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("redis get [%s] fail!, err=%v", key, err))
		return false
	}
	return code == 1
}

func (n *RedisService) Keys(pattern string) ([]string, error) {
	// 使用SCAN命令迭代Redis数据库中的key
	var cursor uint64
	var keys []string
	for {
		var k []string
		var err error
		k, cursor, err = n.redisClient.Scan(cursor, pattern, 0).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func (n *RedisService) Incr(key string) bool {
	_, err := n.redisClient.Incr(key).Result()
	if err != nil {
		n.l.Error(fmt.Sprintf("redis Incr [%s] fail!, err=%v", key, err))
		return false
	}
	return true
}
