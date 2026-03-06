package apperror

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// shortMethod extracts the method name from a full gRPC method path.
// e.g. "/form_templates.v1.FormTemplatesService/ListFormTemplates" → "ListFormTemplates"
func shortMethod(fullMethod string) string {
	if i := strings.LastIndex(fullMethod, "/"); i >= 0 {
		return fullMethod[i+1:]
	}
	return fullMethod
}

// ErrorSanitizerInterceptor returns a gRPC UnaryServerInterceptor that transforms
// raw error messages into user-friendly AppErrors. Internal details are logged
// but never sent to the client.
//
// The Config provides domain-specific error codes, messages, and pattern mappings.
func ErrorSanitizerInterceptor(logger *slog.Logger, cfg *Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		appErr := MapGRPCError(err, cfg)
		grpcCode := httpStatusToGRPCCode(appErr.HTTPStatus)

		attrs := []slog.Attr{
			slog.String("method", info.FullMethod),
			slog.String("code", string(appErr.Code)),
			slog.String("grpc.status", grpcCode.String()),
			slog.Int("http.status", appErr.HTTPStatus),
		}

		short := shortMethod(info.FullMethod)

		if appErr.Internal != "" {
			// Server-side error (5xx) — always ERROR level
			attrs = append(attrs, slog.String("message", appErr.Message))
			logger.LogAttrs(ctx, slog.LevelError,
				fmt.Sprintf("[%s] %s", short, appErr.Internal),
				attrs...,
			)
		} else {
			// Client-side error (4xx) — WARN level
			logger.LogAttrs(ctx, slog.LevelWarn,
				fmt.Sprintf("[%s] %s", short, appErr.Message),
				attrs...,
			)
		}

		// Encode "CODE:message" so the gateway error handler can extract the structured code
		return resp, status.Errorf(grpcCode, "%s:%s", appErr.Code, appErr.Message)
	}
}

func httpStatusToGRPCCode(httpStatus int) codes.Code {
	switch httpStatus {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	case 413:
		return codes.ResourceExhausted
	case 429:
		return codes.ResourceExhausted
	case 501:
		return codes.Unimplemented
	case 503:
		return codes.Unavailable
	case 504:
		return codes.DeadlineExceeded
	default:
		return codes.Internal
	}
}
