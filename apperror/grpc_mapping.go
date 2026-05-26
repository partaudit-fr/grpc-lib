package apperror

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCCodeToHTTPStatus maps gRPC codes to HTTP status codes.
var GRPCCodeToHTTPStatus = map[codes.Code]int{
	codes.OK:                 200,
	codes.Canceled:           499,
	codes.InvalidArgument:    400,
	codes.NotFound:           404,
	codes.AlreadyExists:      409,
	codes.PermissionDenied:   403,
	codes.Unauthenticated:    401,
	codes.ResourceExhausted:  429,
	codes.FailedPrecondition: 400,
	codes.Aborted:            409,
	codes.OutOfRange:         400,
	codes.Unimplemented:      501,
	codes.Internal:           500,
	codes.Unavailable:        503,
	codes.DataLoss:           500,
	codes.DeadlineExceeded:   504,
	codes.Unknown:            500,
}

// grpcCodeToAppCode maps gRPC codes to generic AppError codes (used when no pattern matches).
var grpcCodeToAppCode = map[codes.Code]ErrorCode{
	codes.InvalidArgument:    VALIDATION_FAILED,
	codes.NotFound:           NOT_FOUND,
	codes.AlreadyExists:      CONFLICT,
	codes.PermissionDenied:   FORBIDDEN,
	codes.Unauthenticated:    UNAUTHORIZED,
	codes.FailedPrecondition: VALIDATION_FAILED,
	codes.Aborted:            CONFLICT,
}

// MapGRPCError transforms a gRPC error into an AppError using the provided config
// for domain-specific pattern matching and messages.
// It first tries to match the config's patterns, then falls back to generic gRPC code mapping.
func MapGRPCError(err error, cfg *Config) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		// If the config has a domain-specific message for this code, use it
		// instead of the generic default (e.g. INVALID_CREDENTIALS → "Les identifiants...")
		if cfg != nil {
			if msg := cfg.messageFor(appErr.Code); msg != "" {
				appErr.Message = msg
			}
		}
		return appErr
	}

	st, ok := status.FromError(err)
	if !ok {
		return New(INTERNAL_ERROR, 500)
	}

	msg := strings.ToLower(st.Message())

	// Try domain-specific pattern matching from config
	if cfg != nil {
		for _, pm := range cfg.Patterns {
			if strings.Contains(msg, pm.Pattern) {
				return &AppError{
					Code:       pm.Code,
					Message:    cfg.messageFor(pm.Code),
					HTTPStatus: pm.HTTPStatus,
					Internal:   st.Message(),
				}
			}
		}
	}

	// Fall back to generic gRPC code mapping
	httpStatus := 500
	if s, ok := GRPCCodeToHTTPStatus[st.Code()]; ok {
		httpStatus = s
	}

	appCode := INTERNAL_ERROR
	if c, ok := grpcCodeToAppCode[st.Code()]; ok {
		appCode = c
	}

	return &AppError{
		Code:       appCode,
		Message:    cfg.messageFor(appCode),
		HTTPStatus: httpStatus,
		Internal:   st.Message(),
	}
}
