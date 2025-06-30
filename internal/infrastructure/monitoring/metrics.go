package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"crypto-bubble-map-be/internal/infrastructure/errors"

	"go.uber.org/zap"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a single metric
type Metric struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MetricsCollector collects and manages application metrics
type MetricsCollector struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metric),
		logger:  logger,
	}
}

// Counter increments a counter metric
func (mc *MetricsCollector) Counter(name string, labels map[string]string, description string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.buildKey(name, labels)
	if metric, exists := mc.metrics[key]; exists {
		metric.Value++
		metric.Timestamp = time.Now()
	} else {
		mc.metrics[key] = &Metric{
			Name:        name,
			Type:        MetricTypeCounter,
			Value:       1,
			Labels:      labels,
			Timestamp:   time.Now(),
			Description: description,
		}
	}
}

// Gauge sets a gauge metric value
func (mc *MetricsCollector) Gauge(name string, value float64, labels map[string]string, description string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.buildKey(name, labels)
	mc.metrics[key] = &Metric{
		Name:        name,
		Type:        MetricTypeGauge,
		Value:       value,
		Labels:      labels,
		Timestamp:   time.Now(),
		Description: description,
	}
}

// Histogram records a histogram metric
func (mc *MetricsCollector) Histogram(name string, value float64, labels map[string]string, description string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.buildKey(name, labels)
	if metric, exists := mc.metrics[key]; exists {
		// Simple histogram implementation - in production, use proper buckets
		if metric.Metadata == nil {
			metric.Metadata = make(map[string]interface{})
		}

		count, _ := metric.Metadata["count"].(float64)
		sum, _ := metric.Metadata["sum"].(float64)

		metric.Metadata["count"] = count + 1
		metric.Metadata["sum"] = sum + value
		metric.Value = (sum + value) / (count + 1) // Average
		metric.Timestamp = time.Now()
	} else {
		mc.metrics[key] = &Metric{
			Name:        name,
			Type:        MetricTypeHistogram,
			Value:       value,
			Labels:      labels,
			Timestamp:   time.Now(),
			Description: description,
			Metadata: map[string]interface{}{
				"count": 1.0,
				"sum":   value,
			},
		}
	}
}

// GetMetrics returns all collected metrics
func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*Metric)
	for k, v := range mc.metrics {
		metricCopy := *v
		result[k] = &metricCopy
	}
	return result
}

// GetMetric returns a specific metric
func (mc *MetricsCollector) GetMetric(name string, labels map[string]string) (*Metric, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	key := mc.buildKey(name, labels)
	metric, exists := mc.metrics[key]
	if exists {
		metricCopy := *metric
		return &metricCopy, true
	}
	return nil, false
}

// Reset clears all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics = make(map[string]*Metric)
}

// buildKey creates a unique key for a metric
func (mc *MetricsCollector) buildKey(name string, labels map[string]string) string {
	key := name
	if labels != nil {
		for k, v := range labels {
			key += ":" + k + "=" + v
		}
	}
	return key
}

// PerformanceMonitor monitors application performance
type PerformanceMonitor struct {
	collector *MetricsCollector
	logger    *zap.Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(collector *MetricsCollector, logger *zap.Logger) *PerformanceMonitor {
	return &PerformanceMonitor{
		collector: collector,
		logger:    logger,
	}
}

// TrackDuration tracks the duration of an operation
func (pm *PerformanceMonitor) TrackDuration(operation string, labels map[string]string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		pm.collector.Histogram(
			"operation_duration_seconds",
			duration.Seconds(),
			mergeLabels(labels, map[string]string{"operation": operation}),
			"Duration of operations in seconds",
		)
	}
}

// TrackError tracks error occurrences
func (pm *PerformanceMonitor) TrackError(err error, operation string, labels map[string]string) {
	if err == nil {
		return
	}

	errorLabels := mergeLabels(labels, map[string]string{
		"operation": operation,
		"error":     "true",
	})

	// Track error count
	pm.collector.Counter(
		"operation_errors_total",
		errorLabels,
		"Total number of operation errors",
	)

	// Track error by type if it's an AppError
	if appErr, ok := err.(*errors.AppError); ok {
		errorTypeLabels := mergeLabels(errorLabels, map[string]string{
			"error_code": string(appErr.Code),
		})
		pm.collector.Counter(
			"operation_errors_by_type_total",
			errorTypeLabels,
			"Total number of operation errors by type",
		)
	}
}

// TrackSuccess tracks successful operations
func (pm *PerformanceMonitor) TrackSuccess(operation string, labels map[string]string) {
	successLabels := mergeLabels(labels, map[string]string{
		"operation": operation,
		"error":     "false",
	})

	pm.collector.Counter(
		"operation_success_total",
		successLabels,
		"Total number of successful operations",
	)
}

