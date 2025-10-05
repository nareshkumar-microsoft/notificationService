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

// MockSMSProvider implements the SMSProvider interface for testing and development
type MockSMSProvider struct {
	config    config.SMSProviderConfig
	templates map[string]*SMSTemplate
	sentSMS   []SentSMS
	healthy   bool
	costs     map[string]float64 // Country code to cost mapping
}

// SMSTemplate represents an SMS template
type SMSTemplate struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Message   string            `json:"message"`
	Variables []string          `json:"variables"`
	Category  string            `json:"category"`
	MaxLength int               `json:"max_length"`
	Unicode   bool              `json:"unicode"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// SentSMS represents an SMS that was sent (for mock tracking)
type SentSMS struct {
	ID           uuid.UUID         `json:"id"`
	PhoneNumber  string            `json:"phone_number"`
	CountryCode  string            `json:"country_code,omitempty"`
	Message      string            `json:"message"`
	Unicode      bool              `json:"unicode"`
	SentAt       time.Time         `json:"sent_at"`
	Status       string            `json:"status"`
	DeliveredAt  *time.Time        `json:"delivered_at,omitempty"`
	Cost         float64           `json:"cost"`
	Segments     int               `json:"segments"`
	ProviderData map[string]string `json:"provider_data,omitempty"`
}

// CountryInfo represents information about SMS costs for a country
type CountryInfo struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Cost      float64 `json:"cost"`
	MaxLength int     `json:"max_length"`
	Supported bool    `json:"supported"`
}

// NewMockSMSProvider creates a new mock SMS provider
func NewMockSMSProvider(cfg config.SMSProviderConfig) *MockSMSProvider {
	provider := &MockSMSProvider{
		config:    cfg,
		templates: make(map[string]*SMSTemplate),
		sentSMS:   make([]SentSMS, 0),
		healthy:   true,
		costs:     make(map[string]float64),
	}

	// Load default templates and costs
	provider.loadDefaultTemplates()
	provider.loadDefaultCosts()

	return provider
}

// Send implements the NotificationProvider interface
func (p *MockSMSProvider) Send(ctx context.Context, notification *models.Notification) (*models.NotificationResponse, error) {
	if !p.healthy {
		return nil, errors.NewProviderError("mock-sms", errors.ErrorCodeProviderUnavailable, "provider is unhealthy")
	}

	// Convert generic notification to SMS notification
	smsNotification, err := p.convertToSMSNotification(notification)
	if err != nil {
		return nil, err
	}

	return p.SendSMS(ctx, smsNotification)
}

// SendSMS implements the SMSProvider interface
func (p *MockSMSProvider) SendSMS(ctx context.Context, sms *models.SMSNotification) (*models.NotificationResponse, error) {
	if !p.healthy {
		return nil, errors.NewProviderError("mock-sms", errors.ErrorCodeProviderUnavailable, "provider is unhealthy")
	}

	// Validate SMS
	if err := p.validateSMSNotification(sms); err != nil {
		return nil, err
	}

	// Simulate processing delay
	select {
	case <-ctx.Done():
		return nil, errors.NewNotificationError(errors.ErrorCodeTimeout, "SMS sending timed out")
	case <-time.After(150 * time.Millisecond):
		// Continue processing
	}

	// Calculate segments and cost
	segments := p.calculateSegments(sms.Message, sms.Unicode)
	cost := p.calculateCost(sms.CountryCode, segments)

	// Create sent SMS record
	sentSMS := SentSMS{
		ID:          sms.ID,
		PhoneNumber: sms.PhoneNumber,
		CountryCode: sms.CountryCode,
		Message:     sms.Message,
		Unicode:     sms.Unicode,
		SentAt:      time.Now(),
		Status:      "sent",
		Cost:        cost,
		Segments:    segments,
		ProviderData: map[string]string{
			"provider":     "mock-sms",
			"message_id":   fmt.Sprintf("sms-%s", sms.ID.String()),
			"queue_time":   "150ms",
			"retry_count":  "0",
			"country_code": sms.CountryCode,
		},
	}

	// Simulate delivery (90% success rate)
	if time.Now().UnixNano()%10 < 9 {
		deliveredAt := time.Now().Add(time.Duration(100+time.Now().UnixNano()%500) * time.Millisecond)
		sentSMS.DeliveredAt = &deliveredAt
		sentSMS.Status = "delivered"
		sentSMS.ProviderData["delivery_time"] = deliveredAt.Format(time.RFC3339)
	}

	// Store sent SMS for tracking
	p.sentSMS = append(p.sentSMS, sentSMS)

	// Create response
	now := time.Now()
	response := &models.NotificationResponse{
		ID:         sms.ID,
		Status:     models.StatusSent,
		Message:    fmt.Sprintf("SMS sent to %s (%d segments, $%.4f)", sms.PhoneNumber, segments, cost),
		ProviderID: sentSMS.ProviderData["message_id"],
		SentAt:     &now,
	}

	return response, nil
}

// ValidatePhoneNumber implements the SMSProvider interface
func (p *MockSMSProvider) ValidatePhoneNumber(phoneNumber, countryCode string) error {
	if phoneNumber == "" {
		return errors.NewValidationError("phone_number", "phone number is required")
	}

	// Clean phone number
	cleanNumber := p.cleanPhoneNumber(phoneNumber)

	// Basic validation - should contain only digits
	phoneRegex := regexp.MustCompile(`^\d{7,15}$`)
	if !phoneRegex.MatchString(cleanNumber) {
		return errors.NewValidationError("phone_number", "phone number must contain 7-15 digits")
	}

	// Country-specific validation
	if countryCode != "" {
		if err := p.validateCountryCode(countryCode); err != nil {
			return err
		}

		// Validate number format for specific countries
		if err := p.validatePhoneForCountry(cleanNumber, countryCode); err != nil {
			return err
		}
	}

	return nil
}

// GetSMSCost implements the SMSProvider interface
func (p *MockSMSProvider) GetSMSCost(countryCode string) (float64, error) {
	if countryCode == "" {
		return 0.01, nil // Default cost
	}

	countryCode = strings.ToUpper(countryCode)
	if cost, exists := p.costs[countryCode]; exists {
		return cost, nil
	}

	return 0.0, errors.NewNotificationError(errors.ErrorCodeNotFound, fmt.Sprintf("country code not supported: %s", countryCode))
}

// GetType implements the NotificationProvider interface
func (p *MockSMSProvider) GetType() models.NotificationType {
	return models.NotificationTypeSMS
}

// IsHealthy implements the NotificationProvider interface
func (p *MockSMSProvider) IsHealthy(ctx context.Context) error {
	if !p.healthy {
		return errors.NewProviderError("mock-sms", errors.ErrorCodeProviderUnavailable, "provider is marked as unhealthy")
	}

	// Simulate health check delay
	select {
	case <-ctx.Done():
		return errors.NewNotificationError(errors.ErrorCodeTimeout, "health check timed out")
	case <-time.After(75 * time.Millisecond):
		return nil
	}
}

// GetConfig implements the NotificationProvider interface
func (p *MockSMSProvider) GetConfig() interfaces.ProviderConfig {
	return interfaces.ProviderConfig{
		Name:       "Mock SMS Provider",
		Type:       models.NotificationTypeSMS,
		Enabled:    p.config.Enabled,
		Priority:   2,
		MaxRetries: 3,
		Timeout:    30,
		RateLimit: interfaces.RateLimitConfig{
			Enabled:        true,
			RequestsPerMin: 60,
			BurstSize:      5,
		},
		Settings: map[string]string{
			"provider_type":       "mock",
			"version":             "1.0.0",
			"features":            "templates,validation,cost_calculation,delivery_tracking",
			"supported_countries": "US,UK,CA,AU,DE,FR,IN,BR",
		},
	}
}

// GetTemplate retrieves an SMS template by ID
func (p *MockSMSProvider) GetTemplate(templateID string) (*SMSTemplate, error) {
	template, exists := p.templates[templateID]
	if !exists {
		return nil, errors.NewNotificationError(errors.ErrorCodeTemplateNotFound, fmt.Sprintf("template not found: %s", templateID))
	}
	return template, nil
}

// AddTemplate adds a new SMS template
func (p *MockSMSProvider) AddTemplate(template *SMSTemplate) error {
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Set default max length if not specified
	if template.MaxLength == 0 {
		if template.Unicode {
			template.MaxLength = 70
		} else {
			template.MaxLength = 160
		}
	}

	p.templates[template.ID] = template
	return nil
}

// RenderTemplate renders an SMS template with provided data
func (p *MockSMSProvider) RenderTemplate(templateID string, data map[string]string) (*SMSTemplate, error) {
	template, err := p.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Clone template for rendering
	rendered := &SMSTemplate{
		ID:        template.ID,
		Name:      template.Name,
		Message:   p.replaceVariables(template.Message, data),
		Variables: template.Variables,
		Category:  template.Category,
		MaxLength: template.MaxLength,
		Unicode:   template.Unicode,
		CreatedAt: template.CreatedAt,
		UpdatedAt: template.UpdatedAt,
		Metadata:  template.Metadata,
	}

	return rendered, nil
}

// GetSentSMS returns all sent SMS messages (for testing)
func (p *MockSMSProvider) GetSentSMS() []SentSMS {
	return p.sentSMS
}

// ClearSentSMS clears the sent SMS history (for testing)
func (p *MockSMSProvider) ClearSentSMS() {
	p.sentSMS = make([]SentSMS, 0)
}

// SetHealthy sets the provider health status (for testing)
func (p *MockSMSProvider) SetHealthy(healthy bool) {
	p.healthy = healthy
}

// GetSupportedCountries returns list of supported countries
func (p *MockSMSProvider) GetSupportedCountries() []CountryInfo {
	countries := []CountryInfo{
		{Code: "US", Name: "United States", Cost: 0.0075, MaxLength: 160, Supported: true},
		{Code: "UK", Name: "United Kingdom", Cost: 0.0080, MaxLength: 160, Supported: true},
		{Code: "CA", Name: "Canada", Cost: 0.0070, MaxLength: 160, Supported: true},
		{Code: "AU", Name: "Australia", Cost: 0.0085, MaxLength: 160, Supported: true},
		{Code: "DE", Name: "Germany", Cost: 0.0090, MaxLength: 160, Supported: true},
		{Code: "FR", Name: "France", Cost: 0.0088, MaxLength: 160, Supported: true},
		{Code: "IN", Name: "India", Cost: 0.0050, MaxLength: 160, Supported: true},
		{Code: "BR", Name: "Brazil", Cost: 0.0095, MaxLength: 160, Supported: true},
	}
	return countries
}

// Helper methods

// convertToSMSNotification converts a generic notification to an SMS notification
func (p *MockSMSProvider) convertToSMSNotification(notification *models.Notification) (*models.SMSNotification, error) {
	if notification.Type != models.NotificationTypeSMS {
		return nil, errors.NewValidationError("type", "notification type must be SMS")
	}

	smsNotification := &models.SMSNotification{
		Notification: *notification,
		PhoneNumber:  notification.Recipient,
		Message:      notification.Body,
		Unicode:      p.containsUnicode(notification.Body),
	}

	// Extract country code from metadata if available
	if notification.Metadata != nil {
		if countryCode, exists := notification.Metadata["country_code"]; exists {
			smsNotification.CountryCode = countryCode
		}
	}

	return smsNotification, nil
}

// validateSMSNotification validates an SMS notification
func (p *MockSMSProvider) validateSMSNotification(sms *models.SMSNotification) error {
	// Validate phone number
	if err := p.ValidatePhoneNumber(sms.PhoneNumber, sms.CountryCode); err != nil {
		return err
	}

	// Validate message content
	if sms.Message == "" {
		return errors.NewValidationError("message", "SMS message is required")
	}

	// Check message length
	maxLength := 160
	if sms.Unicode {
		maxLength = 70
	}

	if len(sms.Message) > maxLength*10 { // Allow up to 10 segments
		return errors.NewValidationError("message", fmt.Sprintf("message too long (max %d characters for 10 segments)", maxLength*10))
	}

	return nil
}

// cleanPhoneNumber removes formatting characters from phone number
func (p *MockSMSProvider) cleanPhoneNumber(phoneNumber string) string {
	// Remove common formatting characters
	cleanNumber := strings.ReplaceAll(phoneNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "+", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ".", "")

	return cleanNumber
}

// validateCountryCode validates a country code
func (p *MockSMSProvider) validateCountryCode(countryCode string) error {
	countryCode = strings.ToUpper(countryCode)
	supportedCountries := []string{"US", "UK", "CA", "AU", "DE", "FR", "IN", "BR"}

	for _, supported := range supportedCountries {
		if countryCode == supported {
			return nil
		}
	}

	return errors.NewValidationError("country_code", fmt.Sprintf("country code not supported: %s", countryCode))
}

// validatePhoneForCountry validates phone number format for specific countries
func (p *MockSMSProvider) validatePhoneForCountry(cleanNumber, countryCode string) error {
	countryCode = strings.ToUpper(countryCode)

	switch countryCode {
	case "US", "CA":
		// North American numbers: 10 digits
		if len(cleanNumber) != 10 && len(cleanNumber) != 11 {
			return errors.NewValidationError("phone_number", "US/CA numbers must be 10 or 11 digits")
		}
	case "UK":
		// UK numbers: 10-11 digits
		if len(cleanNumber) < 10 || len(cleanNumber) > 11 {
			return errors.NewValidationError("phone_number", "UK numbers must be 10-11 digits")
		}
	case "AU":
		// Australian numbers: 9-10 digits
		if len(cleanNumber) < 9 || len(cleanNumber) > 10 {
			return errors.NewValidationError("phone_number", "Australian numbers must be 9-10 digits")
		}
	case "DE":
		// German numbers: 10-12 digits
		if len(cleanNumber) < 10 || len(cleanNumber) > 12 {
			return errors.NewValidationError("phone_number", "German numbers must be 10-12 digits")
		}
	case "IN":
		// Indian numbers: 10 digits
		if len(cleanNumber) != 10 {
			return errors.NewValidationError("phone_number", "Indian numbers must be 10 digits")
		}
	}

	return nil
}

// containsUnicode checks if a string contains unicode characters
func (p *MockSMSProvider) containsUnicode(text string) bool {
	for _, r := range text {
		if r > 127 {
			return true
		}
	}
	return false
}

// calculateSegments calculates the number of SMS segments needed
func (p *MockSMSProvider) calculateSegments(message string, unicode bool) int {
	maxLength := 160
	if unicode {
		maxLength = 70
	}

	length := len(message)
	if length <= maxLength {
		return 1
	}

	// For multi-part messages, each segment is slightly shorter
	segmentLength := maxLength - 7 // Account for UDH (User Data Header)
	if unicode {
		segmentLength = 67
	}

	return (length + segmentLength - 1) / segmentLength
}

// calculateCost calculates the cost of sending an SMS
func (p *MockSMSProvider) calculateCost(countryCode string, segments int) float64 {
	baseCost := 0.01 // Default cost per segment

	if countryCode != "" {
		if cost, exists := p.costs[strings.ToUpper(countryCode)]; exists {
			baseCost = cost
		}
	}

	return baseCost * float64(segments)
}

// replaceVariables replaces template variables with provided data
func (p *MockSMSProvider) replaceVariables(template string, data map[string]string) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// loadDefaultTemplates loads default SMS templates
func (p *MockSMSProvider) loadDefaultTemplates() {
	// Verification code template
	verificationTemplate := &SMSTemplate{
		ID:        "verification",
		Name:      "Verification Code",
		Message:   "Your {{service_name}} verification code is: {{code}}. Valid for {{expiry_minutes}} minutes.",
		Variables: []string{"service_name", "code", "expiry_minutes"},
		Category:  "security",
		MaxLength: 160,
		Unicode:   false,
	}

	// Welcome SMS template
	welcomeTemplate := &SMSTemplate{
		ID:        "welcome_sms",
		Name:      "Welcome SMS",
		Message:   "Welcome to {{service_name}}, {{user_name}}! Thanks for joining us.",
		Variables: []string{"service_name", "user_name"},
		Category:  "onboarding",
		MaxLength: 160,
		Unicode:   false,
	}

	// Alert template
	alertTemplate := &SMSTemplate{
		ID:        "alert",
		Name:      "Alert Notification",
		Message:   "ALERT: {{alert_message}} Time: {{timestamp}}",
		Variables: []string{"alert_message", "timestamp"},
		Category:  "alerts",
		MaxLength: 160,
		Unicode:   false,
	}

	// Reminder template
	reminderTemplate := &SMSTemplate{
		ID:        "reminder",
		Name:      "Reminder",
		Message:   "Reminder: {{reminder_text}}. Reply STOP to opt out.",
		Variables: []string{"reminder_text"},
		Category:  "general",
		MaxLength: 160,
		Unicode:   false,
	}

	p.AddTemplate(verificationTemplate)
	p.AddTemplate(welcomeTemplate)
	p.AddTemplate(alertTemplate)
	p.AddTemplate(reminderTemplate)
}

// loadDefaultCosts loads default SMS costs by country
func (p *MockSMSProvider) loadDefaultCosts() {
	p.costs = map[string]float64{
		"US": 0.0075, // United States
		"UK": 0.0080, // United Kingdom
		"CA": 0.0070, // Canada
		"AU": 0.0085, // Australia
		"DE": 0.0090, // Germany
		"FR": 0.0088, // France
		"IN": 0.0050, // India
		"BR": 0.0095, // Brazil
		"MX": 0.0080, // Mexico
		"JP": 0.0120, // Japan
		"KR": 0.0110, // South Korea
		"SG": 0.0100, // Singapore
		"HK": 0.0095, // Hong Kong
		"TH": 0.0085, // Thailand
		"MY": 0.0090, // Malaysia
		"PH": 0.0085, // Philippines
	}
}
