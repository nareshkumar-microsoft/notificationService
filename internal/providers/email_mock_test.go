package providers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
)

func TestNewMockEmailProvider(t *testing.T) {
	cfg := config.EmailProviderConfig{
		Provider: "mock",
		Enabled:  true,
		Settings: map[string]string{
			"default_sender": "test@example.com",
		},
	}
	
	provider := NewMockEmailProvider(cfg)
	
	assert.NotNil(t, provider)
	assert.Equal(t, cfg, provider.config)
	assert.True(t, provider.healthy)
	assert.Len(t, provider.templates, 3) // Default templates loaded
	assert.Empty(t, provider.sentEmails)
}

func TestMockEmailProvider_GetType(t *testing.T) {
	provider := createTestEmailProvider()
	assert.Equal(t, models.NotificationTypeEmail, provider.GetType())
}

func TestMockEmailProvider_IsHealthy(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	// Test healthy provider
	err := provider.IsHealthy(ctx)
	assert.NoError(t, err)
	
	// Test unhealthy provider
	provider.SetHealthy(false)
	err = provider.IsHealthy(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unhealthy")
}

func TestMockEmailProvider_GetConfig(t *testing.T) {
	provider := createTestEmailProvider()
	config := provider.GetConfig()
	
	assert.Equal(t, "Mock Email Provider", config.Name)
	assert.Equal(t, models.NotificationTypeEmail, config.Type)
	assert.True(t, config.Enabled)
	assert.Equal(t, 1, config.Priority)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 30, config.Timeout)
	assert.True(t, config.RateLimit.Enabled)
	assert.Equal(t, 100, config.RateLimit.RequestsPerMin)
}

