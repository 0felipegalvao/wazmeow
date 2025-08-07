package logger

import (
	"os"
	"strings"
	"time"

	"wazmeow/internal/app/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds logger configuration (compatible with app config)
type Config struct {
	Level       string `json:"level"`
	Format      string `json:"format"` // "json" or "console"
	ColorOutput bool   `json:"color_output"`
	TimeFormat  string `json:"time_format"`
}

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
	config Config
}

// New creates a new logger instance with the given configuration
func New(config Config) *Logger {
	// Set global log level
	level := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	// Configure time format
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
	}

	// Use stdout as default output
	output := os.Stdout

	// Configure format
	var logger zerolog.Logger
	switch strings.ToLower(config.Format) {
	case "console", "pretty":
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: config.TimeFormat,
			NoColor:    !config.ColorOutput,
		}).With().Timestamp().Str("service", "wazmeow").Logger()
	case "json":
		logger = zerolog.New(output).With().Timestamp().Str("service", "wazmeow").Logger()
	default:
		logger = zerolog.New(output).With().Timestamp().Str("service", "wazmeow").Logger()
	}

	return &Logger{
		Logger: &logger,
		config: config,
	}
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	return New(Config{
		Level:       "info",
		Format:      "console",
		ColorOutput: true,
		TimeFormat:  time.RFC3339,
	})
}

// NewProduction creates a logger optimized for production
func NewProduction() *Logger {
	return New(Config{
		Level:       "info",
		Format:      "json",
		ColorOutput: false,
		TimeFormat:  time.RFC3339,
	})
}

// NewDevelopment creates a logger optimized for development
func NewDevelopment() *Logger {
	return New(Config{
		Level:       "debug",
		Format:      "console",
		ColorOutput: true,
		TimeFormat:  time.RFC3339,
	})
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	log.Logger = *logger.Logger
}

// WithComponent creates a logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	newLogger := l.Logger.With().Str("component", component).Logger()
	return &Logger{
		Logger: &newLogger,
		config: l.config,
	}
}

// WithService creates a logger with a service field
func (l *Logger) WithService(service string) *Logger {
	newLogger := l.Logger.With().Str("service", service).Logger()
	return &Logger{
		Logger: &newLogger,
		config: l.config,
	}
}

// WithRequestID creates a logger with a request ID field
func (l *Logger) WithRequestID(requestID string) *Logger {
	newLogger := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{
		Logger: &newLogger,
		config: l.config,
	}
}

// WithUserID creates a logger with a user ID field
func (l *Logger) WithUserID(userID string) *Logger {
	newLogger := l.Logger.With().Str("user_id", userID).Logger()
	return &Logger{
		Logger: &newLogger,
		config: l.config,
	}
}

// WithSessionID creates a logger with a session ID field
func (l *Logger) WithSessionID(sessionID string) *Logger {
	newLogger := l.Logger.With().Str("session_id", sessionID).Logger()
	return &Logger{
		Logger: &newLogger,
		config: l.config,
	}
}

// WithFields creates a logger with multiple fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	event := l.Logger.With()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	newLogger := event.Logger()
	return &Logger{
		Logger: &newLogger,
		config: l.config,
	}
}

// parseLogLevel converts string level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

// GetConfig returns the current logger configuration
func (l *Logger) GetConfig() Config {
	return l.config
}

// Clone creates a copy of the logger
func (l *Logger) Clone() *Logger {
	return &Logger{
		Logger: l.Logger,
		config: l.config,
	}
}

// HTTP creates a logger optimized for HTTP requests
func (l *Logger) HTTP() *Logger {
	return l.WithComponent("http")
}

// Database creates a logger optimized for database operations
func (l *Logger) Database() *Logger {
	return l.WithComponent("database")
}

// WhatsApp creates a logger optimized for WhatsApp operations
func (l *Logger) WhatsApp() *Logger {
	return l.WithComponent("whatsapp")
}

// Service creates a logger optimized for service operations
func (l *Logger) Service() *Logger {
	return l.WithComponent("service")
}

// Repository creates a logger optimized for repository operations
func (l *Logger) Repository() *Logger {
	return l.WithComponent("repository")
}

// NewFromAppConfig creates a logger from app configuration
func NewFromAppConfig(appConfig *config.Config) *Logger {
	loggerConfig := Config{
		Level:       appConfig.Logging.Level,
		Format:      appConfig.Logging.Format,
		ColorOutput: appConfig.Logging.ColorOutput,
		TimeFormat:  appConfig.Logging.TimeFormat,
	}

	return New(loggerConfig)
}
