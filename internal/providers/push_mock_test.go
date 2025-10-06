package providers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewMockPushProvider(t *testing.T) {
	cfg := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}

	provider := NewMockPushProvider(cfg)

	assert.NotNil(t, provider)
	assert.Equal(t, cfg, provider.config)
	assert.True(t, provider.healthy)
	assert.NotNil(t, provider.templates)
	assert.NotNil(t, provider.sentPushes)
	assert.NotNil(t, provider.deviceTokens)
	assert.NotNil(t, provider.platformConfigs)

	// Check that default templates are loaded
	assert.True(t, len(provider.templates) > 0)

	// Check that platform configs are loaded
	assert.Contains(t, provider.platformConfigs, "ios")
	assert.Contains(t, provider.platformConfigs, "android")
	assert.Contains(t, provider.platformConfigs, "web")
}

func TestMockPushProvider_SendPush(t *testing.T) {
	provider := createTestPushProvider()
	ctx := context.Background()

	testCases := []struct {
		name             string
		pushNotification *models.PushNotification
		expectError      bool
		errorType        string
	}{
		{
			name: "valid iOS push notification",
			pushNotification: &models.PushNotification{
				Notification: models.Notification{
					ID:       uuid.New(),
					Type:     models.NotificationTypePush,
					Priority: models.PriorityNormal,
				},
				DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
				Platform:    "ios",
				Title:       "Test Notification",
				Message:     "This is a test push notification",
				Sound:       "default",
			},
			expectError: false,
		},
		{
			name: "valid Android push notification",
			pushNotification: &models.PushNotification{
				Notification: models.Notification{
					ID:       uuid.New(),
					Type:     models.NotificationTypePush,
					Priority: models.PriorityHigh,
				},
				DeviceToken: "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
				Platform:    "android",
				Title:       "Android Test",
				Message:     "Android push notification",
				Icon:        "ic_notification",
			},
			expectError: false,
		},
		{
			name: "valid Web push notification",
			pushNotification: &models.PushNotification{
				Notification: models.Notification{
					ID:       uuid.New(),
					Type:     models.NotificationTypePush,
					Priority: models.PriorityLow,
				},
				DeviceToken: "BNJzWlpOQMEK-web-push-token-example-abc123def456ghi789",
				Platform:    "web",
				Title:       "Web Notification",
				Message:     "Web push notification test",
				Icon:        "/icon-192x192.png",
			},
			expectError: false,
		},
		{
			name: "invalid device token - too short",
			pushNotification: &models.PushNotification{
				Notification: models.Notification{
					ID:       uuid.New(),
					Type:     models.NotificationTypePush,
					Priority: models.PriorityNormal,
				},
				DeviceToken: "short",
				Platform:    "ios",
				Title:       "Test",
				Message:     "Test message",
			},
			expectError: true,
			errorType:   "validation",
		},
		{
			name: "unsupported platform",
			pushNotification: &models.PushNotification{
				Notification: models.Notification{
					ID:       uuid.New(),
					Type:     models.NotificationTypePush,
					Priority: models.PriorityNormal,
				},
				DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
				Platform:    "unsupported",
				Title:       "Test",
				Message:     "Test message",
			},
			expectError: true,
			errorType:   "validation",
		},
		{
			name: "empty title and message",
			pushNotification: &models.PushNotification{
				Notification: models.Notification{
					ID:       uuid.New(),
					Type:     models.NotificationTypePush,
					Priority: models.PriorityNormal,
				},
				DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
				Platform:    "ios",
				Title:       "",
				Message:     "",
			},
			expectError: true,
			errorType:   "validation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear previous sent pushes
			provider.ClearSentPushes()

			response, err := provider.SendPush(ctx, tc.pushNotification)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				assert.Empty(t, provider.GetSentPushes())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tc.pushNotification.ID, response.ID)
				assert.Equal(t, models.StatusSent, response.Status)
				assert.NotEmpty(t, response.ProviderID)
				assert.NotNil(t, response.SentAt)

				// Check that push was recorded
				sentPushes := provider.GetSentPushes()
				assert.Len(t, sentPushes, 1)
				assert.Equal(t, tc.pushNotification.ID, sentPushes[0].ID)
				assert.Equal(t, tc.pushNotification.Platform, sentPushes[0].Platform)
			}
		})
	}
}

func TestMockPushProvider_Send(t *testing.T) {
	provider := createTestPushProvider()
	ctx := context.Background()

	notification := &models.Notification{
		ID:        uuid.New(),
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityNormal,
		Recipient: "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
		Subject:   "Test Subject",
		Body:      "Test message body",
		Metadata: map[string]string{
			"platform":     "android",
			"device_token": "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
		},
	}

	response, err := provider.Send(ctx, notification)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, notification.ID, response.ID)
	assert.Equal(t, models.StatusSent, response.Status)
}

