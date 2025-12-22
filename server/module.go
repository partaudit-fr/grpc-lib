package server

import (
	"github.com/partaudit-fr/grpc-lib/healthcheck"
	"github.com/partaudit-fr/grpc-lib/protogen/go/health"
	"go.uber.org/fx"
)

var Module = fx.Options(
	healthcheck.Module,
	fx.Provide(
		newGPRCServer,
	),
	fx.Invoke(
		health.RegisterHealthServiceServer,
	),
)
