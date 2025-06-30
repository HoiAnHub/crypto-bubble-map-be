package repository

import (
	"context"
	"fmt"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoSecurityRepository implements SecurityRepository using MongoDB
type MongoSecurityRepository struct {
	mongo  *database.MongoClient
	logger *zap.Logger
}

// NewMongoSecurityRepository creates a new MongoDB security repository
func NewMongoSecurityRepository(mongo *database.MongoClient, logger *zap.Logger) repository.SecurityRepository {
	return &MongoSecurityRepository{
		mongo:  mongo,
		logger: logger,
	}
}

// GetSecurityAlerts retrieves security alerts with filters
func (r *MongoSecurityRepository) GetSecurityAlerts(ctx context.Context, filters *entity.SecurityAlertFilters, limit, offset int) (*entity.SecurityAlertResult, error) {
	collection := r.mongo.GetCollection("security_alerts")

	// Build filter
	filter := bson.M{}
	if filters != nil {
		if filters.Type != nil {
			filter["type"] = *filters.Type
		}
		if filters.Severity != nil {
			filter["severity"] = *filters.Severity
		}
		if filters.Status != nil {
			filter["status"] = *filters.Status
		}
		if filters.WalletAddress != nil {
			filter["wallet_address"] = *filters.WalletAddress
		}
		if filters.MinConfidence != nil {
			filter["confidence"] = bson.M{"$gte": *filters.MinConfidence}
		}
		if filters.ActionRequired != nil {
			filter["action_required"] = *filters.ActionRequired
		}
		if filters.TimeRange != nil {
			timeFilter := bson.M{}
			if !filters.TimeRange.Start.IsZero() {
				timeFilter["$gte"] = filters.TimeRange.Start
			}
			if !filters.TimeRange.End.IsZero() {
				timeFilter["$lte"] = filters.TimeRange.End
			}
			if len(timeFilter) > 0 {
				filter["timestamp"] = timeFilter
			}
		}
	}

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to count security alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to count security alerts: %w", err)
	}

	// Get alerts with pagination
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error("Failed to find security alerts", zap.Error(err))
		return nil, fmt.Errorf("failed to find security alerts: %w", err)
	}
	defer cursor.Close(ctx)

	var alerts []entity.SecurityAlert
	for cursor.Next(ctx) {
		var alert entity.SecurityAlert
		if err := cursor.Decode(&alert); err != nil {
			r.logger.Error("Failed to decode security alert", zap.Error(err))
			continue
		}
		alerts = append(alerts, alert)
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error while reading security alerts", zap.Error(err))
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	result := &entity.SecurityAlertResult{
		Alerts:  alerts,
		Total:   total,
		HasMore: int64(offset+len(alerts)) < total,
	}

	r.logger.Debug("Retrieved security alerts",
		zap.Int("count", len(alerts)),
		zap.Int64("total", total))

	return result, nil
}

// GetSecurityAlert retrieves a single security alert by ID
func (r *MongoSecurityRepository) GetSecurityAlert(ctx context.Context, alertID string) (*entity.SecurityAlert, error) {
	collection := r.mongo.GetCollection("security_alerts")

	var alert entity.SecurityAlert
	err := collection.FindOne(ctx, bson.M{"id": alertID}).Decode(&alert)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}
		r.logger.Error("Failed to get security alert",
			zap.String("alertID", alertID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get security alert: %w", err)
	}

	return &alert, nil
}

// CreateSecurityAlert creates a new security alert
func (r *MongoSecurityRepository) CreateSecurityAlert(ctx context.Context, alert *entity.SecurityAlert) error {
	collection := r.mongo.GetCollection("security_alerts")

	// Generate ID if not provided
	if alert.ID == "" {
		alert.ID = primitive.NewObjectID().Hex()
	}

	// Set timestamps
	now := time.Now()
	alert.CreatedAt = now
	alert.UpdatedAt = now

	// Set default status if not provided
	if alert.Status == "" {
		alert.Status = entity.AlertStatusActive
	}

	_, err := collection.InsertOne(ctx, alert)
	if err != nil {
		r.logger.Error("Failed to create security alert",
			zap.String("alertID", alert.ID),
			zap.String("type", string(alert.Type)),
			zap.Error(err))
		return fmt.Errorf("failed to create security alert: %w", err)
	}

	r.logger.Info("Created security alert",
		zap.String("alertID", alert.ID),
		zap.String("type", string(alert.Type)),
		zap.String("walletAddress", alert.WalletAddress))

	return nil
}

