package providers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
	"github.com/nareshkumar-microsoft/notificationService/pkg/interfaces"
)

// MockPushProvider implements the PushProvider interface for testing and development
type MockPushProvider struct {
	config          config.PushProviderConfig
	templates       map[string]*PushTemplate
	sentPushes      []SentPush
	healthy         bool
	deviceTokens    map[string]*DeviceInfo
	platformConfigs map[string]*PlatformConfig
}

// PushTemplate represents a push notification template
type PushTemplate struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Platform  string            `json:"platform"` // "all", "ios", "android", "web"
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Icon      string            `json:"icon,omitempty"`
	Sound     string            `json:"sound,omitempty"`
	Variables []string          `json:"variables"`
	Category  string            `json:"category"`
	Actions   []PushAction      `json:"actions,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PushAction represents an action button in a push notification
type PushAction struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon,omitempty"`
}

// SentPush represents a push notification that was sent (for mock tracking)
type SentPush struct {
	ID           uuid.UUID         `json:"id"`
	DeviceToken  string            `json:"device_token"`
	Platform     string            `json:"platform"`
	Title        string            `json:"title"`
	Body         string            `json:"body"`
	Icon         string            `json:"icon,omitempty"`
	Badge        int               `json:"badge,omitempty"`
	Sound        string            `json:"sound,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	ImageURL     string            `json:"image_url,omitempty"`
	ClickAction  string            `json:"click_action,omitempty"`
	SentAt       time.Time         `json:"sent_at"`
	DeliveredAt  *time.Time        `json:"delivered_at,omitempty"`
	Status       string            `json:"status"`
	ProviderData map[string]string `json:"provider_data,omitempty"`
}

// DeviceInfo represents information about a registered device
type DeviceInfo struct {
	Token        string            `json:"token"`
	Platform     string            `json:"platform"`
	AppVersion   string            `json:"app_version,omitempty"`
	OSVersion    string            `json:"os_version,omitempty"`
	RegisteredAt time.Time         `json:"registered_at"`
	LastSeen     time.Time         `json:"last_seen"`
	Active       bool              `json:"active"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// PlatformConfig represents platform-specific configuration
type PlatformConfig struct {
	Platform        string            `json:"platform"`
	MaxPayloadSize  int               `json:"max_payload_size"`
	MaxTitleLength  int               `json:"max_title_length"`
	MaxBodyLength   int               `json:"max_body_length"`
	SupportsBadge   bool              `json:"supports_badge"`
	SupportsSound   bool              `json:"supports_sound"`
	SupportsImage   bool              `json:"supports_image"`
	SupportsActions bool              `json:"supports_actions"`
	DefaultSound    string            `json:"default_sound,omitempty"`
	Settings        map[string]string `json:"settings"`
}

// NewMockPushProvider creates a new mock push provider
func NewMockPushProvider(cfg config.PushProviderConfig) *MockPushProvider {
	provider := &MockPushProvider{
		config:          cfg,
		templates:       make(map[string]*PushTemplate),
		sentPushes:      make([]SentPush, 0),
		healthy:         true,
		deviceTokens:    make(map[string]*DeviceInfo),
		platformConfigs: make(map[string]*PlatformConfig),
	}

	// Load default templates and platform configs
	provider.loadDefaultTemplates()
	provider.loadPlatformConfigs()

	return provider
}

// Send implements the NotificationProvider interface
func (p *MockPushProvider) Send(ctx context.Context, notification *models.Notification) (*models.NotificationResponse, error) {
	if !p.healthy {
		return nil, errors.NewProviderError("mock-push", errors.ErrorCodeProviderUnavailable, "provider is unhealthy")
	}

	// Convert generic notification to push notification
	pushNotification, err := p.convertToPushNotification(notification)
	if err != nil {
		return nil, err
	}

	return p.SendPush(ctx, pushNotification)
}

// SendPush implements the PushProvider interface
func (p *MockPushProvider) SendPush(ctx context.Context, push *models.PushNotification) (*models.NotificationResponse, error) {
	if !p.healthy {
		return nil, errors.NewProviderError("mock-push", errors.ErrorCodeProviderUnavailable, "provider is unhealthy")
	}

	// Format notification for platform first (this includes truncation)
	formattedPush, err := p.formatForPlatform(push)
	if err != nil {
		return nil, err
	}

	// Validate the formatted push notification
	if err := p.validatePushNotification(formattedPush); err != nil {
		return nil, err
	}

	// Simulate processing delay based on platform
	delay := p.getPlatformDelay(push.Platform)
	select {
	case <-ctx.Done():
		return nil, errors.NewNotificationError(errors.ErrorCodeTimeout, "push notification sending timed out")
	case <-time.After(delay):
		// Continue processing
	}

	// Create sent push record
	sentPush := SentPush{
		ID:          push.ID,
		DeviceToken: push.DeviceToken,
		Platform:    push.Platform,
		Title:       formattedPush.Title,
		Body:        formattedPush.Message,
		Icon:        formattedPush.Icon,
		Badge:       formattedPush.Badge,
		Sound:       formattedPush.Sound,
		Data:        formattedPush.Data,
		ImageURL:    formattedPush.ImageURL,
		ClickAction: formattedPush.ClickAction,
		SentAt:      time.Now(),
		Status:      "sent",
		ProviderData: map[string]string{
			"provider":    "mock-push",
			"message_id":  fmt.Sprintf("push-%s", push.ID.String()),
			"platform":    push.Platform,
			"queue_time":  delay.String(),
			"retry_count": "0",
		},
	}

	// Simulate delivery (85% success rate)
	if time.Now().UnixNano()%100 < 85 {
		deliveryDelay := time.Duration(500+time.Now().UnixNano()%2000) * time.Millisecond
		deliveredAt := time.Now().Add(deliveryDelay)
		sentPush.DeliveredAt = &deliveredAt
		sentPush.Status = "delivered"
		sentPush.ProviderData["delivery_time"] = deliveredAt.Format(time.RFC3339)
		sentPush.ProviderData["delivery_delay"] = deliveryDelay.String()
	}

	// Store sent push for tracking
	p.sentPushes = append(p.sentPushes, sentPush)

	// Update device last seen
	p.updateDeviceLastSeen(push.DeviceToken)

	// Create response
	now := time.Now()
	response := &models.NotificationResponse{
		ID:         push.ID,
		Status:     models.StatusSent,
		Message:    fmt.Sprintf("Push notification sent to %s device", push.Platform),
		ProviderID: sentPush.ProviderData["message_id"],
		SentAt:     &now,
	}

	return response, nil
}

// ValidateDeviceToken implements the PushProvider interface
func (p *MockPushProvider) ValidateDeviceToken(token, platform string) error {
	if token == "" {
		return errors.NewValidationError("device_token", "device token is required")
	}

	platform = strings.ToLower(platform)

	switch platform {
	case "ios":
		return p.validateIOSToken(token)
	case "android":
		return p.validateAndroidToken(token)
	case "web":
		return p.validateWebToken(token)
	default:
		return errors.NewValidationError("platform", fmt.Sprintf("unsupported platform: %s", platform))
	}
}

// GetPlatformConfig implements the PushProvider interface
func (p *MockPushProvider) GetPlatformConfig(platform string) interfaces.PlatformConfig {
	config, exists := p.platformConfigs[strings.ToLower(platform)]
	if !exists {
		// Return default config
		return interfaces.PlatformConfig{
			Platform:   platform,
			MaxPayload: 4096,
			Settings:   make(map[string]string),
		}
	}

	return interfaces.PlatformConfig{
		Platform:   config.Platform,
		APIKey:     config.Settings["api_key"],
		ProjectID:  config.Settings["project_id"],
		BundleID:   config.Settings["bundle_id"],
		TeamID:     config.Settings["team_id"],
		MaxPayload: config.MaxPayloadSize,
		Settings:   config.Settings,
	}
}

// GetType implements the NotificationProvider interface
func (p *MockPushProvider) GetType() models.NotificationType {
	return models.NotificationTypePush
}

// IsHealthy implements the NotificationProvider interface
func (p *MockPushProvider) IsHealthy(ctx context.Context) error {
	if !p.healthy {
		return errors.NewProviderError("mock-push", errors.ErrorCodeProviderUnavailable, "provider is marked as unhealthy")
	}

	// Simulate health check delay
	select {
	case <-ctx.Done():
		return errors.NewNotificationError(errors.ErrorCodeTimeout, "health check timed out")
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// GetConfig implements the NotificationProvider interface
func (p *MockPushProvider) GetConfig() interfaces.ProviderConfig {
	return interfaces.ProviderConfig{
		Name:       "Mock Push Provider",
		Type:       models.NotificationTypePush,
		Enabled:    p.config.Enabled,
		Priority:   3,
		MaxRetries: 3,
		Timeout:    45,
		RateLimit: interfaces.RateLimitConfig{
			Enabled:        true,
			RequestsPerMin: 1000,
			BurstSize:      50,
		},
		Settings: map[string]string{
			"provider_type":       "mock",
			"version":             "1.0.0",
			"features":            "templates,validation,platform_specific,device_management",
			"supported_platforms": "ios,android,web",
		},
	}
}

// RegisterDevice registers a device token for push notifications
func (p *MockPushProvider) RegisterDevice(token, platform string, metadata map[string]string) error {
	if err := p.ValidateDeviceToken(token, platform); err != nil {
		return err
	}

	now := time.Now()
	deviceInfo := &DeviceInfo{
		Token:        token,
		Platform:     strings.ToLower(platform),
		RegisteredAt: now,
		LastSeen:     now,
		Active:       true,
		Metadata:     metadata,
	}

	if metadata != nil {
		deviceInfo.AppVersion = metadata["app_version"]
		deviceInfo.OSVersion = metadata["os_version"]
	}

	p.deviceTokens[token] = deviceInfo
	return nil
}

// UnregisterDevice removes a device token
func (p *MockPushProvider) UnregisterDevice(token string) error {
	if _, exists := p.deviceTokens[token]; !exists {
		return errors.NewNotificationError(errors.ErrorCodeNotFound, "device token not found")
	}

	delete(p.deviceTokens, token)
	return nil
}

// GetDeviceInfo returns information about a registered device
func (p *MockPushProvider) GetDeviceInfo(token string) (*DeviceInfo, error) {
	device, exists := p.deviceTokens[token]
	if !exists {
		return nil, errors.NewNotificationError(errors.ErrorCodeNotFound, "device token not found")
	}

	return device, nil
}

// GetTemplate retrieves a push template by ID
func (p *MockPushProvider) GetTemplate(templateID string) (*PushTemplate, error) {
	template, exists := p.templates[templateID]
	if !exists {
		return nil, errors.NewNotificationError(errors.ErrorCodeTemplateNotFound, fmt.Sprintf("template not found: %s", templateID))
	}
	return template, nil
}

// AddTemplate adds a new push template
func (p *MockPushProvider) AddTemplate(template *PushTemplate) error {
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	p.templates[template.ID] = template
	return nil
}

// RenderTemplate renders a push template with provided data
func (p *MockPushProvider) RenderTemplate(templateID string, data map[string]string) (*PushTemplate, error) {
	template, err := p.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Clone template for rendering
	rendered := &PushTemplate{
		ID:        template.ID,
		Name:      template.Name,
		Platform:  template.Platform,
		Title:     p.replaceVariables(template.Title, data),
		Body:      p.replaceVariables(template.Body, data),
		Icon:      template.Icon,
		Sound:     template.Sound,
		Variables: template.Variables,
		Category:  template.Category,
		Actions:   template.Actions,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
		Metadata:  template.Metadata,
	}

	return rendered, nil
}

// GetSentPushes returns all sent push notifications (for testing)
func (p *MockPushProvider) GetSentPushes() []SentPush {
	return p.sentPushes
}

// ClearSentPushes clears the sent push notifications history (for testing)
func (p *MockPushProvider) ClearSentPushes() {
	p.sentPushes = make([]SentPush, 0)
}

// SetHealthy sets the provider health status (for testing)
func (p *MockPushProvider) SetHealthy(healthy bool) {
	p.healthy = healthy
}

// GetSupportedPlatforms returns list of supported platforms
func (p *MockPushProvider) GetSupportedPlatforms() []string {
	return []string{"ios", "android", "web"}
}

// Helper methods

// convertToPushNotification converts a generic notification to a push notification
func (p *MockPushProvider) convertToPushNotification(notification *models.Notification) (*models.PushNotification, error) {
	if notification.Type != models.NotificationTypePush {
		return nil, errors.NewValidationError("type", "notification type must be push")
	}

	// Extract platform and device token from metadata or recipient
	platform := "android" // default
	deviceToken := notification.Recipient

	if notification.Metadata != nil {
		if p, exists := notification.Metadata["platform"]; exists {
			platform = p
		}
		if t, exists := notification.Metadata["device_token"]; exists {
			deviceToken = t
		}
	}

	pushNotification := &models.PushNotification{
		Notification: *notification,
		DeviceToken:  deviceToken,
		Platform:     platform,
		Title:        notification.Subject,
		Message:      notification.Body,
		Sound:        "default",
	}

	return pushNotification, nil
}

// validatePushNotification validates a push notification
func (p *MockPushProvider) validatePushNotification(push *models.PushNotification) error {
	// Validate device token
	if err := p.ValidateDeviceToken(push.DeviceToken, push.Platform); err != nil {
		return err
	}

	// Validate platform
	supportedPlatforms := p.GetSupportedPlatforms()
	platformSupported := false
	for _, supported := range supportedPlatforms {
		if strings.ToLower(push.Platform) == supported {
			platformSupported = true
			break
		}
	}

	if !platformSupported {
		return errors.NewValidationError("platform", fmt.Sprintf("unsupported platform: %s", push.Platform))
	}

	// Validate content
	if push.Title == "" && push.Message == "" {
		return errors.NewValidationError("content", "push notification must have either title or message")
	}

	// Platform-specific validation
	config := p.platformConfigs[strings.ToLower(push.Platform)]
	if config != nil {
		if len(push.Title) > config.MaxTitleLength {
			return errors.NewValidationError("title", fmt.Sprintf("title too long (max %d characters)", config.MaxTitleLength))
		}
		if len(push.Message) > config.MaxBodyLength {
			return errors.NewValidationError("message", fmt.Sprintf("message too long (max %d characters)", config.MaxBodyLength))
		}
	}

	return nil
}

// validateIOSToken validates an iOS device token
func (p *MockPushProvider) validateIOSToken(token string) error {
	// iOS device tokens are typically 64 hexadecimal characters
	if len(token) != 64 {
		return errors.NewValidationError("device_token", "iOS device token must be 64 characters")
	}

	tokenRegex := regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	if !tokenRegex.MatchString(token) {
		return errors.NewValidationError("device_token", "iOS device token must contain only hexadecimal characters")
	}

	return nil
}

// validateAndroidToken validates an Android FCM token
func (p *MockPushProvider) validateAndroidToken(token string) error {
	// Android FCM tokens are base64-encoded strings, typically 140-255 characters
	if len(token) < 140 || len(token) > 255 {
		return errors.NewValidationError("device_token", "Android FCM token must be 140-255 characters")
	}

	// Basic check for base64-like characters
	tokenRegex := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !tokenRegex.MatchString(token) {
		return errors.NewValidationError("device_token", "Android FCM token contains invalid characters")
	}

	return nil
}

// validateWebToken validates a web push token
func (p *MockPushProvider) validateWebToken(token string) error {
	// Web push tokens can vary in format
	if len(token) < 50 || len(token) > 500 {
		return errors.NewValidationError("device_token", "Web push token must be 50-500 characters")
	}

	return nil
}

// formatForPlatform formats a push notification for the specific platform
func (p *MockPushProvider) formatForPlatform(push *models.PushNotification) (*models.PushNotification, error) {
	config := p.platformConfigs[strings.ToLower(push.Platform)]
	if config == nil {
		return push, nil // No platform-specific formatting
	}

	formatted := *push

	// Truncate title and message if too long
	if len(formatted.Title) > config.MaxTitleLength {
		formatted.Title = formatted.Title[:config.MaxTitleLength-3] + "..."
	}
	if len(formatted.Message) > config.MaxBodyLength {
		formatted.Message = formatted.Message[:config.MaxBodyLength-3] + "..."
	}

	// Set default sound if not specified
	if formatted.Sound == "" && config.SupportsSound {
		formatted.Sound = config.DefaultSound
	}

	// Platform-specific adjustments
	switch strings.ToLower(push.Platform) {
	case "ios":
		// iOS specific formatting
		if !config.SupportsBadge {
			formatted.Badge = 0
		}

	case "android":
		// Android specific formatting
		if formatted.Icon == "" {
			formatted.Icon = "ic_notification"
		}

	case "web":
		// Web specific formatting
		if formatted.Icon == "" {
			formatted.Icon = "/icon-192x192.png"
		}
	}

	return &formatted, nil
}

// getPlatformDelay returns processing delay for different platforms
func (p *MockPushProvider) getPlatformDelay(platform string) time.Duration {
	switch strings.ToLower(platform) {
	case "ios":
		return 200 * time.Millisecond // APNs is typically faster
	case "android":
		return 250 * time.Millisecond // FCM processing time
	case "web":
		return 300 * time.Millisecond // Web push can be slower
	default:
		return 200 * time.Millisecond
	}
}

// updateDeviceLastSeen updates the last seen time for a device
func (p *MockPushProvider) updateDeviceLastSeen(token string) {
	if device, exists := p.deviceTokens[token]; exists {
		device.LastSeen = time.Now()
		device.Active = true
	}
}

// replaceVariables replaces template variables with provided data
func (p *MockPushProvider) replaceVariables(template string, data map[string]string) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// loadDefaultTemplates loads default push notification templates
func (p *MockPushProvider) loadDefaultTemplates() {
	// Welcome notification template
	welcomeTemplate := &PushTemplate{
		ID:        "welcome_push",
		Name:      "Welcome Push Notification",
		Platform:  "all",
		Title:     "Welcome to {{app_name}}!",
		Body:      "Hi {{user_name}}, thanks for installing {{app_name}}. Tap to get started!",
		Icon:      "ic_welcome",
		Sound:     "default",
		Variables: []string{"app_name", "user_name"},
		Category:  "onboarding",
	}

	// News alert template
	newsTemplate := &PushTemplate{
		ID:        "news_alert",
		Name:      "News Alert",
		Platform:  "all",
		Title:     "Breaking: {{headline}}",
		Body:      "{{summary}} Tap to read more.",
		Icon:      "ic_news",
		Sound:     "news_alert",
		Variables: []string{"headline", "summary"},
		Category:  "news",
		Actions: []PushAction{
			{ID: "read", Title: "Read Now", Icon: "ic_read"},
			{ID: "save", Title: "Save", Icon: "ic_save"},
		},
	}

	// Promotional template
	promoTemplate := &PushTemplate{
		ID:        "promotion",
		Name:      "Promotional Notification",
		Platform:  "all",
		Title:     "🎉 Special Offer!",
		Body:      "{{offer_text}} Use code {{promo_code}}. Valid until {{expiry_date}}.",
		Icon:      "ic_promotion",
		Sound:     "promotion",
		Variables: []string{"offer_text", "promo_code", "expiry_date"},
		Category:  "marketing",
		Actions: []PushAction{
			{ID: "shop", Title: "Shop Now", Icon: "ic_shop"},
			{ID: "dismiss", Title: "Dismiss", Icon: "ic_close"},
		},
	}

	// Reminder template
	reminderTemplate := &PushTemplate{
		ID:        "reminder",
		Name:      "Reminder Notification",
		Platform:  "all",
		Title:     "Reminder: {{event_title}}",
		Body:      "{{event_description}} Scheduled for {{event_time}}.",
		Icon:      "ic_reminder",
		Sound:     "gentle",
		Variables: []string{"event_title", "event_description", "event_time"},
		Category:  "productivity",
		Actions: []PushAction{
			{ID: "view", Title: "View", Icon: "ic_view"},
			{ID: "snooze", Title: "Snooze", Icon: "ic_snooze"},
		},
	}

	p.AddTemplate(welcomeTemplate)
	p.AddTemplate(newsTemplate)
	p.AddTemplate(promoTemplate)
	p.AddTemplate(reminderTemplate)
}

// loadPlatformConfigs loads platform-specific configurations
func (p *MockPushProvider) loadPlatformConfigs() {
	// iOS (APNs) configuration
	iosConfig := &PlatformConfig{
		Platform:        "ios",
		MaxPayloadSize:  4096,
		MaxTitleLength:  50,
		MaxBodyLength:   200,
		SupportsBadge:   true,
		SupportsSound:   true,
		SupportsImage:   true,
		SupportsActions: true,
		DefaultSound:    "default",
		Settings: map[string]string{
			"service":     "apns",
			"environment": "development",
		},
	}

	// Android (FCM) configuration
	androidConfig := &PlatformConfig{
		Platform:        "android",
		MaxPayloadSize:  4000,
		MaxTitleLength:  65,
		MaxBodyLength:   240,
		SupportsBadge:   false,
		SupportsSound:   true,
		SupportsImage:   true,
		SupportsActions: true,
		DefaultSound:    "default",
		Settings: map[string]string{
			"service":  "fcm",
			"priority": "high",
		},
	}

	// Web Push configuration
	webConfig := &PlatformConfig{
		Platform:        "web",
		MaxPayloadSize:  3072,
		MaxTitleLength:  50,
		MaxBodyLength:   120,
		SupportsBadge:   true,
		SupportsSound:   false,
		SupportsImage:   true,
		SupportsActions: true,
		DefaultSound:    "",
		Settings: map[string]string{
			"service": "web-push",
			"ttl":     "2419200", // 4 weeks
		},
	}

	p.platformConfigs["ios"] = iosConfig
	p.platformConfigs["android"] = androidConfig
	p.platformConfigs["web"] = webConfig
}
