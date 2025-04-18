package queue

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOfficialQueue,
	NewUnOfficialQueue,
	NewRenewMsg,
	NewCallMsg,
	NewRepeatHelp,
	NewGW,
	NewGWRecal,
	NewGWUnknown,
)
