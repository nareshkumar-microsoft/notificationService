package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/providers"
	"github.com/nareshkumar-microsoft/notificationService/internal/utils"
	"github.com/nareshkumar-microsoft/notificationService/pkg/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPushProvider for testing
type MockPushProvider struct {
	mock.Mock
}

func (m *MockPushProvider) Send(ctx context.Context, notification *models.Notification) (*models.NotificationResponse, error) {
	args := m.Called(ctx, notification)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationResponse), args.Error(1)
}

func (m *MockPushProvider) SendPush(ctx context.Context, push *models.PushNotification) (*models.NotificationResponse, error) {
	args := m.Called(ctx, push)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationResponse), args.Error(1)
}

func (m *MockPushProvider) ValidateDeviceToken(token, platform string) error {
	args := m.Called(token, platform)
	return args.Error(0)
}

func (m *MockPushProvider) GetPlatformConfig(platform string) interfaces.PlatformConfig {
	args := m.Called(platform)
	return args.Get(0).(interfaces.PlatformConfig)
}

func (m *MockPushProvider) GetType() models.NotificationType {
	args := m.Called()
	return args.Get(0).(models.NotificationType)
}

func (m *MockPushProvider) IsHealthy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPushProvider) GetConfig() interfaces.ProviderConfig {
	args := m.Called()
	return args.Get(0).(interfaces.ProviderConfig)
}

func TestNewPushService(t *testing.T) {
	provider := &MockPushProvider{}
	config := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("debug")

	service := NewPushService(provider, config, logger)

	assert.NotNil(t, service)
	assert.Equal(t, provider, service.provider)
	assert.Equal(t, config, service.config)
	assert.Equal(t, logger, service.logger)
}

func TestPushService_SendPushNotification(t *testing.T) {
	testCases := []struct {
		name          string
		request       *models.NotificationRequest
		setupMocks    func(*MockPushProvider)
		expectError   bool
		expectedError string
	}{
		{
			name: "successful iOS push notification",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				Subject:   "Test Subject",
				Body:      "Test message body",
				PushData: &models.PushData{
					DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
					Platform:    "ios",
					Title:       "Test Title",
					Sound:       "default",
				},
			},
			setupMocks: func(m *MockPushProvider) {
				m.On("ValidateDeviceToken", mock.AnythingOfType("string"), "ios").Return(nil)
				m.On("GetPlatformConfig", "ios").Return(interfaces.PlatformConfig{
					Platform:   "ios",
					MaxPayload: 4096,
				})
				m.On("SendPush", mock.Anything, mock.AnythingOfType("*models.PushNotification")).Return(
					&models.NotificationResponse{
						ID:         uuid.New(),
						Status:     models.StatusSent,
						Message:    "Push notification sent successfully",
						ProviderID: "test-provider-id",
						SentAt:     &time.Time{},
					}, nil)
			},
			expectError: false,
		},
		{
			name: "successful Android push notification",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityHigh,
				Recipient: "test@example.com",
				Body:      "Test message body",
				PushData: &models.PushData{
					DeviceToken: "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
					Platform:    "android",
					Title:       "Android Test",
					Icon:        "ic_notification",
				},
			},
			setupMocks: func(m *MockPushProvider) {
				m.On("ValidateDeviceToken", mock.AnythingOfType("string"), "android").Return(nil)
				m.On("GetPlatformConfig", "android").Return(interfaces.PlatformConfig{
					Platform:   "android",
					MaxPayload: 4000,
				})
				m.On("SendPush", mock.Anything, mock.AnythingOfType("*models.PushNotification")).Return(
					&models.NotificationResponse{
						ID:         uuid.New(),
						Status:     models.StatusSent,
						Message:    "Push notification sent successfully",
						ProviderID: "test-provider-id",
						SentAt:     &time.Time{},
					}, nil)
			},
			expectError: false,
		},
		{
			name:    "nil request",
			request: nil,
			setupMocks: func(m *MockPushProvider) {
				// No mocks needed
			},
			expectError:   true,
			expectedError: "notification request is required",
		},
		{
			name: "wrong notification type",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypeEmail,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				Body:      "Test message body",
				PushData:  &models.PushData{},
			},
			setupMocks: func(m *MockPushProvider) {
				// No mocks needed
			},
			expectError:   true,
			expectedError: "notification type must be push",
		},
		{
			name: "missing body",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				PushData:  &models.PushData{},
			},
			setupMocks: func(m *MockPushProvider) {
				// No mocks needed
			},
			expectError:   true,
			expectedError: "notification body is required",
		},
		{
			name: "missing push data",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				Body:      "Test message body",
			},
			setupMocks: func(m *MockPushProvider) {
				// No mocks needed
			},
			expectError:   true,
			expectedError: "push data is required for push notifications",
		},
		{
			name: "missing device token",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				Body:      "Test message body",
				PushData: &models.PushData{
					Platform: "ios",
				},
			},
			setupMocks: func(m *MockPushProvider) {
				// No mocks needed
			},
			expectError:   true,
			expectedError: "device token is required",
		},
		{
			name: "missing platform",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				Body:      "Test message body",
				PushData: &models.PushData{
					DeviceToken: "test-token",
				},
			},
			setupMocks: func(m *MockPushProvider) {
				// No mocks needed
			},
			expectError:   true,
			expectedError: "platform is required",
		},
		{
			name: "invalid device token",
			request: &models.NotificationRequest{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "test@example.com",
				Body:      "Test message body",
				PushData: &models.PushData{
					DeviceToken: "invalid-token",
					Platform:    "ios",
				},
			},
			setupMocks: func(m *MockPushProvider) {
				m.On("GetPlatformConfig", "ios").Return(interfaces.PlatformConfig{
					Platform:   "ios",
					MaxPayload: 4096,
				})
				m.On("ValidateDeviceToken", "invalid-token", "ios").Return(assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &MockPushProvider{}
			tc.setupMocks(provider)

			config := config.PushProviderConfig{
				Provider: "mock",
				Enabled:  true,
			}
			logger := utils.NewSimpleLogger("debug")
			service := NewPushService(provider, config, logger)

			ctx := context.Background()
			response, err := service.SendPushNotification(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tc.expectedError != "" {
					assert.Contains(t, err.Error(), tc.expectedError)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, models.StatusSent, response.Status)
			}

			provider.AssertExpectations(t)
		})
	}
}

