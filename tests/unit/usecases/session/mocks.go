package usecases_session

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"

	"wazmeow/internal/domain/session"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// MockSessionRepository is a mock implementation of session.Repository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, sess *session.Session) error {
	args := m.Called(ctx, sess)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id session.SessionID) (*session.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByName(ctx context.Context, name string) (*session.Session, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, sess *session.Session) error {
	args := m.Called(ctx, sess)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id session.SessionID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) List(ctx context.Context, limit, offset int) ([]*session.Session, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*session.Session), args.Int(1), args.Error(2)
}

func (m *MockSessionRepository) GetByStatus(ctx context.Context, status session.Status, limit, offset int) ([]*session.Session, int, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*session.Session), args.Int(1), args.Error(2)
}

func (m *MockSessionRepository) GetActiveCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockSessionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockSessionRepository) Exists(ctx context.Context, id session.SessionID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockSessionRepository) UpdateStatus(ctx context.Context, id session.SessionID, status session.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockLogger is a mock implementation of logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Fatal(msg string) {
	m.Called(msg)
}

func (m *MockLogger) DebugWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockLogger) InfoWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockLogger) WarnWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockLogger) ErrorWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockLogger) FatalWithFields(msg string, fields logger.Fields) {
	m.Called(msg, fields)
}

func (m *MockLogger) DebugWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) InfoWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) WarnWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) ErrorWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) FatalWithError(msg string, err error, fields logger.Fields) {
	m.Called(msg, err, fields)
}

func (m *MockLogger) WithContext(ctx context.Context) logger.Logger {
	return m
}

func (m *MockLogger) WithFields(fields logger.Fields) logger.Logger {
	return m
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	return m
}

func (m *MockLogger) WithError(err error) logger.Logger {
	return m
}

func (m *MockLogger) SetLevel(level logger.Level) {
	m.Called(level)
}

func (m *MockLogger) GetLevel() logger.Level {
	args := m.Called()
	return args.Get(0).(logger.Level)
}

func (m *MockLogger) SetOutput(output io.Writer) {
	m.Called(output)
}

func (m *MockLogger) IsDebugEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLogger) IsInfoEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLogger) IsWarnEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLogger) IsErrorEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockValidator is a mock implementation of validator.Validator
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(s interface{}) error {
	args := m.Called(s)
	return args.Error(0)
}

func (m *MockValidator) ValidateField(field interface{}, tag string) error {
	args := m.Called(field, tag)
	return args.Error(0)
}

func (m *MockValidator) RegisterValidation(tag string, fn validator.ValidationFunc) error {
	args := m.Called(tag, fn)
	return args.Error(0)
}
