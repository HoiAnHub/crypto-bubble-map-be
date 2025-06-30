package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/errors"
	"crypto-bubble-map-be/internal/infrastructure/monitoring"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware handles errors and converts them to appropriate HTTP responses
func ErrorHandlerMiddleware(logger *zap.Logger, monitor *monitoring.PerformanceMonitor) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			handlePanic(c, err, logger, monitor)
		} else if err, ok := recovered.(error); ok {
			handleError(c, err, logger, monitor)
		} else {
			handleUnknownPanic(c, recovered, logger, monitor)
		}
	})
}

// ErrorResponseMiddleware handles application errors and converts them to JSON responses
func ErrorResponseMiddleware(logger *zap.Logger, monitor *monitoring.PerformanceMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err, logger, monitor)
		}
	}
}

// handleError processes application errors
func handleError(c *gin.Context, err error, logger *zap.Logger, monitor *monitoring.PerformanceMonitor) {
	requestID := getRequestID(c)
	userID := getUserID(c)

	// Convert to AppError if it's not already
	var appErr *errors.AppError
	if ae, ok := err.(*errors.AppError); ok {
		appErr = ae
	} else {
		appErr = errors.WrapError(err, errors.ErrCodeInternalServer, "Internal server error")
	}

	// Add request context to error
	appErr.WithRequestID(requestID)
	if userID != nil {
		appErr.WithUserID(*userID)
	}

	// Log the error
	logError(appErr, c, logger)

	// Track error metrics
	if monitor != nil {
		monitor.TrackError(appErr, c.Request.Method+" "+c.FullPath(), map[string]string{
			"status_code": string(rune(appErr.HTTPStatus)),
			"method":      c.Request.Method,
			"path":        c.FullPath(),
		})
	}

	// Return JSON error response
	c.JSON(appErr.HTTPStatus, errors.ToErrorResponse(appErr))
	c.Abort()
}

// handlePanic processes panic recoveries
func handlePanic(c *gin.Context, panicMsg string, logger *zap.Logger, monitor *monitoring.PerformanceMonitor) {
	requestID := getRequestID(c)
	userID := getUserID(c)

	appErr := errors.NewInternalError("Panic recovered", nil).
		WithRequestID(requestID).
		WithMetadata("panic_message", panicMsg).
		WithMetadata("stack_trace", string(debug.Stack()))

	if userID != nil {
		appErr.WithUserID(*userID)
	}

	// Log the panic
	logger.Error("Panic recovered",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("panic_message", panicMsg),
		zap.String("stack_trace", string(debug.Stack())),
	)

	// Track panic metrics
	if monitor != nil {
		monitor.TrackError(appErr, "panic", map[string]string{
			"method": c.Request.Method,
			"path":   c.FullPath(),
		})
	}

	// Return JSON error response
	c.JSON(http.StatusInternalServerError, errors.ToErrorResponse(appErr))
	c.Abort()
}

// handleUnknownPanic processes unknown panic types
func handleUnknownPanic(c *gin.Context, recovered interface{}, logger *zap.Logger, monitor *monitoring.PerformanceMonitor) {
	requestID := getRequestID(c)
	userID := getUserID(c)

	appErr := errors.NewInternalError("Unknown panic recovered", nil).
		WithRequestID(requestID).
		WithMetadata("panic_value", recovered).
		WithMetadata("stack_trace", string(debug.Stack()))

	if userID != nil {
		appErr.WithUserID(*userID)
	}

	// Log the unknown panic
	logger.Error("Unknown panic recovered",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Any("panic_value", recovered),
		zap.String("stack_trace", string(debug.Stack())),
	)

	// Track panic metrics
	if monitor != nil {
		monitor.TrackError(appErr, "unknown_panic", map[string]string{
			"method": c.Request.Method,
			"path":   c.FullPath(),
		})
	}

	// Return JSON error response
	c.JSON(http.StatusInternalServerError, errors.ToErrorResponse(appErr))
	c.Abort()
}

