package server

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	// NewGRPCServer,
	NewHTTPServer,
	// NewCronServer,
	// NewConsumerServer,
)

var ConsumerProviderSet = wire.NewSet(
	// NewGRPCServer,
	// NewCronServer,
	NewHTTPConsumerServer,
	NewConsumerServer,
)

var JobProviderSet = wire.NewSet(
	NewHTTPJobServer,
	NewCronServer,
)

var StaticProviderSet = wire.NewSet(
	NewHTTPStaticServer,
)
