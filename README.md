# Notification Service

A beginner-friendly notification service built in Go that supports multiple notification channels (Email, SMS, Push) with mock providers for learning and development.

## 🚀 Features

- **Multi-Channel Support**: Email, SMS, and Push notifications
- **Provider Pattern**: Easily extensible provider interface
- **Mock Implementations**: Perfect for development and testing
- **Type Safety**: Strongly typed notification system
- **Error Handling**: Comprehensive error types and handling
- **Logging**: Built-in logging for debugging and monitoring
- **Email Templates**: Rich templating system with variable substitution
- **Bulk Operations**: Support for bulk email sending
- **Validation**: Comprehensive input validation and sanitization

## 📁 Project Structure

```
├── cmd/                    # Application entry points
│   ├── demo/              # Main demo application
│   └── email_demo/        # Email-specific demo
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── models/           # Data models and types
│   ├── providers/        # Notification providers
│   ├── services/         # Business logic services
│   └── utils/            # Utility functions and helpers
├── pkg/                  # Public library code
│   ├── errors/          # Custom error types
│   └── interfaces/      # Public interfaces
├── tests/               # Integration tests
└── docs/               # Documentation
```

## 🛠 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/nareshkumar-microsoft/notificationService.git
cd notificationService

# Download dependencies
go mod tidy

# Run tests
go test ./...

# Run email demo
go run ./cmd/email_demo/

# Build the application
go build -o bin/notification-service ./cmd/demo
```

## 📖 Usage Examples

### Email Provider

```go
package main

import (
    "context"
    "github.com/nareshkumar-microsoft/notificationService/internal/config"
    "github.com/nareshkumar-microsoft/notificationService/internal/services"
    "github.com/nareshkumar-microsoft/notificationService/internal/utils"
)

func main() {
    // Create email service
    cfg := config.EmailProviderConfig{
        Provider: "mock",
        Enabled:  true,
    }
    logger := utils.NewSimpleLogger("info")
    service, _ := services.NewEmailService(cfg, logger)
    
    // Send simple email
    request := &services.EmailRequest{
        To:       []string{"user@example.com"},
        Subject:  "Hello World",
        HTMLBody: "<h1>Hello!</h1><p>This is a test email.</p>",
        TextBody: "Hello!\n\nThis is a test email.",
    }
    
    response, err := service.SendEmail(context.Background(), request)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Email sent with ID: %s\n", response.ID)
}
```

### Template Usage

```go
// Send email with template
templateRequest := &services.EmailRequest{
    To:         []string{"user@example.com"},
    TemplateID: "welcome",
    TemplateData: map[string]string{
        "user_name":    "John Doe",
        "service_name": "My App",
    },
}

response, err := service.SendEmail(ctx, templateRequest)
```

### Bulk Email

```go
// Send bulk emails
bulkRequest := &services.BulkEmailRequest{
    Recipients: []services.BulkEmailRecipient{
        {Email: "user1@example.com", Data: map[string]string{"name": "User 1"}},
        {Email: "user2@example.com", Data: map[string]string{"name": "User 2"}},
    },
    TemplateID: "notification",
    TemplateData: map[string]string{
        "message": "Hello {{name}}!",
    },
}

responses, err := service.SendBulkEmail(ctx, bulkRequest)
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific provider tests
go test ./internal/providers/... -v

# Run email service tests
go test ./internal/services/... -v
```

## 🔧 Available Email Templates

The email provider comes with built-in templates:

- **welcome**: User onboarding email
- **password_reset**: Password reset email with secure link
- **notification**: General purpose notification template

## 🤝 Contributing

This project is designed for Hacktoberfest contributions! Each feature is implemented in separate PRs to make it beginner-friendly.

### Contribution Guidelines

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📋 Roadmap

- [x] **PR #1**: Project Foundation & Core Structure
- [x] **PR #2**: Email Notification Provider
- [ ] **PR #3**: SMS Notification Provider
- [ ] **PR #4**: Push Notification Provider
- [ ] **PR #5**: Notification Queue & Batch Processing
- [ ] **PR #6**: REST API & Integration Layer

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built for Hacktoberfest 2024
- Designed to be beginner-friendly while following Go best practices
- Mock implementations make it perfect for learning