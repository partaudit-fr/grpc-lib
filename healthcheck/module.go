package healthcheck

import "go.uber.org/fx"

type healthCheckParams struct {
	fx.In
	Info *Info `optional:"true"`
}

var Module = fx.Options(
	fx.Provide(
		newHealthCheck,
	),
)
