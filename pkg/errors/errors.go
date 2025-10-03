package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents different types of errors that can occur
type ErrorCode string

const (
	// General errors
	ErrorCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrorCodeInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrorCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrorCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrorCodeRateLimited    ErrorCode = "RATE_LIMITED"
	ErrorCodeTimeout        ErrorCode = "TIMEOUT"

	// Provider errors
	ErrorCodeProviderNotFound       ErrorCode = "PROVIDER_NOT_FOUND"
	ErrorCodeProviderUnavailable    ErrorCode = "PROVIDER_UNAVAILABLE"
	ErrorCodeProviderConfiguration  ErrorCode = "PROVIDER_CONFIG_ERROR"
	ErrorCodeProviderAuthentication ErrorCode = "PROVIDER_AUTH_ERROR"

	// Notification errors
	ErrorCodeInvalidRecipient    ErrorCode = "INVALID_RECIPIENT"
	ErrorCodeInvalidNotification ErrorCode = "INVALID_NOTIFICATION"
	ErrorCodeNotificationFailed  ErrorCode = "NOTIFICATION_FAILED"
	ErrorCodeDeliveryFailed      ErrorCode = "DELIVERY_FAILED"
	ErrorCodeTemplateNotFound    ErrorCode = "TEMPLATE_NOT_FOUND"

	// Validation errors
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidEmail     ErrorCode = "INVALID_EMAIL"
	ErrorCodeInvalidPhone     ErrorCode = "INVALID_PHONE"
	ErrorCodeInvalidToken     ErrorCode = "INVALID_TOKEN"

	// Queue errors
	ErrorCodeQueueFull    ErrorCode = "QUEUE_FULL"
	ErrorCodeQueueEmpty   ErrorCode = "QUEUE_EMPTY"
	ErrorCodeQueueTimeout ErrorCode = "QUEUE_TIMEOUT"
)

// NotificationError represents a notification service error
type NotificationError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	StatusCode int               `json:"status_code"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Cause      error             `json:"-"`
}

// Error implements the error interface
func (e *NotificationError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements error unwrapping for error chains
func (e *NotificationError) Unwrap() error {
	return e.Cause
}

// WithMetadata adds metadata to the error
func (e *NotificationError) WithMetadata(key, value string) *NotificationError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// WithCause adds a causing error
func (e *NotificationError) WithCause(cause error) *NotificationError {
	e.Cause = cause
	return e
}

// NewNotificationError creates a new notification error
func NewNotificationError(code ErrorCode, message string) *NotificationError {
	return &NotificationError{
		Code:       code,
		Message:    message,
		StatusCode: getHTTPStatusCode(code),
		Metadata:   make(map[string]string),
	}
}

// NewNotificationErrorWithDetails creates a new notification error with details
func NewNotificationErrorWithDetails(code ErrorCode, message, details string) *NotificationError {
	return &NotificationError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: getHTTPStatusCode(code),
		Metadata:   make(map[string]string),
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *NotificationError {
	return &NotificationError{
		Code:       ErrorCodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Metadata:   make(map[string]string),
		Cause:      cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *NotificationError {
	err := &NotificationError{
		Code:       ErrorCodeValidationFailed,
		Message:    fmt.Sprintf("Validation failed for field '%s': %s", field, message),
		StatusCode: http.StatusBadRequest,
		Metadata:   make(map[string]string),
	}
	err.WithMetadata("field", field)
	return err
}

// NewProviderError creates a new provider error
func NewProviderError(providerName string, code ErrorCode, message string) *NotificationError {
	err := &NotificationError{
		Code:       code,
		Message:    message,
		Details:    fmt.Sprintf("Provider: %s", providerName),
		StatusCode: getHTTPStatusCode(code),
		Metadata:   make(map[string]string),
	}
	err.WithMetadata("provider", providerName)
	return err
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(retryAfter string) *NotificationError {
	err := &NotificationError{
		Code:       ErrorCodeRateLimited,
		Message:    "Rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
		Metadata:   make(map[string]string),
	}
	if retryAfter != "" {
		err.WithMetadata("retry_after", retryAfter)
	}
	return err
}

// IsNotificationError checks if an error is a NotificationError
func IsNotificationError(err error) bool {
	_, ok := err.(*NotificationError)
	return ok
}

// AsNotificationError converts an error to NotificationError if possible
func AsNotificationError(err error) (*NotificationError, bool) {
	if notifErr, ok := err.(*NotificationError); ok {
		return notifErr, true
	}
	return nil, false
}

// WrapError wraps a regular error as an internal notification error
func WrapError(err error, message string) *NotificationError {
	if err == nil {
		return nil
	}

	if notifErr, ok := AsNotificationError(err); ok {
		return notifErr
	}

	return NewInternalError(message, err)
}

// getHTTPStatusCode maps error codes to HTTP status codes
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrorCodeInvalidRequest, ErrorCodeValidationFailed,
		ErrorCodeInvalidEmail, ErrorCodeInvalidPhone, ErrorCodeInvalidToken,
		ErrorCodeInvalidRecipient, ErrorCodeInvalidNotification:
		return http.StatusBadRequest

	case ErrorCodeUnauthorized, ErrorCodeProviderAuthentication:
		return http.StatusUnauthorized

	case ErrorCodeNotFound, ErrorCodeProviderNotFound, ErrorCodeTemplateNotFound:
		return http.StatusNotFound

	case ErrorCodeRateLimited:
		return http.StatusTooManyRequests

	case ErrorCodeTimeout, ErrorCodeQueueTimeout:
		return http.StatusRequestTimeout

	case ErrorCodeProviderUnavailable, ErrorCodeNotificationFailed, ErrorCodeDeliveryFailed:
		return http.StatusServiceUnavailable

	case ErrorCodeQueueFull:
		return http.StatusInsufficientStorage

	default:
		return http.StatusInternalServerError
	}
}

// Common error instances for convenience
var (
	ErrNotificationNotFound = NewNotificationError(ErrorCodeNotFound, "Notification not found")
	ErrInvalidRecipient     = NewNotificationError(ErrorCodeInvalidRecipient, "Invalid recipient")
	ErrProviderNotFound     = NewNotificationError(ErrorCodeProviderNotFound, "Notification provider not found")
	ErrQueueFull            = NewNotificationError(ErrorCodeQueueFull, "Notification queue is full")
	ErrQueueEmpty           = NewNotificationError(ErrorCodeQueueEmpty, "Notification queue is empty")
	ErrInvalidEmail         = NewNotificationError(ErrorCodeInvalidEmail, "Invalid email address")
	ErrInvalidPhone         = NewNotificationError(ErrorCodeInvalidPhone, "Invalid phone number")
	ErrInvalidToken         = NewNotificationError(ErrorCodeInvalidToken, "Invalid device token")
)
