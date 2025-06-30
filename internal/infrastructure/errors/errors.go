package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents different types of errors in the system
type ErrorCode string

const (
	// Database errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery      ErrorCode = "DATABASE_QUERY"
	ErrCodeDatabaseMigration  ErrorCode = "DATABASE_MIGRATION"
	ErrCodeDatabaseTimeout    ErrorCode = "DATABASE_TIMEOUT"

	// Authentication errors
	ErrCodeAuthInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"
	ErrCodeAuthTokenExpired       ErrorCode = "AUTH_TOKEN_EXPIRED"
	ErrCodeAuthTokenInvalid       ErrorCode = "AUTH_TOKEN_INVALID"
	ErrCodeAuthPermissionDenied   ErrorCode = "AUTH_PERMISSION_DENIED"
	ErrCodeAuthUserNotFound       ErrorCode = "AUTH_USER_NOT_FOUND"

	// Validation errors
	ErrCodeValidationFailed    ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput        ErrorCode = "INVALID_INPUT"
	ErrCodeInvalidWalletAddress ErrorCode = "INVALID_WALLET_ADDRESS"
	ErrCodeInvalidNetworkID     ErrorCode = "INVALID_NETWORK_ID"

	// Business logic errors
	ErrCodeWalletNotFound       ErrorCode = "WALLET_NOT_FOUND"
	ErrCodeWalletAlreadyWatched ErrorCode = "WALLET_ALREADY_WATCHED"
	ErrCodeNetworkNotSupported  ErrorCode = "NETWORK_NOT_SUPPORTED"
	ErrCodeInsufficientData     ErrorCode = "INSUFFICIENT_DATA"

	// External service errors
	ErrCodeExternalAPIFailure    ErrorCode = "EXTERNAL_API_FAILURE"
	ErrCodeExternalAPITimeout    ErrorCode = "EXTERNAL_API_TIMEOUT"
	ErrCodeExternalAPIRateLimit  ErrorCode = "EXTERNAL_API_RATE_LIMIT"
	ErrCodeBlockchainRPCFailure  ErrorCode = "BLOCKCHAIN_RPC_FAILURE"

	// System errors
	ErrCodeInternalServer    ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeConfigurationError ErrorCode = "CONFIGURATION_ERROR"
	ErrCodeResourceExhausted  ErrorCode = "RESOURCE_EXHAUSTED"

	// Cache errors
	ErrCodeCacheFailure ErrorCode = "CACHE_FAILURE"
	ErrCodeCacheTimeout ErrorCode = "CACHE_TIMEOUT"

	// AI service errors
	ErrCodeAIServiceFailure ErrorCode = "AI_SERVICE_FAILURE"
	ErrCodeAIServiceTimeout ErrorCode = "AI_SERVICE_TIMEOUT"
	ErrCodeAIQuotaExceeded  ErrorCode = "AI_QUOTA_EXCEEDED"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Cause      error                  `json:"-"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     *uint                  `json:"user_id,omitempty"`
	HTTPStatus int                    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithRequestID adds request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithUserID adds user ID to the error
func (e *AppError) WithUserID(userID uint) *AppError {
	e.UserID = &userID
	return e
}

// WithCause adds the underlying cause error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		Timestamp:  time.Now(),
		HTTPStatus: getHTTPStatusForCode(code),
		Metadata:   make(map[string]interface{}),
	}
}

// NewDatabaseError creates a database-related error
func NewDatabaseError(operation string, cause error) *AppError {
	return NewAppError(
		ErrCodeDatabaseQuery,
		"Database operation failed",
		fmt.Sprintf("Failed to %s", operation),
	).WithCause(cause).WithMetadata("operation", operation)
}

// NewValidationError creates a validation error
func NewValidationError(field string, message string) *AppError {
	return NewAppError(
		ErrCodeValidationFailed,
		"Validation failed",
		message,
	).WithMetadata("field", field)
}

// NewAuthError creates an authentication error
func NewAuthError(code ErrorCode, message string) *AppError {
	return NewAppError(code, message, "Authentication failed")
}

// NewExternalAPIError creates an external API error
func NewExternalAPIError(service string, cause error) *AppError {
	return NewAppError(
		ErrCodeExternalAPIFailure,
		"External API call failed",
		fmt.Sprintf("Failed to call %s API", service),
	).WithCause(cause).WithMetadata("service", service)
}

// NewInternalError creates an internal server error
func NewInternalError(message string, cause error) *AppError {
	return NewAppError(
		ErrCodeInternalServer,
		"Internal server error",
		message,
	).WithCause(cause)
}

// getHTTPStatusForCode maps error codes to HTTP status codes
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	// Authentication errors -> 401 Unauthorized
	case ErrCodeAuthInvalidCredentials, ErrCodeAuthTokenExpired, ErrCodeAuthTokenInvalid:
		return http.StatusUnauthorized

	// Permission errors -> 403 Forbidden
	case ErrCodeAuthPermissionDenied:
		return http.StatusForbidden

	// Not found errors -> 404 Not Found
	case ErrCodeAuthUserNotFound, ErrCodeWalletNotFound:
		return http.StatusNotFound

	// Validation errors -> 400 Bad Request
	case ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeInvalidWalletAddress, ErrCodeInvalidNetworkID:
		return http.StatusBadRequest

	// Conflict errors -> 409 Conflict
	case ErrCodeWalletAlreadyWatched:
		return http.StatusConflict

	// Rate limiting -> 429 Too Many Requests
	case ErrCodeExternalAPIRateLimit:
		return http.StatusTooManyRequests

	// Service unavailable -> 503 Service Unavailable
	case ErrCodeServiceUnavailable, ErrCodeDatabaseTimeout, ErrCodeExternalAPITimeout:
		return http.StatusServiceUnavailable

	// Resource exhausted -> 507 Insufficient Storage
	case ErrCodeResourceExhausted:
		return http.StatusInsufficientStorage

	// All other errors -> 500 Internal Server Error
	default:
		return http.StatusInternalServerError
	}
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code {
		case ErrCodeDatabaseTimeout, ErrCodeExternalAPITimeout, ErrCodeExternalAPIFailure,
			ErrCodeServiceUnavailable, ErrCodeCacheTimeout:
			return true
		default:
			return false
		}
	}
	return false
}

// IsTemporary determines if an error is temporary
func IsTemporary(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Code {
		case ErrCodeDatabaseTimeout, ErrCodeExternalAPITimeout, ErrCodeServiceUnavailable,
			ErrCodeCacheTimeout, ErrCodeResourceExhausted:
			return true
		default:
			return false
		}
	}
	return false
}

// ErrorResponse represents the JSON error response format
type ErrorResponse struct {
	Error struct {
		Code      ErrorCode              `json:"code"`
		Message   string                 `json:"message"`
		Details   string                 `json:"details,omitempty"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
		Timestamp time.Time              `json:"timestamp"`
		RequestID string                 `json:"request_id,omitempty"`
	} `json:"error"`
}

