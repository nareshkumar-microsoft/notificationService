package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/providers"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
	"github.com/nareshkumar-microsoft/notificationService/pkg/interfaces"
)

// SMSService provides SMS notification functionality
type SMSService struct {
	provider interfaces.SMSProvider
	config   config.SMSProviderConfig
	logger   interfaces.Logger
}

// NewSMSService creates a new SMS service
func NewSMSService(cfg config.SMSProviderConfig, logger interfaces.Logger) (*SMSService, error) {
	var provider interfaces.SMSProvider

	switch cfg.Provider {
	case "mock":
		provider = providers.NewMockSMSProvider(cfg)
	default:
		return nil, errors.NewNotificationError(
			errors.ErrorCodeProviderNotFound,
			fmt.Sprintf("unsupported SMS provider: %s", cfg.Provider),
		)
	}

	service := &SMSService{
		provider: provider,
		config:   cfg,
		logger:   logger,
	}

	return service, nil
}

// SendSMS sends an SMS notification
func (s *SMSService) SendSMS(ctx context.Context, request *SMSRequest) (*models.NotificationResponse, error) {
	// Validate request first
	if err := s.validateSMSRequest(request); err != nil {
		s.logger.Errorf("SMS validation failed: %v", err)
		return nil, err
	}

	s.logger.Infof("Sending SMS to %s with message: %s", request.PhoneNumber, truncateMessage(request.Message, 50))

	// Check provider health
	if err := s.provider.IsHealthy(ctx); err != nil {
		s.logger.Errorf("SMS provider health check failed: %v", err)
		return nil, err
	}

	// Create SMS notification
	smsNotification := s.createSMSNotification(request)

	// Apply template if specified
	if request.TemplateID != "" {
		if err := s.applyTemplate(smsNotification, request.TemplateID, request.TemplateData); err != nil {
			s.logger.Errorf("Template application failed: %v", err)
			return nil, err
		}
	}

	// Send SMS
	response, err := s.provider.SendSMS(ctx, smsNotification)
	if err != nil {
		s.logger.Errorf("SMS sending failed: %v", err)
		return nil, err
	}

	s.logger.Infof("SMS sent successfully with ID: %s", response.ID)
	return response, nil
}

// SendBulkSMS sends SMS messages to multiple recipients
func (s *SMSService) SendBulkSMS(ctx context.Context, request *BulkSMSRequest) ([]*models.NotificationResponse, error) {
	s.logger.Infof("Sending bulk SMS to %d recipients", len(request.Recipients))

	if len(request.Recipients) == 0 {
		return nil, errors.NewValidationError("recipients", "at least one recipient is required")
	}

	responses := make([]*models.NotificationResponse, 0, len(request.Recipients))

	for _, recipient := range request.Recipients {
		smsRequest := &SMSRequest{
			PhoneNumber:  recipient.PhoneNumber,
			CountryCode:  recipient.CountryCode,
			Message:      request.Message,
			Unicode:      request.Unicode,
			TemplateID:   request.TemplateID,
			TemplateData: s.mergeTemplateData(request.TemplateData, recipient.Data),
			Priority:     request.Priority,
			Metadata:     request.Metadata,
		}

		response, err := s.SendSMS(ctx, smsRequest)
		if err != nil {
			s.logger.Errorf("Failed to send SMS to %s: %v", recipient.PhoneNumber, err)
			// Continue with other recipients, but record the error
			response = &models.NotificationResponse{
				ID:     uuid.New(),
				Status: models.StatusFailed,
				Error:  err.Error(),
			}
		}

		responses = append(responses, response)
	}

	s.logger.Infof("Bulk SMS completed: %d messages processed", len(responses))
	return responses, nil
}

// GetSMSCost returns the cost of sending an SMS to a specific country
func (s *SMSService) GetSMSCost(countryCode string) (float64, error) {
	return s.provider.GetSMSCost(countryCode)
}

