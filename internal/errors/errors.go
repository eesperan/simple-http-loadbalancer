package errors

import (
	"errors"
	"fmt"
	"time"
)

// ErrorCode represents specific error types
type ErrorCode string

const (
	ErrBackendUnavailable ErrorCode = "BACKEND_UNAVAILABLE"
	ErrConfigInvalid      ErrorCode = "CONFIG_INVALID"
	ErrRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCircuitOpen        ErrorCode = "CIRCUIT_OPEN"
	ErrTimeout            ErrorCode = "TIMEOUT"
	ErrSSLCertificate     ErrorCode = "SSL_CERTIFICATE_ERROR"
)

// LoadBalancerError represents a custom error with context
type LoadBalancerError struct {
	Code      ErrorCode
	Message   string
	Timestamp time.Time
	Err       error
}

func (e *LoadBalancerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v (at %s)", e.Code, e.Message, e.Err, e.Timestamp.Format(time.RFC3339))
	}
	return fmt.Sprintf("[%s] %s (at %s)", e.Code, e.Message, e.Timestamp.Format(time.RFC3339))
}

// New creates a new LoadBalancerError
func New(code ErrorCode, message string, err error) *LoadBalancerError {
	return &LoadBalancerError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

// Is implements error matching for LoadBalancerError
func (e *LoadBalancerError) Is(target error) bool {
	t, ok := target.(*LoadBalancerError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Unwrap returns the underlying error
func (e *LoadBalancerError) Unwrap() error {
	return e.Err
}

// As implements error type assertion for LoadBalancerError
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is provides error comparison functionality
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Wrap wraps an error with additional context
func Wrap(err error, code ErrorCode, message string) *LoadBalancerError {
	return &LoadBalancerError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Err:       err,
	}
}

// GetCode extracts the error code from an error if it's a LoadBalancerError
func GetCode(err error) ErrorCode {
	var lbErr *LoadBalancerError
	if As(err, &lbErr) {
		return lbErr.Code
	}
	return ""
}

// GetMessage extracts the message from an error if it's a LoadBalancerError
func GetMessage(err error) string {
	var lbErr *LoadBalancerError
	if As(err, &lbErr) {
		return lbErr.Message
	}
	return err.Error()
}

// GetTimestamp extracts the timestamp from an error if it's a LoadBalancerError
func GetTimestamp(err error) time.Time {
	var lbErr *LoadBalancerError
	if As(err, &lbErr) {
		return lbErr.Timestamp
	}
	return time.Time{}
}
