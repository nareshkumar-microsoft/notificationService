package services

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/utils"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
)

func TestNewSMSService(t *testing.T) {
	cfg := config.SMSProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("info")

	service, err := NewSMSService(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
}

func TestNewSMSService_UnsupportedProvider(t *testing.T) {
	cfg := config.SMSProviderConfig{
		Provider: "unsupported",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("info")

	service, err := NewSMSService(cfg, logger)
	assert.Error(t, err)
	assert.Nil(t, service)

	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeProviderNotFound, notifErr.Code)
}

func TestSMSService_SendSMS_Success(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	request := &SMSRequest{
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     "Test SMS message",
		Unicode:     false,
		Priority:    models.PriorityNormal,
	}

	response, err := service.SendSMS(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
	assert.Contains(t, response.Message, "SMS sent")
}

func TestSMSService_SendSMS_ValidationErrors(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	tests := []struct {
		name    string
		request *SMSRequest
	}{
		{
			name:    "nil request",
			request: nil,
		},
		{
			name: "empty phone number",
			request: &SMSRequest{
				PhoneNumber: "",
				Message:     "Test",
			},
		},
		{
			name: "invalid phone number",
			request: &SMSRequest{
				PhoneNumber: "invalid",
				Message:     "Test",
			},
		},
		{
			name: "no message and no template",
			request: &SMSRequest{
				PhoneNumber: "1234567890",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := service.SendSMS(ctx, tt.request)
			assert.Error(t, err)
			assert.Nil(t, response)

			notifErr, ok := errors.AsNotificationError(err)
			require.True(t, ok)
			assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
		})
	}
}

func TestSMSService_SendSMS_WithTemplate(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	request := &SMSRequest{
		PhoneNumber: "1234567890",
		CountryCode: "US",
		TemplateID:  "verification",
		TemplateData: map[string]string{
			"service_name":   "TestApp",
			"code":           "123456",
			"expiry_minutes": "10",
		},
		Priority: models.PriorityHigh,
	}

	response, err := service.SendSMS(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
}

func TestSMSService_SendBulkSMS(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	request := &BulkSMSRequest{
		Recipients: []BulkSMSRecipient{
			{PhoneNumber: "1234567890", CountryCode: "US", Data: map[string]string{"name": "User 1"}},
			{PhoneNumber: "1234567891", CountryCode: "US", Data: map[string]string{"name": "User 2"}},
		},
		Message:  "Hello {{name}}!",
		Unicode:  false,
		Priority: models.PriorityNormal,
	}

	responses, err := service.SendBulkSMS(ctx, request)

	require.NoError(t, err)
	assert.Len(t, responses, 2)

	for _, response := range responses {
		assert.Equal(t, models.StatusSent, response.Status)
	}
}

func TestSMSService_SendBulkSMS_NoRecipients(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	request := &BulkSMSRequest{
		Recipients: []BulkSMSRecipient{},
		Message:    "Test",
	}

	responses, err := service.SendBulkSMS(ctx, request)

	assert.Error(t, err)
	assert.Nil(t, responses)

	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeValidationFailed, notifErr.Code)
}

func TestSMSService_GetSMSCost(t *testing.T) {
	service := createTestSMSService()

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
			cost, err := service.GetSMSCost(tt.countryCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectCost, cost)
			}
		})
	}
}