// GetSupportedCountries returns list of supported countries
func (s *SMSService) GetSupportedCountries() []CountryInfo {
	mockProvider, ok := s.provider.(*providers.MockSMSProvider)
	if !ok {
		return []CountryInfo{}
	}

	countries := mockProvider.GetSupportedCountries()
	result := make([]CountryInfo, len(countries))
	for i, country := range countries {
		result[i] = CountryInfo{
			Code:      country.Code,
			Name:      country.Name,
			Cost:      country.Cost,
			MaxLength: country.MaxLength,
			Supported: country.Supported,
		}
	}
	return result
}

// RenderTemplate renders an SMS template with data
func (s *SMSService) RenderTemplate(templateID string, data map[string]string) (*RenderedSMSTemplate, error) {
	mockProvider, ok := s.provider.(*providers.MockSMSProvider)
	if !ok {
		return nil, errors.NewNotificationError(
			errors.ErrorCodeProviderNotFound,
			"template rendering not supported by this provider",
		)
	}

	template, err := mockProvider.RenderTemplate(templateID, data)
	if err != nil {
		return nil, err
	}

	return &RenderedSMSTemplate{
		ID:        template.ID,
		Message:   template.Message,
		MaxLength: template.MaxLength,
		Unicode:   template.Unicode,
		Segments:  calculateSMSSegments(template.Message, template.Unicode),
	}, nil
}

// ValidatePhoneNumber validates a phone number
func (s *SMSService) ValidatePhoneNumber(phoneNumber, countryCode string) error {
	return s.provider.ValidatePhoneNumber(phoneNumber, countryCode)
}

// GetProviderStatus returns the current provider status
func (s *SMSService) GetProviderStatus(ctx context.Context) *ProviderStatus {
	status := &ProviderStatus{
		Name:    s.provider.GetConfig().Name,
		Type:    string(s.provider.GetType()),
		Healthy: true,
	}

	if err := s.provider.IsHealthy(ctx); err != nil {
		status.Healthy = false
		status.Error = err.Error()
	}

	return status
}

// EstimateCost estimates the cost of sending an SMS
func (s *SMSService) EstimateCost(message string, countryCode string, unicode bool) (*SMSCostEstimate, error) {
	segments := calculateSMSSegments(message, unicode)
	costPerSegment, err := s.provider.GetSMSCost(countryCode)
	if err != nil {
		return nil, err
	}

	totalCost := costPerSegment * float64(segments)

	return &SMSCostEstimate{
		Segments:       segments,
		CostPerSegment: costPerSegment,
		TotalCost:      totalCost,
		Unicode:        unicode,
		CountryCode:    countryCode,
		MessageLength:  len(message),
	}, nil
}

// validateSMSRequest validates an SMS request
func (s *SMSService) validateSMSRequest(request *SMSRequest) error {
	if request == nil {
		return errors.NewValidationError("request", "SMS request is required")
	}

	if request.PhoneNumber == "" {
		return errors.NewValidationError("phone_number", "phone number is required")
	}

	// Validate phone number
	if err := s.provider.ValidatePhoneNumber(request.PhoneNumber, request.CountryCode); err != nil {
		return err
	}

	// Validate message content
	if request.Message == "" && request.TemplateID == "" {
		return errors.NewValidationError("message", "SMS message is required when not using a template")
	}

	// Check message length (allow up to 10 segments)
	maxLength := 160 * 10
	if request.Unicode {
		maxLength = 70 * 10
	}

	if len(request.Message) > maxLength {
		return errors.NewValidationError("message", fmt.Sprintf("message too long (max %d characters for 10 segments)", maxLength))
	}

	return nil
}

// createSMSNotification creates an SMS notification from a request
func (s *SMSService) createSMSNotification(request *SMSRequest) *models.SMSNotification {
	now := time.Now()

	notification := &models.SMSNotification{
		Notification: models.Notification{
			ID:         uuid.New(),
			Type:       models.NotificationTypeSMS,
			Status:     models.StatusPending,
			Priority:   request.Priority,
			Recipient:  request.PhoneNumber,
			Subject:    "SMS Notification",
			Body:       request.Message,
			Metadata:   request.Metadata,
			CreatedAt:  now,
			UpdatedAt:  now,
			RetryCount: 0,
			MaxRetries: 3,
		},
		PhoneNumber: request.PhoneNumber,
		CountryCode: request.CountryCode,
		Message:     request.Message,
		Unicode:     request.Unicode,
	}

	// Add country code to metadata if provided
	if notification.Metadata == nil {
		notification.Metadata = make(map[string]string)
	}
	if request.CountryCode != "" {
		notification.Metadata["country_code"] = request.CountryCode
	}

	return notification
}

