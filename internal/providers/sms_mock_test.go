package providers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
)

func TestNewMockSMSProvider(t *testing.T) {
	cfg := config.SMSProviderConfig{
		Provider: "mock",
		Enabled:  true,
		Settings: map[string]string{
			"default_country": "US",
		},
	}

	provider := NewMockSMSProvider(cfg)

	assert.NotNil(t, provider)
	assert.Equal(t, cfg, provider.config)
	assert.True(t, provider.healthy)
	assert.Len(t, provider.templates, 4) // Default templates loaded
	assert.Empty(t, provider.sentSMS)
	assert.NotEmpty(t, provider.costs) // Default costs loaded
}

func TestMockSMSProvider_GetType(t *testing.T) {
	provider := createTestSMSProvider()
	assert.Equal(t, models.NotificationTypeSMS, provider.GetType())
}

func TestMockSMSProvider_IsHealthy(t *testing.T) {
	provider := createTestSMSProvider()
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

func TestMockSMSProvider_GetConfig(t *testing.T) {
	provider := createTestSMSProvider()
	config := provider.GetConfig()

	assert.Equal(t, "Mock SMS Provider", config.Name)
	assert.Equal(t, models.NotificationTypeSMS, config.Type)
	assert.True(t, config.Enabled)
	assert.Equal(t, 2, config.Priority)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 30, config.Timeout)
	assert.True(t, config.RateLimit.Enabled)
	assert.Equal(t, 60, config.RateLimit.RequestsPerMin)
}

func TestMockSMSProvider_ValidatePhoneNumber(t *testing.T) {
	provider := createTestSMSProvider()

	tests := []struct {
		name        string
		phoneNumber string
		countryCode string
		wantErr     bool
	}{
		{"valid US number", "1234567890", "US", false},
		{"valid US number with formatting", "(123) 456-7890", "US", false},
		{"valid UK number", "07123456789", "UK", false},
		{"valid international format", "+1234567890", "", false},
		{"empty phone number", "", "", true},
		{"too short", "123", "", true},
		{"too long", "123456789012345678", "", true},
		{"contains letters", "123abc7890", "", true},
		{"invalid US number", "123456789", "US", true},
		{"unsupported country", "1234567890", "XX", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidatePhoneNumber(tt.phoneNumber, tt.countryCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockSMSProvider_GetSMSCost(t *testing.T) {
	provider := createTestSMSProvider()

	tests := []struct {
		name        string
		countryCode string
		expectCost  float64
		wantErr     bool
	}{
		{"US cost", "US", 0.0075, false},
		{"UK cost", "UK", 0.0080, false},
		{"default cost", "", 0.01, false},
		{"unsupported country", "XX", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := provider.GetSMSCost(tt.countryCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectCost, cost)
			}
		})
	}
}

func TestMockSMSProvider_SendSMS_Success(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	sms := createTestSMSNotification()

	response, err := provider.SendSMS(ctx, sms)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, sms.ID, response.ID)
	assert.Equal(t, models.StatusSent, response.Status)
	assert.Contains(t, response.Message, "SMS sent")
	assert.NotEmpty(t, response.ProviderID)
	assert.NotNil(t, response.SentAt)

	// Check that SMS was recorded
	sentSMS := provider.GetSentSMS()
	assert.Len(t, sentSMS, 1)
	assert.Equal(t, sms.ID, sentSMS[0].ID)
	assert.Equal(t, sms.PhoneNumber, sentSMS[0].PhoneNumber)
	assert.Equal(t, sms.Message, sentSMS[0].Message)
	assert.Equal(t, 1, sentSMS[0].Segments)
	assert.Greater(t, sentSMS[0].Cost, 0.0)
}

func TestMockSMSProvider_SendSMS_ValidationErrors(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	tests := []struct {
		name string
		sms  *models.SMSNotification
	}{
		{
			name: "empty phone number",
			sms: &models.SMSNotification{
				Notification: models.Notification{
					ID:   uuid.New(),
					Type: models.NotificationTypeSMS,
				},
				PhoneNumber: "",
				Message:     "Test message",
			},
		},
		{
			name: "invalid phone number",
			sms: &models.SMSNotification{
				Notification: models.Notification{
					ID:   uuid.New(),
					Type: models.NotificationTypeSMS,
				},
				PhoneNumber: "invalid",
				Message:     "Test message",
			},
		},
		{
			name: "empty message",
			sms: &models.SMSNotification{
				Notification: models.Notification{
					ID:   uuid.New(),
					Type: models.NotificationTypeSMS,
				},
				PhoneNumber: "1234567890",
				Message:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := provider.SendSMS(ctx, tt.sms)
			assert.Error(t, err)
			assert.Nil(t, response)

			// Check that it's a validation error
			notifErr, ok := errors.AsNotificationError(err)
			require.True(t, ok)
			assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
		})
	}
}

func TestMockSMSProvider_SendSMS_UnhealthyProvider(t *testing.T) {
	provider := createTestSMSProvider()
	provider.SetHealthy(false)
	ctx := context.Background()

	sms := createTestSMSNotification()

	response, err := provider.SendSMS(ctx, sms)

	assert.Error(t, err)
	assert.Nil(t, response)

	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeProviderUnavailable, notifErr.Code)
}