// logError logs application errors with appropriate level
func logError(appErr *errors.AppError, c *gin.Context, logger *zap.Logger) {
	fields := []zap.Field{
		zap.String("error_code", string(appErr.Code)),
		zap.String("request_id", appErr.RequestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("remote_addr", c.ClientIP()),
		zap.Int("status_code", appErr.HTTPStatus),
		zap.Time("timestamp", appErr.Timestamp),
	}

	if appErr.UserID != nil {
		fields = append(fields, zap.Uint("user_id", *appErr.UserID))
	}

	if appErr.Metadata != nil && len(appErr.Metadata) > 0 {
		fields = append(fields, zap.Any("metadata", appErr.Metadata))
	}

	if appErr.Cause != nil {
		fields = append(fields, zap.Error(appErr.Cause))
	}

	// Log with appropriate level based on error severity
	switch appErr.HTTPStatus {
	case http.StatusInternalServerError, http.StatusServiceUnavailable:
		logger.Error(appErr.Message, fields...)
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		logger.Warn(appErr.Message, fields...)
	default:
		logger.Info(appErr.Message, fields...)
	}
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// getUserID extracts user ID from context
func getUserID(c *gin.Context) *uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return &id
		}
	}
	return nil
}

// TimeoutMiddleware adds request timeout handling
func TimeoutMiddleware(timeout time.Duration, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		done := make(chan struct{})
		go func() {
			defer close(done)
			c.Next()
		}()

		select {
		case <-done:
			// Request completed normally
			return
		case <-ctx.Done():
			// Request timed out
			if ctx.Err() == context.DeadlineExceeded {
				appErr := errors.NewAppError(
					errors.ErrCodeDatabaseTimeout,
					"Request timeout",
					"Request exceeded maximum allowed time",
				).WithRequestID(getRequestID(c))

				logger.Warn("Request timeout",
					zap.String("request_id", getRequestID(c)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.Duration("timeout", timeout),
				)

				c.JSON(http.StatusRequestTimeout, errors.ToErrorResponse(appErr))
				c.Abort()
			}
		}
	}
}

// RetryableErrorMiddleware handles retryable errors
func RetryableErrorMiddleware(maxRetries int, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			// Reset response writer for retry
			if attempt > 0 {
				// Create a new response writer for retry
				c.Writer = c.Writer
			}

			c.Next()

			// Check if there are any retryable errors
			if len(c.Errors) > 0 {
				err := c.Errors.Last().Err
				if errors.IsRetryable(err) && attempt < maxRetries {
					lastErr = err
					c.Errors = c.Errors[:len(c.Errors)-1] // Remove the error for retry

					logger.Info("Retrying request",
						zap.String("request_id", getRequestID(c)),
						zap.Int("attempt", attempt+1),
						zap.Int("max_retries", maxRetries),
						zap.Error(err),
					)

					// Wait before retry (exponential backoff)
					time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
					continue
				}
			}

			// No retryable errors or max retries reached
			break
		}

		// If we exhausted retries, add the last error back
		if lastErr != nil && len(c.Errors) == 0 {
			c.Error(lastErr)
		}
	}
}

// CircuitBreakerMiddleware implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
	logger       *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
		logger:       logger,
	}
}

// CircuitBreakerMiddleware returns a middleware that implements circuit breaker pattern
func (cb *CircuitBreaker) CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check circuit breaker state
		if cb.state == "open" {
			if time.Since(cb.lastFailTime) > cb.resetTimeout {
				cb.state = "half-open"
				cb.logger.Info("Circuit breaker transitioning to half-open state")
			} else {
				// Circuit is open, reject request
				appErr := errors.NewAppError(
					errors.ErrCodeServiceUnavailable,
					"Service temporarily unavailable",
					"Circuit breaker is open",
				).WithRequestID(getRequestID(c))

				c.JSON(http.StatusServiceUnavailable, errors.ToErrorResponse(appErr))
				c.Abort()
				return
			}
		}

		c.Next()

		// Check if request failed
		if len(c.Errors) > 0 {
			cb.recordFailure()
		} else if cb.state == "half-open" {
			cb.recordSuccess()
		}
	}
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = "open"
		cb.logger.Warn("Circuit breaker opened",
			zap.Int("failures", cb.failures),
			zap.Int("max_failures", cb.maxFailures),
		)
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0
	cb.state = "closed"
	cb.logger.Info("Circuit breaker closed")
}
