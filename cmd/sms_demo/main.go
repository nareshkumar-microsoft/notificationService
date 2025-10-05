package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/providers"
	"github.com/nareshkumar-microsoft/notificationService/internal/services"
	"github.com/nareshkumar-microsoft/notificationService/internal/utils"
)

func main() {
	fmt.Println("📱 SMS Notification Provider Demo")
	fmt.Println("=================================")

	// Create logger
	logger := utils.NewSimpleLogger("info")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Demo 1: Direct provider usage
	fmt.Println("\n📱 Demo 1: Direct SMS Provider Usage")
	fmt.Println("------------------------------------")
	demoDirectProvider(cfg.Providers.SMS, logger)

	// Demo 2: SMS service usage
	fmt.Println("\n📲 Demo 2: SMS Service Usage")
	fmt.Println("----------------------------")
	demoSMSService(cfg.Providers.SMS, logger)

	// Demo 3: Template rendering
	fmt.Println("\n📄 Demo 3: SMS Template Rendering")
	fmt.Println("---------------------------------")
	demoTemplateRendering(cfg.Providers.SMS, logger)

	// Demo 4: Bulk SMS
	fmt.Println("\n📮 Demo 4: Bulk SMS Sending")
	fmt.Println("---------------------------")
	demoBulkSMS(cfg.Providers.SMS, logger)

	// Demo 5: Cost calculation
	fmt.Println("\n💰 Demo 5: Cost Calculation")
	fmt.Println("--------------------------")
	demoCostCalculation(cfg.Providers.SMS, logger)

	// Demo 6: International SMS
	fmt.Println("\n🌍 Demo 6: International SMS")
	fmt.Println("---------------------------")
	demoInternationalSMS(cfg.Providers.SMS, logger)

	// Demo 7: Error handling
	fmt.Println("\n⚠️  Demo 7: Error Handling")
	fmt.Println("-------------------------")
	demoErrorHandling(cfg.Providers.SMS, logger)

	fmt.Println("\n✅ All SMS demos completed successfully!")
}

func demoDirectProvider(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	// Create mock SMS provider
	provider := providers.NewMockSMSProvider(cfg)

	ctx := context.Background()

	// Check provider health
	if err := provider.IsHealthy(ctx); err != nil {
		logger.Errorf("Provider health check failed: %v", err)
		return
	}

	// Create sample SMS notification
	smsNotification := &models.SMSNotification{
		Notification: models.Notification{
			ID:        utils.GenerateNotificationID(),
			Type:      models.NotificationTypeSMS,
			Status:    models.StatusPending,
			Priority:  models.PriorityNormal,
			Recipient: "1234567890",
			Subject:   "Direct Provider Demo",
			Body:      "This SMS was sent directly through the provider.",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     "Hello! This is a direct SMS from the notification service provider.",
		Unicode:     false,
	}

	// Send SMS
	response, err := provider.SendSMS(ctx, smsNotification)
	if err != nil {
		logger.Errorf("Failed to send SMS: %v", err)
		return
	}

	logger.Infof("✅ SMS sent successfully!")
	logger.Infof("   ID: %s", response.ID)
	logger.Infof("   Status: %s", response.Status)
	logger.Infof("   Provider ID: %s", response.ProviderID)
	logger.Infof("   Message: %s", response.Message)

	// Show sent SMS
	sentSMS := provider.GetSentSMS()
	logger.Infof("📊 Total SMS sent: %d", len(sentSMS))
	if len(sentSMS) > 0 {
		sms := sentSMS[len(sentSMS)-1]
		logger.Infof("   Segments: %d", sms.Segments)
		logger.Infof("   Cost: $%.4f", sms.Cost)
		logger.Infof("   Unicode: %t", sms.Unicode)
	}
}

func demoSMSService(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	// Create SMS service
	service, err := services.NewSMSService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create SMS service: %v", err)
		return
	}

	ctx := context.Background()

	// Simple SMS request
	request := &services.SMSRequest{
		PhoneNumber: "1234567890",
		CountryCode: "US",
		Message:     "Welcome to SMS Service! This message was sent through the service layer.",
		Unicode:     false,
		Priority:    models.PriorityNormal,
		Metadata: map[string]string{
			"demo_type": "service_layer",
			"version":   "1.0",
		},
	}

	response, err := service.SendSMS(ctx, request)
	if err != nil {
		logger.Errorf("Failed to send SMS: %v", err)
		return
	}

	logger.Infof("✅ SMS sent through service!")
	logger.Infof("   ID: %s", response.ID)
	logger.Infof("   Status: %s", response.Status)

	// Check provider status
	status := service.GetProviderStatus(ctx)
	logger.Infof("📊 Provider Status:")
	logger.Infof("   Name: %s", status.Name)
	logger.Infof("   Type: %s", status.Type)
	logger.Infof("   Healthy: %t", status.Healthy)
}

