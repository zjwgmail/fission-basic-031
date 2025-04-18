package data

import (
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"

	"fission-basic/internal/conf"
	"fission-basic/kit/sqlx"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewGreeterRepo,
	NewUserRemind,
	NewStudent,
	NewNXCloud,
	NewHelpCode,
	NewUserInfo,
	NewMsg,
	NewWaMsgSend,
	NewWaMsgRetry,
	NewSystemConfig,
	NewInitDB,
	NewActivityInfo,
	NewFeishuReport,
	NewUserJoinGroup,
	NewEmailReport,
	NewRally,
	NewOfficialRally,
	NewUnOfficialRally,
	NewUploadUserInfo,
	NewWaUserScore,
	NewPushEventSendMessage,
	NewWaMsgReceived,
	NewPushEvent4User,
	NewCountLimit,
)

var ConsumerProviderSet = wire.NewSet(
	NewData,
	NewNXCloud,
	NewUserRemind,
	NewRally,
	NewOfficialRally,
	NewUnOfficialRally,
	NewMsg,
	NewHelpCode,
	NewWaMsgSend,
	NewWaMsgRetry,
	NewUserInfo,
	NewSystemConfig,
	NewInitDB,
	NewActivityInfo,
	NewFeishuReport,
	NewUserJoinGroup,
	NewEmailReport,
	NewUploadUserInfo,
	NewWaUserScore,
	NewPushEventSendMessage,
	NewWaMsgReceived,
	NewPushEvent4User,
	NewCountLimit,
)

var JobProviderSet = wire.NewSet(
	NewData,
	NewNXCloud,
	NewRally,
	NewOfficialRally,
	NewUnOfficialRally,
	NewUserRemind,
	NewMsg,
	NewHelpCode,
	NewWaMsgSend,
	NewWaMsgRetry,
	NewSystemConfig,
	NewUserInfo,
	NewInitDB,
	NewActivityInfo,
	NewFeishuReport,
	NewUserJoinGroup,
	NewEmailReport,
	NewUploadUserInfo,
	NewWaUserScore,
	NewPushEventSendMessage,
	NewWaMsgReceived,
	NewPushEvent4User,
	NewCountLimit,
)

type Data struct {
	db sqlx.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}

	//todo zsj 优化参数
	db, err := sqlx.Open(&sqlx.Config{
		DriverName: c.Database.Driver,  // mysql
		Server:     c.Database.Source,  // root:Exa998StgPass1!@tcp(rm-2zewh9h752042ge821o.mysql.rds.aliyuncs.com:3306)/wa-fission-v3.1?parseTime=True
		MaxOpen:    c.Database.MaxOpen, // 最大连接数
		MaxIdle:    c.Database.MaxIdle, // 最大空闲数
	})
	if err != nil {
		panic(err)
	}

	return &Data{db: db}, cleanup, nil
}
