package usecases_session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"wazmeow/internal/domain/session"
	sessionUC "wazmeow/internal/usecases/session"
	"wazmeow/pkg/validator"
)

func TestCreateUseCase(t *testing.T) {
	t.Run("should create session successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)
		mockValidator := new(MockValidator)

		useCase := sessionUC.NewCreateUseCase(mockRepo, mockLogger, mockValidator)

		req := sessionUC.CreateRequest{
			Name: "test-session",
		}

		ctx := context.Background()

		// Mock expectations
		mockValidator.On("Validate", req).Return(nil)
		mockRepo.On("GetByName", ctx, "test-session").Return(nil, session.ErrSessionNotFound)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*session.Session")).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Session)
		assert.Equal(t, "test-session", result.Session.Name())
		assert.Equal(t, session.StatusDisconnected, result.Session.Status())
		assert.False(t, result.Session.IsActive())

		// Verify mocks
		mockValidator.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail with validation error", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)
		mockValidator := new(MockValidator)

		useCase := sessionUC.NewCreateUseCase(mockRepo, mockLogger, mockValidator)

		req := sessionUC.CreateRequest{
			Name: "", // Invalid empty name
		}

		ctx := context.Background()
		validationErr := validator.ValidationErrors{
			validator.ValidationError{
				Field:   "name",
				Tag:     "required",
				Value:   "",
				Message: "name is required",
			},
		}

		// Mock expectations
		mockValidator.On("Validate", req).Return(validationErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), validationErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, validationErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockValidator.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "GetByName")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("should fail when session already exists", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)
		mockValidator := new(MockValidator)

		useCase := sessionUC.NewCreateUseCase(mockRepo, mockLogger, mockValidator)

		req := sessionUC.CreateRequest{
			Name: "existing-session",
		}

		ctx := context.Background()
		existingSession := session.NewSession("existing-session")

		// Mock expectations
		mockValidator.On("Validate", req).Return(nil)
		mockRepo.On("GetByName", ctx, "existing-session").Return(existingSession, nil)
		mockLogger.On("WarnWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionAlreadyExists, err)
		assert.Nil(t, result)

		// Verify mocks
		mockValidator.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("should fail when repository GetByName returns error", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)
		mockValidator := new(MockValidator)

		useCase := sessionUC.NewCreateUseCase(mockRepo, mockLogger, mockValidator)

		req := sessionUC.CreateRequest{
			Name: "test-session",
		}

		ctx := context.Background()
		repoErr := assert.AnError

		// Mock expectations
		mockValidator.On("Validate", req).Return(nil)
		mockRepo.On("GetByName", ctx, "test-session").Return(nil, repoErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), mock.AnythingOfType("*errors.errorString"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockValidator.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("should fail when repository Create returns error", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)
		mockValidator := new(MockValidator)

		useCase := sessionUC.NewCreateUseCase(mockRepo, mockLogger, mockValidator)

		req := sessionUC.CreateRequest{
			Name: "test-session",
		}

		ctx := context.Background()
		createErr := assert.AnError

		// Mock expectations
		mockValidator.On("Validate", req).Return(nil)
		mockRepo.On("GetByName", ctx, "test-session").Return(nil, session.ErrSessionNotFound)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*session.Session")).Return(createErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), createErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, createErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockValidator.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