// TrackDatabaseOperation tracks database operation metrics
func (pm *PerformanceMonitor) TrackDatabaseOperation(database, operation string, duration time.Duration, err error) {
	labels := map[string]string{
		"database":  database,
		"operation": operation,
	}

	// Track duration
	pm.collector.Histogram(
		"database_operation_duration_seconds",
		duration.Seconds(),
		labels,
		"Duration of database operations in seconds",
	)

	// Track success/error
	if err != nil {
		pm.TrackError(err, "database_"+operation, labels)
	} else {
		pm.TrackSuccess("database_"+operation, labels)
	}
}

// TrackAPICall tracks external API call metrics
func (pm *PerformanceMonitor) TrackAPICall(service, endpoint string, statusCode int, duration time.Duration, err error) {
	labels := map[string]string{
		"service":     service,
		"endpoint":    endpoint,
		"status_code": string(rune(statusCode)),
	}

	// Track duration
	pm.collector.Histogram(
		"api_call_duration_seconds",
		duration.Seconds(),
		labels,
		"Duration of external API calls in seconds",
	)

	// Track success/error
	if err != nil || statusCode >= 400 {
		pm.TrackError(err, "api_call", labels)
	} else {
		pm.TrackSuccess("api_call", labels)
	}
}

// TrackCacheOperation tracks cache operation metrics
func (pm *PerformanceMonitor) TrackCacheOperation(operation string, hit bool, duration time.Duration) {
	labels := map[string]string{
		"operation": operation,
		"hit":       "false",
	}

	if hit {
		labels["hit"] = "true"
	}

	// Track duration
	pm.collector.Histogram(
		"cache_operation_duration_seconds",
		duration.Seconds(),
		labels,
		"Duration of cache operations in seconds",
	)

	// Track hit/miss ratio
	pm.collector.Counter(
		"cache_operations_total",
		labels,
		"Total number of cache operations",
	)
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	collector *MetricsCollector
	logger    *zap.Logger
	startTime time.Time
}

// NewSystemMetrics creates a new system metrics collector
func NewSystemMetrics(collector *MetricsCollector, logger *zap.Logger) *SystemMetrics {
	return &SystemMetrics{
		collector: collector,
		logger:    logger,
		startTime: time.Now(),
	}
}

// UpdateSystemMetrics updates system-level metrics
func (sm *SystemMetrics) UpdateSystemMetrics(ctx context.Context) {
	// Update uptime
	uptime := time.Since(sm.startTime)
	sm.collector.Gauge(
		"system_uptime_seconds",
		uptime.Seconds(),
		nil,
		"System uptime in seconds",
	)

	// Update timestamp
	sm.collector.Gauge(
		"system_timestamp",
		float64(time.Now().Unix()),
		nil,
		"Current system timestamp",
	)
}

// TrackUserActivity tracks user activity metrics
func (sm *SystemMetrics) TrackUserActivity(userID uint, action string) {
	labels := map[string]string{
		"action": action,
	}

	sm.collector.Counter(
		"user_activity_total",
		labels,
		"Total user activity events",
	)
}

// TrackWalletAnalysis tracks wallet analysis metrics
func (sm *SystemMetrics) TrackWalletAnalysis(networkID string, walletType string, riskScore float64) {
	labels := map[string]string{
		"network":     networkID,
		"wallet_type": walletType,
	}

	sm.collector.Counter(
		"wallet_analysis_total",
		labels,
		"Total wallet analyses performed",
	)

	sm.collector.Histogram(
		"wallet_risk_score",
		riskScore,
		labels,
		"Distribution of wallet risk scores",
	)
}

// mergeLabels merges multiple label maps
func mergeLabels(labelMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, labels := range labelMaps {
		for k, v := range labels {
			result[k] = v
		}
	}
	return result
}

// MetricsExporter exports metrics in various formats
type MetricsExporter struct {
	collector *MetricsCollector
	logger    *zap.Logger
}

// NewMetricsExporter creates a new metrics exporter
func NewMetricsExporter(collector *MetricsCollector, logger *zap.Logger) *MetricsExporter {
	return &MetricsExporter{
		collector: collector,
		logger:    logger,
	}
}

// ExportPrometheus exports metrics in Prometheus format
func (me *MetricsExporter) ExportPrometheus() string {
	metrics := me.collector.GetMetrics()
	var output string

	for _, metric := range metrics {
		// Add help comment
		if metric.Description != "" {
			output += "# HELP " + metric.Name + " " + metric.Description + "\n"
		}

		// Add type comment
		output += "# TYPE " + metric.Name + " " + string(metric.Type) + "\n"

		// Add metric line
		metricLine := metric.Name
		if len(metric.Labels) > 0 {
			metricLine += "{"
			first := true
			for k, v := range metric.Labels {
				if !first {
					metricLine += ","
				}
				metricLine += k + `="` + v + `"`
				first = false
			}
			metricLine += "}"
		}
		metricLine += " " + fmt.Sprintf("%.6f", metric.Value)
		metricLine += " " + fmt.Sprintf("%d", metric.Timestamp.UnixMilli())
		output += metricLine + "\n"
	}

	return output
}

// ExportJSON exports metrics in JSON format
func (me *MetricsExporter) ExportJSON() map[string]*Metric {
	return me.collector.GetMetrics()
}
