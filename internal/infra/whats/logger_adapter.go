package whats

import (
	"fmt"

	waLog "go.mau.fi/whatsmeow/util/log"
	"wazmeow/pkg/logger"
)

// LoggerAdapter adapts our logger to whatsmeow's logger interface
type LoggerAdapter struct {
	logger logger.Logger
	module string
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(log logger.Logger, module string) waLog.Logger {
	return &LoggerAdapter{
		logger: log,
		module: module,
	}
}

// Debugf implements waLog.Logger
func (l *LoggerAdapter) Debugf(msg string, args ...interface{}) {
	l.logger.DebugWithFields(fmt.Sprintf(msg, args...), logger.Fields{
		"module": l.module,
	})
}

// Infof implements waLog.Logger
func (l *LoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.InfoWithFields(fmt.Sprintf(msg, args...), logger.Fields{
		"module": l.module,
	})
}

// Warnf implements waLog.Logger
func (l *LoggerAdapter) Warnf(msg string, args ...interface{}) {
	l.logger.WarnWithFields(fmt.Sprintf(msg, args...), logger.Fields{
		"module": l.module,
	})
}

// Errorf implements waLog.Logger
func (l *LoggerAdapter) Errorf(msg string, args ...interface{}) {
	l.logger.ErrorWithFields(fmt.Sprintf(msg, args...), logger.Fields{
		"module": l.module,
	})
}

// Sub implements waLog.Logger
func (l *LoggerAdapter) Sub(module string) waLog.Logger {
	return NewLoggerAdapter(l.logger, fmt.Sprintf("%s/%s", l.module, module))
}