func TestMockPushProvider_ValidateDeviceToken(t *testing.T) {
	provider := createTestPushProvider()

	testCases := []struct {
		name        string
		token       string
		platform    string
		expectError bool
	}{
		{
			name:        "valid iOS token",
			token:       "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			platform:    "ios",
			expectError: false,
		},
		{
			name:        "valid Android token",
			token:       "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
			platform:    "android",
			expectError: false,
		},
		{
			name:        "valid Web token",
			token:       "BNJzWlpOQMEK-web-push-token-example-abc123def456ghi789",
			platform:    "web",
			expectError: false,
		},
		{
			name:        "iOS token too short",
			token:       "a1b2c3d4",
			platform:    "ios",
			expectError: true,
		},
		{
			name:        "iOS token invalid characters",
			token:       "g1h2i3j4k5l6789012345678901234567890123456789012345678901234567890123456",
			platform:    "ios",
			expectError: true,
		},
		{
			name:        "Android token too short",
			token:       "short",
			platform:    "android",
			expectError: true,
		},
		{
			name:        "Android token too long",
			token:       "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-ABC123DEF456GHI789JKL012MNO345PQR678STU901VWX234YZA_EXTRA_LONG_TEXT_TO_EXCEED_255_CHARACTERS_LIMIT_FOR_ANDROID_FCM_TOKENS_0123456789",
			platform:    "android",
			expectError: true,
		},
		{
			name:        "Web token too short",
			token:       "short",
			platform:    "web",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			platform:    "ios",
			expectError: true,
		},
		{
			name:        "unsupported platform",
			token:       "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			platform:    "unsupported",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := provider.ValidateDeviceToken(tc.token, tc.platform)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockPushProvider_GetPlatformConfig(t *testing.T) {
	provider := createTestPushProvider()

	testCases := []struct {
		platform string
		expected string
	}{
		{"ios", "ios"},
		{"android", "android"},
		{"web", "web"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.platform, func(t *testing.T) {
			config := provider.GetPlatformConfig(tc.platform)
			assert.Equal(t, tc.expected, config.Platform)
			assert.True(t, config.MaxPayload > 0)
		})
	}
}

func TestMockPushProvider_RegisterUnregisterDevice(t *testing.T) {
	provider := createTestPushProvider()

	deviceToken := "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678"
	platform := "ios"
	metadata := map[string]string{
		"app_version": "1.0.0",
		"os_version":  "15.0",
	}

	// Test registration
	err := provider.RegisterDevice(deviceToken, platform, metadata)
	assert.NoError(t, err)

	// Test getting device info
	deviceInfo, err := provider.GetDeviceInfo(deviceToken)
	assert.NoError(t, err)
	assert.NotNil(t, deviceInfo)
	assert.Equal(t, deviceToken, deviceInfo.Token)
	assert.Equal(t, platform, deviceInfo.Platform)
	assert.Equal(t, "1.0.0", deviceInfo.AppVersion)
	assert.Equal(t, "15.0", deviceInfo.OSVersion)
	assert.True(t, deviceInfo.Active)

	// Test unregistration
	err = provider.UnregisterDevice(deviceToken)
	assert.NoError(t, err)

	// Test getting device info after unregistration
	_, err = provider.GetDeviceInfo(deviceToken)
	assert.Error(t, err)

	// Test unregistering non-existent device
	err = provider.UnregisterDevice("non-existent")
	assert.Error(t, err)
}

func TestMockPushProvider_Templates(t *testing.T) {
	provider := createTestPushProvider()

	// Test getting existing template
	template, err := provider.GetTemplate("welcome_push")
	assert.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "welcome_push", template.ID)
	assert.Equal(t, "Welcome Push Notification", template.Name)

	// Test getting non-existent template
	_, err = provider.GetTemplate("non-existent")
	assert.Error(t, err)

	// Test adding new template
	newTemplate := &PushTemplate{
		ID:        "test_template",
		Name:      "Test Template",
		Platform:  "all",
		Title:     "Test {{title}}",
		Body:      "Hello {{name}}!",
		Variables: []string{"title", "name"},
		Category:  "test",
	}

	err = provider.AddTemplate(newTemplate)
	assert.NoError(t, err)

	// Test getting the new template
	retrievedTemplate, err := provider.GetTemplate("test_template")
	assert.NoError(t, err)
	assert.Equal(t, newTemplate.ID, retrievedTemplate.ID)
	assert.Equal(t, newTemplate.Name, retrievedTemplate.Name)

	// Test rendering template
	data := map[string]string{
		"title": "Important",
		"name":  "John",
	}

	renderedTemplate, err := provider.RenderTemplate("test_template", data)
	assert.NoError(t, err)
	assert.Equal(t, "Test Important", renderedTemplate.Title)
	assert.Equal(t, "Hello John!", renderedTemplate.Body)
}

func TestMockPushProvider_GetType(t *testing.T) {
	provider := createTestPushProvider()
	assert.Equal(t, models.NotificationTypePush, provider.GetType())
}

func TestMockPushProvider_IsHealthy(t *testing.T) {
	provider := createTestPushProvider()
	ctx := context.Background()

	// Test healthy provider
	err := provider.IsHealthy(ctx)
	assert.NoError(t, err)

	// Test unhealthy provider
	provider.SetHealthy(false)
	err = provider.IsHealthy(ctx)
	assert.Error(t, err)

	// Test health check timeout
	provider.SetHealthy(true)
	ctx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	err = provider.IsHealthy(ctx)
	assert.NoError(t, err) // Should complete within timeout
}

func TestMockPushProvider_GetConfig(t *testing.T) {
	provider := createTestPushProvider()

	config := provider.GetConfig()
	assert.Equal(t, "Mock Push Provider", config.Name)
	assert.Equal(t, models.NotificationTypePush, config.Type)
	assert.True(t, config.Enabled)
	assert.Equal(t, 3, config.Priority)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 45, config.Timeout)
}

