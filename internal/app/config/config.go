package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Logging  LoggingConfig  `json:"logging"`
	Webhook  WebhookConfig  `json:"webhook"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	APIKey       string        `json:"api_key,omitempty"`
	EnableCORS   bool          `json:"enable_cors"`
	TLS          TLSConfig     `json:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password,omitempty"`
	Name            string        `json:"name"`
	SSLMode         string        `json:"ssl_mode"`
	Debug           bool          `json:"debug"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// WhatsAppConfig holds WhatsApp client configuration
type WhatsAppConfig struct {
	Debug       bool   `json:"debug"`
	LogLevel    string `json:"log_level"`
	OSName      string `json:"os_name"`
	Timeout     int    `json:"timeout"`
	RetryCount  int    `json:"retry_count"`
	AutoConnect bool   `json:"auto_connect"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level       string `json:"level"`
	Format      string `json:"format"` // "json" or "console"
	ColorOutput bool   `json:"color_output"`
	TimeFormat  string `json:"time_format"`
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	GlobalURL string        `json:"global_url"`
	Timeout   time.Duration `json:"timeout"`
	Retries   int           `json:"retries"`
	Events    []string      `json:"events"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Warn().Err(err).Msg("Could not load .env file (it may not exist)")
	}

	config := &Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		WhatsApp: loadWhatsAppConfig(),
		Logging:  loadLoggingConfig(),
		Webhook:  loadWebhookConfig(),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host:         getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		Port:         getEnvAsIntOrDefault("SERVER_PORT", 8080),
		ReadTimeout:  getEnvAsDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout: getEnvAsDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getEnvAsDurationOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
		APIKey:       os.Getenv("WAZMEOW_API_KEY"),
		EnableCORS:   getEnvAsBoolOrDefault("SERVER_ENABLE_CORS", true),
		TLS: TLSConfig{
			Enabled:  getEnvAsBoolOrDefault("TLS_ENABLED", false),
			CertFile: os.Getenv("TLS_CERT_FILE"),
			KeyFile:  os.Getenv("TLS_KEY_FILE"),
		},
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            getEnvOrDefault("DB_HOST", "localhost"),
		Port:            getEnvAsIntOrDefault("DB_PORT", 5432),
		User:            getEnvOrDefault("DB_USER", "wazmeow"),
		Password:        os.Getenv("DB_PASSWORD"),
		Name:            getEnvOrDefault("DB_NAME", "wazmeow"),
		SSLMode:         getEnvOrDefault("DB_SSL_MODE", "disable"),
		Debug:           getEnvAsBoolOrDefault("DB_DEBUG", false),
		MaxOpenConns:    getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvAsDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}
}

func loadWhatsAppConfig() WhatsAppConfig {
	return WhatsAppConfig{
		Debug:       getEnvAsBoolOrDefault("WHATSAPP_DEBUG", false),
		LogLevel:    getEnvOrDefault("WHATSAPP_LOG_LEVEL", "INFO"),
		OSName:      getEnvOrDefault("WHATSAPP_OS_NAME", "WazMeow"),
		Timeout:     getEnvAsIntOrDefault("WHATSAPP_TIMEOUT", 30),
		RetryCount:  getEnvAsIntOrDefault("WHATSAPP_RETRY_COUNT", 3),
		AutoConnect: getEnvAsBoolOrDefault("WHATSAPP_AUTO_CONNECT", true),
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:       getEnvOrDefault("LOG_LEVEL", "info"),
		Format:      getEnvOrDefault("LOG_FORMAT", "console"),
		ColorOutput: getEnvAsBoolOrDefault("LOG_COLOR_OUTPUT", true),
		TimeFormat:  getEnvOrDefault("LOG_TIME_FORMAT", "2006-01-02 15:04:05"),
	}
}

func loadWebhookConfig() WebhookConfig {
	eventsStr := getEnvOrDefault("WEBHOOK_EVENTS", "message,presence,receipt")
	events := strings.Split(eventsStr, ",")
	for i, event := range events {
		events[i] = strings.TrimSpace(event)
	}

	return WebhookConfig{
		GlobalURL: os.Getenv("WEBHOOK_GLOBAL_URL"),
		Timeout:   getEnvAsDurationOrDefault("WEBHOOK_TIMEOUT", 10*time.Second),
		Retries:   getEnvAsIntOrDefault("WEBHOOK_RETRIES", 3),
		Events:    events,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate TLS config
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" || c.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not provided")
		}
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate logging config
	if !isValidLogLevel(c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}
	if c.Logging.Format != "json" && c.Logging.Format != "console" {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	return nil
}

// SetupLogger configures the global logger based on configuration
func (c *Config) SetupLogger() {
	// Set log level
	level, err := zerolog.ParseLevel(c.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	if c.Logging.Format == "json" {
		log.Logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Str("service", "wazmeow").
			Logger()
	} else {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: c.Logging.TimeFormat,
			NoColor:    !c.Logging.ColorOutput,
		}

		output.FormatLevel = func(i interface{}) string {
			if i == nil {
				return ""
			}
			lvl := strings.ToUpper(i.(string))
			switch lvl {
			case "DEBUG":
				return "\x1b[34m" + lvl + "\x1b[0m"
			case "INFO":
				return "\x1b[32m" + lvl + "\x1b[0m"
			case "WARN":
				return "\x1b[33m" + lvl + "\x1b[0m"
			case "ERROR", "FATAL", "PANIC":
				return "\x1b[31m" + lvl + "\x1b[0m"
			default:
				return lvl
			}
		}

		log.Logger = zerolog.New(output).
			With().
			Timestamp().
			Str("service", "wazmeow").
			Logger()
	}
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode,
	)
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func isValidLogLevel(level string) bool {
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	for _, validLevel := range validLevels {
		if strings.ToLower(level) == validLevel {
			return true
		}
	}
	return false
}
