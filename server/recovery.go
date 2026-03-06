package server

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor catches panics in gRPC handlers, logs them with full stack trace,
// and returns an Internal error instead of crashing the server.
func RecoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.ErrorContext(ctx, fmt.Sprintf("[PANIC] %v", r),
					slog.String("method", info.FullMethod),
					slog.String("stack", stack),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