func TestMockEmailProvider_ValidateEmailAddress(t *testing.T) {
	provider := createTestEmailProvider()
	
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with numbers", "user123@example.com", false},
		{"empty email", "", true},
		{"invalid format - missing @", "invalid-email", true},
		{"invalid format - missing domain", "test@", true},
		{"invalid format - missing TLD", "test@example", true},
		{"invalid format - spaces", "test @example.com", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateEmailAddress(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockEmailProvider_SendEmail_Success(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	email := createTestEmailNotification()
	
	response, err := provider.SendEmail(ctx, email)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, email.ID, response.ID)
	assert.Equal(t, models.StatusSent, response.Status)
	assert.Contains(t, response.Message, "successfully sent")
	assert.NotEmpty(t, response.ProviderID)
	assert.NotNil(t, response.SentAt)
	
	// Check that email was recorded
	sentEmails := provider.GetSentEmails()
	assert.Len(t, sentEmails, 1)
	assert.Equal(t, email.ID, sentEmails[0].ID)
	assert.Equal(t, email.To, sentEmails[0].To)
	assert.Equal(t, email.Subject, sentEmails[0].Subject)
}

func TestMockEmailProvider_SendEmail_ValidationErrors(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	tests := []struct {
		name  string
		email *models.EmailNotification
	}{
		{
			name: "empty recipients",
			email: &models.EmailNotification{
				Notification: models.Notification{
					ID:      uuid.New(),
					Type:    models.NotificationTypeEmail,
					Subject: "Test",
				},
				To:       []string{},
				From:     "sender@example.com",
				HTMLBody: "Test body",
			},
		},
		{
			name: "invalid recipient email",
			email: &models.EmailNotification{
				Notification: models.Notification{
					ID:      uuid.New(),
					Type:    models.NotificationTypeEmail,
					Subject: "Test",
				},
				To:       []string{"invalid-email"},
				From:     "sender@example.com",
				HTMLBody: "Test body",
			},
		},
		{
			name: "invalid sender email",
			email: &models.EmailNotification{
				Notification: models.Notification{
					ID:      uuid.New(),
					Type:    models.NotificationTypeEmail,
					Subject: "Test",
				},
				To:       []string{"recipient@example.com"},
				From:     "invalid-sender",
				HTMLBody: "Test body",
			},
		},
		{
			name: "empty subject",
			email: &models.EmailNotification{
				Notification: models.Notification{
					ID:   uuid.New(),
					Type: models.NotificationTypeEmail,
				},
				To:       []string{"recipient@example.com"},
				From:     "sender@example.com",
				HTMLBody: "Test body",
			},
		},
		{
			name: "empty body",
			email: &models.EmailNotification{
				Notification: models.Notification{
					ID:      uuid.New(),
					Type:    models.NotificationTypeEmail,
					Subject: "Test",
				},
				To:   []string{"recipient@example.com"},
				From: "sender@example.com",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := provider.SendEmail(ctx, tt.email)
			assert.Error(t, err)
			assert.Nil(t, response)
			
			// Check that it's a validation error
			notifErr, ok := errors.AsNotificationError(err)
			require.True(t, ok)
			assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
		})
	}
}

func TestMockEmailProvider_SendEmail_UnhealthyProvider(t *testing.T) {
	provider := createTestEmailProvider()
	provider.SetHealthy(false)
	ctx := context.Background()
	
	email := createTestEmailNotification()
	
	response, err := provider.SendEmail(ctx, email)
	
	assert.Error(t, err)
	assert.Nil(t, response)
	
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeProviderUnavailable, notifErr.Code)
}

func TestMockEmailProvider_Send_GenericNotification(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	notification := &models.Notification{
		ID:        uuid.New(),
		Type:      models.NotificationTypeEmail,
		Status:    models.StatusPending,
		Priority:  models.PriorityNormal,
		Recipient: "test@example.com",
		Subject:   "Test Notification",
		Body:      "Test body content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	response, err := provider.Send(ctx, notification)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, notification.ID, response.ID)
	assert.Equal(t, models.StatusSent, response.Status)
	
	// Check that email was recorded
	sentEmails := provider.GetSentEmails()
	assert.Len(t, sentEmails, 1)
	assert.Equal(t, notification.Recipient, sentEmails[0].To[0])
}

func TestMockEmailProvider_Send_WrongType(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	notification := &models.Notification{
		ID:   uuid.New(),
		Type: models.NotificationTypeSMS, // Wrong type
	}
	
	response, err := provider.Send(ctx, notification)
	
	assert.Error(t, err)
	assert.Nil(t, response)
	
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
}

func TestMockEmailProvider_Templates(t *testing.T) {
	provider := createTestEmailProvider()
	
	// Test getting templates
	templates := provider.GetEmailTemplates()
	assert.Len(t, templates, 3) // welcome, password_reset, notification
	
	// Test getting specific template
	welcomeTemplate, err := provider.GetTemplate("welcome")
	require.NoError(t, err)
	assert.Equal(t, "welcome", welcomeTemplate.ID)
	assert.Equal(t, "Welcome Email", welcomeTemplate.Name)
	assert.Contains(t, welcomeTemplate.Subject, "{{service_name}}")
	assert.Contains(t, welcomeTemplate.Variables, "user_name")
	assert.Contains(t, welcomeTemplate.Variables, "service_name")
	
	// Test getting non-existent template
	_, err = provider.GetTemplate("non-existent")
	assert.Error(t, err)
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeTemplateNotFound, notifErr.Code)
}

func TestMockEmailProvider_AddTemplate(t *testing.T) {
	provider := createTestEmailProvider()
	
	newTemplate := &EmailTemplate{
		Name:      "Test Template",
		Subject:   "Test Subject {{name}}",
		HTMLBody:  "<h1>Hello {{name}}</h1>",
		TextBody:  "Hello {{name}}",
		Variables: []string{"name"},
		Category:  "test",
	}
	
	err := provider.AddTemplate(newTemplate)
	require.NoError(t, err)
	
	// Check that template was added
	assert.NotEmpty(t, newTemplate.ID) // ID should be generated
	assert.False(t, newTemplate.CreatedAt.IsZero())
	assert.False(t, newTemplate.UpdatedAt.IsZero())
	
	// Verify we can retrieve it
	retrieved, err := provider.GetTemplate(newTemplate.ID)
	require.NoError(t, err)
	assert.Equal(t, newTemplate.Name, retrieved.Name)
	assert.Equal(t, newTemplate.Subject, retrieved.Subject)
}

func TestMockEmailProvider_RenderTemplate(t *testing.T) {
	provider := createTestEmailProvider()
	
	data := map[string]string{
		"user_name":    "John Doe",
		"user_email":   "john@example.com",
		"service_name": "Test Service",
	}
	
	rendered, err := provider.RenderTemplate("welcome", data)
	require.NoError(t, err)
	
	assert.Equal(t, "welcome", rendered.ID)
	assert.Contains(t, rendered.Subject, "Test Service")
	assert.Contains(t, rendered.Subject, "John Doe")
	assert.Contains(t, rendered.HTMLBody, "John Doe")
	assert.Contains(t, rendered.HTMLBody, "john@example.com")
	assert.Contains(t, rendered.TextBody, "Test Service")
	
	// Test with non-existent template
	_, err = provider.RenderTemplate("non-existent", data)
	assert.Error(t, err)
}

func TestMockEmailProvider_ComplexEmail(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	email := &models.EmailNotification{
		Notification: models.Notification{
			ID:       uuid.New(),
			Type:     models.NotificationTypeEmail,
			Subject:  "Complex Email Test",
			Metadata: map[string]string{"campaign": "test"},
		},
		To:      []string{"recipient1@example.com", "recipient2@example.com"},
		CC:      []string{"cc@example.com"},
		BCC:     []string{"bcc@example.com"},
		From:    "sender@example.com",
		ReplyTo: "reply@example.com",
		HTMLBody: "<html><body><h1>Test</h1><p>This is a test email with <strong>formatting</strong>.</p></body></html>",
		TextBody: "Test\n\nThis is a test email with formatting.",
		Headers: map[string]string{
			"X-Priority":     "1",
			"X-Campaign-ID":  "test-123",
			"X-Mailer":       "NotificationService",
		},
		Attachments: []models.EmailAttachment{
			{
				Filename:    "test.txt",
				Content:     []byte("Hello, World!"),
				ContentType: "text/plain",
				Size:        13,
			},
		},
	}
	
	response, err := provider.SendEmail(ctx, email)
	
	require.NoError(t, err)
	assert.NotNil(t, response)
	
	// Verify sent email details
	sentEmails := provider.GetSentEmails()
	require.Len(t, sentEmails, 1)
	
	sentEmail := sentEmails[0]
	assert.Equal(t, email.To, sentEmail.To)
	assert.Equal(t, email.CC, sentEmail.CC)
	assert.Equal(t, email.BCC, sentEmail.BCC)
	assert.Equal(t, email.From, sentEmail.From)
	assert.Equal(t, email.Subject, sentEmail.Subject)
	assert.Equal(t, email.HTMLBody, sentEmail.HTMLBody)
	assert.Equal(t, email.TextBody, sentEmail.TextBody)
	assert.Equal(t, email.Headers, sentEmail.Headers)
	assert.Equal(t, "sent", sentEmail.Status)
	assert.Contains(t, sentEmail.ProviderData, "provider")
	assert.Contains(t, sentEmail.ProviderData, "message_id")
}

func TestMockEmailProvider_ClearSentEmails(t *testing.T) {
	provider := createTestEmailProvider()
	ctx := context.Background()
	
	// Send an email
	email := createTestEmailNotification()
	_, err := provider.SendEmail(ctx, email)
	require.NoError(t, err)
	
	// Verify email was sent
	assert.Len(t, provider.GetSentEmails(), 1)
	
	// Clear sent emails
	provider.ClearSentEmails()
	assert.Empty(t, provider.GetSentEmails())
}

// Helper functions

func createTestEmailProvider() *MockEmailProvider {
	cfg := config.EmailProviderConfig{
		Provider: "mock",
		Enabled:  true,
		Settings: map[string]string{
			"default_sender": "noreply@test.com",
		},
	}
	return NewMockEmailProvider(cfg)
}

func createTestEmailNotification() *models.EmailNotification {
	return &models.EmailNotification{
		Notification: models.Notification{
			ID:        uuid.New(),
			Type:      models.NotificationTypeEmail,
			Status:    models.StatusPending,
			Priority:  models.PriorityNormal,
			Recipient: "test@example.com",
			Subject:   "Test Email",
			Body:      "Test body content",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		To:       []string{"test@example.com"},
		From:     "sender@example.com",
		HTMLBody: "<p>Test HTML content</p>",
		TextBody: "Test text content",
	}
}