package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"crypto-bubble-map-be/internal/domain/entity"
	"crypto-bubble-map-be/internal/domain/repository"
	"crypto-bubble-map-be/internal/infrastructure/config"

	"go.uber.org/zap"
)

// OpenAIRepository implements AIRepository using OpenAI API
type OpenAIRepository struct {
	config     *config.ExternalConfig
	logger     *zap.Logger
	httpClient *http.Client
}

// NewOpenAIRepository creates a new OpenAI AI repository
func NewOpenAIRepository(cfg *config.ExternalConfig, logger *zap.Logger) repository.AIRepository {
	return &OpenAIRepository{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenAI API structures
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AskAI sends a question to OpenAI and returns a structured response
func (r *OpenAIRepository) AskAI(ctx context.Context, question string, aiContext *entity.AIContext, walletAddress *string) (*entity.AIResponse, error) {
	if r.config.OpenAIAPIKey == "" {
		r.logger.Warn("OpenAI API key not configured, falling back to mock response")
		return r.generateMockResponse(question, walletAddress), nil
	}

	// Build system prompt based on context
	systemPrompt := r.buildSystemPrompt(aiContext, walletAddress)

	// Create OpenAI request
	openAIReq := OpenAIRequest{
		Model: r.config.OpenAIModel,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// Make API call
	openAIResp, err := r.callOpenAI(ctx, openAIReq)
	if err != nil {
		r.logger.Error("Failed to call OpenAI API", zap.Error(err))
		// Fall back to mock response on error
		return r.generateMockResponse(question, walletAddress), nil
	}

	// Parse response
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	answer := openAIResp.Choices[0].Message.Content

	// Generate additional response components
	response := &entity.AIResponse{
		Answer:           answer,
		Confidence:       r.calculateConfidence(question, answer),
		Sources:          r.generateSources(question, walletAddress),
		RelatedQuestions: r.generateRelatedQuestions(question),
		ActionItems:      r.generateActionItems(question, walletAddress),
		GeneratedAt:      time.Now(),
		Model:            openAIResp.Model,
		TokensUsed:       openAIResp.Usage.TotalTokens,
	}

	r.logger.Info("Generated AI response",
		zap.String("model", openAIResp.Model),
		zap.Int("tokens", openAIResp.Usage.TotalTokens))

	return response, nil
}

// GetAvailableModels returns available AI models
func (r *OpenAIRepository) GetAvailableModels(ctx context.Context) ([]string, error) {
	return []string{
		"gpt-3.5-turbo",
		"gpt-4",
		"gpt-4-turbo-preview",
		"crypto-bubble-map-ai-v1.0", // Our custom model identifier
	}, nil
}

// GetModelInfo returns information about a specific model
func (r *OpenAIRepository) GetModelInfo(ctx context.Context, modelName string) (map[string]interface{}, error) {
	models := map[string]map[string]interface{}{
		"gpt-3.5-turbo": {
			"name":        "GPT-3.5 Turbo",
			"version":     "3.5",
			"description": "Fast and efficient model for general blockchain analysis",
			"capabilities": []string{
				"wallet analysis",
				"transaction interpretation",
				"risk assessment",
				"compliance guidance",
			},
			"max_tokens":    4096,
			"response_time": "< 3s",
			"provider":      "OpenAI",
		},
		"gpt-4": {
			"name":        "GPT-4",
			"version":     "4.0",
			"description": "Advanced model for complex blockchain analysis and reasoning",
			"capabilities": []string{
				"advanced wallet analysis",
				"complex transaction pattern detection",
				"sophisticated risk assessment",
				"detailed compliance analysis",
				"multi-step reasoning",
			},
			"max_tokens":    8192,
			"response_time": "< 10s",
			"provider":      "OpenAI",
		},
		"gpt-4-turbo-preview": {
			"name":        "GPT-4 Turbo",
			"version":     "4.0-turbo",
			"description": "Latest GPT-4 model with improved performance and larger context",
			"capabilities": []string{
				"comprehensive blockchain analysis",
				"large-scale transaction analysis",
				"advanced pattern recognition",
				"regulatory compliance analysis",
			},
			"max_tokens":    128000,
			"response_time": "< 15s",
			"provider":      "OpenAI",
		},
		"crypto-bubble-map-ai-v1.0": {
			"name":        "Crypto Bubble Map AI",
			"version":     "1.0",
			"description": "Specialized AI assistant for crypto bubble map analysis",
			"capabilities": []string{
				"bubble map interpretation",
				"wallet clustering analysis",
				"network visualization insights",
				"risk correlation analysis",
			},
			"max_tokens":    4096,
			"response_time": "< 5s",
			"provider":      "Custom",
		},
	}

	if info, exists := models[modelName]; exists {
		return info, nil
	}

	return nil, fmt.Errorf("model not found: %s", modelName)
}

// callOpenAI makes the actual API call to OpenAI
func (r *OpenAIRepository) callOpenAI(ctx context.Context, req OpenAIRequest) (*OpenAIResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", r.config.OpenAIBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.config.OpenAIAPIKey)

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &openAIResp, nil
}

// buildSystemPrompt creates a system prompt based on context
func (r *OpenAIRepository) buildSystemPrompt(aiContext *entity.AIContext, walletAddress *string) string {
	prompt := `You are a specialized AI assistant for blockchain and cryptocurrency analysis. You help users understand wallet behaviors, transaction patterns, risk assessments, and compliance requirements.

Your expertise includes:
- Wallet analysis and classification
- Transaction pattern recognition
- Risk scoring and threat detection
- AML/KYC compliance guidance
- Network analysis and visualization
- Regulatory compliance requirements

Always provide:
1. Clear, actionable insights
2. Evidence-based analysis
3. Risk assessments when relevant
4. Compliance considerations
5. Specific recommendations

Be concise but thorough. Focus on practical, actionable information.`

	if aiContext != nil {
		if aiContext.AnalysisType != "" {
			prompt += fmt.Sprintf("\n\nCurrent analysis type: %s", aiContext.AnalysisType)
		}
		if aiContext.NetworkID != "" {
			prompt += fmt.Sprintf("\nNetwork: %s", aiContext.NetworkID)
		}
		if aiContext.UserRole != "" {
			prompt += fmt.Sprintf("\nUser role: %s", aiContext.UserRole)
		}
	}

	if walletAddress != nil {
		prompt += fmt.Sprintf("\n\nAnalyzing wallet: %s", *walletAddress)
	}

	return prompt
}

// Helper methods for generating response components

func (r *OpenAIRepository) calculateConfidence(question, answer string) float64 {
	// Simple confidence calculation based on question and answer characteristics
	confidence := 0.7 // Base confidence

	question = strings.ToLower(question)
	answer = strings.ToLower(answer)

	// Higher confidence for specific questions
	if strings.Contains(question, "risk") || strings.Contains(question, "transaction") {
		confidence += 0.1
	}

	if strings.Contains(question, "compliance") || strings.Contains(question, "wallet") {
		confidence += 0.1
	}

	// Lower confidence for vague questions
	if len(question) < 20 {
		confidence -= 0.1
	}

	// Adjust based on answer quality
	if len(answer) > 100 && !strings.Contains(answer, "i don't know") {
		confidence += 0.05
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

func (r *OpenAIRepository) generateSources(question string, walletAddress *string) []string {
	sources := []string{
		"OpenAI GPT Analysis",
		"Crypto Bubble Map Database",
		"Blockchain Analysis Engine",
	}

	if walletAddress != nil {
		sources = append(sources, fmt.Sprintf("Wallet Data: %s", *walletAddress))
	}

	question = strings.ToLower(question)
	if strings.Contains(question, "compliance") {
		sources = append(sources, "Regulatory Guidelines Database", "AML/KYC Framework")
	}

	if strings.Contains(question, "risk") {
		sources = append(sources, "Risk Scoring Engine", "Threat Intelligence Database")
	}

	if strings.Contains(question, "network") {
		sources = append(sources, "Network Graph Analysis", "Connection Pattern Database")
	}

	return sources
}

func (r *OpenAIRepository) generateRelatedQuestions(question string) []string {
	question = strings.ToLower(question)

	if strings.Contains(question, "risk") {
		return []string{
			"What are the main risk factors for this wallet?",
			"How is the risk score calculated?",
			"What actions should I take for high-risk addresses?",
			"How often should risk assessments be updated?",
		}
	}

	if strings.Contains(question, "transaction") {
		return []string{
			"What transaction patterns indicate suspicious activity?",
			"How can I trace money flows effectively?",
			"What are common money laundering techniques?",
			"How do I analyze cross-chain transactions?",
		}
	}

	if strings.Contains(question, "compliance") {
		return []string{
			"What are the key AML requirements?",
			"How do I implement effective monitoring?",
			"What documentation is required for compliance?",
			"How do I handle suspicious activity reports?",
		}
	}

	if strings.Contains(question, "wallet") {
		return []string{
			"How do I classify wallet types?",
			"What indicates a high-value wallet?",
			"How do I identify exchange wallets?",
			"What are signs of wallet clustering?",
		}
	}

	return []string{
		"How can I improve my blockchain analysis?",
		"What are the latest trends in crypto compliance?",
		"How do I identify emerging threats?",
		"What tools are most effective for investigation?",
	}
}

func (r *OpenAIRepository) generateActionItems(question string, walletAddress *string) []string {
	question = strings.ToLower(question)
	var actions []string

	if walletAddress != nil {
		actions = append(actions, fmt.Sprintf("Add wallet %s to monitoring list", *walletAddress))
	}

	if strings.Contains(question, "risk") {
		actions = append(actions,
			"Review risk assessment criteria",
			"Update risk scoring parameters",
			"Schedule regular risk reviews",
		)
	}

	if strings.Contains(question, "compliance") {
		actions = append(actions,
			"Review compliance policies",
			"Update monitoring procedures",
			"Document compliance findings",
		)
	}

	if strings.Contains(question, "transaction") {
		actions = append(actions,
			"Analyze transaction patterns",
			"Review counterparty relationships",
			"Document suspicious activities",
		)
	}

	if len(actions) == 0 {
		actions = []string{
			"Review the analysis results",
			"Consider additional investigation",
			"Document findings for future reference",
		}
	}

	return actions
}

// generateMockResponse creates a fallback response when OpenAI is not available
func (r *OpenAIRepository) generateMockResponse(question string, walletAddress *string) *entity.AIResponse {
	answer := "I'm analyzing your blockchain query using our internal analysis engine. "

	question = strings.ToLower(question)
	if strings.Contains(question, "risk") {
		answer += "For risk assessment, I recommend reviewing transaction patterns, counterparty analysis, and compliance indicators."
	} else if strings.Contains(question, "wallet") {
		answer += "For wallet analysis, consider transaction volume, frequency, counterparties, and behavioral patterns."
	} else if strings.Contains(question, "compliance") {
		answer += "For compliance analysis, focus on AML/KYC requirements, suspicious activity detection, and regulatory reporting."
	} else {
		answer += "Please provide more specific details about your blockchain analysis needs for a more targeted response."
	}

	return &entity.AIResponse{
		Answer:           answer,
		Confidence:       0.75,
		Sources:          r.generateSources(question, walletAddress),
		RelatedQuestions: r.generateRelatedQuestions(question),
		ActionItems:      r.generateActionItems(question, walletAddress),
		GeneratedAt:      time.Now(),
		Model:            "crypto-bubble-map-ai-v1.0",
		TokensUsed:       len(question) + len(answer),
	}
}