func demoTemplateRendering(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewSMSService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create SMS service: %v", err)
		return
	}

	ctx := context.Background()

	// Show available templates
	logger.Infof("📄 Available SMS templates:")
	templateIDs := []string{"verification", "welcome_sms", "alert", "reminder"}
	for _, templateID := range templateIDs {
		logger.Infof("   - %s", templateID)
	}

	// Render verification template
	templateData := map[string]string{
		"service_name":   "DemoApp",
		"code":           "987654",
		"expiry_minutes": "15",
	}

	rendered, err := service.RenderTemplate("verification", templateData)
	if err != nil {
		logger.Errorf("Failed to render template: %v", err)
		return
	}

	logger.Infof("✅ Template rendered successfully!")
	logger.Infof("   Message: %s", rendered.Message)
	logger.Infof("   Segments: %d", rendered.Segments)
	logger.Infof("   Unicode: %t", rendered.Unicode)

	// Send SMS with template
	templateRequest := &services.SMSRequest{
		PhoneNumber:  "1234567890",
		CountryCode:  "US",
		TemplateID:   "verification",
		TemplateData: templateData,
		Priority:     models.PriorityHigh,
	}

	response, err := service.SendSMS(ctx, templateRequest)
	if err != nil {
		logger.Errorf("Failed to send templated SMS: %v", err)
		return
	}

	logger.Infof("✅ Templated SMS sent!")
	logger.Infof("   ID: %s", response.ID)
}

func demoBulkSMS(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewSMSService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create SMS service: %v", err)
		return
	}

	ctx := context.Background()

	// Bulk SMS request
	bulkRequest := &services.BulkSMSRequest{
		Recipients: []services.BulkSMSRecipient{
			{
				PhoneNumber: "1234567890",
				CountryCode: "US",
				Data: map[string]string{
					"user_name": "Alice",
					"user_id":   "001",
				},
			},
			{
				PhoneNumber: "1234567891",
				CountryCode: "US",
				Data: map[string]string{
					"user_name": "Bob",
					"user_id":   "002",
				},
			},
			{
				PhoneNumber: "07123456789",
				CountryCode: "UK",
				Data: map[string]string{
					"user_name": "Charlie",
					"user_id":   "003",
				},
			},
		},
		TemplateID: "welcome_sms",
		TemplateData: map[string]string{
			"service_name": "BulkDemo App",
		},
		Priority: models.PriorityNormal,
		Metadata: map[string]string{
			"campaign_type": "bulk_demo",
			"batch_id":      "sms-demo-001",
		},
	}

	responses, err := service.SendBulkSMS(ctx, bulkRequest)
	if err != nil {
		logger.Errorf("Failed to send bulk SMS: %v", err)
		return
	}

	logger.Infof("✅ Bulk SMS completed!")
	logger.Infof("   Total SMS: %d", len(responses))

	successCount := 0
	for i, response := range responses {
		if response.Status == models.StatusSent {
			successCount++
		}
		logger.Infof("   SMS %d: %s", i+1, response.Status)
	}

	logger.Infof("📊 Success rate: %d/%d (%.1f%%)",
		successCount, len(responses),
		float64(successCount)/float64(len(responses))*100)
}

func demoCostCalculation(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewSMSService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create SMS service: %v", err)
		return
	}

	// Show supported countries and costs
	logger.Infof("💰 SMS Costs by Country:")
	countries := service.GetSupportedCountries()
	for _, country := range countries {
		logger.Infof("   %s (%s): $%.4f per segment", country.Name, country.Code, country.Cost)
	}

	// Test cost estimation for different scenarios
	testMessages := []struct {
		name    string
		message string
		country string
		unicode bool
	}{
		{
			name:    "Short US SMS",
			message: "Hello world!",
			country: "US",
			unicode: false,
		},
		{
			name:    "Long UK SMS",
			message: strings.Repeat("This is a test message. ", 10),
			country: "UK",
			unicode: false,
		},
		{
			name:    "Unicode emoji SMS",
			message: "Hello 🌍! Welcome to our app 🎉🚀",
			country: "DE",
			unicode: true,
		},
	}

	logger.Infof("\n📊 Cost Estimates:")
	for _, test := range testMessages {
		estimate, err := service.EstimateCost(test.message, test.country, test.unicode)
		if err != nil {
			logger.Errorf("Failed to estimate cost for %s: %v", test.name, err)
			continue
		}

		logger.Infof("   %s:", test.name)
		logger.Infof("     Length: %d chars", estimate.MessageLength)
		logger.Infof("     Segments: %d", estimate.Segments)
		logger.Infof("     Cost per segment: $%.4f", estimate.CostPerSegment)
		logger.Infof("     Total cost: $%.4f", estimate.TotalCost)
		logger.Infof("     Unicode: %t", estimate.Unicode)
	}
}

