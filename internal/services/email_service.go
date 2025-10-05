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

// EmailService provides email notification functionality
type EmailService struct {
	provider interfaces.EmailProvider
	config   config.EmailProviderConfig
	logger   interfaces.Logger
}

// NewEmailService creates a new email service
func NewEmailService(cfg config.EmailProviderConfig, logger interfaces.Logger) (*EmailService, error) {
	var provider interfaces.EmailProvider

	switch cfg.Provider {
	case "mock":
		provider = providers.NewMockEmailProvider(cfg)
	default:
		return nil, errors.NewNotificationError(
			errors.ErrorCodeProviderNotFound,
			fmt.Sprintf("unsupported email provider: %s", cfg.Provider),
		)
	}

	service := &EmailService{
		provider: provider,
		config:   cfg,
		logger:   logger,
	}

	return service, nil
}

// SendEmail sends an email notification
func (s *EmailService) SendEmail(ctx context.Context, request *EmailRequest) (*models.NotificationResponse, error) {
	// Validate request first
	if err := s.validateEmailRequest(request); err != nil {
		s.logger.Errorf("Email validation failed: %v", err)
		return nil, err
	}

	s.logger.Infof("Sending email to %v with subject: %s", request.To, request.Subject)

	// Check provider health
	if err := s.provider.IsHealthy(ctx); err != nil {
		s.logger.Errorf("Email provider health check failed: %v", err)
		return nil, err
	}

	// Create email notification
	emailNotification := s.createEmailNotification(request)

	// Apply template if specified
	if request.TemplateID != "" {
		if err := s.applyTemplate(emailNotification, request.TemplateID, request.TemplateData); err != nil {
			s.logger.Errorf("Template application failed: %v", err)
			return nil, err
		}
	}

	// Send email
	response, err := s.provider.SendEmail(ctx, emailNotification)
	if err != nil {
		s.logger.Errorf("Email sending failed: %v", err)
		return nil, err
	}

	s.logger.Infof("Email sent successfully with ID: %s", response.ID)
	return response, nil
}

// SendBulkEmail sends emails to multiple recipients
func (s *EmailService) SendBulkEmail(ctx context.Context, request *BulkEmailRequest) ([]*models.NotificationResponse, error) {
	s.logger.Infof("Sending bulk email to %d recipients", len(request.Recipients))

	if len(request.Recipients) == 0 {
		return nil, errors.NewValidationError("recipients", "at least one recipient is required")
	}

	responses := make([]*models.NotificationResponse, 0, len(request.Recipients))

	for _, recipient := range request.Recipients {
		emailRequest := &EmailRequest{
			To:           []string{recipient.Email},
			Subject:      request.Subject,
			HTMLBody:     request.HTMLBody,
			TextBody:     request.TextBody,
			From:         request.From,
			ReplyTo:      request.ReplyTo,
			Headers:      request.Headers,
			TemplateID:   request.TemplateID,
			TemplateData: s.mergeTemplateData(request.TemplateData, recipient.Data),
			Priority:     request.Priority,
			Metadata:     request.Metadata,
		}

		response, err := s.SendEmail(ctx, emailRequest)
		if err != nil {
			s.logger.Errorf("Failed to send email to %s: %v", recipient.Email, err)
			// Continue with other recipients, but record the error
			response = &models.NotificationResponse{
				ID:     uuid.New(),
				Status: models.StatusFailed,
				Error:  err.Error(),
			}
		}

		responses = append(responses, response)
	}

	s.logger.Infof("Bulk email completed: %d emails processed", len(responses))
	return responses, nil
}

// GetEmailTemplates returns available email templates
func (s *EmailService) GetEmailTemplates() []interfaces.EmailTemplate {
	return s.provider.GetEmailTemplates()
}

// RenderTemplate renders an email template with data
func (s *EmailService) RenderTemplate(templateID string, data map[string]string) (*RenderedTemplate, error) {
	mockProvider, ok := s.provider.(*providers.MockEmailProvider)
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

	return &RenderedTemplate{
		ID:       template.ID,
		Subject:  template.Subject,
		HTMLBody: template.HTMLBody,
		TextBody: template.TextBody,
	}, nil
}

// ValidateEmailAddress validates an email address
func (s *EmailService) ValidateEmailAddress(email string) error {
	return s.provider.ValidateEmailAddress(email)
}

