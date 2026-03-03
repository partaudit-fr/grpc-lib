package apperror

// PatternMapping maps an error message pattern to a structured error code.
type PatternMapping struct {
	Pattern    string
	Code       ErrorCode
	HTTPStatus int
}

// Config holds domain-specific error codes, messages, and pattern mappings.
// Each service provides its own Config via fx dependency injection.
type Config struct {
	// Messages maps domain-specific error codes to user-friendly messages.
	// These are merged with the generic default messages.
	Messages map[ErrorCode]string

	// Patterns maps known error message substrings to structured error codes.
	// These are checked before falling back to gRPC code-based mapping.
	Patterns []PatternMapping
}

// messageFor returns the user-friendly message for the given code,
// checking the config's custom messages first, then the generic defaults.
func (c *Config) messageFor(code ErrorCode) string {
	if c != nil && c.Messages != nil {
		if msg, ok := c.Messages[code]; ok {
			return msg
		}
	}
	if msg, ok := defaultMessages[code]; ok {
		return msg
	}
	return defaultMessages[INTERNAL_ERROR]
}

// isKnownCode returns true if the code exists in either the config's custom
// messages or the generic defaults.
func (c *Config) isKnownCode(code ErrorCode) bool {
	if c != nil && c.Messages != nil {
		if _, ok := c.Messages[code]; ok {
			return true
		}
	}
	_, ok := defaultMessages[code]
	return ok
}
