package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nareshkumar-microsoft/notificationService/internal/config"
	"github.com/nareshkumar-microsoft/notificationService/internal/models"
	"github.com/nareshkumar-microsoft/notificationService/pkg/errors"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print startup banner
	printBanner()

	// Print configuration summary
	printConfigSummary(cfg)

	// Demonstrate basic functionality
	demonstrateFoundation()

	// Wait for interrupt signal
	waitForShutdown()

	fmt.Println("Notification Service shutting down...")
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════════╗
║                    🔔 Notification Service                        ║
║                                                                  ║
║  A beginner-friendly notification service built in Go           ║
║  Supports Email, SMS, and Push notifications                    ║
║  Perfect for Hacktoberfest contributions!                       ║
╚══════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printConfigSummary(cfg *config.Config) {
	fmt.Printf("📊 Configuration Summary:\n")
	fmt.Printf("  Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  Database: %s\n", cfg.Database.Type)
	fmt.Printf("  Queue: %s (workers: %d)\n", cfg.Queue.Type, cfg.Queue.Workers)
	fmt.Printf("  Providers:\n")
	fmt.Printf("    📧 Email: %s (enabled: %t)\n", cfg.Providers.Email.Provider, cfg.Providers.Email.Enabled)
	fmt.Printf("    📱 SMS: %s (enabled: %t)\n", cfg.Providers.SMS.Provider, cfg.Providers.SMS.Enabled)
	fmt.Printf("    🔔 Push: %s (enabled: %t)\n", cfg.Providers.Push.Provider, cfg.Providers.Push.Enabled)
	fmt.Println()
}

func demonstrateFoundation() {
	fmt.Println("🧪 Demonstrating Foundation Components:")

	// Demonstrate notification types
	fmt.Println("\n1. Notification Types:")
	types := []models.NotificationType{
		models.NotificationTypeEmail,
		models.NotificationTypeSMS,
		models.NotificationTypePush,
	}
	for _, t := range types {
		fmt.Printf("   ✓ %s\n", t)
	}

	// Demonstrate statuses
	fmt.Println("\n2. Notification Statuses:")
	statuses := []models.NotificationStatus{
		models.StatusPending,
		models.StatusSent,
		models.StatusDelivered,
		models.StatusFailed,
		models.StatusRetrying,
	}
	for _, s := range statuses {
		fmt.Printf("   ✓ %s\n", s)
	}

	// Demonstrate priorities
	fmt.Println("\n3. Priority Levels:")
	priorities := []models.Priority{
		models.PriorityLow,
		models.PriorityNormal,
		models.PriorityHigh,
		models.PriorityUrgent,
	}
	for _, p := range priorities {
		fmt.Printf("   ✓ %s\n", p)
	}

	// Demonstrate error handling
	fmt.Println("\n4. Error Handling:")
	demoErrors := []struct {
		name string
		err  error
	}{
		{"Validation Error", errors.NewValidationError("email", "invalid format")},
		{"Provider Error", errors.NewProviderError("mock", errors.ErrorCodeProviderUnavailable, "service unavailable")},
		{"Rate Limit Error", errors.NewRateLimitError("60")},
	}

	for _, demo := range demoErrors {
		fmt.Printf("   ✓ %s: %s\n", demo.name, demo.err.Error())
	}

	fmt.Println("\n✅ Foundation is ready! All core components are working.")
	fmt.Println("   Next PRs will add concrete implementations for each provider.")
}

func waitForShutdown() {
	fmt.Println("\n🚀 Service is running! Press Ctrl+C to stop...")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wait for signal in a goroutine
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutdown signal received...")
		cancel()
	}()

	// Wait for context cancellation
	<-ctx.Done()
}
