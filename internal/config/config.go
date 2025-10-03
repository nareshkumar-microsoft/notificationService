package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the main configuration for the notification service
type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Logger    LoggerConfig    `json:"logger"`
	Queue     QueueConfig     `json:"queue"`
	Providers ProvidersConfig `json:"providers"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	EnableCORS   bool          `json:"enable_cors"`
	EnableTLS    bool          `json:"enable_tls"`
	CertFile     string        `json:"cert_file,omitempty"`
	KeyFile      string        `json:"key_file,omitempty"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type         string        `json:"type"` // "memory", "postgres", "mysql", etc.
	Host         string        `json:"host,omitempty"`
	Port         int           `json:"port,omitempty"`
	Database     string        `json:"database,omitempty"`
	Username     string        `json:"username,omitempty"`
	Password     string        `json:"password,omitempty"`
	MaxOpenConns int           `json:"max_open_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
	SSLMode      string        `json:"ssl_mode,omitempty"`
}

// LoggerConfig represents logging configuration
type LoggerConfig struct {
	Level      string `json:"level"`  // "debug", "info", "warn", "error"
	Format     string `json:"format"` // "json", "text"
	Output     string `json:"output"` // "stdout", "stderr", "file"
	Filename   string `json:"filename,omitempty"`
	MaxSize    int    `json:"max_size,omitempty"`    // megabytes
	MaxBackups int    `json:"max_backups,omitempty"` // number of backups
	MaxAge     int    `json:"max_age,omitempty"`     // days
	Compress   bool   `json:"compress,omitempty"`
}

// QueueConfig represents queue configuration
type QueueConfig struct {
	Type           string        `json:"type"` // "memory", "redis", "rabbitmq", etc.
	MaxSize        int           `json:"max_size"`
	Workers        int           `json:"workers"`
	BatchSize      int           `json:"batch_size"`
	ProcessTimeout time.Duration `json:"process_timeout"`
	RetryDelay     time.Duration `json:"retry_delay"`
	MaxRetries     int           `json:"max_retries"`
	// Redis specific
	RedisURL      string `json:"redis_url,omitempty"`
	RedisPassword string `json:"redis_password,omitempty"`
	RedisDB       int    `json:"redis_db,omitempty"`
}

// ProvidersConfig represents configuration for all notification providers
type ProvidersConfig struct {
	Email EmailProviderConfig `json:"email"`
	SMS   SMSProviderConfig   `json:"sms"`
	Push  PushProviderConfig  `json:"push"`
}

// EmailProviderConfig represents email provider configuration
type EmailProviderConfig struct {
	Provider string            `json:"provider"` // "mock", "smtp", "sendgrid", "ses", etc.
	Enabled  bool              `json:"enabled"`
	Settings map[string]string `json:"settings"`

	// SMTP specific settings
	SMTPHost     string `json:"smtp_host,omitempty"`
	SMTPPort     int    `json:"smtp_port,omitempty"`
	SMTPUsername string `json:"smtp_username,omitempty"`
	SMTPPassword string `json:"smtp_password,omitempty"`
	SMTPUseTLS   bool   `json:"smtp_use_tls,omitempty"`

	// SendGrid specific
	SendGridAPIKey string `json:"sendgrid_api_key,omitempty"`

	// AWS SES specific
	SESRegion          string `json:"ses_region,omitempty"`
	SESAccessKeyID     string `json:"ses_access_key_id,omitempty"`
	SESSecretAccessKey string `json:"ses_secret_access_key,omitempty"`
}

// SMSProviderConfig represents SMS provider configuration
type SMSProviderConfig struct {
	Provider string            `json:"provider"` // "mock", "twilio", "nexmo", etc.
	Enabled  bool              `json:"enabled"`
	Settings map[string]string `json:"settings"`

	// Twilio specific
	TwilioAccountSID string `json:"twilio_account_sid,omitempty"`
	TwilioAuthToken  string `json:"twilio_auth_token,omitempty"`
	TwilioFromNumber string `json:"twilio_from_number,omitempty"`

	// Nexmo specific
	NexmoAPIKey    string `json:"nexmo_api_key,omitempty"`
	NexmoAPISecret string `json:"nexmo_api_secret,omitempty"`
	NexmoFromName  string `json:"nexmo_from_name,omitempty"`
}

// PushProviderConfig represents push notification provider configuration
type PushProviderConfig struct {
	Provider string            `json:"provider"` // "mock", "fcm", "apns", etc.
	Enabled  bool              `json:"enabled"`
	Settings map[string]string `json:"settings"`

	// FCM specific
	FCMServerKey string `json:"fcm_server_key,omitempty"`
	FCMProjectID string `json:"fcm_project_id,omitempty"`

	// APNS specific
	APNSKeyID      string `json:"apns_key_id,omitempty"`
	APNSTeamID     string `json:"apns_team_id,omitempty"`
	APNSBundleID   string `json:"apns_bundle_id,omitempty"`
	APNSKeyFile    string `json:"apns_key_file,omitempty"`
	APNSProduction bool   `json:"apns_production,omitempty"`
}

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			EnableCORS:   getEnvBool("SERVER_ENABLE_CORS", true),
			EnableTLS:    getEnvBool("SERVER_ENABLE_TLS", false),
			CertFile:     getEnv("SERVER_CERT_FILE", ""),
			KeyFile:      getEnv("SERVER_KEY_FILE", ""),
		},
		Database: DatabaseConfig{
			Type:         getEnv("DB_TYPE", "memory"),
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			Database:     getEnv("DB_NAME", "notifications"),
			Username:     getEnv("DB_USERNAME", ""),
			Password:     getEnv("DB_PASSWORD", ""),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvDuration("DB_MAX_LIFETIME", 5*time.Minute),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			Filename:   getEnv("LOG_FILENAME", ""),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
			Compress:   getEnvBool("LOG_COMPRESS", true),
		},
		Queue: QueueConfig{
			Type:           getEnv("QUEUE_TYPE", "memory"),
			MaxSize:        getEnvInt("QUEUE_MAX_SIZE", 10000),
			Workers:        getEnvInt("QUEUE_WORKERS", 5),
			BatchSize:      getEnvInt("QUEUE_BATCH_SIZE", 10),
			ProcessTimeout: getEnvDuration("QUEUE_PROCESS_TIMEOUT", 30*time.Second),
			RetryDelay:     getEnvDuration("QUEUE_RETRY_DELAY", 5*time.Second),
			MaxRetries:     getEnvInt("QUEUE_MAX_RETRIES", 3),
			RedisURL:       getEnv("REDIS_URL", ""),
			RedisPassword:  getEnv("REDIS_PASSWORD", ""),
			RedisDB:        getEnvInt("REDIS_DB", 0),
		},
		Providers: ProvidersConfig{
			Email: EmailProviderConfig{
				Provider:           getEnv("EMAIL_PROVIDER", "mock"),
				Enabled:            getEnvBool("EMAIL_ENABLED", true),
				Settings:           make(map[string]string),
				SMTPHost:           getEnv("SMTP_HOST", ""),
				SMTPPort:           getEnvInt("SMTP_PORT", 587),
				SMTPUsername:       getEnv("SMTP_USERNAME", ""),
				SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
				SMTPUseTLS:         getEnvBool("SMTP_USE_TLS", true),
				SendGridAPIKey:     getEnv("SENDGRID_API_KEY", ""),
				SESRegion:          getEnv("SES_REGION", ""),
				SESAccessKeyID:     getEnv("SES_ACCESS_KEY_ID", ""),
				SESSecretAccessKey: getEnv("SES_SECRET_ACCESS_KEY", ""),
			},
			SMS: SMSProviderConfig{
				Provider:         getEnv("SMS_PROVIDER", "mock"),
				Enabled:          getEnvBool("SMS_ENABLED", true),
				Settings:         make(map[string]string),
				TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
				TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
				TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
				NexmoAPIKey:      getEnv("NEXMO_API_KEY", ""),
				NexmoAPISecret:   getEnv("NEXMO_API_SECRET", ""),
				NexmoFromName:    getEnv("NEXMO_FROM_NAME", ""),
			},
			Push: PushProviderConfig{
				Provider:       getEnv("PUSH_PROVIDER", "mock"),
				Enabled:        getEnvBool("PUSH_ENABLED", true),
				Settings:       make(map[string]string),
				FCMServerKey:   getEnv("FCM_SERVER_KEY", ""),
				FCMProjectID:   getEnv("FCM_PROJECT_ID", ""),
				APNSKeyID:      getEnv("APNS_KEY_ID", ""),
				APNSTeamID:     getEnv("APNS_TEAM_ID", ""),
				APNSBundleID:   getEnv("APNS_BUNDLE_ID", ""),
				APNSKeyFile:    getEnv("APNS_KEY_FILE", ""),
				APNSProduction: getEnvBool("APNS_PRODUCTION", false),
			},
		},
	}

	return config, nil
}

// LoadConfigFromFile loads configuration from a JSON file
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfigToFile saves configuration to a JSON file
func SaveConfigToFile(config *Config, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Utility functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
