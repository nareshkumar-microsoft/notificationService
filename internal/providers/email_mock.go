package providers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/utils"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
	"github.com/nareshkumar-microsoft/notificationService/pkg/interfaces"
)

// MockEmailProvider implements the EmailProvider interface for testing and development
type MockEmailProvider struct {
	config     config.EmailProviderConfig
	templates  map[string]*EmailTemplate
	sentEmails []SentEmail
	healthy    bool
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Subject   string            `json:"subject"`
	HTMLBody  string            `json:"html_body"`
	TextBody  string            `json:"text_body"`
	Variables []string          `json:"variables"`
	Category  string            `json:"category"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SentEmail represents an email that was sent (for mock tracking)
type SentEmail struct {
	ID           uuid.UUID         `json:"id"`
	To           []string          `json:"to"`
	CC           []string          `json:"cc,omitempty"`
	BCC          []string          `json:"bcc,omitempty"`
	From         string            `json:"from"`
	Subject      string            `json:"subject"`
	HTMLBody     string            `json:"html_body,omitempty"`
	TextBody     string            `json:"text_body,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	SentAt       time.Time         `json:"sent_at"`
	Status       string            `json:"status"`
	ProviderData map[string]string `json:"provider_data,omitempty"`
}

// NewMockEmailProvider creates a new mock email provider
func NewMockEmailProvider(cfg config.EmailProviderConfig) *MockEmailProvider {
	provider := &MockEmailProvider{
		config:     cfg,
		templates:  make(map[string]*EmailTemplate),
		sentEmails: make([]SentEmail, 0),
		healthy:    true,
	}

	// Load default templates
	provider.loadDefaultTemplates()

	return provider
}

// Send implements the NotificationProvider interface
func (p *MockEmailProvider) Send(ctx context.Context, notification *models.Notification) (*models.NotificationResponse, error) {
	if !p.healthy {
		return nil, errors.NewProviderError("mock-email", errors.ErrorCodeProviderUnavailable, "provider is unhealthy")
	}

	// Convert generic notification to email notification
	emailNotification, err := p.convertToEmailNotification(notification)
	if err != nil {
		return nil, err
	}

	return p.SendEmail(ctx, emailNotification)
}

// SendEmail implements the EmailProvider interface
func (p *MockEmailProvider) SendEmail(ctx context.Context, email *models.EmailNotification) (*models.NotificationResponse, error) {
	if !p.healthy {
		return nil, errors.NewProviderError("mock-email", errors.ErrorCodeProviderUnavailable, "provider is unhealthy")
	}

	// Validate email
	if err := p.validateEmailNotification(email); err != nil {
		return nil, err
	}

	// Simulate processing delay
	select {
	case <-ctx.Done():
		return nil, errors.NewNotificationError(errors.ErrorCodeTimeout, "email sending timed out")
	case <-time.After(100 * time.Millisecond):
		// Continue processing
	}

	// Create sent email record
	sentEmail := SentEmail{
		ID:       email.ID,
		To:       email.To,
		CC:       email.CC,
		BCC:      email.BCC,
		From:     email.From,
		Subject:  email.Subject,
		HTMLBody: email.HTMLBody,
		TextBody: email.TextBody,
		Headers:  email.Headers,
		SentAt:   time.Now(),
		Status:   "sent",
		ProviderData: map[string]string{
			"provider":    "mock-email",
			"message_id":  fmt.Sprintf("mock-%s", email.ID.String()),
			"queue_time":  "100ms",
			"retry_count": "0",
		},
	}

	// Store sent email for tracking
	p.sentEmails = append(p.sentEmails, sentEmail)

	// Create response
	now := time.Now()
	response := &models.NotificationResponse{
		ID:         email.ID,
		Status:     models.StatusSent,
		Message:    fmt.Sprintf("Email successfully sent to %d recipients", len(email.To)),
		ProviderID: sentEmail.ProviderData["message_id"],
		SentAt:     &now,
	}

	return response, nil
}

// ValidateEmailAddress implements the EmailProvider interface
func (p *MockEmailProvider) ValidateEmailAddress(email string) error {
	return utils.ValidateEmailAddress(email)
}

