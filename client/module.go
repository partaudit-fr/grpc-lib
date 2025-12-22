package client

import (
	"github.com/partaudit-fr/grpc-lib/protogen/gateway/go/health"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		newGRPCClientConn,
		newServeMux,
	),
	fx.Invoke(
		startHTTPClient,
		health.RegisterHealthServiceHandler,
	),
)
