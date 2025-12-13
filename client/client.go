package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/disco07/grpc-lib/marshal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"

	"github.com/disco07/grpc-lib/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/stats"
)

// healthCheckFilter returns a filter function that excludes health check endpoints from tracing
func healthCheckFilter(r *http.Request) bool {
	// Return false to exclude from tracing, true to include
	return !strings.HasPrefix(r.URL.Path, "/health")
}

// grpcHealthCheckFilter filters out health check methods from gRPC tracing
func grpcHealthCheckFilter(info *stats.RPCTagInfo) bool {
	// Return false to exclude from tracing, true to include
	if info.FullMethodName != "" && strings.Contains(info.FullMethodName, "HealthService") {
		return false
	}
	return true
}

func newGRPCClientConn(lc fx.Lifecycle, grpcServerConfig server.GRPCConfigServer, tmp propagation.TextMapPropagator) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		grpcServerConfig.Host(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithPropagators(tmp),
			otelgrpc.WithFilter(grpcHealthCheckFilter),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("could not connect to order service: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			fmt.Println("Closing gRPC connection")
			return conn.Close()
		},
	})

	return conn, nil
}

func newServeMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		marshal.WithMultipartFormMarshaler(),
	)
}

func startHTTPClient(lc fx.Lifecycle, mux *runtime.ServeMux, config GRPCConfigClient, tmp propagation.TextMapPropagator) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				fmt.Println("API gateway server is running on " + fmt.Sprintf(":%d", config.Port()))
				// Wrap the handler with OpenTelemetry instrumentation and CORS
				// Use WithFilter to exclude health check endpoints from tracing
				handler := otelhttp.NewHandler(
					withCORS(mux),
					config.ServiceName(),
					otelhttp.WithPropagators(tmp),
					otelhttp.WithFilter(healthCheckFilter),
				)
				if err := http.ListenAndServe(
					fmt.Sprintf(":%d", config.Port()),
					handler,
				); err != nil {
					log.Fatalf("gateway server closed abruptly: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Stopping HTTP server")
			return nil
		},
	})
}
