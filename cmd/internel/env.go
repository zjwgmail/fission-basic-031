package internel

import (
	"github.com/go-kratos/kratos/v2/log"
	"os"
	"time"

	"fission-basic/internal/conf"
)

const (
	NX_AK            = `NX_AK`
	NX_SK            = `NX_SK`
	REDIS_ADDRESS    = `REDIS_ADDRESS`
	REDIS_PASSWORD   = `REDIS_PASSWORD`
	DATA_SOURCE_LINK = `DATA_SOURCE_LINK`
)

// ReplaceEnv 直接在bc上操作，一般不要这么干
func ReplaceEnv(bc *conf.Bootstrap) error {
	log.Infof("replace env start")
	log.Infof("replace time %v", time.Now())
	log.Infof("replace local time %v", time.Now().Format("2006-01-02 15:04:05 -0700 MST"))
	nxAK := os.Getenv(NX_AK)
	if nxAK != "" {
		bc.Data.Nx.Ak = nxAK
	}

	nxSK := os.Getenv(NX_SK)
	if nxSK != "" {
		log.Infof("replace nx sk nxSk %v", nxSK)
		bc.Data.Nx.Sk = nxSK
	}

	redisAddress := os.Getenv(REDIS_ADDRESS)
	if redisAddress != "" {
		log.Infof("replace redis address redisAddress %v", redisAddress)
		bc.Data.Redis.Addr = redisAddress
	}

	redisPassword := os.Getenv(REDIS_PASSWORD)
	if redisPassword != "" {
		log.Infof("replace redis password redisPassword %v", redisPassword)
		bc.Data.Redis.Password = redisPassword
	}

	dataSourceLink := os.Getenv(DATA_SOURCE_LINK)
	if dataSourceLink != "" {
		log.Infof("replace data source link dataSourceLink %v", dataSourceLink)
		bc.Data.Database.Source = dataSourceLink
	}

	log.Infow("bootstrap", bc)
	return nil
}