func TestMockSMSProvider_Send_GenericNotification(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	notification := &models.Notification{
		ID:        uuid.New(),
		Type:      models.NotificationTypeSMS,
		Status:    models.StatusPending,
		Priority:  models.PriorityNormal,
		Recipient: "1234567890",
		Subject:   "Test SMS",
		Body:      "Test SMS message content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]string{
			"country_code": "US",
		},
	}

	response, err := provider.Send(ctx, notification)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, notification.ID, response.ID)
	assert.Equal(t, models.StatusSent, response.Status)

	// Check that SMS was recorded
	sentSMS := provider.GetSentSMS()
	assert.Len(t, sentSMS, 1)
	assert.Equal(t, notification.Recipient, sentSMS[0].PhoneNumber)
	assert.Equal(t, "US", sentSMS[0].CountryCode)
}

func TestMockSMSProvider_Send_WrongType(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	notification := &models.Notification{
		ID:   uuid.New(),
		Type: models.NotificationTypeEmail, // Wrong type
	}

	response, err := provider.Send(ctx, notification)

	assert.Error(t, err)
	assert.Nil(t, response)

	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
}

func TestMockSMSProvider_Templates(t *testing.T) {
	provider := createTestSMSProvider()

	// Test getting specific template
	verificationTemplate, err := provider.GetTemplate("verification")
	require.NoError(t, err)
	assert.Equal(t, "verification", verificationTemplate.ID)
	assert.Equal(t, "Verification Code", verificationTemplate.Name)
	assert.Contains(t, verificationTemplate.Message, "{{code}}")
	assert.Contains(t, verificationTemplate.Variables, "code")
	assert.Contains(t, verificationTemplate.Variables, "service_name")

	// Test getting non-existent template
	_, err = provider.GetTemplate("non-existent")
	assert.Error(t, err)
	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeTemplateNotFound, notifErr.Code)
}

