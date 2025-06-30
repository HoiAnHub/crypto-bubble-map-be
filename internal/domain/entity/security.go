package entity

import (
	"time"
)

// AlertType represents different types of security alerts
type AlertType string

const (
	AlertTypePhishing   AlertType = "PHISHING"
	AlertTypeMEV        AlertType = "MEV"
	AlertTypeLaundering AlertType = "LAUNDERING"
	AlertTypeSanctions  AlertType = "SANCTIONS"
	AlertTypeScam       AlertType = "SCAM"
	AlertTypeSuspicious AlertType = "SUSPICIOUS"
)

// AlertStatus represents the status of a security alert
type AlertStatus string

const (
	AlertStatusActive        AlertStatus = "ACTIVE"
	AlertStatusResolved      AlertStatus = "RESOLVED"
	AlertStatusInvestigating AlertStatus = "INVESTIGATING"
)

// SecurityAlert represents a security alert in the system
type SecurityAlert struct {
	ID                   string                 `json:"id"`
	Type                 AlertType              `json:"type"`
	Severity             AlertSeverity          `json:"severity"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	WalletAddress        string                 `json:"wallet_address"`
	Timestamp            time.Time              `json:"timestamp"`
	Status               AlertStatus            `json:"status"`
	Confidence           int                    `json:"confidence"` // 0-100
	RelatedTransactions  []string               `json:"related_transactions"`
	ActionRequired       bool                   `json:"action_required"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	
	// Resolution fields
	ResolvedAt           *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy           *string                `json:"resolved_by,omitempty"`
	Resolution           *string                `json:"resolution,omitempty"`
	ResolutionNotes      *string                `json:"resolution_notes,omitempty"`
	
	// Investigation fields
	InvestigatedAt       *time.Time             `json:"investigated_at,omitempty"`
	InvestigatedBy       *string                `json:"investigated_by,omitempty"`
	InvestigationNotes   *string                `json:"investigation_notes,omitempty"`
	
	// System fields
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// SecurityAlertFilters represents filters for security alert queries
type SecurityAlertFilters struct {
	Type          *AlertType     `json:"type,omitempty"`
	Severity      *AlertSeverity `json:"severity,omitempty"`
	Status        *AlertStatus   `json:"status,omitempty"`
	WalletAddress *string        `json:"wallet_address,omitempty"`
	TimeRange     *TimeRange     `json:"time_range,omitempty"`
	MinConfidence *int           `json:"min_confidence,omitempty"`
	ActionRequired *bool         `json:"action_required,omitempty"`
}

// SecurityAlertResult represents paginated security alert results
type SecurityAlertResult struct {
	Alerts  []SecurityAlert `json:"alerts"`
	HasMore bool            `json:"has_more"`
	Total   int64           `json:"total"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ID               string                 `json:"id"`
	WalletAddress    string                 `json:"wallet_address"`
	ReportType       ComplianceReportType   `json:"report_type"`
	GeneratedAt      time.Time              `json:"generated_at"`
	GeneratedBy      string                 `json:"generated_by"`
	Summary          ComplianceSummary      `json:"summary"`
	Findings         []ComplianceFinding    `json:"findings"`
	Recommendations  []string               `json:"recommendations"`
	RiskAssessment   ComplianceRiskAssessment `json:"risk_assessment"`
	RegulatoryFlags  []RegulatoryFlag       `json:"regulatory_flags"`
	TimeRange        TimeRange              `json:"time_range"`
	Status           ComplianceReportStatus `json:"status"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceReportType represents different types of compliance reports
type ComplianceReportType string

const (
	ComplianceReportTypeAML        ComplianceReportType = "AML"
	ComplianceReportTypeKYC        ComplianceReportType = "KYC"
	ComplianceReportTypeSAR        ComplianceReportType = "SAR" // Suspicious Activity Report
	ComplianceReportTypeCTR        ComplianceReportType = "CTR" // Currency Transaction Report
	ComplianceReportTypeOFAC       ComplianceReportType = "OFAC"
	ComplianceReportTypeRiskAssessment ComplianceReportType = "RISK_ASSESSMENT"
)

// ComplianceReportStatus represents the status of a compliance report
type ComplianceReportStatus string

const (
	ComplianceReportStatusDraft     ComplianceReportStatus = "DRAFT"
	ComplianceReportStatusPending   ComplianceReportStatus = "PENDING"
	ComplianceReportStatusApproved  ComplianceReportStatus = "APPROVED"
	ComplianceReportStatusSubmitted ComplianceReportStatus = "SUBMITTED"
	ComplianceReportStatusRejected  ComplianceReportStatus = "REJECTED"
)

// ComplianceSummary represents a summary of compliance findings
type ComplianceSummary struct {
	TotalTransactions    int64   `json:"total_transactions"`
	TotalVolume          string  `json:"total_volume"`
	TotalVolumeUSD       float64 `json:"total_volume_usd"`
	HighRiskTransactions int64   `json:"high_risk_transactions"`
	SuspiciousPatterns   int64   `json:"suspicious_patterns"`
	RegulatoryViolations int64   `json:"regulatory_violations"`
	OverallRiskScore     float64 `json:"overall_risk_score"`
	ComplianceScore      float64 `json:"compliance_score"`
}

// ComplianceFinding represents a specific compliance finding
type ComplianceFinding struct {
	ID                   string                 `json:"id"`
	Type                 ComplianceFindingType  `json:"type"`
	Severity             AlertSeverity          `json:"severity"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Evidence             []string               `json:"evidence"`
	RelatedTransactions  []string               `json:"related_transactions"`
	RegulatoryReference  *string                `json:"regulatory_reference,omitempty"`
	Recommendation       string                 `json:"recommendation"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceFindingType represents different types of compliance findings
type ComplianceFindingType string

const (
	ComplianceFindingTypeStructuring       ComplianceFindingType = "STRUCTURING"
	ComplianceFindingTypeSmurfing          ComplianceFindingType = "SMURFING"
	ComplianceFindingTypeLayering          ComplianceFindingType = "LAYERING"
	ComplianceFindingTypeIntegration       ComplianceFindingType = "INTEGRATION"
	ComplianceFindingTypeUnusualPattern    ComplianceFindingType = "UNUSUAL_PATTERN"
	ComplianceFindingTypeHighRiskJurisdiction ComplianceFindingType = "HIGH_RISK_JURISDICTION"
	ComplianceFindingTypeSanctionsViolation ComplianceFindingType = "SANCTIONS_VIOLATION"
	ComplianceFindingTypeThresholdViolation ComplianceFindingType = "THRESHOLD_VIOLATION"
)

// ComplianceRiskAssessment represents a risk assessment for compliance
type ComplianceRiskAssessment struct {
	OverallRisk          RiskLevel              `json:"overall_risk"`
	GeographicRisk       RiskLevel              `json:"geographic_risk"`
	TransactionRisk      RiskLevel              `json:"transaction_risk"`
	CounterpartyRisk     RiskLevel              `json:"counterparty_risk"`
	ProductRisk          RiskLevel              `json:"product_risk"`
	RiskFactors          []string               `json:"risk_factors"`
	MitigatingFactors    []string               `json:"mitigating_factors"`
	RecommendedActions   []string               `json:"recommended_actions"`
	NextReviewDate       time.Time              `json:"next_review_date"`
}

// RegulatoryFlag represents a regulatory flag or violation
type RegulatoryFlag struct {
	Type                 RegulatoryFlagType     `json:"type"`
	Jurisdiction         string                 `json:"jurisdiction"`
	Regulation           string                 `json:"regulation"`
	Description          string                 `json:"description"`
	Severity             AlertSeverity          `json:"severity"`
	RequiredAction       string                 `json:"required_action"`
	Deadline             *time.Time             `json:"deadline,omitempty"`
	Status               RegulatoryFlagStatus   `json:"status"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// RegulatoryFlagType represents different types of regulatory flags
type RegulatoryFlagType string

const (
	RegulatoryFlagTypeOFAC         RegulatoryFlagType = "OFAC"
	RegulatoryFlagTypeEU_SANCTIONS RegulatoryFlagType = "EU_SANCTIONS"
	RegulatoryFlagTypeUN_SANCTIONS RegulatoryFlagType = "UN_SANCTIONS"
	RegulatoryFlagTypeAML          RegulatoryFlagType = "AML"
	RegulatoryFlagTypeKYC          RegulatoryFlagType = "KYC"
	RegulatoryFlagTypeCTR          RegulatoryFlagType = "CTR"
	RegulatoryFlagTypeSAR          RegulatoryFlagType = "SAR"
	RegulatoryFlagTypeFATF         RegulatoryFlagType = "FATF"
)

// RegulatoryFlagStatus represents the status of a regulatory flag
type RegulatoryFlagStatus string

const (
	RegulatoryFlagStatusActive    RegulatoryFlagStatus = "ACTIVE"
	RegulatoryFlagStatusResolved  RegulatoryFlagStatus = "RESOLVED"
	RegulatoryFlagStatusExempted  RegulatoryFlagStatus = "EXEMPTED"
	RegulatoryFlagStatusPending   RegulatoryFlagStatus = "PENDING"
)

// AIResponse represents a response from the AI assistant
type AIResponse struct {
	Answer           string   `json:"answer"`
	Confidence       float64  `json:"confidence"`
	Sources          []string `json:"sources"`
	RelatedQuestions []string `json:"related_questions"`
	ActionItems      []string `json:"action_items"`
	GeneratedAt      time.Time `json:"generated_at"`
	Model            string   `json:"model"`
	TokensUsed       int      `json:"tokens_used"`
}

// AIContext represents context for AI queries
type AIContext struct {
	AnalysisType string  `json:"analysis_type,omitempty"`
	Timeframe    string  `json:"timeframe,omitempty"`
	NetworkID    string  `json:"network_id,omitempty"`
	UserRole     string  `json:"user_role,omitempty"`
}

// Helper methods for SecurityAlert
func (sa *SecurityAlert) IsActive() bool {
	return sa.Status == AlertStatusActive
}

func (sa *SecurityAlert) IsResolved() bool {
	return sa.Status == AlertStatusResolved
}

func (sa *SecurityAlert) IsHighPriority() bool {
	return sa.Severity == AlertSeverityHigh || sa.Severity == AlertSeverityCritical
}

func (sa *SecurityAlert) RequiresAction() bool {
	return sa.ActionRequired && sa.IsActive()
}

func (sa *SecurityAlert) Resolve(resolvedBy, resolution, notes string) {
	sa.Status = AlertStatusResolved
	now := time.Now()
	sa.ResolvedAt = &now
	sa.ResolvedBy = &resolvedBy
	sa.Resolution = &resolution
	if notes != "" {
		sa.ResolutionNotes = &notes
	}
	sa.UpdatedAt = now
}

func (sa *SecurityAlert) StartInvestigation(investigatedBy, notes string) {
	sa.Status = AlertStatusInvestigating
	now := time.Now()
	sa.InvestigatedAt = &now
	sa.InvestigatedBy = &investigatedBy
	if notes != "" {
		sa.InvestigationNotes = &notes
	}
	sa.UpdatedAt = now
}

// Helper methods for ComplianceReport
func (cr *ComplianceReport) IsComplete() bool {
	return cr.Status == ComplianceReportStatusApproved || cr.Status == ComplianceReportStatusSubmitted
}

func (cr *ComplianceReport) RequiresApproval() bool {
	return cr.Status == ComplianceReportStatusPending
}

func (cr *ComplianceReport) GetHighSeverityFindings() []ComplianceFinding {
	var highSeverity []ComplianceFinding
	for _, finding := range cr.Findings {
		if finding.Severity == AlertSeverityHigh || finding.Severity == AlertSeverityCritical {
			highSeverity = append(highSeverity, finding)
		}
	}
	return highSeverity
}

func (cr *ComplianceReport) GetActiveRegulatoryFlags() []RegulatoryFlag {
	var active []RegulatoryFlag
	for _, flag := range cr.RegulatoryFlags {
		if flag.Status == RegulatoryFlagStatusActive {
			active = append(active, flag)
		}
	}
	return active
}

// Helper methods for RegulatoryFlag
func (rf *RegulatoryFlag) IsActive() bool {
	return rf.Status == RegulatoryFlagStatusActive
}

func (rf *RegulatoryFlag) IsOverdue() bool {
	return rf.Deadline != nil && time.Now().After(*rf.Deadline) && rf.IsActive()
}

func (rf *RegulatoryFlag) DaysUntilDeadline() int {
	if rf.Deadline == nil {
		return -1
	}
	return int(time.Until(*rf.Deadline).Hours() / 24)
}
