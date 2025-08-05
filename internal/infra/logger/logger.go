package logger

import (
	"wazmeow/internal/infra/config"
	"wazmeow/pkg/logger"
)

// New creates a new logger instance based on configuration
func New(cfg *config.LogConfig) logger.Logger {
	loggerConfig := &logger.Config{
		Level:         cfg.Level,
		Output:        cfg.Output,
		ConsoleFormat: cfg.ConsoleFormat,
		FileFormat:    cfg.FileFormat,
		TimeFormat:    cfg.TimeFormat,
		Caller:        cfg.Caller,
		StackTrace:    cfg.StackTrace,
		FilePath:      cfg.FilePath,
		MaxSize:       cfg.MaxSize,
		MaxBackups:    cfg.MaxBackups,
		MaxAge:        cfg.MaxAge,
	}

	return logger.New(loggerConfig)
}

// NewDefault creates a logger with default configuration
func NewDefault() logger.Logger {
	return New(&config.LogConfig{
		Level:         "info",
		Output:        "console",
		ConsoleFormat: "console",
		FileFormat:    "json",
		TimeFormat:    "2006-01-02T15:04:05.000Z07:00",
		Caller:        false,
		StackTrace:    false,
		FilePath:      "./logs/wazmeow.log",
		MaxSize:       100,
		MaxBackups:    3,
		MaxAge:        28,
	})
}
