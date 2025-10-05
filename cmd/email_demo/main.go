package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/internal/providers"
	"github.com/nareshkumar-microsoft/notificationService/internal/services"
	"github.com/nareshkumar-microsoft/notificationService/internal/utils"
)

func main() {
	fmt.Println("🔔 Email Notification Provider Demo")
	fmt.Println("=====================================")
	
	// Create logger
	logger := utils.NewSimpleLogger("info")
	
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Demo 1: Direct provider usage
	fmt.Println("\n📧 Demo 1: Direct Email Provider Usage")
	fmt.Println("--------------------------------------")
	demoDirectProvider(cfg.Providers.Email, logger)
	
	// Demo 2: Email service usage
	fmt.Println("\n📬 Demo 2: Email Service Usage")
	fmt.Println("------------------------------")
	demoEmailService(cfg.Providers.Email, logger)
	
	// Demo 3: Template rendering
	fmt.Println("\n📄 Demo 3: Template Rendering")
	fmt.Println("-----------------------------")
	demoTemplateRendering(cfg.Providers.Email, logger)
	
	// Demo 4: Bulk email
	fmt.Println("\n📮 Demo 4: Bulk Email Sending")
	fmt.Println("-----------------------------")
	demoBulkEmail(cfg.Providers.Email, logger)
	
	// Demo 5: Error handling
	fmt.Println("\n⚠️  Demo 5: Error Handling")
	fmt.Println("-------------------------")
	demoErrorHandling(cfg.Providers.Email, logger)
	
	fmt.Println("\n✅ All email demos completed successfully!")
}

func demoDirectProvider(cfg config.EmailProviderConfig, logger *utils.SimpleLogger) {
	// Create mock email provider
	provider := providers.NewMockEmailProvider(cfg)
	
	ctx := context.Background()
	
	// Check provider health
	if err := provider.IsHealthy(ctx); err != nil {
		logger.Errorf("Provider health check failed: %v", err)
		return
	}
	
	// Create sample email notification
	emailNotification := &models.EmailNotification{
		Notification: models.Notification{
			ID:        utils.GenerateNotificationID(),
			Type:      models.NotificationTypeEmail,
			Status:    models.StatusPending,
			Priority:  models.PriorityNormal,
			Recipient: "demo@example.com",
			Subject:   "Direct Provider Demo",
			Body:      "This email was sent directly through the provider.",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		To:       []string{"demo@example.com", "demo2@example.com"},
		From:     "noreply@notification-service.com",
		HTMLBody: "<h1>Hello from Direct Provider!</h1><p>This is a demonstration of the email provider.</p>",
		TextBody: "Hello from Direct Provider!\n\nThis is a demonstration of the email provider.",
		Headers: map[string]string{
			"X-Demo": "Direct-Provider",
		},
	}
	
	// Send email
	response, err := provider.SendEmail(ctx, emailNotification)
	if err != nil {
		logger.Errorf("Failed to send email: %v", err)
		return
	}
	
	logger.Infof("✅ Email sent successfully!")
	logger.Infof("   ID: %s", response.ID)
	logger.Infof("   Status: %s", response.Status)
	logger.Infof("   Provider ID: %s", response.ProviderID)
	logger.Infof("   Message: %s", response.Message)
	
	// Show sent emails
	sentEmails := provider.GetSentEmails()
	logger.Infof("📊 Total emails sent: %d", len(sentEmails))
}

func demoEmailService(cfg config.EmailProviderConfig, logger *utils.SimpleLogger) {
	// Create email service
	service, err := services.NewEmailService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create email service: %v", err)
		return
	}
	
	ctx := context.Background()
	
	// Simple email request
	request := &services.EmailRequest{
		To:       []string{"user@example.com"},
		Subject:  "Email Service Demo",
		HTMLBody: "<h2>Welcome to Email Service!</h2><p>This email was sent through the email service layer.</p>",
		TextBody: "Welcome to Email Service!\n\nThis email was sent through the email service layer.",
		Priority: models.PriorityNormal,
		Metadata: map[string]string{
			"demo_type": "service_layer",
			"version":   "1.0",
		},
	}
	
	response, err := service.SendEmail(ctx, request)
	if err != nil {
		logger.Errorf("Failed to send email: %v", err)
		return
	}
	
	logger.Infof("✅ Email sent through service!")
	logger.Infof("   ID: %s", response.ID)
	logger.Infof("   Status: %s", response.Status)
	
	// Check provider status
	status := service.GetProviderStatus(ctx)
	logger.Infof("📊 Provider Status:")
	logger.Infof("   Name: %s", status.Name)
	logger.Infof("   Type: %s", status.Type)
	logger.Infof("   Healthy: %t", status.Healthy)
}

