package logger

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level represents the logging level
type Level int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production
	DebugLevel Level = iota
	// InfoLevel is the default logging priority
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual human review
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs
	ErrorLevel
	// FatalLevel logs a message, then calls os.Exit(1)
	FatalLevel
)

// String returns the string representation of the Level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// Fields represents a map of key-value pairs for structured logging
type Fields map[string]interface{}

// Logger defines the interface for structured logging
type Logger interface {
	// Logging methods with message only
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)

	// Logging methods with message and fields
	DebugWithFields(msg string, fields Fields)
	InfoWithFields(msg string, fields Fields)
	WarnWithFields(msg string, fields Fields)
	ErrorWithFields(msg string, fields Fields)
	FatalWithFields(msg string, fields Fields)

	// Logging methods with message, error and fields
	DebugWithError(msg string, err error, fields Fields)
	InfoWithError(msg string, err error, fields Fields)
	WarnWithError(msg string, err error, fields Fields)
	ErrorWithError(msg string, err error, fields Fields)
	FatalWithError(msg string, err error, fields Fields)

	// Context-aware logging
	WithContext(ctx context.Context) Logger
	WithFields(fields Fields) Logger
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger

	// Configuration
	SetLevel(level Level)
	GetLevel() Level
	SetOutput(output io.Writer)

	// Utility methods
	IsDebugEnabled() bool
	IsInfoEnabled() bool
	IsWarnEnabled() bool
	IsErrorEnabled() bool
}

// Config represents logger configuration
type Config struct {
	Level         string `json:"level" yaml:"level" env:"LOG_LEVEL" default:"info"`
	Output        string `json:"output" yaml:"output" env:"LOG_OUTPUT" default:"console"`                         // "console", "file", or "dual"
	ConsoleFormat string `json:"console_format" yaml:"console_format" env:"LOG_CONSOLE_FORMAT" default:"console"` // "console" (colorido) or "json" (estruturado)
	FileFormat    string `json:"file_format" yaml:"file_format" env:"LOG_FILE_FORMAT" default:"json"`             // "console" or "json" (recomendado: json)
	TimeFormat    string `json:"time_format" yaml:"time_format" env:"LOG_TIME_FORMAT" default:"2006-01-02T15:04:05.000Z07:00"`
	Caller        bool   `json:"caller" yaml:"caller" env:"LOG_CALLER" default:"false"`
	StackTrace    bool   `json:"stack_trace" yaml:"stack_trace" env:"LOG_STACK_TRACE" default:"false"`
	FilePath      string `json:"file_path" yaml:"file_path" env:"LOG_FILE_PATH" default:"./logs/wazmeow.log"` // Path for file logging
	MaxSize       int    `json:"max_size" yaml:"max_size" env:"LOG_MAX_SIZE" default:"100"`                   // Maximum size in MB before rotation
	MaxBackups    int    `json:"max_backups" yaml:"max_backups" env:"LOG_MAX_BACKUPS" default:"3"`            // Maximum number of backup files
	MaxAge        int    `json:"max_age" yaml:"max_age" env:"LOG_MAX_AGE" default:"28"`                       // Maximum age in days

}

// ParseLevel parses a string level into a Level
func ParseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return InfoLevel, &InvalidLevelError{Level: level}
	}
}

// InvalidLevelError represents an error for invalid log levels
type InvalidLevelError struct {
	Level string
}

// Error implements the error interface
func (e *InvalidLevelError) Error() string {
	return "invalid log level: " + e.Level
}

// ContextKey represents a key for context values
type ContextKey string

const (
	// ContextKeyRequestID is the context key for request ID
	ContextKeyRequestID ContextKey = "request_id"
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeySessionID is the context key for session ID
	ContextKeySessionID ContextKey = "session_id"
	// ContextKeyCorrelationID is the context key for correlation ID
	ContextKeyCorrelationID ContextKey = "correlation_id"
)