func TestPushService_SendBulkPushNotifications(t *testing.T) {
	provider := &MockPushProvider{}
	config := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(provider, config, logger)

	// Test successful bulk send
	t.Run("successful bulk send", func(t *testing.T) {
		requests := []*models.NotificationRequest{
			{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "user1@example.com",
				Body:      "Message 1",
				PushData: &models.PushData{
					DeviceToken: "token1",
					Platform:    "ios",
				},
			},
			{
				Type:      models.NotificationTypePush,
				Priority:  models.PriorityNormal,
				Recipient: "user2@example.com",
				Body:      "Message 2",
				PushData: &models.PushData{
					DeviceToken: "token2",
					Platform:    "android",
				},
			},
		}

		// Setup mocks for both requests
		provider.On("ValidateDeviceToken", "token1", "ios").Return(nil)
		provider.On("GetPlatformConfig", "ios").Return(interfaces.PlatformConfig{
			Platform:   "ios",
			MaxPayload: 4096,
		})
		provider.On("SendPush", mock.Anything, mock.AnythingOfType("*models.PushNotification")).Return(
			&models.NotificationResponse{
				ID:         uuid.New(),
				Status:     models.StatusSent,
				Message:    "Sent",
				ProviderID: "id1",
			}, nil).Once()

		provider.On("ValidateDeviceToken", "token2", "android").Return(nil)
		provider.On("GetPlatformConfig", "android").Return(interfaces.PlatformConfig{
			Platform:   "android",
			MaxPayload: 4000,
		})
		provider.On("SendPush", mock.Anything, mock.AnythingOfType("*models.PushNotification")).Return(
			&models.NotificationResponse{
				ID:         uuid.New(),
				Status:     models.StatusSent,
				Message:    "Sent",
				ProviderID: "id2",
			}, nil).Once()

		ctx := context.Background()
		responses, err := service.SendBulkPushNotifications(ctx, requests)

		assert.NoError(t, err)
		assert.Len(t, responses, 2)
		provider.AssertExpectations(t)
	})

	// Test empty requests
	t.Run("empty requests", func(t *testing.T) {
		ctx := context.Background()
		responses, err := service.SendBulkPushNotifications(ctx, []*models.NotificationRequest{})

		assert.Error(t, err)
		assert.Nil(t, responses)
		assert.Contains(t, err.Error(), "at least one notification request is required")
	})
}

func TestPushService_ValidateDeviceToken(t *testing.T) {
	provider := &MockPushProvider{}
	config := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(provider, config, logger)

	provider.On("ValidateDeviceToken", "valid-token", "ios").Return(nil)
	provider.On("ValidateDeviceToken", "invalid-token", "ios").Return(assert.AnError)

	err := service.ValidateDeviceToken("valid-token", "ios")
	assert.NoError(t, err)

	err = service.ValidateDeviceToken("invalid-token", "ios")
	assert.Error(t, err)

	provider.AssertExpectations(t)
}

func TestPushService_GetPlatformConfig(t *testing.T) {
	provider := &MockPushProvider{}
	config := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(provider, config, logger)

	expectedConfig := interfaces.PlatformConfig{
		Platform:   "ios",
		MaxPayload: 4096,
	}

	provider.On("GetPlatformConfig", "ios").Return(expectedConfig)

	config_result := service.GetPlatformConfig("ios")
	assert.Equal(t, expectedConfig, config_result)

	provider.AssertExpectations(t)
}

