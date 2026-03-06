package server

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs the details of each gRPC request via slog.
// Logs are structured and exported to any configured slog handler (stdout, SigNoz/OTEL, etc.).
func LoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	logger = logger.WithGroup("grpc")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if info.FullMethod == "/health.HealthService/Check" {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		code := status.Code(err)

		attrs := []slog.Attr{
			slog.String("method", info.FullMethod),
			slog.String("status", code.String()),
			slog.Duration("duration", duration),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		var level slog.Level
		switch {
		case code == codes.OK:
			level = slog.LevelInfo
		case code == codes.Internal || code == codes.Unavailable || code == codes.DataLoss:
			level = slog.LevelError
		default:
			level = slog.LevelWarn
		}

		logger.LogAttrs(ctx, level, "request", attrs...)

		return resp, err
	}
}
