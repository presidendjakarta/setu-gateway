package types

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ErrCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeRateLimited      ErrorCode = "RATE_LIMITED"
	ErrCodeCircuitOpen      ErrorCode = "CIRCUIT_OPEN"
	ErrCodeTimeout          ErrorCode = "TIMEOUT"
	ErrCodeBadGateway       ErrorCode = "BAD_GATEWAY"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeRouteNotFound    ErrorCode = "ROUTE_NOT_FOUND"
	ErrCodeAuthFailed       ErrorCode = "AUTH_FAILED"
	ErrCodePluginError      ErrorCode = "PLUGIN_ERROR"
)

// GatewayError represents a structured gateway error
type GatewayError struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	Details    string         `json:"details,omitempty"`
	StatusCode int            `json:"status_code"`
	Err        error          `json:"-"`
}

// Error implements error interface
func (e *GatewayError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap implements errors.Unwrap for error chaining
func (e *GatewayError) Unwrap() error {
	return e.Err
}

// NewGatewayError creates a new GatewayError
func NewGatewayError(code ErrorCode, message string, statusCode int) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewGatewayErrorWithDetails creates a new GatewayError with details
func NewGatewayErrorWithDetails(code ErrorCode, message, details string, statusCode int) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// WrapError wraps an existing error with gateway error context
func WrapError(code ErrorCode, message string, statusCode int, err error) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Predefined errors
var (
	ErrRouteNotFound = &GatewayError{
		Code:       ErrCodeRouteNotFound,
		Message:    "No route matched the request",
		StatusCode: http.StatusNotFound,
	}

	ErrUnauthorized = &GatewayError{
		Code:       ErrCodeUnauthorized,
		Message:    "Authentication required",
		StatusCode: http.StatusUnauthorized,
	}

	ErrForbidden = &GatewayError{
		Code:       ErrCodeForbidden,
		Message:    "Access denied",
		StatusCode: http.StatusForbidden,
	}

	ErrRateLimited = &GatewayError{
		Code:       ErrCodeRateLimited,
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
	}

	ErrCircuitOpen = &GatewayError{
		Code:       ErrCodeCircuitOpen,
		Message:    "Service temporarily unavailable",
		StatusCode: http.StatusServiceUnavailable,
	}

	ErrTimeout = &GatewayError{
		Code:       ErrCodeTimeout,
		Message:    "Request timeout",
		StatusCode: http.StatusGatewayTimeout,
	}

	ErrBadGateway = &GatewayError{
		Code:       ErrCodeBadGateway,
		Message:    "Bad gateway",
		StatusCode: http.StatusBadGateway,
	}

	ErrServiceUnavailable = &GatewayError{
		Code:       ErrCodeServiceUnavailable,
		Message:    "Service unavailable",
		StatusCode: http.StatusServiceUnavailable,
	}
)

// IsGatewayError checks if an error is a GatewayError
func IsGatewayError(err error) bool {
	var gwErr *GatewayError
	return errors.As(err, &gwErr)
}

// GetGatewayError extracts GatewayError from error chain
func GetGatewayError(err error) (*GatewayError, bool) {
	var gwErr *GatewayError
	if errors.As(err, &gwErr) {
		return gwErr, true
	}
	return nil, false
}

// ToHTTPStatusCode converts GatewayError to HTTP status code
func ToHTTPStatusCode(err error) int {
	if gwErr, ok := GetGatewayError(err); ok {
		return gwErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code from an error
func GetErrorCode(err error) ErrorCode {
	if gwErr, ok := GetGatewayError(err); ok {
		return gwErr.Code
	}
	return ErrCodeInternal
}