func TestMockSMSProvider_AddTemplate(t *testing.T) {
	provider := createTestSMSProvider()

	newTemplate := &SMSTemplate{
		Name:      "Test Template",
		Message:   "Hello {{name}}, your order {{order_id}} is ready!",
		Variables: []string{"name", "order_id"},
		Category:  "orders",
		MaxLength: 160,
		Unicode:   false,
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
	assert.Equal(t, newTemplate.Message, retrieved.Message)
	assert.Equal(t, 160, retrieved.MaxLength)
}

func TestMockSMSProvider_RenderTemplate(t *testing.T) {
	provider := createTestSMSProvider()

	data := map[string]string{
		"service_name":   "TestApp",
		"code":           "123456",
		"expiry_minutes": "10",
	}

	rendered, err := provider.RenderTemplate("verification", data)
	require.NoError(t, err)

	assert.Equal(t, "verification", rendered.ID)
	assert.Contains(t, rendered.Message, "TestApp")
	assert.Contains(t, rendered.Message, "123456")
	assert.Contains(t, rendered.Message, "10")
	assert.NotContains(t, rendered.Message, "{{") // No unresolved variables

	// Test with non-existent template
	_, err = provider.RenderTemplate("non-existent", data)
	assert.Error(t, err)
}

func TestMockSMSProvider_UnicodeHandling(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	// Unicode message
	sms := &models.SMSNotification{
		Notification: models.Notification{
			ID:      uuid.New(),
			Type:    models.NotificationTypeSMS,
			Subject: "Unicode Test",
		},
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     "Hello 🌍! Welcome to TestApp 🎉",
		Unicode:     true,
	}

	response, err := provider.SendSMS(ctx, sms)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// Check sent SMS details
	sentSMS := provider.GetSentSMS()
	require.Len(t, sentSMS, 1)
	assert.True(t, sentSMS[0].Unicode)
	assert.Equal(t, 1, sentSMS[0].Segments) // Should fit in one unicode segment
}

func TestMockSMSProvider_MultiSegmentMessage(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	// Long message that requires multiple segments
	longMessage := strings.Repeat("This is a test message. ", 20) // ~500 characters

	sms := &models.SMSNotification{
		Notification: models.Notification{
			ID:   uuid.New(),
			Type: models.NotificationTypeSMS,
		},
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     longMessage,
		Unicode:     false,
	}

	response, err := provider.SendSMS(ctx, sms)

	require.NoError(t, err)
	assert.NotNil(t, response)

	// Check sent SMS details
	sentSMS := provider.GetSentSMS()
	require.Len(t, sentSMS, 1)
	assert.Greater(t, sentSMS[0].Segments, 1)  // Should be multiple segments
	assert.Greater(t, sentSMS[0].Cost, 0.0075) // Cost should be higher for multiple segments
}

func TestMockSMSProvider_CountrySpecificValidation(t *testing.T) {
	provider := createTestSMSProvider()

	tests := []struct {
		name        string
		phoneNumber string
		countryCode string
		wantErr     bool
	}{
		{"valid US number", "1234567890", "US", false},
		{"valid US number with country code", "11234567890", "US", false},
		{"invalid US number - too short", "123456789", "US", true},
		{"valid UK number", "07123456789", "UK", false},
		{"valid Indian number", "9876543210", "IN", false},
		{"invalid Indian number", "987654321", "IN", true},
		{"valid German number", "1234567890", "DE", false},
		{"invalid German number - too short", "123456789", "DE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidatePhoneNumber(tt.phoneNumber, tt.countryCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockSMSProvider_GetSupportedCountries(t *testing.T) {
	provider := createTestSMSProvider()

	countries := provider.GetSupportedCountries()

	assert.GreaterOrEqual(t, len(countries), 8) // At least 8 countries

	// Check for specific countries
	var usFound, ukFound bool
	for _, country := range countries {
		if country.Code == "US" {
			usFound = true
			assert.Equal(t, "United States", country.Name)
			assert.Greater(t, country.Cost, 0.0)
			assert.True(t, country.Supported)
		}
		if country.Code == "UK" {
			ukFound = true
			assert.Equal(t, "United Kingdom", country.Name)
		}
	}

	assert.True(t, usFound, "US should be in supported countries")
	assert.True(t, ukFound, "UK should be in supported countries")
}

func TestMockSMSProvider_ClearSentSMS(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	// Send an SMS
	sms := createTestSMSNotification()
	_, err := provider.SendSMS(ctx, sms)
	require.NoError(t, err)

	// Verify SMS was sent
	assert.Len(t, provider.GetSentSMS(), 1)

	// Clear sent SMS
	provider.ClearSentSMS()
	assert.Empty(t, provider.GetSentSMS())
}

func TestMockSMSProvider_CostCalculation(t *testing.T) {
	provider := createTestSMSProvider()
	ctx := context.Background()

	tests := []struct {
		name          string
		countryCode   string
		messageLength int
		expectedCost  float64
	}{
		{"US single segment", "US", 100, 0.0075},
		{"US double segment", "US", 300, 0.0150}, // 2 * 0.0075
		{"UK single segment", "UK", 100, 0.0080},
		{"Default single segment", "", 100, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := strings.Repeat("a", tt.messageLength)

			sms := &models.SMSNotification{
				Notification: models.Notification{
					ID:   uuid.New(),
					Type: models.NotificationTypeSMS,
				},
				PhoneNumber: "1234567890",
				CountryCode: tt.countryCode,
				Message:     message,
				Unicode:     false,
			}

			_, err := provider.SendSMS(ctx, sms)
			require.NoError(t, err)

			sentSMS := provider.GetSentSMS()
			require.Len(t, sentSMS, 1)
			assert.Equal(t, tt.expectedCost, sentSMS[0].Cost)

			// Clear for next test
			provider.ClearSentSMS()
		})
	}
}

// Helper functions

func createTestSMSProvider() *MockSMSProvider {
	cfg := config.SMSProviderConfig{
		Provider: "mock",
		Enabled:  true,
		Settings: map[string]string{
			"default_country": "US",
		},
	}
	return NewMockSMSProvider(cfg)
}

func createTestSMSNotification() *models.SMSNotification {
	return &models.SMSNotification{
		Notification: models.Notification{
			ID:        uuid.New(),
			Type:      models.NotificationTypeSMS,
			Status:    models.StatusPending,
			Priority:  models.PriorityNormal,
			Recipient: "1234567890",
			Subject:   "Test SMS",
			Body:      "Test SMS message",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     "Test SMS message",
		Unicode:     false,
	}
}