// ToErrorResponse converts an AppError to ErrorResponse
func ToErrorResponse(err *AppError) ErrorResponse {
	return ErrorResponse{
		Error: struct {
			Code      ErrorCode              `json:"code"`
			Message   string                 `json:"message"`
			Details   string                 `json:"details,omitempty"`
			Metadata  map[string]interface{} `json:"metadata,omitempty"`
			Timestamp time.Time              `json:"timestamp"`
			RequestID string                 `json:"request_id,omitempty"`
		}{
			Code:      err.Code,
			Message:   err.Message,
			Details:   err.Details,
			Metadata:  err.Metadata,
			Timestamp: err.Timestamp,
			RequestID: err.RequestID,
		},
	}
}

// WrapError wraps a standard error into an AppError
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return NewAppError(code, message, err.Error()).WithCause(err)
}

// ErrorMetrics represents error metrics for monitoring
type ErrorMetrics struct {
	Code      ErrorCode `json:"code"`
	Count     int64     `json:"count"`
	LastSeen  time.Time `json:"last_seen"`
	Service   string    `json:"service"`
	Operation string    `json:"operation"`
}

// ErrorCollector collects error metrics
type ErrorCollector struct {
	metrics map[string]*ErrorMetrics
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		metrics: make(map[string]*ErrorMetrics),
	}
}

// Collect records an error occurrence
func (ec *ErrorCollector) Collect(err *AppError, service, operation string) {
	key := fmt.Sprintf("%s:%s:%s", err.Code, service, operation)
	
	if metric, exists := ec.metrics[key]; exists {
		metric.Count++
		metric.LastSeen = time.Now()
	} else {
		ec.metrics[key] = &ErrorMetrics{
			Code:      err.Code,
			Count:     1,
			LastSeen:  time.Now(),
			Service:   service,
			Operation: operation,
		}
	}
}

// GetMetrics returns all collected metrics
func (ec *ErrorCollector) GetMetrics() map[string]*ErrorMetrics {
	return ec.metrics
}

// Reset clears all metrics
func (ec *ErrorCollector) Reset() {
	ec.metrics = make(map[string]*ErrorMetrics)
}