// Entry represents a log entry
type Entry struct {
	Level     Level                  `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Caller    string                 `json:"caller,omitempty"`
}

// Hook defines the interface for log hooks
type Hook interface {
	// Levels returns the log levels that this hook should be triggered for
	Levels() []Level
	// Fire is called when a log entry is written
	Fire(entry *Entry) error
}

// HookLogger extends Logger with hook support
type HookLogger interface {
	Logger
	AddHook(hook Hook)
	RemoveHook(hook Hook)
}

// Factory defines the interface for creating loggers
type Factory interface {
	// Create creates a new logger with the given configuration
	Create(config *Config) (Logger, error)
	// CreateWithName creates a new logger with the given name and configuration
	CreateWithName(name string, config *Config) (Logger, error)
}

// Noop logger that does nothing (useful for testing)
type NoopLogger struct{}

func (n *NoopLogger) Debug(msg string)                                    {}
func (n *NoopLogger) Info(msg string)                                     {}
func (n *NoopLogger) Warn(msg string)                                     {}
func (n *NoopLogger) Error(msg string)                                    {}
func (n *NoopLogger) Fatal(msg string)                                    {}
func (n *NoopLogger) DebugWithFields(msg string, fields Fields)           {}
func (n *NoopLogger) InfoWithFields(msg string, fields Fields)            {}
func (n *NoopLogger) WarnWithFields(msg string, fields Fields)            {}
func (n *NoopLogger) ErrorWithFields(msg string, fields Fields)           {}
func (n *NoopLogger) FatalWithFields(msg string, fields Fields)           {}
func (n *NoopLogger) DebugWithError(msg string, err error, fields Fields) {}
func (n *NoopLogger) InfoWithError(msg string, err error, fields Fields)  {}
func (n *NoopLogger) WarnWithError(msg string, err error, fields Fields)  {}
func (n *NoopLogger) ErrorWithError(msg string, err error, fields Fields) {}
func (n *NoopLogger) FatalWithError(msg string, err error, fields Fields) {}
func (n *NoopLogger) WithContext(ctx context.Context) Logger              { return n }
func (n *NoopLogger) WithFields(fields Fields) Logger                     { return n }
func (n *NoopLogger) WithField(key string, value interface{}) Logger      { return n }
func (n *NoopLogger) WithError(err error) Logger                          { return n }
func (n *NoopLogger) SetLevel(level Level)                                {}
func (n *NoopLogger) GetLevel() Level                                     { return InfoLevel }
func (n *NoopLogger) SetOutput(output io.Writer)                          {}
func (n *NoopLogger) IsDebugEnabled() bool                                { return false }
func (n *NoopLogger) IsInfoEnabled() bool                                 { return false }
func (n *NoopLogger) IsWarnEnabled() bool                                 { return false }
func (n *NoopLogger) IsErrorEnabled() bool                                { return false }

// ZerologLogger implements Logger using zerolog
type ZerologLogger struct {
	logger zerolog.Logger
	level  Level
}

// New creates a new logger with the given configuration
func New(config *Config) Logger {
	// Parse level
	level, err := ParseLevel(config.Level)
	if err != nil {
		level = InfoLevel
	}

	// Set global level
	zerolog.SetGlobalLevel(parseZerologLevel(level))

	// Configure outputs based on config.Output
	var writers []io.Writer

	switch config.Output {
	case "console":
		// Only console output
		if config.ConsoleFormat == "console" {
			// Colorized console output
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		} else {
			// JSON to console
			writers = append(writers, os.Stdout)
		}

	case "file":
		// Only file output
		fileWriter := createFileWriter(config)
		if config.FileFormat == "console" {
			// Console format to file
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        fileWriter,
				TimeFormat: time.RFC3339,
				NoColor:    true, // No colors in file
			})
		} else {
			// JSON to file
			writers = append(writers, fileWriter)
		}

	case "dual":
		// Both console and file output

		// Console output
		if config.ConsoleFormat == "console" {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		} else {
			writers = append(writers, os.Stdout)
		}

		// File output
		fileWriter := createFileWriter(config)
		if config.FileFormat == "console" {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        fileWriter,
				TimeFormat: time.RFC3339,
				NoColor:    true, // No colors in file
			})
		} else {
			writers = append(writers, fileWriter)
		}

	default:
		// Default to console
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	// Create multi-writer if multiple outputs
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = zerolog.MultiLevelWriter(writers...)
	}

	// Create logger
	logger := zerolog.New(output).With().Timestamp().Logger()

	if config.Caller {
		logger = logger.With().Caller().Logger()
	}

	return &ZerologLogger{
		logger: logger,
		level:  level,
	}
}

// createFileWriter creates a file writer with rotation support
func createFileWriter(config *Config) io.Writer {
	// Ensure directory exists
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// If we can't create the directory, fall back to stdout
		return os.Stdout
	}

	return &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize,    // megabytes
		MaxBackups: config.MaxBackups, // number of backups
		MaxAge:     config.MaxAge,     // days
		Compress:   true,              // compress rotated files
	}
}

// parseZerologLevel converts our Level to zerolog.Level
func parseZerologLevel(level Level) zerolog.Level {
	switch level {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Implement Logger interface methods

func (z *ZerologLogger) Debug(msg string) {
	z.logger.Debug().Msg(msg)
}

func (z *ZerologLogger) Info(msg string) {
	z.logger.Info().Msg(msg)
}

func (z *ZerologLogger) Warn(msg string) {
	z.logger.Warn().Msg(msg)
}

func (z *ZerologLogger) Error(msg string) {
	z.logger.Error().Msg(msg)
}

func (z *ZerologLogger) Fatal(msg string) {
	z.logger.Fatal().Msg(msg)
}

func (z *ZerologLogger) DebugWithFields(msg string, fields Fields) {
	event := z.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) InfoWithFields(msg string, fields Fields) {
	event := z.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) WarnWithFields(msg string, fields Fields) {
	event := z.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) ErrorWithFields(msg string, fields Fields) {
	event := z.logger.Error()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) FatalWithFields(msg string, fields Fields) {
	event := z.logger.Fatal()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) DebugWithError(msg string, err error, fields Fields) {
	event := z.logger.Debug().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) InfoWithError(msg string, err error, fields Fields) {
	event := z.logger.Info().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) WarnWithError(msg string, err error, fields Fields) {
	event := z.logger.Warn().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) ErrorWithError(msg string, err error, fields Fields) {
	event := z.logger.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) FatalWithError(msg string, err error, fields Fields) {
	event := z.logger.Fatal().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (z *ZerologLogger) WithContext(ctx context.Context) Logger {
	newLogger := z.logger.With().Logger()

	// Extract common context values
	if requestID := ctx.Value(ContextKeyRequestID); requestID != nil {
		newLogger = newLogger.With().Interface("request_id", requestID).Logger()
	}
	if userID := ctx.Value(ContextKeyUserID); userID != nil {
		newLogger = newLogger.With().Interface("user_id", userID).Logger()
	}
	if sessionID := ctx.Value(ContextKeySessionID); sessionID != nil {
		newLogger = newLogger.With().Interface("session_id", sessionID).Logger()
	}
	if correlationID := ctx.Value(ContextKeyCorrelationID); correlationID != nil {
		newLogger = newLogger.With().Interface("correlation_id", correlationID).Logger()
	}

	return &ZerologLogger{
		logger: newLogger,
		level:  z.level,
	}
}

func (z *ZerologLogger) WithFields(fields Fields) Logger {
	event := z.logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}

	return &ZerologLogger{
		logger: event.Logger(),
		level:  z.level,
	}
}

func (z *ZerologLogger) WithField(key string, value interface{}) Logger {
	return &ZerologLogger{
		logger: z.logger.With().Interface(key, value).Logger(),
		level:  z.level,
	}
}

func (z *ZerologLogger) WithError(err error) Logger {
	return &ZerologLogger{
		logger: z.logger.With().Err(err).Logger(),
		level:  z.level,
	}
}

func (z *ZerologLogger) SetLevel(level Level) {
	z.level = level
	zerolog.SetGlobalLevel(parseZerologLevel(level))
}

func (z *ZerologLogger) GetLevel() Level {
	return z.level
}

func (z *ZerologLogger) SetOutput(output io.Writer) {
	z.logger = z.logger.Output(output)
}

func (z *ZerologLogger) IsDebugEnabled() bool {
	return z.level <= DebugLevel
}

func (z *ZerologLogger) IsInfoEnabled() bool {
	return z.level <= InfoLevel
}

func (z *ZerologLogger) IsWarnEnabled() bool {
	return z.level <= WarnLevel
}

func (z *ZerologLogger) IsErrorEnabled() bool {
	return z.level <= ErrorLevel
}
