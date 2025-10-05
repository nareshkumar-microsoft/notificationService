# Contributing to Notification Service

Welcome to the Notification Service project! 🎉 This project is designed to be **beginner-friendly** and **Hacktoberfest-ready**. We're excited to have you contribute!

## 🎯 Project Goals

This notification service is built to:
- **Educate**: Teach Go best practices and clean architecture
- **Include**: Welcome developers of all skill levels
- **Scale**: Provide a foundation that can grow into a production system
- **Inspire**: Show how to build maintainable, testable code

## 🗺️ Contribution Roadmap

We have planned **6 progressive PRs** that build the notification service step by step:

### ✅ PR #1: Project Foundation & Core Structure (COMPLETED)
- [x] Go module setup with proper dependencies
- [x] Core interfaces and type definitions  
- [x] Basic folder structure following Go conventions
- [x] Comprehensive error handling system
- [x] Configuration management
- [x] Validation utilities
- [x] Complete test coverage

### 🚧 Upcoming PRs:

#### 📧 PR #2: Email Notification Provider
**Difficulty: Beginner**
- [ ] Mock SMTP service implementation
- [ ] Email templates (HTML/text)
- [ ] Email validation and sanitization
- [ ] Provider interface implementation
- [ ] Comprehensive unit tests

#### 📱 PR #3: SMS Notification Provider  
**Difficulty: Beginner-Intermediate**
- [ ] Mock SMS gateway integration
- [ ] Phone number validation (international formats)
- [ ] SMS templates and character limits
- [ ] Delivery status tracking
- [ ] Cost calculation utilities

#### 🔔 PR #4: Push Notification Provider
**Difficulty: Intermediate**
- [ ] Mock FCM/APNs integration  
- [ ] Device token management
- [ ] Platform-specific payload formatting
- [ ] Retry mechanisms
- [ ] Badge and sound handling

#### ⚡ PR #5: Notification Queue & Batch Processing
**Difficulty: Intermediate-Advanced**
- [ ] In-memory queue implementation
- [ ] Batch processing capabilities
- [ ] Priority-based handling
- [ ] Rate limiting and throttling
- [ ] Scheduled notifications

#### 🌐 PR #6: REST API & Integration Layer
**Difficulty: Advanced**
- [ ] HTTP server with Gin/Echo
- [ ] RESTful endpoints
- [ ] Request validation middleware
- [ ] API documentation (OpenAPI)
- [ ] Integration tests

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** (optional) - For using the Makefile

### Setting Up Your Development Environment

1. **Fork the repository**
   ```bash
   # Click the "Fork" button on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/notificationService.git
   cd notificationService
   ```

2. **Set up the project**
   ```bash
   # Download dependencies
   make deps
   
   # Run all checks
   make check-all
   
   # Try the demo
   make demo
   ```

3. **Verify everything works**
   ```bash
   # Run tests
   make test
   
   # Check code formatting
   make fmt
   
   # Run linting (optional)
   make lint
   ```

## 🛠️ Development Workflow

### For Each Contribution:

1. **Create a feature branch**
   ```bash
   git checkout -b feature/email-provider
   # or
   git checkout -b fix/validation-bug
   ```

2. **Make your changes**
   - Follow Go conventions and best practices
   - Add comprehensive tests for new functionality
   - Update documentation as needed
   - Ensure code is well-commented

3. **Test your changes**
   ```bash
   make check-all  # Runs fmt, vet, test
   ```

4. **Commit with descriptive messages**
   ```bash
   git add .
   git commit -m "feat: implement email provider with SMTP mock
   
   - Add MockSMTPProvider struct with Send method
   - Implement email validation and sanitization
   - Add comprehensive test coverage (>90%)
   - Include HTML and text template support
   
   Closes #issue-number"
   ```

5. **Push and create PR**
   ```bash
   git push origin feature/email-provider
   # Then create a Pull Request on GitHub
   ```

## 🎯 Contribution Guidelines

### Code Quality Standards

- **Test Coverage**: Aim for >85% test coverage on new code
- **Documentation**: Add godoc comments for all public functions
- **Error Handling**: Use our custom error types consistently
- **Validation**: Validate all inputs using our validation utilities
- **Naming**: Follow Go naming conventions (PascalCase for exports, camelCase for private)

### Code Style

- **Formatting**: Use `go fmt` (or `make fmt`)
- **Linting**: Address all `go vet` issues
- **Imports**: Group imports (standard, third-party, local)
- **Comments**: Explain *why*, not just *what*

### Testing Requirements

- **Unit Tests**: Test all public functions
- **Table Tests**: Use table-driven tests for multiple scenarios
- **Mocks**: Create mocks for external dependencies
- **Edge Cases**: Test error conditions and edge cases
- **Examples**: Include example usage in tests

### Example Test Structure:
```go
func TestEmailProvider_Send(t *testing.T) {
    tests := []struct {
        name           string
        notification   *models.EmailNotification
        expectedStatus models.NotificationStatus
        expectedError  error
    }{
        {
            name: "valid email sends successfully",
            notification: &models.EmailNotification{
                // ... test data
            },
            expectedStatus: models.StatusSent,
            expectedError:  nil,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 📝 PR Requirements

### What Makes a Good PR:

✅ **Clear Description**: Explain what the PR does and why  
✅ **Focused Scope**: One feature or fix per PR  
✅ **Tests Included**: Comprehensive test coverage  
✅ **Documentation**: Update relevant docs  
✅ **No Breaking Changes**: Maintain backward compatibility  
✅ **Clean History**: Squash commits if needed  

### PR Template:
```markdown
## Description
Brief description of what this PR accomplishes.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated  
- [ ] All tests pass locally
- [ ] Manual testing performed

## Checklist
- [ ] Code follows Go best practices
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes
```

## 🐛 Reporting Issues

Found a bug? Have a suggestion? Please create an issue with:

- **Clear title** describing the problem
- **Steps to reproduce** the issue
- **Expected behavior** vs actual behavior
- **Environment details** (Go version, OS, etc.)
- **Relevant code snippets** or error messages

## 💡 Feature Requests

We love new ideas! For feature requests:

- **Check existing issues** to avoid duplicates
- **Describe the use case** - why is this needed?
- **Propose a solution** if you have one in mind
- **Consider the scope** - does it fit our roadmap?

## 🏆 Hacktoberfest Participation

This project is **Hacktoberfest-friendly**! Here's how to participate:

1. **Register** for [Hacktoberfest](https://hacktoberfest.com)
2. **Find an issue** labeled `hacktoberfest` or `good first issue`
3. **Comment** on the issue to let us know you're working on it
4. **Create your PR** following our guidelines
5. **Celebrate** your contribution! 🎉

### Hacktoberfest Labels:
- `hacktoberfest` - Suitable for Hacktoberfest
- `good first issue` - Perfect for beginners
- `help wanted` - We'd love assistance with this
- `beginner-friendly` - Great for learning

## 🙏 Recognition

All contributors will be:
- **Listed** in our README contributors section
- **Credited** in release notes for significant contributions
- **Welcomed** into our community of developers

## 📞 Getting Help

Need help? We're here for you:

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Comments**: Ask questions in PR reviews

## 🎉 Thank You!

Thank you for considering contributing to the Notification Service! Your contributions help make this project better for everyone. Whether you're fixing a typo, adding a major feature, or improving documentation, every contribution matters.

Happy coding! 🚀