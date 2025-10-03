package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
)

// GenerateNotificationID generates a unique notification ID
func GenerateNotificationID() uuid.UUID {
	return uuid.New()
}

// ValidateEmailAddress validates an email address format
func ValidateEmailAddress(email string) error {
	if email == "" {
		return errors.NewValidationError("email", "email address is required")
	}

	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.NewValidationError("email", "invalid email address format")
	}

	return nil
}

// ValidatePhoneNumber validates a phone number format
func ValidatePhoneNumber(phoneNumber string, countryCode string) error {
	if phoneNumber == "" {
		return errors.NewValidationError("phone_number", "phone number is required")
	}

	// Remove common formatting characters
	cleanNumber := strings.ReplaceAll(phoneNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "+", "")

	// Basic phone number validation - should contain only digits
	phoneRegex := regexp.MustCompile(`^\d{7,15}$`)
	if !phoneRegex.MatchString(cleanNumber) {
		return errors.NewValidationError("phone_number", "invalid phone number format")
	}

	return nil
}

// ValidateDeviceToken validates a device token for push notifications
func ValidateDeviceToken(token string, platform string) error {
	if token == "" {
		return errors.NewValidationError("device_token", "device token is required")
	}

	switch strings.ToLower(platform) {
	case "ios":
		// iOS device tokens are typically 64 hexadecimal characters
		if len(token) != 64 {
			return errors.NewValidationError("device_token", "invalid iOS device token length")
		}
		tokenRegex := regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
		if !tokenRegex.MatchString(token) {
			return errors.NewValidationError("device_token", "invalid iOS device token format")
		}

	case "android":
		// Android FCM tokens are base64-encoded strings
		if len(token) < 140 || len(token) > 255 {
			return errors.NewValidationError("device_token", "invalid Android device token length")
		}

	case "web":
		// Web push tokens vary in format but should be non-empty
		if len(token) < 10 {
			return errors.NewValidationError("device_token", "invalid web push token")
		}

	default:
		return errors.NewValidationError("platform", "unsupported platform")
	}

	return nil
}

// ValidateNotificationRequest validates a notification request
func ValidateNotificationRequest(request *models.NotificationRequest) error {
	if request == nil {
		return errors.NewValidationError("request", "notification request is required")
	}

	// Validate basic fields
	if request.Type == "" {
		return errors.NewValidationError("type", "notification type is required")
	}

	if request.Recipient == "" {
		return errors.NewValidationError("recipient", "recipient is required")
	}

	if request.Body == "" {
		return errors.NewValidationError("body", "notification body is required")
	}

	// Validate priority
	if !IsValidPriority(request.Priority) {
		return errors.NewValidationError("priority", "invalid priority level")
	}

	// Type-specific validation
	switch request.Type {
	case models.NotificationTypeEmail:
		return validateEmailRequest(request)
	case models.NotificationTypeSMS:
		return validateSMSRequest(request)
	case models.NotificationTypePush:
		return validatePushRequest(request)
	default:
		return errors.NewValidationError("type", "unsupported notification type")
	}
}

// validateEmailRequest validates email-specific fields
func validateEmailRequest(request *models.NotificationRequest) error {
	if err := ValidateEmailAddress(request.Recipient); err != nil {
		return err
	}

	if request.EmailData != nil {
		// Validate additional email addresses
		for _, email := range request.EmailData.To {
			if err := ValidateEmailAddress(email); err != nil {
				return errors.NewValidationError("to", fmt.Sprintf("invalid email in 'to' field: %s", email))
			}
		}

		for _, email := range request.EmailData.CC {
			if err := ValidateEmailAddress(email); err != nil {
				return errors.NewValidationError("cc", fmt.Sprintf("invalid email in 'cc' field: %s", email))
			}
		}

		for _, email := range request.EmailData.BCC {
			if err := ValidateEmailAddress(email); err != nil {
				return errors.NewValidationError("bcc", fmt.Sprintf("invalid email in 'bcc' field: %s", email))
			}
		}

		if request.EmailData.From != "" {
			if err := ValidateEmailAddress(request.EmailData.From); err != nil {
				return errors.NewValidationError("from", "invalid sender email address")
			}
		}

		if request.EmailData.ReplyTo != "" {
			if err := ValidateEmailAddress(request.EmailData.ReplyTo); err != nil {
				return errors.NewValidationError("reply_to", "invalid reply-to email address")
			}
		}
	}

	return nil
}

