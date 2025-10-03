package interfaces

import (
	"context"

	"github.com/nareshkumar-microsoft/notificationService/internal/models"
)

// NotificationProvider defines the interface that all notification providers must implement
type NotificationProvider interface {
	// Send sends a notification and returns the response
	Send(ctx context.Context, notification *models.Notification) (*models.NotificationResponse, error)

	// GetType returns the type of notifications this provider handles
	GetType() models.NotificationType

	// IsHealthy checks if the provider is healthy and ready to send notifications
	IsHealthy(ctx context.Context) error

	// GetConfig returns the provider configuration
	GetConfig() ProviderConfig
}

// EmailProvider defines the interface for email notification providers
type EmailProvider interface {
	NotificationProvider

	// SendEmail sends an email notification with email-specific features
	SendEmail(ctx context.Context, email *models.EmailNotification) (*models.NotificationResponse, error)

	// ValidateEmailAddress validates an email address format
	ValidateEmailAddress(email string) error

	// GetEmailTemplates returns available email templates
	GetEmailTemplates() []EmailTemplate
}

// SMSProvider defines the interface for SMS notification providers
type SMSProvider interface {
	NotificationProvider

	// SendSMS sends an SMS notification with SMS-specific features
	SendSMS(ctx context.Context, sms *models.SMSNotification) (*models.NotificationResponse, error)

	// ValidatePhoneNumber validates a phone number format
	ValidatePhoneNumber(phoneNumber, countryCode string) error

	// GetSMSCost returns the cost of sending an SMS to a specific country
	GetSMSCost(countryCode string) (float64, error)
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	NotificationProvider

	// SendPush sends a push notification with push-specific features
	SendPush(ctx context.Context, push *models.PushNotification) (*models.NotificationResponse, error)

	// ValidateDeviceToken validates a device token for the specific platform
	ValidateDeviceToken(token, platform string) error

	// GetPlatformConfig returns platform-specific configuration
	GetPlatformConfig(platform string) PlatformConfig
}

// NotificationService defines the main service interface
type NotificationService interface {
	// SendNotification sends a notification using the appropriate provider
	SendNotification(ctx context.Context, request *models.NotificationRequest) (*models.NotificationResponse, error)

	// GetNotificationStatus gets the current status of a notification
	GetNotificationStatus(ctx context.Context, notificationID string) (*models.Notification, error)

	// RegisterProvider registers a new notification provider
	RegisterProvider(provider NotificationProvider) error

	// GetProvider gets a provider by type
	GetProvider(notificationType models.NotificationType) (NotificationProvider, error)

	// ListProviders returns all registered providers
	ListProviders() map[models.NotificationType]NotificationProvider

	// HealthCheck performs a health check on all providers
	HealthCheck(ctx context.Context) map[models.NotificationType]error
}

// NotificationRepository defines the interface for notification storage
type NotificationRepository interface {
	// Save saves a notification to storage
	Save(ctx context.Context, notification *models.Notification) error

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id string) (*models.Notification, error)

	// Update updates a notification in storage
	Update(ctx context.Context, notification *models.Notification) error

	// List retrieves notifications with pagination and filtering
	List(ctx context.Context, filters NotificationFilters) ([]*models.Notification, error)

	// Delete soft deletes a notification
	Delete(ctx context.Context, id string) error

	// GetPendingNotifications gets all pending notifications for processing
	GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error)
}

// Logger defines the interface for logging
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// ProviderConfig represents the configuration for a notification provider
type ProviderConfig struct {
	Name       string                  `json:"name"`
	Type       models.NotificationType `json:"type"`
	Enabled    bool                    `json:"enabled"`
	Priority   int                     `json:"priority"`
	MaxRetries int                     `json:"max_retries"`
	Timeout    int                     `json:"timeout_seconds"`
	RateLimit  RateLimitConfig         `json:"rate_limit"`
	Settings   map[string]string       `json:"settings"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool `json:"enabled"`
	RequestsPerMin int  `json:"requests_per_minute"`
	BurstSize      int  `json:"burst_size"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Subject   string   `json:"subject"`
	HTMLBody  string   `json:"html_body"`
	TextBody  string   `json:"text_body"`
	Variables []string `json:"variables"`
	Category  string   `json:"category"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// PlatformConfig represents platform-specific configuration for push notifications
type PlatformConfig struct {
	Platform   string            `json:"platform"`
	APIKey     string            `json:"api_key"`
	ProjectID  string            `json:"project_id,omitempty"`
	BundleID   string            `json:"bundle_id,omitempty"`
	TeamID     string            `json:"team_id,omitempty"`
	MaxPayload int               `json:"max_payload_size"`
	Settings   map[string]string `json:"settings"`
}

// NotificationFilters represents filters for querying notifications
type NotificationFilters struct {
	Type      *models.NotificationType   `json:"type,omitempty"`
	Status    *models.NotificationStatus `json:"status,omitempty"`
	Priority  *models.Priority           `json:"priority,omitempty"`
	Recipient string                     `json:"recipient,omitempty"`
	DateFrom  *string                    `json:"date_from,omitempty"`
	DateTo    *string                    `json:"date_to,omitempty"`
	Limit     int                        `json:"limit"`
	Offset    int                        `json:"offset"`
	SortBy    string                     `json:"sort_by"`
	SortOrder string                     `json:"sort_order"`
}
