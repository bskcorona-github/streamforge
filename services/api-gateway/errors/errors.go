package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// ErrorCode represents a specific error code
type ErrorCode string

const (
	// Authentication errors
	ErrorCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrorCodeInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	ErrorCodeTokenExpired        ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeTokenInvalid        ErrorCode = "TOKEN_INVALID"
	ErrorCodeInsufficientPerms   ErrorCode = "INSUFFICIENT_PERMISSIONS"

	// Validation errors
	ErrorCodeValidationFailed    ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingRequired     ErrorCode = "MISSING_REQUIRED_FIELD"
	ErrorCodeInvalidFormat       ErrorCode = "INVALID_FORMAT"

	// Resource errors
	ErrorCodeNotFound            ErrorCode = "RESOURCE_NOT_FOUND"
	ErrorCodeAlreadyExists       ErrorCode = "RESOURCE_ALREADY_EXISTS"
	ErrorCodeConflict            ErrorCode = "RESOURCE_CONFLICT"
	ErrorCodeDeleted             ErrorCode = "RESOURCE_DELETED"

	// Database errors
	ErrorCodeDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrorCodeConnectionFailed    ErrorCode = "DATABASE_CONNECTION_FAILED"
	ErrorCodeTransactionFailed   ErrorCode = "DATABASE_TRANSACTION_FAILED"
	ErrorCodeConstraintViolation ErrorCode = "DATABASE_CONSTRAINT_VIOLATION"

	// Service errors
	ErrorCodeServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeInternalError       ErrorCode = "INTERNAL_ERROR"
	ErrorCodeTimeout             ErrorCode = "TIMEOUT"
	ErrorCodeRateLimitExceeded   ErrorCode = "RATE_LIMIT_EXCEEDED"

	// External service errors
	ErrorCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeExternalTimeout      ErrorCode = "EXTERNAL_SERVICE_TIMEOUT"
	ErrorCodeExternalUnavailable  ErrorCode = "EXTERNAL_SERVICE_UNAVAILABLE"
)

// Error represents a structured error
type Error struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	Stack      []StackFrame           `json:"stack,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Cause      error                  `json:"-"`
}

// StackFrame represents a stack frame
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error   *Error `json:"error"`
	Success bool   `json:"success"`
}

// New creates a new Error
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
		HTTPStatus: getHTTPStatus(code),
		Stack:     getStackTrace(),
	}
}

// NewWithDetails creates a new Error with details
func NewWithDetails(code ErrorCode, message, details string) *Error {
	err := New(code, message)
	err.Details = details
	return err
}

// NewWithContext creates a new Error with context
func NewWithContext(code ErrorCode, message string, context map[string]interface{}) *Error {
	err := New(code, message)
	err.Context = context
	return err
}

// Wrap wraps an existing error
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if e, ok := err.(*Error); ok {
		appErr = e
	} else {
		appErr = New(code, message)
		appErr.Cause = err
	}

	appErr.Stack = getStackTrace()
	return appErr
}

// WrapWithContext wraps an existing error with context
func WrapWithContext(err error, code ErrorCode, message string, context map[string]interface{}) *Error {
	appErr := Wrap(err, code, message)
	appErr.Context = context
	return appErr
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the cause error
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithRequestID adds a request ID to the error
func (e *Error) WithRequestID(requestID string) *Error {
	e.RequestID = requestID
	return e
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// getHTTPStatus returns the HTTP status code for an error code
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrorCodeUnauthorized, ErrorCodeInvalidCredentials, ErrorCodeTokenExpired, ErrorCodeTokenInvalid:
		return http.StatusUnauthorized
	case ErrorCodeInsufficientPerms:
		return http.StatusForbidden
	case ErrorCodeValidationFailed, ErrorCodeInvalidInput, ErrorCodeMissingRequired, ErrorCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrorCodeNotFound, ErrorCodeDeleted:
		return http.StatusNotFound
	case ErrorCodeAlreadyExists, ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeDatabaseError, ErrorCodeConnectionFailed, ErrorCodeTransactionFailed, ErrorCodeConstraintViolation:
		return http.StatusInternalServerError
	case ErrorCodeServiceUnavailable, ErrorCodeInternalError:
		return http.StatusInternalServerError
	case ErrorCodeTimeout, ErrorCodeExternalTimeout:
		return http.StatusRequestTimeout
	case ErrorCodeRateLimitExceeded:
		return http.StatusTooManyRequests
	case ErrorCodeExternalServiceError, ErrorCodeExternalUnavailable:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// getStackTrace returns the current stack trace
func getStackTrace() []StackFrame {
	var frames []StackFrame
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}
	return frames
}

// Common error constructors
func Unauthorized(message string) *Error {
	return New(ErrorCodeUnauthorized, message)
}

func InvalidCredentials() *Error {
	return New(ErrorCodeInvalidCredentials, "Invalid username or password")
}

func TokenExpired() *Error {
	return New(ErrorCodeTokenExpired, "Token has expired")
}

func TokenInvalid() *Error {
	return New(ErrorCodeTokenInvalid, "Invalid token")
}

func InsufficientPermissions(resource string) *Error {
	return NewWithContext(ErrorCodeInsufficientPerms, "Insufficient permissions", map[string]interface{}{
		"resource": resource,
	})
}

func ValidationFailed(field, reason string) *Error {
	return NewWithContext(ErrorCodeValidationFailed, "Validation failed", map[string]interface{}{
		"field":  field,
		"reason": reason,
	})
}

func InvalidInput(field, reason string) *Error {
	return NewWithContext(ErrorCodeInvalidInput, "Invalid input", map[string]interface{}{
		"field":  field,
		"reason": reason,
	})
}

func MissingRequiredField(field string) *Error {
	return NewWithContext(ErrorCodeMissingRequired, "Missing required field", map[string]interface{}{
		"field": field,
	})
}

func NotFound(resource, id string) *Error {
	return NewWithContext(ErrorCodeNotFound, "Resource not found", map[string]interface{}{
		"resource": resource,
		"id":       id,
	})
}

func AlreadyExists(resource, identifier string) *Error {
	return NewWithContext(ErrorCodeAlreadyExists, "Resource already exists", map[string]interface{}{
		"resource":   resource,
		"identifier": identifier,
	})
}

func DatabaseError(operation string, err error) *Error {
	return WrapWithContext(err, ErrorCodeDatabaseError, "Database operation failed", map[string]interface{}{
		"operation": operation,
	})
}

func ServiceUnavailable(service string) *Error {
	return NewWithContext(ErrorCodeServiceUnavailable, "Service unavailable", map[string]interface{}{
		"service": service,
	})
}

func InternalError(message string) *Error {
	return New(ErrorCodeInternalError, message)
}

func Timeout(operation string, duration time.Duration) *Error {
	return NewWithContext(ErrorCodeTimeout, "Operation timed out", map[string]interface{}{
		"operation": operation,
		"duration":  duration.String(),
	})
}

func RateLimitExceeded(limit int, window time.Duration) *Error {
	return NewWithContext(ErrorCodeRateLimitExceeded, "Rate limit exceeded", map[string]interface{}{
		"limit":  limit,
		"window": window.String(),
	})
}

func ExternalServiceError(service, operation string, err error) *Error {
	return WrapWithContext(err, ErrorCodeExternalServiceError, "External service error", map[string]interface{}{
		"service":   service,
		"operation": operation,
	})
} 