func TestSMSService_GetSupportedCountries(t *testing.T) {
	service := createTestSMSService()

	countries := service.GetSupportedCountries()

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

func TestSMSService_RenderTemplate(t *testing.T) {
	service := createTestSMSService()

	data := map[string]string{
		"service_name":   "TestApp",
		"code":           "123456",
		"expiry_minutes": "10",
	}

	rendered, err := service.RenderTemplate("verification", data)
	require.NoError(t, err)

	assert.NotNil(t, rendered)
	assert.Equal(t, "verification", rendered.ID)
	assert.Contains(t, rendered.Message, "TestApp")
	assert.Contains(t, rendered.Message, "123456")
	assert.Contains(t, rendered.Message, "10")
	assert.Equal(t, 1, rendered.Segments)
	assert.False(t, rendered.Unicode)
}

func TestSMSService_RenderTemplate_NotFound(t *testing.T) {
	service := createTestSMSService()

	rendered, err := service.RenderTemplate("non-existent", map[string]string{})

	assert.Error(t, err)
	assert.Nil(t, rendered)

	notifErr, ok := errors.AsNotificationError(err)
	require.True(t, ok)
	assert.Equal(t, errors.ErrorCodeTemplateNotFound, notifErr.Code)
}

func TestSMSService_ValidatePhoneNumber(t *testing.T) {
	service := createTestSMSService()

	tests := []struct {
		phoneNumber string
		countryCode string
		wantErr     bool
	}{
		{"1234567890", "US", false},
		{"invalid", "", true},
		{"", "", true},
		{"123456789", "US", true}, // Too short for US
	}

	for _, tt := range tests {
		t.Run(tt.phoneNumber, func(t *testing.T) {
			err := service.ValidatePhoneNumber(tt.phoneNumber, tt.countryCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMSService_GetProviderStatus(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	status := service.GetProviderStatus(ctx)

	assert.NotNil(t, status)
	assert.Equal(t, "Mock SMS Provider", status.Name)
	assert.Equal(t, "sms", status.Type)
	assert.True(t, status.Healthy)
	assert.Empty(t, status.Error)
}

func TestSMSService_EstimateCost(t *testing.T) {
	service := createTestSMSService()

	tests := []struct {
		name         string
		message      string
		countryCode  string
		unicode      bool
		expectedSegs int
		expectedCost float64
	}{
		{
			name:         "Short message US",
			message:      "Hello world",
			countryCode:  "US",
			unicode:      false,
			expectedSegs: 1,
			expectedCost: 0.0075,
		},
		{
			name:         "Long message US",
			message:      "This is a very long message that will require multiple SMS segments to send properly and completely to the recipient.",
			countryCode:  "US",
			unicode:      false,
			expectedSegs: 1, // This message is 128 chars, fits in one segment
			expectedCost: 0.0075,
		},
		{
			name:         "Unicode message",
			message:      "Hello 🌍!",
			countryCode:  "UK",
			unicode:      true,
			expectedSegs: 1,
			expectedCost: 0.0080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimate, err := service.EstimateCost(tt.message, tt.countryCode, tt.unicode)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedSegs, estimate.Segments)
			assert.Equal(t, tt.expectedCost, estimate.TotalCost)
			assert.Equal(t, tt.unicode, estimate.Unicode)
			assert.Equal(t, tt.countryCode, estimate.CountryCode)
			assert.Equal(t, len(tt.message), estimate.MessageLength)
		})
	}
}

func TestSMSService_UnicodeHandling(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	request := &SMSRequest{
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     "Hello 🌍! Welcome to TestApp 🎉",
		Unicode:     true,
		Priority:    models.PriorityNormal,
	}

	response, err := service.SendSMS(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
}

func TestSMSService_LongMessage(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	// Create a message that will require multiple segments
	longMessage := "This is a test message that is intentionally very long to test the multi-segment SMS functionality. " +
		"It should be split into multiple segments when sent via SMS. " +
		"This allows us to test both the segmentation logic and cost calculation for multi-part messages."

	request := &SMSRequest{
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     longMessage,
		Unicode:     false,
		Priority:    models.PriorityNormal,
	}

	response, err := service.SendSMS(ctx, request)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)

	// Estimate cost to verify segmentation
	estimate, err := service.EstimateCost(longMessage, "US", false)
	require.NoError(t, err)
	assert.Greater(t, estimate.Segments, 1)       // Should be multiple segments
	assert.Greater(t, estimate.TotalCost, 0.0075) // Should cost more than single segment
}

func TestSMSService_CountrySpecific(t *testing.T) {
	service := createTestSMSService()
	ctx := context.Background()

	tests := []struct {
		name        string
		phoneNumber string
		countryCode string
	}{
		{"US number", "1234567890", "US"},
		{"UK number", "07123456789", "UK"},
		{"German number", "1234567890", "DE"},
		{"Indian number", "9876543210", "IN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &SMSRequest{
				PhoneNumber: tt.phoneNumber,
				CountryCode: tt.countryCode,
				Message:     "Test message for " + tt.name,
				Priority:    models.PriorityNormal,
			}

			response, err := service.SendSMS(ctx, request)
			require.NoError(t, err)
			assert.Equal(t, models.StatusSent, response.Status)
		})
	}
}

func TestCalculateSMSSegments(t *testing.T) {
	tests := []struct {
		name             string
		message          string
		unicode          bool
		expectedSegments int
	}{
		{"Short text", "Hello", false, 1},
		{"Single segment", "This is a test message that fits in one SMS segment.", false, 1},
		{"Two segments", strings.Repeat("This is a very long message. ", 10), false, 2}, // 300+ chars
		{"Short unicode", "Hello 🌍", true, 1},
		{"Long unicode", strings.Repeat("This is unicode text. ", 4), true, 2}, // Simpler unicode test
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := calculateSMSSegments(tt.message, tt.unicode)
			assert.Equal(t, tt.expectedSegments, segments)
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		message   string
		maxLength int
		expected  string
	}{
		{"Short", 10, "Short"},
		{"This is a long message", 10, "This is..."},
		{"Exact length", 12, "Exact length"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := truncateMessage(tt.message, tt.maxLength)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), tt.maxLength)
		})
	}
}

// Helper function
func createTestSMSService() *SMSService {
	cfg := config.SMSProviderConfig{
		Provider: "mock",
		Enabled:  true,
		Settings: map[string]string{
			"default_country": "US",
		},
	}
	logger := utils.NewSimpleLogger("info")

	service, err := NewSMSService(cfg, logger)
	if err != nil {
		panic(err)
	}

	return service
}
