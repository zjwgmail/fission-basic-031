package queue

import (
	"testing"
	"time"

	"github.com/go-redis/redis"

	"fission-basic/internal/conf"
)

func TestSend(t *testing.T) {
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

	o := NewOfficialQueue(cli, &conf.Data{
		Queue: &conf.Data_Queue{
			// OfficialKey: "offical_msg",
			OfficialKey: "un_official_msg_zsj",
		},
	})

	// {WaID:85266831315 RallyCode:dj4pr NickName:一哥 Channel:a Language:02 SendTime:1735470750 Generation:02}
	// 85266831315|a0201djggk|a|02|1735470750|02|zhang xxxxxxx
	// d.WaID, d.RallyCode, d.Channel, d.Language, d.SendTime, d.Generation, d.NickName)

	// id,del,create_time,update_time,wa_id,rally_code,state,channel,language,generation,nickname,send_time
	// "1","2","2025-02-24 10:35:55","2025-02-24 10:35:55","85266831320","dj4a2","1","a","02","2","Name-0","1735470750"

	err := o.SendBack(&RallyData{
		WaID:       "85266831320",
		RallyCode:  "dj4a2",
		NickName:   "Name-0",
		Channel:    "a",
		Language:   "02",
		SendTime:   1735470750,
		Generation: "02",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log("ok")
}
