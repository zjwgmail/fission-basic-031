package service

import (
	"fission-basic/internal/pkg/redis"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewGreeterService,
	NewStudentService,
	NewNxCloudService,
	NewHelpCodeService,
	redis.NewRedisService,
	NewCDKService,
	NewInitService,
	NewImageService,
	NewUploadService,
)

var JobProviderSet = wire.NewSet(
	NewTaskService,
	NewCronService,
	NewUserRemindService,
	NewHelpCodeService,
	NewRetryService,
	NewInitService,
	NewNxCloudService,
	NewUploadService,
)

var ConsumerProviderSet = wire.NewSet(
	NewTaskService,
	NewHelpCodeService,
	NewInitService,
	NewNxCloudService,
)
