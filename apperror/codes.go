package apperror

// ErrorCode represents an application error code.
type ErrorCode string

// Generic error codes shared across all services.
const (
	INTERNAL_ERROR    ErrorCode = "INTERNAL_ERROR"
	UNAUTHORIZED      ErrorCode = "UNAUTHORIZED"
	FORBIDDEN         ErrorCode = "FORBIDDEN"
	NOT_FOUND         ErrorCode = "NOT_FOUND"
	VALIDATION_FAILED ErrorCode = "VALIDATION_FAILED"
	CONFLICT          ErrorCode = "CONFLICT"
)