func demoInternationalSMS(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewSMSService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create SMS service: %v", err)
		return
	}

	ctx := context.Background()

	// Test international SMS to different countries
	internationalTests := []struct {
		name        string
		phoneNumber string
		countryCode string
		message     string
	}{
		{
			name:        "United States",
			phoneNumber: "1234567890",
			countryCode: "US",
			message:     "Hello from the US! 🇺🇸",
		},
		{
			name:        "United Kingdom",
			phoneNumber: "07123456789",
			countryCode: "UK",
			message:     "Greetings from the UK! 🇬🇧",
		},
		{
			name:        "Germany",
			phoneNumber: "1234567890",
			countryCode: "DE",
			message:     "Guten Tag aus Deutschland! 🇩🇪",
		},
		{
			name:        "India",
			phoneNumber: "9876543210",
			countryCode: "IN",
			message:     "नमस्ते भारत से! 🇮🇳",
		},
	}

	logger.Infof("🌍 Sending international SMS:")
	for _, test := range internationalTests {
		request := &services.SMSRequest{
			PhoneNumber: test.phoneNumber,
			CountryCode: test.countryCode,
			Message:     test.message,
			Unicode:     true, // Enable unicode for international characters
			Priority:    models.PriorityNormal,
			Metadata: map[string]string{
				"demo_type": "international",
				"country":   test.name,
			},
		}

		response, err := service.SendSMS(ctx, request)
		if err != nil {
			logger.Errorf("   ❌ %s: %v", test.name, err)
			continue
		}

		logger.Infof("   ✅ %s: %s", test.name, response.Status)

		// Get cost for this country
		cost, _ := service.GetSMSCost(test.countryCode)
		logger.Infof("      Cost: $%.4f per segment", cost)
	}
}

func demoErrorHandling(cfg config.SMSProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewSMSService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create SMS service: %v", err)
		return
	}

	ctx := context.Background()

	// Test various error scenarios
	errorTests := []struct {
		name    string
		request *services.SMSRequest
	}{
		{
			name: "Empty phone number",
			request: &services.SMSRequest{
				PhoneNumber: "",
				Message:     "Test",
			},
		},
		{
			name: "Invalid phone number",
			request: &services.SMSRequest{
				PhoneNumber: "invalid-phone",
				Message:     "Test",
			},
		},
		{
			name: "Missing message",
			request: &services.SMSRequest{
				PhoneNumber: "1234567890",
			},
		},
		{
			name: "Unsupported country",
			request: &services.SMSRequest{
				PhoneNumber: "1234567890",
				CountryCode: "XX",
				Message:     "Test",
			},
		},
		{
			name: "Non-existent template",
			request: &services.SMSRequest{
				PhoneNumber: "1234567890",
				TemplateID:  "non-existent",
			},
		},
	}

	for _, test := range errorTests {
		logger.Infof("🧪 Testing: %s", test.name)

		_, err := service.SendSMS(ctx, test.request)
		if err != nil {
			logger.Infof("   ❌ Expected error: %v", err)
		} else {
			logger.Infof("   ⚠️  Unexpected success")
		}
	}

	// Test phone number validation
	logger.Infof("🧪 Testing phone number validation:")
	validationTests := []struct {
		phone   string
		country string
	}{
		{"1234567890", "US"},
		{"invalid-phone", ""},
		{"", ""},
		{"07123456789", "UK"},
		{"123", ""},
		{"9876543210", "IN"},
	}

	for _, test := range validationTests {
		err := service.ValidatePhoneNumber(test.phone, test.country)
		if err != nil {
			logger.Infof("   ❌ '%s' (%s): %v", test.phone, test.country, err)
		} else {
			logger.Infof("   ✅ '%s' (%s): valid", test.phone, test.country)
		}
	}
}
