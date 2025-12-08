package client

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/disco07/grpc-lib/marshal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"

	"github.com/disco07/grpc-lib/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newGRPCClientConn(lc fx.Lifecycle, grpcServerConfig server.GRPCConfigServer, tmp propagation.TextMapPropagator) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		grpcServerConfig.Host(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithPropagators(tmp))),
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
				handler := otelhttp.NewHandler(
					withCORS(mux),
					config.ServiceName(),
					otelhttp.WithPropagators(tmp),
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