// applyTemplate applies a template to an SMS notification
func (s *SMSService) applyTemplate(sms *models.SMSNotification, templateID string, data map[string]string) error {
	mockProvider, ok := s.provider.(*providers.MockSMSProvider)
	if !ok {
		return errors.NewNotificationError(
			errors.ErrorCodeProviderNotFound,
			"template rendering not supported by this provider",
		)
	}

	template, err := mockProvider.RenderTemplate(templateID, data)
	if err != nil {
		return err
	}

	// Apply template content
	sms.Message = template.Message
	sms.Body = template.Message
	sms.Unicode = template.Unicode

	return nil
}

// mergeTemplateData merges global and recipient-specific template data
func (s *SMSService) mergeTemplateData(global, recipient map[string]string) map[string]string {
	merged := make(map[string]string)

	// Add global data first
	for key, value := range global {
		merged[key] = value
	}

	// Override with recipient-specific data
	for key, value := range recipient {
		merged[key] = value
	}

	return merged
}

// Helper functions

// truncateMessage truncates a message to a maximum length for logging
func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength-3] + "..."
}

// calculateSMSSegments calculates the number of SMS segments needed
func calculateSMSSegments(message string, unicode bool) int {
	maxLength := 160
	if unicode {
		maxLength = 70
	}

	length := len(message)
	if length <= maxLength {
		return 1
	}

	// For multi-part messages, each segment is slightly shorter
	segmentLength := maxLength - 7 // Account for UDH (User Data Header)
	if unicode {
		segmentLength = 67
	}

	return (length + segmentLength - 1) / segmentLength
}

// Request and response types

// SMSRequest represents a request to send an SMS
type SMSRequest struct {
	PhoneNumber  string            `json:"phone_number" validate:"required"`
	CountryCode  string            `json:"country_code,omitempty"`
	Message      string            `json:"message,omitempty"`
	Unicode      bool              `json:"unicode"`
	TemplateID   string            `json:"template_id,omitempty"`
	TemplateData map[string]string `json:"template_data,omitempty"`
	Priority     models.Priority   `json:"priority"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// BulkSMSRequest represents a request to send SMS to multiple recipients
type BulkSMSRequest struct {
	Recipients   []BulkSMSRecipient `json:"recipients" validate:"required,min=1"`
	Message      string             `json:"message,omitempty"`
	Unicode      bool               `json:"unicode"`
	TemplateID   string             `json:"template_id,omitempty"`
	TemplateData map[string]string  `json:"template_data,omitempty"`
	Priority     models.Priority    `json:"priority"`
	Metadata     map[string]string  `json:"metadata,omitempty"`
}

// BulkSMSRecipient represents a recipient in a bulk SMS request
type BulkSMSRecipient struct {
	PhoneNumber string            `json:"phone_number" validate:"required"`
	CountryCode string            `json:"country_code,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
}

// RenderedSMSTemplate represents a rendered SMS template
type RenderedSMSTemplate struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	MaxLength int    `json:"max_length"`
	Unicode   bool   `json:"unicode"`
	Segments  int    `json:"segments"`
}

// CountryInfo represents information about SMS support for a country
type CountryInfo struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Cost      float64 `json:"cost"`
	MaxLength int     `json:"max_length"`
	Supported bool    `json:"supported"`
}

// SMSCostEstimate represents a cost estimate for an SMS
type SMSCostEstimate struct {
	Segments       int     `json:"segments"`
	CostPerSegment float64 `json:"cost_per_segment"`
	TotalCost      float64 `json:"total_cost"`
	Unicode        bool    `json:"unicode"`
	CountryCode    string  `json:"country_code"`
	MessageLength  int     `json:"message_length"`
}
