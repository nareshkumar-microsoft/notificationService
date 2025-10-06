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
	fmt.Println("🚀 Push Notification Service Demo")
	fmt.Println("=================================")

	// Initialize logger
	logger := utils.NewSimpleLogger("info")

	// Create push provider configuration
	pushConfig := config.PushProviderConfig{
		Provider: "mock",
		Enabled:  true,
	}

	// Create push provider
	provider := providers.NewMockPushProvider(pushConfig)

	// Create push service
	service := services.NewPushService(provider, pushConfig, logger)

	// Demo scenarios
	fmt.Println("\n📱 Running Push Notification Demo Scenarios...")

	ctx := context.Background()

	// Scenario 1: iOS Push Notification
	fmt.Println("\n1️⃣ iOS Push Notification")
	fmt.Println("------------------------")

	iosRequest := &models.NotificationRequest{
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityHigh,
		Recipient: "ios-user@example.com",
		Subject:   "iOS Notification",
		Body:      "This is a push notification for iOS device!",
		PushData: &models.PushData{
			DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			Platform:    "ios",
			Title:       "🍎 iOS Alert",
			Sound:       "default",
			Badge:       5,
			Data: map[string]string{
				"action":  "view_details",
				"item_id": "12345",
			},
		},
	}

	response, err := service.SendPushNotification(ctx, iosRequest)
	if err != nil {
		log.Printf("❌ Failed to send iOS push: %v", err)
	} else {
		fmt.Printf("✅ iOS push sent successfully!\n")
		fmt.Printf("   📍 ID: %s\n", response.ID)
		fmt.Printf("   📍 Provider ID: %s\n", response.ProviderID)
		fmt.Printf("   📍 Status: %s\n", response.Status)
	}

	// Scenario 2: Android Push Notification
	fmt.Println("\n2️⃣ Android Push Notification")
	fmt.Println("----------------------------")

	androidRequest := &models.NotificationRequest{
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityNormal,
		Recipient: "android-user@example.com",
		Subject:   "Android Notification",
		Body:      "This is a push notification for Android device!",
		PushData: &models.PushData{
			DeviceToken: "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
			Platform:    "android",
			Title:       "🤖 Android Alert",
			Icon:        "ic_notification",
			ImageURL:    "https://example.com/image.png",
			ClickAction: "OPEN_ACTIVITY",
			Data: map[string]string{
				"type":     "promotional",
				"offer_id": "SAVE20",
			},
		},
	}

	response, err = service.SendPushNotification(ctx, androidRequest)
	if err != nil {
		log.Printf("❌ Failed to send Android push: %v", err)
	} else {
		fmt.Printf("✅ Android push sent successfully!\n")
		fmt.Printf("   📍 ID: %s\n", response.ID)
		fmt.Printf("   📍 Provider ID: %s\n", response.ProviderID)
		fmt.Printf("   📍 Status: %s\n", response.Status)
	}

	// Scenario 3: Web Push Notification
	fmt.Println("\n3️⃣ Web Push Notification")
	fmt.Println("------------------------")

	webRequest := &models.NotificationRequest{
		Type:      models.NotificationTypePush,
		Priority:  models.PriorityLow,
		Recipient: "web-user@example.com",
		Subject:   "Web Notification",
		Body:      "This is a push notification for web browser!",
		PushData: &models.PushData{
			DeviceToken: "BNJzWlpOQMEK-web-push-token-example-abc123def456ghi789jkl012",
			Platform:    "web",
			Title:       "🌐 Web Alert",
			Icon:        "/icon-192x192.png",
			ImageURL:    "https://example.com/banner.jpg",
			Data: map[string]string{
				"url":      "https://example.com/news",
				"category": "news",
			},
		},
	}

	response, err = service.SendPushNotification(ctx, webRequest)
	if err != nil {
		log.Printf("❌ Failed to send Web push: %v", err)
	} else {
		fmt.Printf("✅ Web push sent successfully!\n")
		fmt.Printf("   📍 ID: %s\n", response.ID)
		fmt.Printf("   📍 Provider ID: %s\n", response.ProviderID)
		fmt.Printf("   📍 Status: %s\n", response.Status)
	}

	// Scenario 4: Using Push Templates
	fmt.Println("\n4️⃣ Template-based Push Notification")
	fmt.Println("----------------------------------")

	template, err := provider.RenderTemplate("welcome_push", map[string]string{
		"app_name":  "Demo App",
		"user_name": "John Doe",
	})
	if err != nil {
		log.Printf("❌ Failed to render template: %v", err)
	} else {
		templateRequest := &models.NotificationRequest{
			Type:      models.NotificationTypePush,
			Priority:  models.PriorityNormal,
			Recipient: "template-user@example.com",
			Subject:   template.Title,
			Body:      template.Body,
			PushData: &models.PushData{
				DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
				Platform:    "ios",
				Title:       template.Title,
				Icon:        template.Icon,
				Sound:       template.Sound,
			},
		}

		response, err = service.SendPushNotification(ctx, templateRequest)
		if err != nil {
			log.Printf("❌ Failed to send template push: %v", err)
		} else {
			fmt.Printf("✅ Template push sent successfully!\n")
			fmt.Printf("   📍 Template: %s\n", template.Name)
			fmt.Printf("   📍 Title: %s\n", template.Title)
			fmt.Printf("   📍 Body: %s\n", template.Body)
			fmt.Printf("   📍 ID: %s\n", response.ID)
		}
	}

	// Scenario 5: Bulk Push Notifications
	fmt.Println("\n5️⃣ Bulk Push Notifications")
	fmt.Println("-------------------------")

	bulkRequests := []*models.NotificationRequest{
		{
			Type:      models.NotificationTypePush,
			Priority:  models.PriorityNormal,
			Recipient: "user1@example.com",
			Body:      "Bulk notification 1",
			PushData: &models.PushData{
				DeviceToken: "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
				Platform:    "ios",
				Title:       "Bulk Alert 1",
			},
		},
		{
			Type:      models.NotificationTypePush,
			Priority:  models.PriorityNormal,
			Recipient: "user2@example.com",
			Body:      "Bulk notification 2",
			PushData: &models.PushData{
				DeviceToken: "eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-",
				Platform:    "android",
				Title:       "Bulk Alert 2",
			},
		},
		{
			Type:      models.NotificationTypePush,
			Priority:  models.PriorityNormal,
			Recipient: "user3@example.com",
			Body:      "Bulk notification 3",
			PushData: &models.PushData{
				DeviceToken: "BNJzWlpOQMEK-web-push-token-example-abc123def456ghi789jkl012",
				Platform:    "web",
				Title:       "Bulk Alert 3",
			},
		},
	}

	responses, err := service.SendBulkPushNotifications(ctx, bulkRequests)
	if err != nil {
		log.Printf("❌ Bulk send had errors: %v", err)
	}

	fmt.Printf("✅ Bulk send completed! Sent %d out of %d notifications\n",
		len(responses), len(bulkRequests))

	// Scenario 6: Device Registration and Management
	fmt.Println("\n6️⃣ Device Management")
	fmt.Println("-------------------")

	deviceToken := "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678"

	// Register device
	err = service.RegisterDevice(deviceToken, "ios", map[string]string{
		"app_version": "1.2.3",
		"os_version":  "15.0",
		"device_name": "iPhone 13",
	})
	if err != nil {
		log.Printf("❌ Failed to register device: %v", err)
	} else {
		fmt.Printf("✅ Device registered successfully\n")
	}

	// Get platform config
	config := service.GetPlatformConfig("ios")
	fmt.Printf("📱 iOS Platform Config:\n")
	fmt.Printf("   📍 Max Payload: %d bytes\n", config.MaxPayload)
	fmt.Printf("   📍 Bundle ID: %s\n", config.BundleID)

	// Unregister device
	err = service.UnregisterDevice(deviceToken)
	if err != nil {
		log.Printf("❌ Failed to unregister device: %v", err)
	} else {
		fmt.Printf("✅ Device unregistered successfully\n")
	}

	// Scenario 7: Platform-specific Features
	fmt.Println("\n7️⃣ Platform-specific Features")
	fmt.Println("----------------------------")

	// Long content that will be truncated per platform
	longTitle := "This is a very long notification title that will be truncated differently based on the platform capabilities and limits"
	longBody := "This is a very long notification message that will be truncated based on platform-specific limits. iOS typically allows shorter messages compared to Android, and web push has its own constraints. The mock provider will automatically handle platform-specific formatting and truncation."

	platforms := []string{"ios", "android", "web"}
	tokens := []string{
		"a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",                                                                                  // iOS
		"eHQq_abc123def456ghi789jkl012mno345pqr678stu901vwx234yzaBCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-", // Android
		"BNJzWlpOQMEK-web-push-token-example-abc123def456ghi789jkl012",                                                                                      // Web
	}

	for i, platform := range platforms {
		fmt.Printf("\n%s formatting:\n", platform)

		request := &models.NotificationRequest{
			Type:      models.NotificationTypePush,
			Priority:  models.PriorityNormal,
			Recipient: fmt.Sprintf("%s-user@example.com", platform),
			Body:      longBody,
			PushData: &models.PushData{
				DeviceToken: tokens[i],
				Platform:    platform,
				Title:       longTitle,
			},
		}

		response, err = service.SendPushNotification(ctx, request)
		if err != nil {
			log.Printf("❌ Failed to send %s push: %v", platform, err)
		} else {
			fmt.Printf("✅ %s push sent with platform-specific formatting\n", platform)
		}
	}

	// Scenario 8: Health Check and Provider Info
	fmt.Println("\n8️⃣ Service Health and Provider Info")
	fmt.Println("----------------------------------")

	// Health check
	err = service.HealthCheck(ctx)
	if err != nil {
		log.Printf("❌ Health check failed: %v", err)
	} else {
		fmt.Printf("✅ Push service is healthy\n")
	}

	// Provider info
	providerConfig := service.GetProvider().GetConfig()
	fmt.Printf("📊 Provider Information:\n")
	fmt.Printf("   📍 Name: %s\n", providerConfig.Name)
	fmt.Printf("   📍 Type: %s\n", providerConfig.Type)
	fmt.Printf("   📍 Enabled: %t\n", providerConfig.Enabled)
	fmt.Printf("   📍 Max Retries: %d\n", providerConfig.MaxRetries)
	fmt.Printf("   📍 Timeout: %d seconds\n", providerConfig.Timeout)
	fmt.Printf("   📍 Rate Limit: %d req/min\n", providerConfig.RateLimit.RequestsPerMin)

	// Supported platforms
	platforms = service.GetSupportedPlatforms()
	fmt.Printf("   📍 Supported Platforms: %v\n", platforms)

	// Check sent push notifications
	fmt.Println("\n📊 Push Statistics")
	fmt.Println("----------------")

	sentPushes := provider.GetSentPushes()
	fmt.Printf("📈 Total pushes sent: %d\n", len(sentPushes))

	// Count by platform
	platformCounts := make(map[string]int)
	deliveredCount := 0

	for _, push := range sentPushes {
		platformCounts[push.Platform]++
		if push.DeliveredAt != nil {
			deliveredCount++
		}
	}

	fmt.Printf("📊 Breakdown by platform:\n")
	for platform, count := range platformCounts {
		fmt.Printf("   📍 %s: %d notifications\n", platform, count)
	}

	fmt.Printf("📦 Delivery success rate: %.1f%% (%d/%d)\n",
		float64(deliveredCount)/float64(len(sentPushes))*100,
		deliveredCount, len(sentPushes))

	// Wait for any async delivery updates
	time.Sleep(1 * time.Second)

	fmt.Println("\n🎉 Push Notification Demo completed successfully!")
	fmt.Println("================================================")
	fmt.Println("This demo showcased:")
	fmt.Println("• Multi-platform push notifications (iOS, Android, Web)")
	fmt.Println("• Template-based notifications")
	fmt.Println("• Bulk notification sending")
	fmt.Println("• Device registration and management")
	fmt.Println("• Platform-specific formatting and validation")
	fmt.Println("• Service health monitoring")
	fmt.Println("• Comprehensive statistics and reporting")
	fmt.Println("\nThe mock provider simulates real push notification services")
	fmt.Println("with delivery tracking, rate limiting, and error handling!")
}
