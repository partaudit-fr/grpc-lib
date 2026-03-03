package apperror

import "fmt"

// AppError is a structured application error with a code, user-friendly message,
// HTTP status, and internal details (never exposed to the client).
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"status"`
	Internal   string    `json:"-"` // never serialized to the client, used for logging
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates an AppError with the default message for the given code.
func New(code ErrorCode, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    defaultMessageFor(code),
		HTTPStatus: httpStatus,
	}
}

// NewWithMessage creates an AppError with a custom user-facing message.
func NewWithMessage(code ErrorCode, httpStatus int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap creates an AppError that wraps an internal error for logging.
func Wrap(code ErrorCode, httpStatus int, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    defaultMessageFor(code),
		HTTPStatus: httpStatus,
		Internal:   internal.Error(),
	}
}

// defaultMessageFor returns the message from the generic defaults only.
// For config-aware lookups, use Config.messageFor().
func defaultMessageFor(code ErrorCode) string {
	if msg, ok := defaultMessages[code]; ok {
		return msg
	}
	return defaultMessages[INTERNAL_ERROR]
}
