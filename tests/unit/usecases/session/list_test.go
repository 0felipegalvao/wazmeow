package usecases_session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"wazmeow/internal/domain/session"
	sessionUC "wazmeow/internal/usecases/session"
)

func TestListUseCase(t *testing.T) {
	t.Run("should list sessions successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListRequest{
			Limit:  10,
			Offset: 0,
		}

		ctx := context.Background()

		// Create test sessions
		sessions := []*session.Session{
			session.NewSession("session-1"),
			session.NewSession("session-2"),
			session.NewSession("session-3"),
		}
		totalCount := 3

		// Mock expectations
		mockRepo.On("List", ctx, 10, 0).Return(sessions, totalCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sessions, result.Sessions)
		assert.Equal(t, totalCount, result.Total)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, 0, result.Offset)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should use default limit when not provided", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListRequest{
			Limit:  0, // Should default to 10
			Offset: 0,
		}

		ctx := context.Background()

		sessions := []*session.Session{}
		totalCount := 0

		// Mock expectations - should use default limit of 10
		mockRepo.On("List", ctx, 10, 0).Return(sessions, totalCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 10, result.Limit)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should cap limit at maximum", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListRequest{
			Limit:  200, // Should be capped at 100
			Offset: 0,
		}

		ctx := context.Background()

		sessions := []*session.Session{}
		totalCount := 0

		// Mock expectations - should use maximum limit of 100
		mockRepo.On("List", ctx, 100, 0).Return(sessions, totalCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 100, result.Limit)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle negative offset", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListRequest{
			Limit:  10,
			Offset: -5, // Should default to 0
		}

		ctx := context.Background()

		sessions := []*session.Session{}
		totalCount := 0

		// Mock expectations - should use offset 0
		mockRepo.On("List", ctx, 10, 0).Return(sessions, totalCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Offset)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when repository returns error", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListRequest{
			Limit:  10,
			Offset: 0,
		}

		ctx := context.Background()
		repoErr := assert.AnError

		// Mock expectations
		mockRepo.On("List", ctx, 10, 0).Return(nil, 0, repoErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), repoErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle empty result", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListRequest{
			Limit:  10,
			Offset: 0,
		}

		ctx := context.Background()

		sessions := []*session.Session{}
		totalCount := 0

		// Mock expectations
		mockRepo.On("List", ctx, 10, 0).Return(sessions, totalCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Sessions)
		assert.Equal(t, 0, result.Total)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListUseCaseByStatus(t *testing.T) {
	t.Run("should list sessions by status successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListByStatusRequest{
			Status: session.StatusConnected,
			Limit:  10,
			Offset: 0,
		}

		ctx := context.Background()

		// Create test sessions
		sessions := []*session.Session{
			session.NewSession("connected-session-1"),
			session.NewSession("connected-session-2"),
		}
		totalCount := 2

		// Mock expectations
		mockRepo.On("GetByStatus", ctx, session.StatusConnected, 10, 0).Return(sessions, totalCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.ExecuteByStatus(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sessions, result.Sessions)
		assert.Equal(t, totalCount, result.Total)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when repository returns error", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.ListByStatusRequest{
			Status: session.StatusConnected,
			Limit:  10,
			Offset: 0,
		}

		ctx := context.Background()
		repoErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByStatus", ctx, session.StatusConnected, 10, 0).Return(nil, 0, repoErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), repoErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.ExecuteByStatus(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestListUseCaseGetActiveCount(t *testing.T) {
	t.Run("should get active count successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.GetActiveCountRequest{}
		ctx := context.Background()
		activeCount := 5

		// Mock expectations
		mockRepo.On("GetActiveCount", ctx).Return(activeCount, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.ExecuteGetActiveCount(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, activeCount, result.Count)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when repository returns error", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewListUseCase(mockRepo, mockLogger)

		req := sessionUC.GetActiveCountRequest{}
		ctx := context.Background()
		repoErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetActiveCount", ctx).Return(0, repoErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), repoErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.ExecuteGetActiveCount(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