// GetProviderStatus returns the current provider status
func (s *EmailService) GetProviderStatus(ctx context.Context) *ProviderStatus {
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

// validateEmailRequest validates an email request
func (s *EmailService) validateEmailRequest(request *EmailRequest) error {
	if request == nil {
		return errors.NewValidationError("request", "email request is required")
	}

	if len(request.To) == 0 {
		return errors.NewValidationError("to", "at least one recipient is required")
	}

	// Validate all email addresses
	for _, email := range request.To {
		if err := s.provider.ValidateEmailAddress(email); err != nil {
			return errors.NewValidationError("to", fmt.Sprintf("invalid email address: %s", email))
		}
	}

	for _, email := range request.CC {
		if err := s.provider.ValidateEmailAddress(email); err != nil {
			return errors.NewValidationError("cc", fmt.Sprintf("invalid email address: %s", email))
		}
	}

	for _, email := range request.BCC {
		if err := s.provider.ValidateEmailAddress(email); err != nil {
			return errors.NewValidationError("bcc", fmt.Sprintf("invalid email address: %s", email))
		}
	}

	if request.From != "" {
		if err := s.provider.ValidateEmailAddress(request.From); err != nil {
			return errors.NewValidationError("from", "invalid sender email address")
		}
	}

	if request.ReplyTo != "" {
		if err := s.provider.ValidateEmailAddress(request.ReplyTo); err != nil {
			return errors.NewValidationError("reply_to", "invalid reply-to email address")
		}
	}

	// Validate content
	if request.Subject == "" && request.TemplateID == "" {
		return errors.NewValidationError("subject", "email subject is required when not using a template")
	}

	if request.HTMLBody == "" && request.TextBody == "" && request.TemplateID == "" {
		return errors.NewValidationError("body", "email must have either HTML body, text body, or template")
	}

	return nil
}

// createEmailNotification creates an email notification from a request
func (s *EmailService) createEmailNotification(request *EmailRequest) *models.EmailNotification {
	now := time.Now()

	notification := &models.EmailNotification{
		Notification: models.Notification{
			ID:         uuid.New(),
			Type:       models.NotificationTypeEmail,
			Status:     models.StatusPending,
			Priority:   request.Priority,
			Recipient:  request.To[0], // Primary recipient
			Subject:    request.Subject,
			Body:       request.TextBody,
			Metadata:   request.Metadata,
			CreatedAt:  now,
			UpdatedAt:  now,
			RetryCount: 0,
			MaxRetries: 3,
		},
		To:          request.To,
		CC:          request.CC,
		BCC:         request.BCC,
		From:        request.From,
		ReplyTo:     request.ReplyTo,
		HTMLBody:    request.HTMLBody,
		TextBody:    request.TextBody,
		Attachments: request.Attachments,
		Headers:     request.Headers,
	}

	// Set default sender if not provided
	if notification.From == "" {
		notification.From = s.getDefaultSender()
	}

	return notification
}

// applyTemplate applies a template to an email notification
func (s *EmailService) applyTemplate(email *models.EmailNotification, templateID string, data map[string]string) error {
	mockProvider, ok := s.provider.(*providers.MockEmailProvider)
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
	email.Subject = template.Subject
	email.HTMLBody = template.HTMLBody
	email.TextBody = template.TextBody
	email.Body = template.TextBody

	return nil
}

// mergeTemplateData merges global and recipient-specific template data
func (s *EmailService) mergeTemplateData(global, recipient map[string]string) map[string]string {
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

// getDefaultSender returns the default sender email address
func (s *EmailService) getDefaultSender() string {
	if sender, exists := s.config.Settings["default_sender"]; exists {
		return sender
	}
	return "noreply@notification-service.local"
}

// EmailRequest represents a request to send an email
type EmailRequest struct {
	To           []string                 `json:"to" validate:"required,min=1"`
	CC           []string                 `json:"cc,omitempty"`
	BCC          []string                 `json:"bcc,omitempty"`
	From         string                   `json:"from,omitempty"`
	ReplyTo      string                   `json:"reply_to,omitempty"`
	Subject      string                   `json:"subject,omitempty"`
	HTMLBody     string                   `json:"html_body,omitempty"`
	TextBody     string                   `json:"text_body,omitempty"`
	Attachments  []models.EmailAttachment `json:"attachments,omitempty"`
	Headers      map[string]string        `json:"headers,omitempty"`
	TemplateID   string                   `json:"template_id,omitempty"`
	TemplateData map[string]string        `json:"template_data,omitempty"`
	Priority     models.Priority          `json:"priority"`
	Metadata     map[string]string        `json:"metadata,omitempty"`
}

// BulkEmailRequest represents a request to send emails to multiple recipients
type BulkEmailRequest struct {
	Recipients   []BulkEmailRecipient `json:"recipients" validate:"required,min=1"`
	Subject      string               `json:"subject,omitempty"`
	HTMLBody     string               `json:"html_body,omitempty"`
	TextBody     string               `json:"text_body,omitempty"`
	From         string               `json:"from,omitempty"`
	ReplyTo      string               `json:"reply_to,omitempty"`
	Headers      map[string]string    `json:"headers,omitempty"`
	TemplateID   string               `json:"template_id,omitempty"`
	TemplateData map[string]string    `json:"template_data,omitempty"`
	Priority     models.Priority      `json:"priority"`
	Metadata     map[string]string    `json:"metadata,omitempty"`
}

// BulkEmailRecipient represents a recipient in a bulk email request
type BulkEmailRecipient struct {
	Email string            `json:"email" validate:"required,email"`
	Data  map[string]string `json:"data,omitempty"`
}

// RenderedTemplate represents a rendered email template
type RenderedTemplate struct {
	ID       string `json:"id"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

// ProviderStatus represents the status of an email provider
type ProviderStatus struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Healthy bool   `json:"healthy"`
	Error   string `json:"error,omitempty"`
}
