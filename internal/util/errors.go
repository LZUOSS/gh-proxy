package util

import "fmt"

// Common error types for the GitHub reverse proxy

var (
	// ErrInvalidConfig indicates a configuration error
	ErrInvalidConfig = &ProxyError{
		Code:    "INVALID_CONFIG",
		Message: "invalid configuration",
	}

	// ErrRateLimited indicates the client has exceeded rate limits
	ErrRateLimited = &ProxyError{
		Code:    "RATE_LIMITED",
		Message: "rate limit exceeded",
	}

	// ErrUnauthorized indicates authentication failure
	ErrUnauthorized = &ProxyError{
		Code:    "UNAUTHORIZED",
		Message: "unauthorized access",
	}

	// ErrProxyUnavailable indicates the proxy service is unavailable
	ErrProxyUnavailable = &ProxyError{
		Code:    "PROXY_UNAVAILABLE",
		Message: "proxy service unavailable",
	}

	// ErrUpstreamTimeout indicates upstream server timeout
	ErrUpstreamTimeout = &ProxyError{
		Code:    "UPSTREAM_TIMEOUT",
		Message: "upstream server timeout",
	}

	// ErrInvalidRequest indicates an invalid client request
	ErrInvalidRequest = &ProxyError{
		Code:    "INVALID_REQUEST",
		Message: "invalid request",
	}

	// ErrSSRFDetected indicates a potential SSRF attack
	ErrSSRFDetected = &ProxyError{
		Code:    "SSRF_DETECTED",
		Message: "potential SSRF attack detected",
	}

	// ErrCacheMiss indicates the requested resource is not in cache
	ErrCacheMiss = &ProxyError{
		Code:    "CACHE_MISS",
		Message: "cache miss",
	}

	// ErrCacheWriteFailed indicates a cache write operation failed
	ErrCacheWriteFailed = &ProxyError{
		Code:    "CACHE_WRITE_FAILED",
		Message: "failed to write to cache",
	}
)

// ProxyError represents an error with a code and message
type ProxyError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *ProxyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements error unwrapping
func (e *ProxyError) Unwrap() error {
	return e.Err
}

// WithError returns a new ProxyError with the given underlying error
func (e *ProxyError) WithError(err error) *ProxyError {
	return &ProxyError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

// WithMessage returns a new ProxyError with a custom message
func (e *ProxyError) WithMessage(msg string) *ProxyError {
	return &ProxyError{
		Code:    e.Code,
		Message: msg,
		Err:     e.Err,
	}
}

// NewProxyError creates a new ProxyError
func NewProxyError(code, message string) *ProxyError {
	return &ProxyError{
		Code:    code,
		Message: message,
	}
}
