package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
	"github.com/nareshkumar-microsoft/notificationService/pkg/interfaces"
)

// PushService implements push notification business logic
type PushService struct {
	provider interfaces.PushProvider
	config   config.PushProviderConfig
	logger   interfaces.Logger
}

// NewPushService creates a new push notification service
func NewPushService(provider interfaces.PushProvider, config config.PushProviderConfig, logger interfaces.Logger) *PushService {
	return &PushService{
		provider: provider,
		config:   config,
		logger:   logger,
	}
}

// SendPushNotification sends a push notification with validation and processing
func (s *PushService) SendPushNotification(ctx context.Context, request *models.NotificationRequest) (*models.NotificationResponse, error) {
	// Validate request first
	if err := s.validatePushRequest(request); err != nil {
		s.logger.Errorf("Push request validation failed: %v", err)
		return nil, err
	}

	s.logger.Infof("Processing push notification request for recipient: %s", request.Recipient)

	// Convert request to push notification
	pushNotification, err := s.convertRequestToPushNotification(request)
	if err != nil {
		s.logger.Errorf("Failed to convert request to push notification: %v", err)
		return nil, err
	}

	// Pre-processing: apply templates, enrich data, etc.
	if err := s.preprocessPushNotification(pushNotification); err != nil {
		s.logger.Errorf("Push notification preprocessing failed: %v", err)
		return nil, err
	}

	// Validate device token and platform
	if err := s.provider.ValidateDeviceToken(pushNotification.DeviceToken, pushNotification.Platform); err != nil {
		s.logger.Errorf("Device token validation failed: %v", err)
		return nil, err
	}

	// Send notification through provider
	s.logger.Debugf("Sending push notification via provider")
	response, err := s.provider.SendPush(ctx, pushNotification)
	if err != nil {
		s.logger.Errorf("Failed to send push notification: %v", err)
		return nil, err
	}

	s.logger.Infof("Push notification sent successfully. ID: %s, Provider ID: %s",
		response.ID, response.ProviderID)

	return response, nil
}

// SendBulkPushNotifications sends multiple push notifications efficiently
func (s *PushService) SendBulkPushNotifications(ctx context.Context, requests []*models.NotificationRequest) ([]*models.NotificationResponse, error) {
	s.logger.Infof("Processing bulk push notification request with %d notifications", len(requests))

	if len(requests) == 0 {
		return nil, errors.NewValidationError("requests", "at least one notification request is required")
	}

	responses := make([]*models.NotificationResponse, 0, len(requests))
	var errors []error

	for i, request := range requests {
		s.logger.Debugf("Processing bulk notification %d/%d", i+1, len(requests))

		response, err := s.SendPushNotification(ctx, request)
		if err != nil {
			s.logger.Warnf("Failed to send bulk notification %d: %v", i+1, err)
			errors = append(errors, fmt.Errorf("notification %d: %w", i+1, err))
			continue
		}

		responses = append(responses, response)
	}

	if len(errors) > 0 {
		s.logger.Warnf("Bulk push operation completed with %d errors out of %d notifications",
			len(errors), len(requests))

		// Return partial success with error details
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}

		return responses, fmt.Errorf("bulk push operation had errors: %s",
			strings.Join(errorMessages, "; "))
	}

	s.logger.Infof("Bulk push operation completed successfully for %d notifications", len(responses))
	return responses, nil
}

// ValidateDeviceToken validates a device token for the given platform
func (s *PushService) ValidateDeviceToken(deviceToken, platform string) error {
	return s.provider.ValidateDeviceToken(deviceToken, platform)
}

// GetPlatformConfig returns platform-specific configuration
func (s *PushService) GetPlatformConfig(platform string) interfaces.PlatformConfig {
	return s.provider.GetPlatformConfig(platform)
}

// RegisterDevice registers a device for push notifications
func (s *PushService) RegisterDevice(deviceToken, platform string, metadata map[string]string) error {
	s.logger.Infof("Registering device for platform %s", platform)

	// Validate device token
	if err := s.ValidateDeviceToken(deviceToken, platform); err != nil {
		s.logger.Errorf("Device token validation failed during registration: %v", err)
		return err
	}

	// If provider supports device registration, use it
	if deviceProvider, ok := s.provider.(interface {
		RegisterDevice(token, platform string, metadata map[string]string) error
	}); ok {
		if err := deviceProvider.RegisterDevice(deviceToken, platform, metadata); err != nil {
			s.logger.Errorf("Device registration failed: %v", err)
			return err
		}
	}

	s.logger.Infof("Device registered successfully for platform %s", platform)
	return nil
}