func TestPushService_HealthCheck(t *testing.T) {
	provider := &MockPushProvider{}
	config := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(provider, config, logger)

	ctx := context.Background()

	// Test healthy provider
	provider.On("IsHealthy", mock.Anything).Return(nil).Once()
	err := service.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test unhealthy provider
	provider.On("IsHealthy", mock.Anything).Return(assert.AnError).Once()
	err = service.HealthCheck(ctx)
	assert.Error(t, err)

	provider.AssertExpectations(t)
}

func TestPushService_GetProvider(t *testing.T) {
	provider := &MockPushProvider{}
	config := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}
	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(provider, config, logger)

	assert.Equal(t, provider, service.GetProvider())
}

// Integration test with real mock provider
func TestPushService_Integration(t *testing.T) {
	// Use the real mock provider instead of the test mock
	realProvider := providers.NewMockPushProvider(config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	})

	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(realProvider, config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}, logger)

	ctx := context.Background()

	// Test successful push notification
	request := &models.NotificationRequest{
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityNormal,
		Recipient: "test@example.com",
		Subject:   "Test Subject",
		Body:      "Test message body",
		PushData: &models.PushData{
			DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			Platform:    "ios",
			Title:       "Test Title",
			Sound:       "default",
			Data: map[string]string{
				"custom_key": "custom_value",
			},
		},
	}

	response, err := service.SendPushNotification(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusSent, response.Status)
	assert.NotEmpty(t, response.ProviderID)

	// Check that the notification was recorded by the mock provider
	sentPushes := realProvider.GetSentPushes()
	assert.Len(t, sentPushes, 1)
	assert.Equal(t, response.ID, sentPushes[0].ID)
	assert.Equal(t, "ios", sentPushes[0].Platform)
	assert.Equal(t, "Test Title", sentPushes[0].Title)
	assert.Equal(t, "Test message body", sentPushes[0].Body)

	// Test device registration
	err = service.RegisterDevice("a1b2c3d4e5f67890123456789012345678901234567890123456789012345678", "ios", map[string]string{
		"app_version": "1.0.0",
	})
	assert.NoError(t, err)

	// Test device unregistration
	err = service.UnregisterDevice("a1b2c3d4e5f67890123456789012345678901234567890123456789012345678")
	assert.NoError(t, err)

	// Test health check
	err = service.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestPushService_PayloadSizeValidation(t *testing.T) {
	realProvider := providers.NewMockPushProvider(config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	})

	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(realProvider, config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}, logger)

	ctx := context.Background()

	// Create a request with very large payload
	largeData := make(map[string]string)
	for i := 0; i < 100; i++ {
		largeData[fmt.Sprintf("key_%d", i)] = strings.Repeat("a", 100)
	}

	request := &models.NotificationRequest{
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityNormal,
		Recipient: "test@example.com",
		Body:      "Test message body",
		PushData: &models.PushData{
			DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			Platform:    "ios",
			Title:       "Test Title",
			Data:        largeData,
		},
	}

	_, err := service.SendPushNotification(ctx, request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payload too large")
}

func TestPushService_ConvertRequestToPushNotification(t *testing.T) {
	provider := providers.NewMockPushProvider(config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	})

	logger := utils.NewSimpleLogger("debug")
	service := NewPushService(provider, config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}, logger)

	request := &models.NotificationRequest{
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityHigh,
		Recipient: "test@example.com",
		Subject:   "Test Subject",
		Body:      "Test message body",
		Metadata: map[string]string{
			"source": "test",
		},
		PushData: &models.PushData{
			DeviceToken: "test-token",
			Platform:    "ios",
			Title:       "Push Title",
			Icon:        "ic_test",
			Badge:       5,
			Sound:       "custom",
			Data: map[string]string{
				"action": "view",
			},
		},
	}

	// Use reflection to access private method or test through public interface
	pushNotification, err := service.convertRequestToPushNotification(request)

	assert.NoError(t, err)
	assert.NotNil(t, pushNotification)
	assert.Equal(t, models.NotificationTypePush, pushNotification.Type)
	assert.Equal(t, models.PriorityHigh, pushNotification.Priority)
	assert.Equal(t, "test@example.com", pushNotification.Recipient)
	assert.Equal(t, "Test Subject", pushNotification.Subject)
	assert.Equal(t, "Test message body", pushNotification.Body)
	assert.Equal(t, "test-token", pushNotification.DeviceToken)
	assert.Equal(t, "ios", pushNotification.Platform)
	assert.Equal(t, "Push Title", pushNotification.Title)
	assert.Equal(t, "Test message body", pushNotification.Message)
	assert.Equal(t, "ic_test", pushNotification.Icon)
	assert.Equal(t, 5, pushNotification.Badge)
	assert.Equal(t, "custom", pushNotification.Sound)

	// Check that metadata and data are merged
	assert.Contains(t, pushNotification.Data, "source")
	assert.Contains(t, pushNotification.Data, "action")
	assert.Equal(t, "test", pushNotification.Data["source"])
	assert.Equal(t, "view", pushNotification.Data["action"])
}
