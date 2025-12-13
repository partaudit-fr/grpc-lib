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

func newGPRCServer(lifecycle fx.Lifecycle, logger *slog.Logger, config GRPCConfigServer, tmp propagation.TextMapPropagator) grpc.ServiceRegistrar {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port()))
	if err != nil {
		logger.Warn(err.Error())
	}

	keepaliveOptions := grpc.KeepaliveParams(keepalive.ServerParameters{
		Time:    time.Minute,
		Timeout: 3 * time.Second,
	})

	keepaliveEnforcementOptions := grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             30 * time.Second,
		PermitWithoutStream: true,
	})

	server := grpc.NewServer(
		keepaliveOptions,
		keepaliveEnforcementOptions,
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithPropagators(tmp),
			otelgrpc.WithFilter(grpcHealthCheckFilter),
		)),
		grpc.UnaryInterceptor(LoggingInterceptor()),
	)

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				reflection.Register(server)

				if err = server.Serve(listener); err != nil {
					logger.Warn(err.Error())
				}
			}()

			logger.Info(fmt.Sprintf("%s://%s", listener.Addr().Network(), listener.Addr().String()))

			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.GracefulStop()

			return nil
		},
	})

	return server
}