// UnregisterDevice removes a device from push notifications
func (s *PushService) UnregisterDevice(deviceToken string) error {
	s.logger.Infof("Unregistering device with token: %s", deviceToken[:8]+"...")

	// If provider supports device unregistration, use it
	if deviceProvider, ok := s.provider.(interface {
		UnregisterDevice(token string) error
	}); ok {
		if err := deviceProvider.UnregisterDevice(deviceToken); err != nil {
			s.logger.Errorf("Device unregistration failed: %v", err)
			return err
		}
	}

	s.logger.Infof("Device unregistered successfully")
	return nil
}

// GetDeliveryReport gets delivery report for a push notification
func (s *PushService) GetDeliveryReport(notificationID uuid.UUID) (*models.DeliveryStatus, error) {
	s.logger.Debugf("Retrieving delivery report for notification: %s", notificationID)

	// If provider supports delivery reports, use it
	if reportProvider, ok := s.provider.(interface {
		GetDeliveryReport(notificationID uuid.UUID) (*models.DeliveryStatus, error)
	}); ok {
		return reportProvider.GetDeliveryReport(notificationID)
	}

	// Return a default status if provider doesn't support delivery reports
	return &models.DeliveryStatus{
		NotificationID: notificationID,
		Status:         models.StatusSent,
		StatusDetails:  "Provider does not support delivery reports",
		UpdatedAt:      time.Now(),
	}, nil
}

// GetSupportedPlatforms returns list of supported push platforms
func (s *PushService) GetSupportedPlatforms() []string {
	if platformProvider, ok := s.provider.(interface {
		GetSupportedPlatforms() []string
	}); ok {
		return platformProvider.GetSupportedPlatforms()
	}

	// Return default platforms
	return []string{"ios", "android", "web"}
}

// HealthCheck performs a health check on the push service
func (s *PushService) HealthCheck(ctx context.Context) error {
	s.logger.Debug("Performing push service health check")

	if err := s.provider.IsHealthy(ctx); err != nil {
		s.logger.Errorf("Push provider health check failed: %v", err)
		return fmt.Errorf("push provider unhealthy: %w", err)
	}

	s.logger.Debug("Push service health check passed")
	return nil
}

// GetProvider returns the underlying push provider
func (s *PushService) GetProvider() interfaces.PushProvider {
	return s.provider
}

// Helper methods

// validatePushRequest validates a push notification request
func (s *PushService) validatePushRequest(request *models.NotificationRequest) error {
	if request == nil {
		return errors.NewValidationError("request", "notification request is required")
	}

	if request.Type != models.NotificationTypePush {
		return errors.NewValidationError("type", "notification type must be push")
	}

	if request.Body == "" {
		return errors.NewValidationError("body", "notification body is required")
	}

	if request.PushData == nil {
		return errors.NewValidationError("push_data", "push data is required for push notifications")
	}

	pushData := request.PushData

	if pushData.DeviceToken == "" {
		return errors.NewValidationError("device_token", "device token is required")
	}

	if pushData.Platform == "" {
		return errors.NewValidationError("platform", "platform is required")
	}

	// Validate platform
	supportedPlatforms := s.GetSupportedPlatforms()
	platformSupported := false
	for _, supported := range supportedPlatforms {
		if strings.ToLower(pushData.Platform) == supported {
			platformSupported = true
			break
		}
	}

	if !platformSupported {
		return errors.NewValidationError("platform",
			fmt.Sprintf("unsupported platform: %s. Supported: %v",
				pushData.Platform, supportedPlatforms))
	}

	return nil
}