// UpdateSecurityAlert updates an existing security alert
func (r *MongoSecurityRepository) UpdateSecurityAlert(ctx context.Context, alert *entity.SecurityAlert) error {
	collection := r.mongo.GetCollection("security_alerts")

	// Update timestamp
	alert.UpdatedAt = time.Now()

	filter := bson.M{"id": alert.ID}
	update := bson.M{"$set": alert}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update security alert",
			zap.String("alertID", alert.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update security alert: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("security alert not found")
	}

	r.logger.Info("Updated security alert",
		zap.String("alertID", alert.ID))

	return nil
}

// AcknowledgeSecurityAlert acknowledges a security alert
func (r *MongoSecurityRepository) AcknowledgeSecurityAlert(ctx context.Context, alertID string) error {
	collection := r.mongo.GetCollection("security_alerts")

	now := time.Now()
	filter := bson.M{"id": alertID}
	update := bson.M{
		"$set": bson.M{
			"acknowledged_at": now,
			"updated_at":      now,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to acknowledge security alert",
			zap.String("alertID", alertID),
			zap.Error(err))
		return fmt.Errorf("failed to acknowledge security alert: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("security alert not found")
	}

	r.logger.Info("Acknowledged security alert",
		zap.String("alertID", alertID))

	return nil
}

// ResolveSecurityAlert resolves a security alert
func (r *MongoSecurityRepository) ResolveSecurityAlert(ctx context.Context, alertID, resolution, notes string) error {
	collection := r.mongo.GetCollection("security_alerts")

	now := time.Now()
	filter := bson.M{"id": alertID}
	update := bson.M{
		"$set": bson.M{
			"status":           entity.AlertStatusResolved,
			"resolved_at":      now,
			"resolution":       resolution,
			"resolution_notes": notes,
			"updated_at":       now,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to resolve security alert",
			zap.String("alertID", alertID),
			zap.Error(err))
		return fmt.Errorf("failed to resolve security alert: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("security alert not found")
	}

	r.logger.Info("Resolved security alert",
		zap.String("alertID", alertID),
		zap.String("resolution", resolution))

	return nil
}

// GetComplianceReports retrieves compliance reports with filters
func (r *MongoSecurityRepository) GetComplianceReports(ctx context.Context, filters map[string]interface{}) ([]entity.ComplianceReport, error) {
	collection := r.mongo.GetCollection("compliance_reports")

	// Build filter from map
	filter := bson.M{}
	if filters != nil {
		if reportType, ok := filters["report_type"].(string); ok {
			filter["report_type"] = reportType
		}
		if status, ok := filters["status"].(string); ok {
			filter["status"] = status
		}
		if walletAddress, ok := filters["wallet_address"].(string); ok {
			filter["wallet_address"] = walletAddress
		}
		if generatedBy, ok := filters["generated_by"].(string); ok {
			filter["generated_by"] = generatedBy
		}
		if timeRange, ok := filters["time_range"].(map[string]interface{}); ok {
			timeFilter := bson.M{}
			if start, ok := timeRange["start"].(time.Time); ok {
				timeFilter["$gte"] = start
			}
			if end, ok := timeRange["end"].(time.Time); ok {
				timeFilter["$lte"] = end
			}
			if len(timeFilter) > 0 {
				filter["generated_at"] = timeFilter
			}
		}
	}

	// Get reports with default sorting
	findOptions := options.Find().
		SetSort(bson.D{{Key: "generated_at", Value: -1}}).
		SetLimit(100) // Default limit

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.logger.Error("Failed to find compliance reports", zap.Error(err))
		return nil, fmt.Errorf("failed to find compliance reports: %w", err)
	}
	defer cursor.Close(ctx)

	var reports []entity.ComplianceReport
	for cursor.Next(ctx) {
		var report entity.ComplianceReport
		if err := cursor.Decode(&report); err != nil {
			r.logger.Error("Failed to decode compliance report", zap.Error(err))
			continue
		}
		reports = append(reports, report)
	}

	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error while reading compliance reports", zap.Error(err))
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	r.logger.Debug("Retrieved compliance reports",
		zap.Int("count", len(reports)))

	return reports, nil
}

// GetComplianceReport retrieves a single compliance report by ID
func (r *MongoSecurityRepository) GetComplianceReport(ctx context.Context, reportID string) (*entity.ComplianceReport, error) {
	collection := r.mongo.GetCollection("compliance_reports")

	var report entity.ComplianceReport
	err := collection.FindOne(ctx, bson.M{"id": reportID}).Decode(&report)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, nil
		}
		r.logger.Error("Failed to get compliance report",
			zap.String("reportID", reportID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get compliance report: %w", err)
	}

	return &report, nil
}

// CreateComplianceReport creates a new compliance report
func (r *MongoSecurityRepository) CreateComplianceReport(ctx context.Context, report *entity.ComplianceReport) error {
	collection := r.mongo.GetCollection("compliance_reports")

	// Generate ID if not provided
	if report.ID == "" {
		report.ID = primitive.NewObjectID().Hex()
	}

	// Set timestamp
	if report.GeneratedAt.IsZero() {
		report.GeneratedAt = time.Now()
	}

	// Set default status if not provided
	if report.Status == "" {
		report.Status = entity.ComplianceReportStatusDraft
	}

	_, err := collection.InsertOne(ctx, report)
	if err != nil {
		r.logger.Error("Failed to create compliance report",
			zap.String("reportID", report.ID),
			zap.String("type", string(report.ReportType)),
			zap.Error(err))
		return fmt.Errorf("failed to create compliance report: %w", err)
	}

	r.logger.Info("Created compliance report",
		zap.String("reportID", report.ID),
		zap.String("type", string(report.ReportType)),
		zap.String("walletAddress", report.WalletAddress))

	return nil
}

// UpdateComplianceReport updates an existing compliance report
func (r *MongoSecurityRepository) UpdateComplianceReport(ctx context.Context, report *entity.ComplianceReport) error {
	collection := r.mongo.GetCollection("compliance_reports")

	filter := bson.M{"id": report.ID}
	update := bson.M{"$set": report}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update compliance report",
			zap.String("reportID", report.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update compliance report: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("compliance report not found")
	}

	r.logger.Info("Updated compliance report",
		zap.String("reportID", report.ID))

	return nil
}

// GenerateComplianceReport generates a new compliance report for a wallet
func (r *MongoSecurityRepository) GenerateComplianceReport(ctx context.Context, walletAddress string, reportType entity.ComplianceReportType, timeRange entity.TimeRange) (*entity.ComplianceReport, error) {
	// Create a new compliance report
	report := &entity.ComplianceReport{
		ID:            primitive.NewObjectID().Hex(),
		WalletAddress: walletAddress,
		ReportType:    reportType,
		GeneratedAt:   time.Now(),
		GeneratedBy:   "system", // This could be passed as parameter
		TimeRange:     timeRange,
		Status:        entity.ComplianceReportStatusDraft,
		Summary: entity.ComplianceSummary{
			TotalTransactions:    0,
			TotalVolume:          "0",
			TotalVolumeUSD:       0,
			HighRiskTransactions: 0,
			SuspiciousPatterns:   0,
			RegulatoryViolations: 0,
			OverallRiskScore:     0,
			ComplianceScore:      100, // Start with perfect score
		},
		Findings:        []entity.ComplianceFinding{},
		Recommendations: []string{},
		RiskAssessment: entity.ComplianceRiskAssessment{
			OverallRisk:        entity.RiskLevelLow,
			GeographicRisk:     entity.RiskLevelLow,
			TransactionRisk:    entity.RiskLevelLow,
			CounterpartyRisk:   entity.RiskLevelLow,
			ProductRisk:        entity.RiskLevelLow,
			RiskFactors:        []string{},
			MitigatingFactors:  []string{},
			RecommendedActions: []string{},
			NextReviewDate:     time.Now().AddDate(0, 3, 0), // 3 months from now
		},
		RegulatoryFlags: []entity.RegulatoryFlag{},
		Metadata:        make(map[string]interface{}),
	}

	// TODO: Implement actual analysis logic here
	// This would involve:
	// 1. Fetching transaction data for the wallet in the time range
	// 2. Analyzing patterns and risks
	// 3. Checking against regulatory databases
	// 4. Generating findings and recommendations
	// For now, we'll create a basic report

	// Add some basic analysis based on report type
	switch reportType {
	case entity.ComplianceReportTypeAML:
		report.Recommendations = append(report.Recommendations,
			"Continue monitoring for unusual transaction patterns",
			"Verify source of funds for large transactions",
			"Maintain transaction records for regulatory compliance")
	case entity.ComplianceReportTypeKYC:
		report.Recommendations = append(report.Recommendations,
			"Verify customer identity documentation",
			"Update customer information regularly",
			"Monitor for changes in transaction behavior")
	case entity.ComplianceReportTypeSAR:
		report.Recommendations = append(report.Recommendations,
			"File suspicious activity report if patterns persist",
			"Enhanced monitoring recommended",
			"Consider customer due diligence review")
	case entity.ComplianceReportTypeRiskAssessment:
		report.Recommendations = append(report.Recommendations,
			"Regular risk assessment updates recommended",
			"Monitor for changes in risk profile",
			"Implement appropriate risk controls")
	}

	// Save the report
	if err := r.CreateComplianceReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save generated compliance report: %w", err)
	}

	r.logger.Info("Generated compliance report",
		zap.String("reportID", report.ID),
		zap.String("walletAddress", walletAddress),
		zap.String("reportType", string(reportType)))

	return report, nil
}
