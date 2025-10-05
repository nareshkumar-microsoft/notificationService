package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/utils"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
)

func TestNewEmailService(t *testing.T) {
	cfg := config.EmailProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("info")
	
	service, err := NewEmailService(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
}

func TestNewEmailService_UnsupportedProvider(t *testing.T) {
	cfg := config.EmailProviderConfig{
		Provider: "unsupported",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("info")
	
	service, err := NewEmailService(cfg, logger)
	assert.Error(t, err)
	assert.Nil(t, service)
	
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeProviderNotFound, notifErr.Code)
}

func TestEmailService_SendEmail_Success(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	request := &EmailRequest{
		To:       []string{"test@example.com"},
		Subject:  "Test Email",
		HTMLBody: "<p>Test content</p>",
		TextBody: "Test content",
		Priority: models.PriorityNormal,
	}
	
	response, err := service.SendEmail(ctx, request)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
	assert.Contains(t, response.Message, "successfully sent")
}

func TestEmailService_SendEmail_ValidationErrors(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		request *EmailRequest
	}{
		{
			name:    "nil request",
			request: nil,
		},
		{
			name: "empty recipients",
			request: &EmailRequest{
				To:      []string{},
				Subject: "Test",
			},
		},
		{
			name: "invalid email",
			request: &EmailRequest{
				To:      []string{"invalid-email"},
				Subject: "Test",
			},
		},
		{
			name: "no content and no template",
			request: &EmailRequest{
				To: []string{"test@example.com"},
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.SendEmail(ctx, tt.request)
			assert.Error(t, err)
			assert.Nil(t, response)
			
			notifErr, ok := errors.AsNotificationError(err)
			require.True(t, ok)
			assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
		})
	}
}

func TestEmailService_SendEmail_WithTemplate(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	request := &EmailRequest{
		To:         []string{"test@example.com"},
		TemplateID: "welcome",
		TemplateData: map[string]string{
			"user_name":    "John Doe",
			"user_email":   "john@example.com",
			"service_name": "Test Service",
		},
		Priority: models.PriorityNormal,
	}
	
	response, err := service.SendEmail(ctx, request)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
}

func TestEmailService_SendBulkEmail(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	request := &BulkEmailRequest{
		Recipients: []BulkEmailRecipient{
			{Email: "user1@example.com", Data: map[string]string{"name": "User 1"}},
			{Email: "user2@example.com", Data: map[string]string{"name": "User 2"}},
		},
		Subject:  "Bulk Email Test",
		TextBody: "Hello {{name}}!",
		Priority: models.PriorityNormal,
	}
	
	responses, err := service.SendBulkEmail(ctx, request)
	
	require.NoError(t, err)
	assert.Len(t, responses, 2)
	
	for _, response := range responses {
		assert.Equal(t, models.StatusSent, response.Status)
	}
}

func TestEmailService_SendBulkEmail_NoRecipients(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	request := &BulkEmailRequest{
		Recipients: []BulkEmailRecipient{},
		Subject:    "Test",
		TextBody:   "Test",
	}
	
	responses, err := service.SendBulkEmail(ctx, request)
	
	assert.Error(t, err)
	assert.Nil(t, responses)
	
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
}

func TestEmailService_GetEmailTemplates(t *testing.T) {
	service := createTestEmailService()
	
	templates := service.GetEmailTemplates()
	
	assert.Len(t, templates, 3) // Default templates: welcome, password_reset, notification
	
	// Find welcome template
	var welcomeFound bool
	for _, template := range templates {
		if template.ID == "welcome" {
			welcomeFound = true
			assert.Equal(t, "Welcome Email", template.Name)
			assert.Contains(t, template.Subject, "{{service_name}}")
			break
		}
	}
	
	assert.True(t, welcomeFound, "Welcome template should be found")
}

func TestEmailService_RenderTemplate(t *testing.T) {
	service := createTestEmailService()
	
	data := map[string]string{
		"user_name":    "John Doe",
		"user_email":   "john@example.com",
		"service_name": "Test Service",
	}
	
	rendered, err := service.RenderTemplate("welcome", data)
	
	require.NoError(t, err)
	assert.NotNil(t, rendered)
	assert.Equal(t, "welcome", rendered.ID)
	assert.Contains(t, rendered.Subject, "John Doe")
	assert.Contains(t, rendered.Subject, "Test Service")
	assert.Contains(t, rendered.HTMLBody, "John Doe")
	assert.Contains(t, rendered.TextBody, "john@example.com")
}

func TestEmailService_RenderTemplate_NotFound(t *testing.T) {
	service := createTestEmailService()
	
	rendered, err := service.RenderTemplate("non-existent", map[string]string{})
	
	assert.Error(t, err)
	assert.Nil(t, rendered)
	
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeTemplateNotFound, notifErr.Code)
}

func TestEmailService_ValidateEmailAddress(t *testing.T) {
	service := createTestEmailService()
	
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"valid@example.com", false},
		{"invalid-email", true},
		{"", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := service.ValidateEmailAddress(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailService_GetProviderStatus(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	status := service.GetProviderStatus(ctx)
	
	assert.NotNil(t, status)
	assert.Equal(t, "Mock Email Provider", status.Name)
	assert.Equal(t, "email", status.Type)
	assert.True(t, status.Healthy)
	assert.Empty(t, status.Error)
}

func TestEmailService_ComplexEmail(t *testing.T) {
	service := createTestEmailService()
	ctx := context.Background()
	
	request := &EmailRequest{
		To:      []string{"recipient1@example.com", "recipient2@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		From:    "sender@example.com",
		ReplyTo: "reply@example.com",
		Subject: "Complex Email Test",
		HTMLBody: "<html><body><h1>Test</h1><p>Complex email with formatting.</p></body></html>",
		TextBody: "Test\n\nComplex email with formatting.",
		Headers: map[string]string{
			"X-Priority":    "1",
			"X-Campaign-ID": "test-123",
		},
		Attachments: []models.EmailAttachment{
			{
				Filename:    "test.txt",
				Content:     []byte("Hello, World!"),
				ContentType: "text/plain",
				Size:        13,
			},
		},
		Priority: models.PriorityHigh,
		Metadata: map[string]string{
			"campaign": "test",
			"source":   "api",
		},
	}
	
	response, err := service.SendEmail(ctx, request)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
}

// Helper function
func createTestEmailService() *EmailService {
	cfg := config.EmailProviderConfig{
		Provider: "mock",
		Enabled:  true,
		Settings: map[string]string{
			"default_sender": "noreply@test.com",
		},
	}
	logger := utils.NewSimpleLogger("info")
	
	service, err := NewEmailService(cfg, logger)
	if err != nil {
		panic(err)
	}
	
	return service
}