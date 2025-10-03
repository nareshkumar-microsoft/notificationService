package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRetrying  NotificationStatus = "retrying"
)

// Priority represents the priority level of a notification
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Notification represents a generic notification
type Notification struct {
	ID          uuid.UUID          `json:"id"`
	Type        NotificationType   `json:"type"`
	Status      NotificationStatus `json:"status"`
	Priority    Priority           `json:"priority"`
	Recipient   string             `json:"recipient"`
	Subject     string             `json:"subject,omitempty"`
	Body        string             `json:"body"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ScheduledAt *time.Time         `json:"scheduled_at,omitempty"`
	SentAt      *time.Time         `json:"sent_at,omitempty"`
	DeliveredAt *time.Time         `json:"delivered_at,omitempty"`
	FailedAt    *time.Time         `json:"failed_at,omitempty"`
	ErrorMsg    string             `json:"error_message,omitempty"`
	RetryCount  int                `json:"retry_count"`
	MaxRetries  int                `json:"max_retries"`
}

// EmailNotification represents an email notification with specific fields
type EmailNotification struct {
	Notification
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	From        string            `json:"from"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	HTMLBody    string            `json:"html_body,omitempty"`
	TextBody    string            `json:"text_body,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// SMSNotification represents an SMS notification with specific fields
type SMSNotification struct {
	Notification
	PhoneNumber string `json:"phone_number"`
	CountryCode string `json:"country_code,omitempty"`
	Message     string `json:"message"`
	Unicode     bool   `json:"unicode"`
}

// PushNotification represents a push notification with specific fields
type PushNotification struct {
	Notification
	DeviceToken string            `json:"device_token"`
	Platform    string            `json:"platform"` // "ios", "android", "web"
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	Icon        string            `json:"icon,omitempty"`
	Badge       int               `json:"badge,omitempty"`
	Sound       string            `json:"sound,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	ClickAction string            `json:"click_action,omitempty"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	Type        NotificationType  `json:"type" validate:"required,oneof=email sms push"`
	Priority    Priority          `json:"priority" validate:"required,oneof=low normal high urgent"`
	Recipient   string            `json:"recipient" validate:"required"`
	Subject     string            `json:"subject,omitempty"`
	Body        string            `json:"body" validate:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`
	MaxRetries  int               `json:"max_retries,omitempty"`

	// Type-specific fields
	EmailData *EmailData `json:"email_data,omitempty"`
	SMSData   *SMSData   `json:"sms_data,omitempty"`
	PushData  *PushData  `json:"push_data,omitempty"`
}

// EmailData contains email-specific request data
type EmailData struct {
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	From        string            `json:"from"`
	ReplyTo     string            `json:"reply_to,omitempty"`
	HTMLBody    string            `json:"html_body,omitempty"`
	TextBody    string            `json:"text_body,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// SMSData contains SMS-specific request data
type SMSData struct {
	PhoneNumber string `json:"phone_number"`
	CountryCode string `json:"country_code,omitempty"`
	Unicode     bool   `json:"unicode"`
}

// PushData contains push notification-specific request data
type PushData struct {
	DeviceToken string            `json:"device_token"`
	Platform    string            `json:"platform" validate:"required,oneof=ios android web"`
	Title       string            `json:"title"`
	Icon        string            `json:"icon,omitempty"`
	Badge       int               `json:"badge,omitempty"`
	Sound       string            `json:"sound,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	ClickAction string            `json:"click_action,omitempty"`
}

// NotificationResponse represents the response after sending a notification
type NotificationResponse struct {
	ID         uuid.UUID          `json:"id"`
	Status     NotificationStatus `json:"status"`
	Message    string             `json:"message"`
	ProviderID string             `json:"provider_id,omitempty"`
	SentAt     *time.Time         `json:"sent_at,omitempty"`
	Error      string             `json:"error,omitempty"`
}

// DeliveryStatus represents the delivery status of a notification
type DeliveryStatus struct {
	NotificationID uuid.UUID          `json:"notification_id"`
	Status         NotificationStatus `json:"status"`
	StatusDetails  string             `json:"status_details,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at"`
	ProviderData   map[string]string  `json:"provider_data,omitempty"`
}
