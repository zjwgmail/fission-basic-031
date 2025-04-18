package biz

import (
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGreeterUsecase,
	NewStudentUsecase,
	NewNxCLoudUsecase,
	NewHelpCodeUsecase,
	NewActivityInfoUsecase,
	NewUserInfoUsecase,
	NewMsgUsecase,
	NewWaMsgService,
	NewImageGenerate,
	NewCdkUsecase,
	NewSystemConfigUsecase,
	NewInit,
	NewActivityJob,
	NewResendJob,
	NewResendRetryJob,
	NewFeishuReportJob,
	NewEmailReportJob,
	NewDrainageJob,
)

var ConsumerProviderSet = wire.NewSet(
	NewNxCLoudUsecase,
	NewOfficialRallyUsecase,
	NewUnOfficialRallyUsecase,
	NewMsgUsecase,
	NewHelpCodeUsecase,
	NewActivityInfoUsecase,
	NewWaMsgService,
	NewImageGenerate,
	NewCdkUsecase,
	NewSystemConfigUsecase,
	NewActivityJob,
	NewResendJob,
	NewResendRetryJob,
	NewFeishuReportJob,
	NewEmailReportJob,
	NewDrainageJob,
	NewInit,
)

var JobProviderSet = wire.NewSet(
	NewNxCLoudUsecase,
	NewOfficialRallyUsecase,
	NewUnOfficialRallyUsecase,
	NewUserRemindUsecase,
	NewMsgUsecase,
	NewHelpCodeUsecase,
	NewActivityInfoUsecase,
	NewWaMsgService,
	NewImageGenerate,
	NewCdkUsecase,
	NewSystemConfigUsecase,
	NewActivityJob,
	NewResendJob,
	NewResendRetryJob,
	NewFeishuReportJob,
	NewEmailReportJob,
	NewInit,
	NewDrainageJob,
)