// GetEmailTemplates implements the EmailProvider interface
func (p *MockEmailProvider) GetEmailTemplates() []interfaces.EmailTemplate {
	templates := make([]interfaces.EmailTemplate, 0, len(p.templates))
	for _, template := range p.templates {
		templates = append(templates, interfaces.EmailTemplate{
			ID:        template.ID,
			Name:      template.Name,
			Subject:   template.Subject,
			HTMLBody:  template.HTMLBody,
			TextBody:  template.TextBody,
			Variables: template.Variables,
			Category:  template.Category,
			CreatedAt: template.CreatedAt.Format(time.RFC3339),
			UpdatedAt: template.UpdatedAt.Format(time.RFC3339),
		})
	}
	return templates
}

// GetType implements the NotificationProvider interface
func (p *MockEmailProvider) GetType() models.NotificationType {
	return models.NotificationTypeEmail
}

// IsHealthy implements the NotificationProvider interface
func (p *MockEmailProvider) IsHealthy(ctx context.Context) error {
	if !p.healthy {
		return errors.NewProviderError("mock-email", errors.ErrorCodeProviderUnavailable, "provider is marked as unhealthy")
	}

	// Simulate health check delay
	select {
	case <-ctx.Done():
		return errors.NewNotificationError(errors.ErrorCodeTimeout, "health check timed out")
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

// GetConfig implements the NotificationProvider interface
func (p *MockEmailProvider) GetConfig() interfaces.ProviderConfig {
	return interfaces.ProviderConfig{
		Name:       "Mock Email Provider",
		Type:       models.NotificationTypeEmail,
		Enabled:    p.config.Enabled,
		Priority:   1,
		MaxRetries: 3,
		Timeout:    30,
		RateLimit: interfaces.RateLimitConfig{
			Enabled:        true,
			RequestsPerMin: 100,
			BurstSize:      10,
		},
		Settings: map[string]string{
			"provider_type": "mock",
			"version":       "1.0.0",
			"features":      "templates,validation,tracking",
		},
	}
}

// GetTemplate retrieves an email template by ID
func (p *MockEmailProvider) GetTemplate(templateID string) (*EmailTemplate, error) {
	template, exists := p.templates[templateID]
	if !exists {
		return nil, errors.NewNotificationError(errors.ErrorCodeTemplateNotFound, fmt.Sprintf("template not found: %s", templateID))
	}
	return template, nil
}

// AddTemplate adds a new email template
func (p *MockEmailProvider) AddTemplate(template *EmailTemplate) error {
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	p.templates[template.ID] = template
	return nil
}

// RenderTemplate renders an email template with provided data
func (p *MockEmailProvider) RenderTemplate(templateID string, data map[string]string) (*EmailTemplate, error) {
	template, err := p.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Clone template for rendering
	rendered := &EmailTemplate{
		ID:        template.ID,
		Name:      template.Name,
		Subject:   p.replaceVariables(template.Subject, data),
		HTMLBody:  p.replaceVariables(template.HTMLBody, data),
		TextBody:  p.replaceVariables(template.TextBody, data),
		Variables: template.Variables,
		Category:  template.Category,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
		Metadata:  template.Metadata,
	}

	return rendered, nil
}

// GetSentEmails returns all sent emails (for testing)
func (p *MockEmailProvider) GetSentEmails() []SentEmail {
	return p.sentEmails
}

// ClearSentEmails clears the sent emails history (for testing)
func (p *MockEmailProvider) ClearSentEmails() {
	p.sentEmails = make([]SentEmail, 0)
}

// SetHealthy sets the provider health status (for testing)
func (p *MockEmailProvider) SetHealthy(healthy bool) {
	p.healthy = healthy
}

// convertToEmailNotification converts a generic notification to an email notification
func (p *MockEmailProvider) convertToEmailNotification(notification *models.Notification) (*models.EmailNotification, error) {
	if notification.Type != models.NotificationTypeEmail {
		return nil, errors.NewValidationError("type", "notification type must be email")
	}

	emailNotification := &models.EmailNotification{
		Notification: *notification,
		To:           []string{notification.Recipient},
		From:         p.getDefaultSender(),
		HTMLBody:     notification.Body,
		TextBody:     notification.Body,
	}

	return emailNotification, nil
}

// validateEmailNotification validates an email notification
func (p *MockEmailProvider) validateEmailNotification(email *models.EmailNotification) error {
	// Validate To addresses
	if len(email.To) == 0 {
		return errors.NewValidationError("to", "at least one recipient is required")
	}

	for _, addr := range email.To {
		if err := p.ValidateEmailAddress(addr); err != nil {
			return errors.NewValidationError("to", fmt.Sprintf("invalid email address: %s", addr))
		}
	}

	// Validate CC addresses
	for _, addr := range email.CC {
		if err := p.ValidateEmailAddress(addr); err != nil {
			return errors.NewValidationError("cc", fmt.Sprintf("invalid email address: %s", addr))
		}
	}

	// Validate BCC addresses
	for _, addr := range email.BCC {
		if err := p.ValidateEmailAddress(addr); err != nil {
			return errors.NewValidationError("bcc", fmt.Sprintf("invalid email address: %s", addr))
		}
	}

	// Validate From address
	if email.From != "" {
		if err := p.ValidateEmailAddress(email.From); err != nil {
			return errors.NewValidationError("from", "invalid sender email address")
		}
	}

	// Validate ReplyTo address
	if email.ReplyTo != "" {
		if err := p.ValidateEmailAddress(email.ReplyTo); err != nil {
			return errors.NewValidationError("reply_to", "invalid reply-to email address")
		}
	}

	// Validate content
	if email.Subject == "" {
		return errors.NewValidationError("subject", "email subject is required")
	}

	if email.HTMLBody == "" && email.TextBody == "" {
		return errors.NewValidationError("body", "email must have either HTML or text body")
	}

	return nil
}

// getDefaultSender returns the default sender email address
func (p *MockEmailProvider) getDefaultSender() string {
	if sender, exists := p.config.Settings["default_sender"]; exists {
		return sender
	}
	return "noreply@notification-service.local"
}

// replaceVariables replaces template variables with provided data
func (p *MockEmailProvider) replaceVariables(template string, data map[string]string) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// loadDefaultTemplates loads default email templates
func (p *MockEmailProvider) loadDefaultTemplates() {
	// Welcome email template
	welcomeTemplate := &EmailTemplate{
		ID:      "welcome",
		Name:    "Welcome Email",
		Subject: "Welcome to {{service_name}}, {{user_name}}!",
		HTMLBody: `
			<html>
				<body>
					<h1>Welcome {{user_name}}!</h1>
					<p>Thank you for joining {{service_name}}. We're excited to have you on board!</p>
					<p>Your account email: {{user_email}}</p>
					<p>Best regards,<br>The {{service_name}} Team</p>
				</body>
			</html>
		`,
		TextBody: `
			Welcome {{user_name}}!
			
			Thank you for joining {{service_name}}. We're excited to have you on board!
			
			Your account email: {{user_email}}
			
			Best regards,
			The {{service_name}} Team
		`,
		Variables: []string{"user_name", "user_email", "service_name"},
		Category:  "onboarding",
	}

	// Password reset template
	resetTemplate := &EmailTemplate{
		ID:      "password_reset",
		Name:    "Password Reset",
		Subject: "Reset your {{service_name}} password",
		HTMLBody: `
			<html>
				<body>
					<h1>Password Reset Request</h1>
					<p>Hi {{user_name}},</p>
					<p>You requested a password reset for your {{service_name}} account.</p>
					<p><a href="{{reset_link}}" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a></p>
					<p>This link will expire in {{expiry_time}}.</p>
					<p>If you didn't request this reset, please ignore this email.</p>
				</body>
			</html>
		`,
		TextBody: `
			Password Reset Request
			
			Hi {{user_name}},
			
			You requested a password reset for your {{service_name}} account.
			
			Reset your password: {{reset_link}}
			
			This link will expire in {{expiry_time}}.
			
			If you didn't request this reset, please ignore this email.
		`,
		Variables: []string{"user_name", "service_name", "reset_link", "expiry_time"},
		Category:  "security",
	}

	// Notification email template
	notificationTemplate := &EmailTemplate{
		ID:      "notification",
		Name:    "General Notification",
		Subject: "{{notification_title}}",
		HTMLBody: `
			<html>
				<body>
					<h2>{{notification_title}}</h2>
					<p>{{notification_message}}</p>
					<p><em>Sent at {{timestamp}}</em></p>
				</body>
			</html>
		`,
		TextBody: `
			{{notification_title}}
			
			{{notification_message}}
			
			Sent at {{timestamp}}
		`,
		Variables: []string{"notification_title", "notification_message", "timestamp"},
		Category:  "general",
	}

	p.AddTemplate(welcomeTemplate)
	p.AddTemplate(resetTemplate)
	p.AddTemplate(notificationTemplate)
}