// convertRequestToPushNotification converts a notification request to a push notification
func (s *PushService) convertRequestToPushNotification(request *models.NotificationRequest) (*models.PushNotification, error) {
	pushData := request.PushData

	// Create base notification
	now := time.Now()
	notification := models.Notification{
		ID:          uuid.New(),
		Type:        models.NotificationTypePush,
		Status:      models.StatusPending,
		Priority:    request.Priority,
		Recipient:   request.Recipient,
		Subject:     request.Subject,
		Body:        request.Body,
		Metadata:    request.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
		ScheduledAt: request.ScheduledAt,
		RetryCount:  0,
		MaxRetries:  request.MaxRetries,
	}

	// Set default max retries if not specified
	if notification.MaxRetries == 0 {
		notification.MaxRetries = 3
	}

	// Create push notification
	pushNotification := &models.PushNotification{
		Notification: notification,
		DeviceToken:  pushData.DeviceToken,
		Platform:     strings.ToLower(pushData.Platform),
		Title:        pushData.Title,
		Message:      request.Body,
		Icon:         pushData.Icon,
		Badge:        pushData.Badge,
		Sound:        pushData.Sound,
		Data:         pushData.Data,
		ImageURL:     pushData.ImageURL,
		ClickAction:  pushData.ClickAction,
	}

	// Use subject as title if title is not provided
	if pushNotification.Title == "" && request.Subject != "" {
		pushNotification.Title = request.Subject
	}

	// Set default sound if not provided
	if pushNotification.Sound == "" {
		pushNotification.Sound = "default"
	}

	// Initialize data map if nil
	if pushNotification.Data == nil {
		pushNotification.Data = make(map[string]string)
	}

	// Add metadata to data
	for key, value := range request.Metadata {
		pushNotification.Data[key] = value
	}

	return pushNotification, nil
}

// preprocessPushNotification performs preprocessing on the push notification
func (s *PushService) preprocessPushNotification(push *models.PushNotification) error {
	// Apply platform-specific configurations
	platformConfig := s.GetPlatformConfig(push.Platform)

	// Validate payload size
	if s.estimatePayloadSize(push) > platformConfig.MaxPayload {
		return errors.NewValidationError("payload",
			fmt.Sprintf("payload too large for platform %s (max: %d bytes)",
				push.Platform, platformConfig.MaxPayload))
	}

	// Platform-specific preprocessing
	switch push.Platform {
	case "ios":
		return s.preprocessIOSNotification(push, platformConfig)
	case "android":
		return s.preprocessAndroidNotification(push, platformConfig)
	case "web":
		return s.preprocessWebNotification(push, platformConfig)
	}

	return nil
}

// preprocessIOSNotification handles iOS-specific preprocessing
func (s *PushService) preprocessIOSNotification(push *models.PushNotification, config interfaces.PlatformConfig) error {
	// Truncate title and message if too long
	if len(push.Title) > 50 {
		push.Title = push.Title[:47] + "..."
	}
	if len(push.Message) > 200 {
		push.Message = push.Message[:197] + "..."
	}

	// Ensure badge is non-negative
	if push.Badge < 0 {
		push.Badge = 0
	}

	// Set default sound if empty
	if push.Sound == "" {
		push.Sound = "default"
	}

	return nil
}

// preprocessAndroidNotification handles Android-specific preprocessing
func (s *PushService) preprocessAndroidNotification(push *models.PushNotification, config interfaces.PlatformConfig) error {
	// Truncate title and message if too long
	if len(push.Title) > 65 {
		push.Title = push.Title[:62] + "..."
	}
	if len(push.Message) > 240 {
		push.Message = push.Message[:237] + "..."
	}

	// Set default icon if empty
	if push.Icon == "" {
		push.Icon = "ic_notification"
	}

	// Android doesn't use badge, so reset it
	push.Badge = 0

	return nil
}

// preprocessWebNotification handles Web push-specific preprocessing
func (s *PushService) preprocessWebNotification(push *models.PushNotification, config interfaces.PlatformConfig) error {
	// Truncate title and message if too long
	if len(push.Title) > 50 {
		push.Title = push.Title[:47] + "..."
	}
	if len(push.Message) > 120 {
		push.Message = push.Message[:117] + "..."
	}

	// Set default icon if empty
	if push.Icon == "" {
		push.Icon = "/icon-192x192.png"
	}

	// Web push doesn't typically use sound
	push.Sound = ""

	return nil
}

// estimatePayloadSize estimates the size of the push notification payload
func (s *PushService) estimatePayloadSize(push *models.PushNotification) int {
	size := 0

	// Basic fields
	size += len(push.Title) + len(push.Message) + len(push.Icon) + len(push.Sound)
	size += len(push.ImageURL) + len(push.ClickAction)

	// Data map
	for key, value := range push.Data {
		size += len(key) + len(value)
	}

	// Add overhead for JSON structure (~200 bytes)
	size += 200

	return size
}
