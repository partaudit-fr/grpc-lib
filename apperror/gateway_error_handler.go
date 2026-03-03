package apperror

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// ErrorResponse is the standardized JSON error format returned by the gateway.
type ErrorResponse struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Status  int               `json:"status"`
	Details []json.RawMessage `json:"details,omitempty"`
}

// NewCustomHTTPErrorHandler creates a gRPC-Gateway error handler that returns
// the standardized error format: {"code": "...", "message": "...", "status": ...}
//
// The Config provides domain-specific error codes and messages.
func NewCustomHTTPErrorHandler(cfg *Config) runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		st, _ := status.FromError(err)

		var appCode ErrorCode
		var message string

		// Try to extract the structured code from the interceptor's "CODE:message" format
		if st != nil {
			parts := strings.SplitN(st.Message(), ":", 2)
			if len(parts) == 2 {
				candidate := ErrorCode(parts[0])
				if cfg.isKnownCode(candidate) {
					appCode = candidate
					message = parts[1]
				}
			}
		}

		// Fall back to MapGRPCError if we couldn't extract a structured code
		if appCode == "" {
			appErr := MapGRPCError(err, cfg)
			appCode = appErr.Code
			message = appErr.Message
		}

		httpStatus := 500
		if s, ok := GRPCCodeToHTTPStatus[st.Code()]; ok {
			httpStatus = s
		}

		var details []json.RawMessage
		if st != nil {
			for _, d := range st.Details() {
				if raw, marshalErr := json.Marshal(d); marshalErr == nil {
					details = append(details, raw)
				}
			}
		}

		resp := ErrorResponse{
			Code:    appCode,
			Message: message,
			Status:  httpStatus,
			Details: details,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		json.NewEncoder(w).Encode(resp)
	}
}
