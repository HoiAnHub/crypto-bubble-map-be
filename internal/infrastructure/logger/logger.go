package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level       string `json:"level"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// NewLogger creates a new logger instance
func NewLogger(cfg *Config) (*Logger, error) {
	var zapConfig zap.Config

	// Configure based on environment
	if cfg.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.DisableStacktrace = true
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Configure output paths
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	// Enable debug mode if specified
	if cfg.Debug {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		zapConfig.Development = true
	}

	// Build logger
	zapLogger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewDefaultLogger creates a logger with default configuration
func NewDefaultLogger() (*Logger, error) {
	cfg := &Config{
		Level:       "info",
		Environment: "development",
		Debug:       false,
	}
	return NewLogger(cfg)
}

// parseLogLevel parses string log level to zapcore.Level
func parseLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "warn", "warning":
		return zap.WarnLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "fatal":
		return zap.FatalLevel, nil
	case "panic":
		return zap.PanicLevel, nil
	default:
		return zap.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(fields...)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return l.WithFields(zap.Error(err))
}

// WithRequestID adds a request ID field to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithFields(zap.String("request_id", requestID))
}

// WithUserID adds a user ID field to the logger
func (l *Logger) WithUserID(userID uint) *Logger {
	return l.WithFields(zap.Uint("user_id", userID))
}

// WithWalletAddress adds a wallet address field to the logger
func (l *Logger) WithWalletAddress(address string) *Logger {
	return l.WithFields(zap.String("wallet_address", address))
}

// WithComponent adds a component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithFields(zap.String("component", component))
}

// Logging methods - delegate to underlying zap.Logger
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.Logger.Panic(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close closes the logger
func (l *Logger) Close() error {
	return l.Sync()
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(cfg *Config) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// Create default logger if not initialized
		logger, err := NewDefaultLogger()
		if err != nil {
			// Fallback to basic logger
			zapLogger, _ := zap.NewDevelopment()
			globalLogger = &Logger{Logger: zapLogger}
		} else {
			globalLogger = logger
		}
	}
	return globalLogger
}

// Global logging functions

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetGlobalLogger().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetGlobalLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

// Panic logs a panic message and panics
func Panic(msg string, fields ...zap.Field) {
	GetGlobalLogger().Panic(msg, fields...)
}

// Middleware logger for HTTP requests
type HTTPLogger struct {
	logger *Logger
}

// NewHTTPLogger creates a new HTTP logger
func NewHTTPLogger(logger *Logger) *HTTPLogger {
	return &HTTPLogger{logger: logger}
}

// LogRequest logs HTTP request details
func (h *HTTPLogger) LogRequest(method, path, userAgent, clientIP string, statusCode int, duration int64, requestID string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
		zap.String("request_id", requestID),
	}

	if statusCode >= 500 {
		h.logger.Error("HTTP request completed with server error", fields...)
	} else if statusCode >= 400 {
		h.logger.Warn("HTTP request completed with client error", fields...)
	} else {
		h.logger.Info("HTTP request completed", fields...)
	}
}

// LogError logs HTTP error details
func (h *HTTPLogger) LogError(err error, method, path, requestID string) {
	h.logger.Error("HTTP request error",
		zap.Error(err),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("request_id", requestID),
	)
}

// Database logger
type DatabaseLogger struct {
	logger *Logger
}

// NewDatabaseLogger creates a new database logger
func NewDatabaseLogger(logger *Logger) *DatabaseLogger {
	return &DatabaseLogger{logger: logger}
}

// LogQuery logs database query details
func (d *DatabaseLogger) LogQuery(query string, duration int64, err error) {
	fields := []zap.Field{
		zap.String("query", query),
		zap.Int64("duration_ms", duration),
	}

	if err != nil {
		d.logger.Error("Database query failed", append(fields, zap.Error(err))...)
	} else {
		d.logger.Debug("Database query executed", fields...)
	}
}

// LogConnection logs database connection events
func (d *DatabaseLogger) LogConnection(database, event string, err error) {
	fields := []zap.Field{
		zap.String("database", database),
		zap.String("event", event),
	}

	if err != nil {
		d.logger.Error("Database connection event failed", append(fields, zap.Error(err))...)
	} else {
		d.logger.Info("Database connection event", fields...)
	}
}

// GraphQL logger
type GraphQLLogger struct {
	logger *Logger
}

// NewGraphQLLogger creates a new GraphQL logger
func NewGraphQLLogger(logger *Logger) *GraphQLLogger {
	return &GraphQLLogger{logger: logger}
}

// LogQuery logs GraphQL query details
func (g *GraphQLLogger) LogQuery(operationName, query string, variables map[string]interface{}, duration int64, err error, requestID string) {
	fields := []zap.Field{
		zap.String("operation_name", operationName),
		zap.String("query", query),
		zap.Any("variables", variables),
		zap.Int64("duration_ms", duration),
		zap.String("request_id", requestID),
	}

	if err != nil {
		g.logger.Error("GraphQL query failed", append(fields, zap.Error(err))...)
	} else {
		g.logger.Info("GraphQL query executed", fields...)
	}
}

// LogSubscription logs GraphQL subscription events
func (g *GraphQLLogger) LogSubscription(operationName string, event string, clientID string) {
	g.logger.Info("GraphQL subscription event",
		zap.String("operation_name", operationName),
		zap.String("event", event),
		zap.String("client_id", clientID),
	)
}

// Security logger
type SecurityLogger struct {
	logger *Logger
}

// NewSecurityLogger creates a new security logger
func NewSecurityLogger(logger *Logger) *SecurityLogger {
	return &SecurityLogger{logger: logger}
}

// LogAuthAttempt logs authentication attempts
func (s *SecurityLogger) LogAuthAttempt(email, clientIP string, success bool, reason string) {
	fields := []zap.Field{
		zap.String("email", email),
		zap.String("client_ip", clientIP),
		zap.Bool("success", success),
		zap.String("reason", reason),
	}

	if success {
		s.logger.Info("Authentication successful", fields...)
	} else {
		s.logger.Warn("Authentication failed", fields...)
	}
}

// LogSecurityEvent logs security-related events
func (s *SecurityLogger) LogSecurityEvent(eventType, description, clientIP string, userID *uint) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("description", description),
		zap.String("client_ip", clientIP),
	}

	if userID != nil {
		fields = append(fields, zap.Uint("user_id", *userID))
	}

	s.logger.Warn("Security event", fields...)
}

// LogRateLimitExceeded logs rate limit violations
func (s *SecurityLogger) LogRateLimitExceeded(clientIP, endpoint string, limit int64) {
	s.logger.Warn("Rate limit exceeded",
		zap.String("client_ip", clientIP),
		zap.String("endpoint", endpoint),
		zap.Int64("limit", limit),
	)
}

// Cleanup function to be called on application shutdown
func Cleanup() {
	if globalLogger != nil {
		globalLogger.Close()
	}
}

// SetupLogRotation sets up log rotation (if needed)
func SetupLogRotation(logFile string, maxSize, maxBackups, maxAge int) error {
	// This would implement log rotation using lumberjack or similar
	// For now, we'll just log to stdout/stderr
	return nil
}

// Level returns the current log level
func (l *Logger) Level() zapcore.Level {
	return l.Logger.Level()
}

// Core returns the underlying zapcore.Core
func (l *Logger) Core() zapcore.Core {
	return l.Logger.Core()
}

// GetLogLevel returns the current log level
func GetLogLevel() zapcore.Level {
	if globalLogger != nil {
		return globalLogger.Level()
	}
	return zap.InfoLevel
}

// SetLogLevel sets the log level dynamically
func SetLogLevel(level zapcore.Level) {
	if globalLogger != nil {
		globalLogger.Core().Enabled(level)
	}
}
