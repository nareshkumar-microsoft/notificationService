package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationTypes(t *testing.T) {
	tests := []struct {
		name         string
		notification NotificationType
		expected     string
	}{
		{"Email type", NotificationTypeEmail, "email"},
		{"SMS type", NotificationTypeSMS, "sms"},
		{"Push type", NotificationTypePush, "push"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.notification))
		})
	}
}

func TestNotificationStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   NotificationStatus
		expected string
	}{
		{"Pending status", StatusPending, "pending"},
		{"Sent status", StatusSent, "sent"},
		{"Delivered status", StatusDelivered, "delivered"},
		{"Failed status", StatusFailed, "failed"},
		{"Retrying status", StatusRetrying, "retrying"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority Priority
		expected string
	}{
		{"Low priority", PriorityLow, "low"},
		{"Normal priority", PriorityNormal, "normal"},
		{"High priority", PriorityHigh, "high"},
		{"Urgent priority", PriorityUrgent, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.priority))
		})
	}
}

func TestNotificationCreation(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	notification := &Notification{
		ID:         id,
		Type:       NotificationTypeEmail,
		Status:     StatusPending,
		Priority:   PriorityNormal,
		Recipient:  "test@example.com",
		Subject:    "Test Subject",
		Body:       "Test Body",
		CreatedAt:  now,
		UpdatedAt:  now,
		RetryCount: 0,
		MaxRetries: 3,
	}

	assert.Equal(t, id, notification.ID)
	assert.Equal(t, NotificationTypeEmail, notification.Type)
	assert.Equal(t, StatusPending, notification.Status)
	assert.Equal(t, PriorityNormal, notification.Priority)
	assert.Equal(t, "test@example.com", notification.Recipient)
	assert.Equal(t, "Test Subject", notification.Subject)
	assert.Equal(t, "Test Body", notification.Body)
	assert.Equal(t, 0, notification.RetryCount)
	assert.Equal(t, 3, notification.MaxRetries)
}

func TestEmailNotification(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	emailNotification := &EmailNotification{
		Notification: Notification{
			ID:        id,
			Type:      NotificationTypeEmail,
			Status:    StatusPending,
			Priority:  PriorityNormal,
			Recipient: "test@example.com",
			Subject:   "Test Email",
			Body:      "Test Body",
			CreatedAt: now,
			UpdatedAt: now,
		},
		To:       []string{"test@example.com", "test2@example.com"},
		From:     "sender@example.com",
		HTMLBody: "<h1>Test HTML Body</h1>",
		TextBody: "Test Text Body",
	}

	assert.Equal(t, NotificationTypeEmail, emailNotification.Type)
	assert.Len(t, emailNotification.To, 2)
	assert.Contains(t, emailNotification.To, "test@example.com")
	assert.Contains(t, emailNotification.To, "test2@example.com")
	assert.Equal(t, "sender@example.com", emailNotification.From)
	assert.Equal(t, "<h1>Test HTML Body</h1>", emailNotification.HTMLBody)
	assert.Equal(t, "Test Text Body", emailNotification.TextBody)
}

func TestSMSNotification(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	smsNotification := &SMSNotification{
		Notification: Notification{
			ID:        id,
			Type:      NotificationTypeSMS,
			Status:    StatusPending,
			Priority:  PriorityHigh,
			Recipient: "+1234567890",
			Body:      "Test SMS Message",
			CreatedAt: now,
			UpdatedAt: now,
		},
		PhoneNumber: "+1234567890",
		CountryCode: "US",
		Message:     "Test SMS Message",
		Unicode:     false,
	}

	assert.Equal(t, NotificationTypeSMS, smsNotification.Type)
	assert.Equal(t, "+1234567890", smsNotification.PhoneNumber)
	assert.Equal(t, "US", smsNotification.CountryCode)
	assert.Equal(t, "Test SMS Message", smsNotification.Message)
	assert.False(t, smsNotification.Unicode)
}

func TestPushNotification(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	pushNotification := &PushNotification{
		Notification: Notification{
			ID:        id,
			Type:      NotificationTypePush,
			Status:    StatusPending,
			Priority:  PriorityUrgent,
			Recipient: "device_token_123",
			Body:      "Test Push Message",
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeviceToken: "device_token_123",
		Platform:    "ios",
		Title:       "Test Push Title",
		Message:     "Test Push Message",
		Badge:       1,
		Sound:       "default",
	}

	assert.Equal(t, NotificationTypePush, pushNotification.Type)
	assert.Equal(t, "device_token_123", pushNotification.DeviceToken)
	assert.Equal(t, "ios", pushNotification.Platform)
	assert.Equal(t, "Test Push Title", pushNotification.Title)
	assert.Equal(t, "Test Push Message", pushNotification.Message)
	assert.Equal(t, 1, pushNotification.Badge)
	assert.Equal(t, "default", pushNotification.Sound)
}

func TestNotificationRequest(t *testing.T) {
	request := &NotificationRequest{
		Type:      NotificationTypeEmail,
		Priority:  PriorityNormal,
		Recipient: "test@example.com",
		Subject:   "Test Request",
		Body:      "Test Body",
		Metadata: map[string]string{
			"campaign_id": "123",
			"user_id":     "456",
		},
		MaxRetries: 5,
		EmailData: &EmailData{
			To:       []string{"test@example.com"},
			From:     "sender@example.com",
			HTMLBody: "<p>HTML Body</p>",
			TextBody: "Text Body",
		},
	}

	assert.Equal(t, NotificationTypeEmail, request.Type)
	assert.Equal(t, PriorityNormal, request.Priority)
	assert.Equal(t, "test@example.com", request.Recipient)
	assert.Equal(t, 5, request.MaxRetries)
	require.NotNil(t, request.EmailData)
	assert.Len(t, request.EmailData.To, 1)
	assert.Contains(t, request.EmailData.To, "test@example.com")
	assert.Equal(t, "sender@example.com", request.EmailData.From)
	assert.Contains(t, request.Metadata, "campaign_id")
	assert.Equal(t, "123", request.Metadata["campaign_id"])
}

func TestNotificationResponse(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	response := &NotificationResponse{
		ID:         id,
		Status:     StatusSent,
		Message:    "Successfully sent",
		ProviderID: "provider_123",
		SentAt:     &now,
	}

	assert.Equal(t, id, response.ID)
	assert.Equal(t, StatusSent, response.Status)
	assert.Equal(t, "Successfully sent", response.Message)
	assert.Equal(t, "provider_123", response.ProviderID)
	require.NotNil(t, response.SentAt)
	assert.Equal(t, now, *response.SentAt)
}
