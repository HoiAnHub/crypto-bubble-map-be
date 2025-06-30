package middleware

import (
	"bytes"
	"io"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestLoggingMiddleware logs HTTP requests and responses
func RequestLoggingMiddleware(logger *zap.Logger, monitor *monitoring.PerformanceMonitor) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom logging is handled in the middleware below
		return ""
	})
}

// DetailedLoggingMiddleware provides detailed request/response logging
func DetailedLoggingMiddleware(logger *zap.Logger, monitor *monitoring.PerformanceMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Start timer
		start := time.Now()

		// Capture request body if needed (for debugging)
		var requestBody []byte
		if c.Request.Body != nil && shouldLogRequestBody(c) {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response writer wrapper to capture response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get response body if needed
		var responseBody []byte
		if shouldLogResponseBody(c) {
			responseBody = responseWriter.body.Bytes()
		}

		// Log request details
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
			zap.Int64("request_size", c.Request.ContentLength),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Add user ID if available
		if userID := getUserID(c); userID != nil {
			fields = append(fields, zap.Uint("user_id", *userID))
		}

		// Add request body for debugging (be careful with sensitive data)
		if len(requestBody) > 0 && len(requestBody) < 1024 { // Limit size
			fields = append(fields, zap.String("request_body", string(requestBody)))
		}

		// Add response body for debugging (be careful with size)
		if len(responseBody) > 0 && len(responseBody) < 1024 { // Limit size
			fields = append(fields, zap.String("response_body", string(responseBody)))
		}

		// Add error information if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log with appropriate level based on status code
		switch {
		case c.Writer.Status() >= 500:
			logger.Error("HTTP Request", fields...)
		case c.Writer.Status() >= 400:
			logger.Warn("HTTP Request", fields...)
		default:
			logger.Info("HTTP Request", fields...)
		}

		// Track metrics
		if monitor != nil {
			labels := map[string]string{
				"method":      c.Request.Method,
				"path":        c.FullPath(),
				"status_code": string(rune(c.Writer.Status())),
			}

			// Track request duration
			monitor.TrackDuration("http_request", labels)

			// Track request count
			if c.Writer.Status() >= 400 {
				monitor.TrackError(nil, "http_request", labels)
			} else {
				monitor.TrackSuccess("http_request", labels)
			}
		}
	}
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldLogRequestBody determines if request body should be logged
func shouldLogRequestBody(c *gin.Context) bool {
	// Only log for certain content types and methods
	contentType := c.GetHeader("Content-Type")
	method := c.Request.Method

	// Log JSON requests for POST, PUT, PATCH
	if (method == "POST" || method == "PUT" || method == "PATCH") &&
		(contentType == "application/json" || contentType == "application/json; charset=utf-8") {
		return true
	}

	return false
}

// shouldLogResponseBody determines if response body should be logged
func shouldLogResponseBody(c *gin.Context) bool {
	// Only log for errors or specific endpoints
	if c.Writer.Status() >= 400 {
		return true
	}

	// Log for specific debug endpoints
	if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
		return true
	}

	return false
}

// CorrelationIDMiddleware adds correlation ID to requests
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID is provided in header
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			// Generate new correlation ID
			correlationID = uuid.New().String()
		}

		// Set in context and response header
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// SecurityLoggingMiddleware logs security-related events
func SecurityLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log authentication attempts
		if c.Request.URL.Path == "/auth/login" || c.Request.URL.Path == "/auth/register" {
			logger.Info("Authentication attempt",
				zap.String("request_id", getRequestID(c)),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
			)
		}

		// Log admin actions
		if isAdminEndpoint(c.Request.URL.Path) {
			logger.Info("Admin action",
				zap.String("request_id", getRequestID(c)),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()),
				zap.Any("user_id", getUserID(c)),
			)
		}

		c.Next()

		// Log failed authentication
		if c.Writer.Status() == 401 {
			logger.Warn("Authentication failed",
				zap.String("request_id", getRequestID(c)),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
			)
		}

		// Log authorization failures
		if c.Writer.Status() == 403 {
			logger.Warn("Authorization failed",
				zap.String("request_id", getRequestID(c)),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Any("user_id", getUserID(c)),
			)
		}
	}
}

// isAdminEndpoint checks if the endpoint requires admin privileges
func isAdminEndpoint(path string) bool {
	adminPaths := []string{
		"/admin/",
		"/api/admin/",
		"/health/detailed",
		"/metrics",
		"/debug/",
	}

	for _, adminPath := range adminPaths {
		if len(path) >= len(adminPath) && path[:len(adminPath)] == adminPath {
			return true
		}
	}

	return false
}

// AuditLoggingMiddleware logs audit events
func AuditLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit logging for read-only operations
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Capture request for audit
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		c.Next()

		// Log audit event for successful modifications
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			fields := []zap.Field{
				zap.String("request_id", getRequestID(c)),
				zap.String("action", c.Request.Method+" "+c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.Time("timestamp", time.Now()),
			}

			if userID := getUserID(c); userID != nil {
				fields = append(fields, zap.Uint("user_id", *userID))
			}

			// Add request body for audit trail (be careful with sensitive data)
			if len(requestBody) > 0 && !containsSensitiveData(c.Request.URL.Path) {
				fields = append(fields, zap.String("request_data", string(requestBody)))
			}

			logger.Info("Audit event", fields...)
		}
	}
}

// containsSensitiveData checks if the endpoint might contain sensitive data
func containsSensitiveData(path string) bool {
	sensitivePaths := []string{
		"/auth/login",
		"/auth/register",
		"/auth/password",
		"/user/password",
	}

	for _, sensitivePath := range sensitivePaths {
		if path == sensitivePath {
			return true
		}
	}

	return false
}

// PerformanceLoggingMiddleware logs performance metrics
func PerformanceLoggingMiddleware(logger *zap.Logger, slowThreshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		// Log slow requests
		if duration > slowThreshold {
			logger.Warn("Slow request detected",
				zap.String("request_id", getRequestID(c)),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("duration", duration),
				zap.Duration("threshold", slowThreshold),
				zap.String("client_ip", c.ClientIP()),
			)
		}

		// Log very slow requests as errors
		if duration > slowThreshold*3 {
			logger.Error("Very slow request detected",
				zap.String("request_id", getRequestID(c)),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("duration", duration),
				zap.String("client_ip", c.ClientIP()),
			)
		}
	}
}
