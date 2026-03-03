package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
)

// grpcHealthCheckFilter filters out health check methods from tracing
func grpcHealthCheckFilter(info *stats.RPCTagInfo) bool {
	// Return false to exclude from tracing, true to include
	if info.FullMethodName != "" && strings.Contains(info.FullMethodName, "HealthService") {
		return false
	}
	return true
}

// serverParams holds dependencies for the gRPC server, injected via fx.
type serverParams struct {
	fx.In

	Lifecycle    fx.Lifecycle
	Logger       *slog.Logger
	Config       GRPCConfigServer
	Propagator   propagation.TextMapPropagator
	Interceptors []grpc.UnaryServerInterceptor `group:"grpc_interceptors"`
}

func newGPRCServer(params serverParams) grpc.ServiceRegistrar {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", params.Config.Port()))
	if err != nil {
		params.Logger.Warn(err.Error())
	}

	keepaliveOptions := grpc.KeepaliveParams(keepalive.ServerParameters{
		Time:    time.Minute,
		Timeout: 3 * time.Second,
	})

	keepaliveEnforcementOptions := grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             30 * time.Second,
		PermitWithoutStream: true,
	})

	// Build interceptor chain: LoggingInterceptor first, then any custom interceptors
	interceptors := []grpc.UnaryServerInterceptor{LoggingInterceptor()}
	interceptors = append(interceptors, params.Interceptors...)

	server := grpc.NewServer(
		keepaliveOptions,
		keepaliveEnforcementOptions,
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithPropagators(params.Propagator),
			otelgrpc.WithFilter(grpcHealthCheckFilter),
		)),
		grpc.ChainUnaryInterceptor(interceptors...),
	)

	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				reflection.Register(server)

				if err = server.Serve(listener); err != nil {
					params.Logger.Warn(err.Error())
				}
			}()

			params.Logger.Info(fmt.Sprintf("%s://%s", listener.Addr().Network(), listener.Addr().String()))

			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.GracefulStop()

			return nil
		},
	})

	return server
}
