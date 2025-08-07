package logger

import (
	"time"
	"wazmeow/internal/app/config"

	"github.com/rs/zerolog/log"
)

// ExampleUsage demonstrates how to use the logger in different parts of the application
func ExampleUsage() {
	// 1. Using with app configuration
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:       "debug",
			Format:      "console",
			ColorOutput: true,
			TimeFormat:  "2006-01-02 15:04:05",
		},
	}

	// Create logger from app config
	appLogger := NewFromAppConfig(cfg)
	SetGlobalLogger(appLogger)

	// 2. Basic logging
	log.Info().Msg("Application started")
	log.Debug().Str("version", "1.0.0").Msg("Debug information")
	log.Error().Err(nil).Msg("Something went wrong")

	// 3. Component-specific loggers
	httpLogger := appLogger.HTTP()
	httpLogger.Info().
		Str("method", "GET").
		Str("path", "/api/v1/sessions").
		Int("status", 200).
		Msg("HTTP request processed")

	dbLogger := appLogger.Database()
	dbLogger.Debug().
		Str("query", "SELECT * FROM sessions").
		Dur("duration", 0).
		Msg("Database query executed")

	whatsappLogger := appLogger.WhatsApp()
	whatsappLogger.Info().
		Str("session_id", "session-123").
		Str("status", "connected").
		Msg("WhatsApp session status changed")

	// 4. Contextual loggers
	sessionLogger := appLogger.WithSessionID("session-123")
	sessionLogger.Info().Msg("Processing session operation")

	requestLogger := appLogger.WithRequestID("req-456")
	requestLogger.Info().Msg("Processing HTTP request")

	// 5. Multiple fields
	fieldsLogger := appLogger.WithFields(map[string]interface{}{
		"user_id":    "user-789",
		"session_id": "session-123",
		"operation":  "send_message",
	})
	fieldsLogger.Info().Msg("User operation completed")

	// 6. Different log levels
	log.Trace().Msg("Very detailed debug information")
	log.Debug().Msg("Debug information")
	log.Info().Msg("General information")
	log.Warn().Msg("Warning message")
	log.Error().Msg("Error occurred")

	// 7. Structured logging with events
	log.Info().
		Str("event", "session_created").
		Str("session_id", "session-123").
		Str("user_id", "user-789").
		Time("timestamp", time.Now()).
		Msg("Session created successfully")
}