func demoTemplateRendering(cfg config.EmailProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewEmailService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create email service: %v", err)
		return
	}
	
	ctx := context.Background()
	
	// List available templates
	templates := service.GetEmailTemplates()
	logger.Infof("📄 Available templates: %d", len(templates))
	for _, template := range templates {
		logger.Infof("   - %s: %s (%s)", template.ID, template.Name, template.Category)
	}
	
	// Render welcome template
	templateData := map[string]string{
		"user_name":    "John Doe",
		"user_email":   "john.doe@example.com",
		"service_name": "Notification Service Demo",
	}
	
	rendered, err := service.RenderTemplate("welcome", templateData)
	if err != nil {
		logger.Errorf("Failed to render template: %v", err)
		return
	}
	
	logger.Infof("✅ Template rendered successfully!")
	logger.Infof("   Subject: %s", rendered.Subject)
	
	// Send email with template
	templateRequest := &services.EmailRequest{
		To:           []string{"john.doe@example.com"},
		TemplateID:   "welcome",
		TemplateData: templateData,
		Priority:     models.PriorityNormal,
	}
	
	response, err := service.SendEmail(ctx, templateRequest)
	if err != nil {
		logger.Errorf("Failed to send templated email: %v", err)
		return
	}
	
	logger.Infof("✅ Templated email sent!")
	logger.Infof("   ID: %s", response.ID)
}

func demoBulkEmail(cfg config.EmailProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewEmailService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create email service: %v", err)
		return
	}
	
	ctx := context.Background()
	
	// Bulk email request
	bulkRequest := &services.BulkEmailRequest{
		Recipients: []services.BulkEmailRecipient{
			{
				Email: "user1@example.com",
				Data: map[string]string{
					"user_name": "Alice Johnson",
					"user_id":   "001",
				},
			},
			{
				Email: "user2@example.com",
				Data: map[string]string{
					"user_name": "Bob Smith",
					"user_id":   "002",
				},
			},
			{
				Email: "user3@example.com",
				Data: map[string]string{
					"user_name": "Carol Davis",
					"user_id":   "003",
				},
			},
		},
		TemplateID: "notification",
		TemplateData: map[string]string{
			"notification_title":   "Bulk Email Demo",
			"notification_message": "Hello {{user_name}}! Your user ID is {{user_id}}.",
			"timestamp":           time.Now().Format("2006-01-02 15:04:05"),
		},
		Priority: models.PriorityNormal,
		Metadata: map[string]string{
			"campaign_type": "bulk_demo",
			"batch_id":      "demo-001",
		},
	}
	
	responses, err := service.SendBulkEmail(ctx, bulkRequest)
	if err != nil {
		logger.Errorf("Failed to send bulk email: %v", err)
		return
	}
	
	logger.Infof("✅ Bulk email completed!")
	logger.Infof("   Total emails: %d", len(responses))
	
	successCount := 0
	for i, response := range responses {
		if response.Status == models.StatusSent {
			successCount++
		}
		logger.Infof("   Email %d: %s (%s)", i+1, response.Status, response.Message)
	}
	
	logger.Infof("📊 Success rate: %d/%d (%.1f%%)", 
		successCount, len(responses), 
		float64(successCount)/float64(len(responses))*100)
}

func demoErrorHandling(cfg config.EmailProviderConfig, logger *utils.SimpleLogger) {
	service, err := services.NewEmailService(cfg, logger)
	if err != nil {
		logger.Errorf("Failed to create email service: %v", err)
		return
	}
	
	ctx := context.Background()
	
	// Test various error scenarios
	errorTests := []struct {
		name    string
		request *services.EmailRequest
	}{
		{
			name: "Empty recipients",
			request: &services.EmailRequest{
				To:      []string{},
				Subject: "Test",
			},
		},
		{
			name: "Invalid email address",
			request: &services.EmailRequest{
				To:      []string{"invalid-email"},
				Subject: "Test",
			},
		},
		{
			name: "Missing content",
			request: &services.EmailRequest{
				To: []string{"test@example.com"},
			},
		},
		{
			name: "Non-existent template",
			request: &services.EmailRequest{
				To:         []string{"test@example.com"},
				TemplateID: "non-existent",
			},
		},
	}
	
	for _, test := range errorTests {
		logger.Infof("🧪 Testing: %s", test.name)
		
		_, err := service.SendEmail(ctx, test.request)
		if err != nil {
			logger.Infof("   ❌ Expected error: %v", err)
		} else {
			logger.Infof("   ⚠️  Unexpected success")
		}
	}
	
	// Test provider validation
	logger.Infof("🧪 Testing email validation:")
	validationTests := []string{
		"valid@example.com",
		"invalid-email",
		"",
		"user@domain.co.uk",
		"test@",
	}
	
	for _, email := range validationTests {
		err := service.ValidateEmailAddress(email)
		if err != nil {
			logger.Infof("   ❌ '%s': %v", email, err)
		} else {
			logger.Infof("   ✅ '%s': valid", email)
		}
	}
}