func TestMockPushProvider_GetSupportedPlatforms(t *testing.T) {
	provider := createTestPushProvider()

	platforms := provider.GetSupportedPlatforms()
	assert.Contains(t, platforms, "ios")
	assert.Contains(t, platforms, "android")
	assert.Contains(t, platforms, "web")
	assert.Len(t, platforms, 3)
}

func TestMockPushProvider_SendPushWithContext(t *testing.T) {
	provider := createTestPushProvider()

	pushNotification := &models.PushNotification{
		Notification: models.Notification{
			ID:       uuid.New(),
			Type:     models.NotificationTypePush,
			Priority: models.PriorityNormal,
		},
		DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
		Platform:    "ios",
		Title:       "Test Notification",
		Message:     "This is a test",
	}

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	response, err := provider.SendPush(ctx, pushNotification)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Test with cancelled context
	ctx, cancel = context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = provider.SendPush(ctx, pushNotification)
	assert.Error(t, err)
}

func TestMockPushProvider_PlatformSpecificFormatting(t *testing.T) {
	provider := createTestPushProvider()
	ctx := context.Background()

	// Test iOS formatting
	iosNotification := &models.PushNotification{
		Notification: models.Notification{
			ID:       uuid.New(),
			Type:     models.NotificationTypePush,
			Priority: models.PriorityNormal,
		},
		DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
		Platform:    "ios",
		Title:       "This is a very long title that should be truncated for iOS platform because it exceeds the maximum allowed length",
		Message:     "This is a very long message that should be truncated for iOS platform because it exceeds the maximum allowed length for push notification messages on iOS devices which typically support shorter text",
		Sound:       "",
	}

	_, err := provider.SendPush(ctx, iosNotification)
	assert.NoError(t, err)

	sentPushes := provider.GetSentPushes()
	assert.Len(t, sentPushes, 1)

	sentPush := sentPushes[0]
	assert.True(t, len(sentPush.Title) <= 50)
	assert.True(t, len(sentPush.Body) <= 200)
	assert.Equal(t, "default", sentPush.Sound)

	// Test Android formatting
	provider.ClearSentPushes()

	androidNotification := &models.PushNotification{
		Notification: models.Notification{
			ID:       uuid.New(),
			Type:     models.NotificationTypePush,
			Priority: models.PriorityNormal,
		},
		DeviceToken: "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
		Platform:    "android",
		Title:       "This is a very long title that should be truncated for Android platform",
		Message:     "This is a very long message that should be truncated for Android platform because it exceeds the maximum allowed length for push notification messages on Android devices which support slightly longer text than iOS",
		Icon:        "",
	}

	_, err = provider.SendPush(ctx, androidNotification)
	assert.NoError(t, err)

	sentPushes = provider.GetSentPushes()
	assert.Len(t, sentPushes, 1)

	sentPush = sentPushes[0]
	assert.True(t, len(sentPush.Title) <= 65)
	assert.True(t, len(sentPush.Body) <= 240)
	assert.Equal(t, "ic_notification", sentPush.Icon)
	assert.Equal(t, 0, sentPush.Badge) // Android doesn't use badge
}

func TestMockPushProvider_DeliverySimulation(t *testing.T) {
	provider := createTestPushProvider()
	ctx := context.Background()

	// Send multiple notifications to test delivery simulation
	var responses []*models.NotificationResponse
	for i := 0; i < 20; i++ {
		pushNotification := &models.PushNotification{
			Notification: models.Notification{
				ID:       uuid.New(),
				Type:     models.NotificationTypePush,
				Priority: models.PriorityNormal,
			},
			DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			Platform:    "ios",
			Title:       "Test Notification",
			Message:     "This is a test",
		}

		response, err := provider.SendPush(ctx, pushNotification)
		assert.NoError(t, err)
		responses = append(responses, response)
	}

	// Check delivery simulation (should have ~85% success rate)
	sentPushes := provider.GetSentPushes()
	assert.Len(t, sentPushes, 20)

	deliveredCount := 0
	for _, push := range sentPushes {
		if push.DeliveredAt != nil {
			deliveredCount++
			assert.Equal(t, "delivered", push.Status)
		} else {
			assert.Equal(t, "sent", push.Status)
		}
	}

	// Should have some delivered notifications (probabilistic test)
	assert.True(t, deliveredCount > 0, "Expected at least some notifications to be delivered")
}

// Helper function to create a test push provider
func createTestPushProvider() *MockPushProvider {
	cfg := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	return NewMockPushProvider(cfg)
}