// validateSMSRequest validates SMS-specific fields
func validateSMSRequest(request *models.NotificationRequest) error {
	phoneNumber := request.Recipient
	countryCode := ""

	if request.SMSData != nil {
		if request.SMSData.PhoneNumber != "" {
			phoneNumber = request.SMSData.PhoneNumber
		}
		countryCode = request.SMSData.CountryCode
	}

	return ValidatePhoneNumber(phoneNumber, countryCode)
}

// validatePushRequest validates push notification-specific fields
func validatePushRequest(request *models.NotificationRequest) error {
	if request.PushData == nil {
		return errors.NewValidationError("push_data", "push notification data is required")
	}

	deviceToken := request.Recipient
	if request.PushData.DeviceToken != "" {
		deviceToken = request.PushData.DeviceToken
	}

	platform := request.PushData.Platform
	if platform == "" {
		return errors.NewValidationError("platform", "platform is required for push notifications")
	}

	return ValidateDeviceToken(deviceToken, platform)
}

// IsValidPriority checks if a priority level is valid
func IsValidPriority(priority models.Priority) bool {
	switch priority {
	case models.PriorityLow, models.PriorityNormal, models.PriorityHigh, models.PriorityUrgent:
		return true
	default:
		return false
	}
}

// IsValidNotificationType checks if a notification type is valid
func IsValidNotificationType(notificationType models.NotificationType) bool {
	switch notificationType {
	case models.NotificationTypeEmail, models.NotificationTypeSMS, models.NotificationTypePush:
		return true
	default:
		return false
	}
}

// IsValidNotificationStatus checks if a notification status is valid
func IsValidNotificationStatus(status models.NotificationStatus) bool {
	switch status {
	case models.StatusPending, models.StatusSent, models.StatusDelivered, models.StatusFailed, models.StatusRetrying:
		return true
	default:
		return false
	}
}

// FormatPhoneNumber formats a phone number for display
func FormatPhoneNumber(phoneNumber string, countryCode string) string {
	// Remove formatting
	cleanNumber := strings.ReplaceAll(phoneNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "+", "")

	// Add country code if provided
	if countryCode != "" && !strings.HasPrefix(cleanNumber, countryCode) {
		cleanNumber = countryCode + cleanNumber
	}

	// Add + prefix
	if !strings.HasPrefix(cleanNumber, "+") {
		cleanNumber = "+" + cleanNumber
	}

	return cleanNumber
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// IsScheduledNotification checks if a notification is scheduled for the future
func IsScheduledNotification(notification *models.Notification) bool {
	return notification.ScheduledAt != nil && notification.ScheduledAt.After(time.Now())
}

// ShouldRetryNotification determines if a notification should be retried
func ShouldRetryNotification(notification *models.Notification) bool {
	return notification.Status == models.StatusFailed &&
		notification.RetryCount < notification.MaxRetries
}

// CalculateNextRetryTime calculates when to retry a failed notification
func CalculateNextRetryTime(retryCount int, baseDelay time.Duration) time.Time {
	// Exponential backoff: baseDelay * 2^retryCount
	delay := baseDelay
	for i := 0; i < retryCount; i++ {
		delay *= 2
	}

	// Cap the delay at 1 hour
	if delay > time.Hour {
		delay = time.Hour
	}

	return time.Now().Add(delay)
}

// SanitizeString removes potentially harmful characters from strings
func SanitizeString(s string) string {
	// Remove null bytes and control characters
	result := strings.ReplaceAll(s, "\x00", "")
	result = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(result, "")
	return strings.TrimSpace(result)
}

// CreateNotificationFromRequest creates a Notification from a NotificationRequest
func CreateNotificationFromRequest(request *models.NotificationRequest) *models.Notification {
	now := time.Now()
	notification := &models.Notification{
		ID:         GenerateNotificationID(),
		Type:       request.Type,
		Status:     models.StatusPending,
		Priority:   request.Priority,
		Recipient:  request.Recipient,
		Subject:    request.Subject,
		Body:       request.Body,
		Metadata:   request.Metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
		RetryCount: 0,
		MaxRetries: 3, // default
	}

	if request.ScheduledAt != nil {
		notification.ScheduledAt = request.ScheduledAt
	}

	if request.MaxRetries > 0 {
		notification.MaxRetries = request.MaxRetries
	}

	return notification
}
