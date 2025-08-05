package usecases_session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
	sessionUC "wazmeow/internal/usecases/session"
)

func TestResolveUseCase(t *testing.T) {
	t.Run("should resolve session by ID successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		sess := session.NewSession("test-session")
		identifier, err := session.NewSessionIdentifier(sess.ID().String())
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, "id", result.IdentifierType)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should resolve session by name successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		sess := session.NewSession("test-session")
		identifier, err := session.NewSessionIdentifier("test-session")
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByName", ctx, "test-session").Return(sess, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, "name", result.IdentifierType)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail with invalid identifier", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		// Create empty identifier (invalid)
		identifier := session.SessionIdentifier{}

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()

		// Mock expectations
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify mocks
		mockLogger.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "GetByID")
		mockRepo.AssertNotCalled(t, "GetByName")
	})

	t.Run("should fail when session not found by ID", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		sessionID := session.NewSessionID()
		identifier, err := session.NewSessionIdentifier(sessionID.String())
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sessionID).Return(nil, session.ErrSessionNotFound)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()
		mockLogger.On("WarnWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when session not found by name", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		identifier, err := session.NewSessionIdentifier("non-existent-session")
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByName", ctx, "non-existent-session").Return(nil, session.ErrSessionNotFound)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()
		mockLogger.On("WarnWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when repository returns error for ID lookup", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		sessionID := session.NewSessionID()
		identifier, err := session.NewSessionIdentifier(sessionID.String())
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()
		repoErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByID", ctx, sessionID).Return(nil, repoErr)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), repoErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve session by ID")
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when repository returns error for name lookup", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		identifier, err := session.NewSessionIdentifier("test-session")
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()
		repoErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByName", ctx, "test-session").Return(nil, repoErr)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), repoErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve session by name")
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should log resolution attempt", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

		sess := session.NewSession("test-session")
		identifier, err := session.NewSessionIdentifier("test-session")
		require.NoError(t, err)

		req := sessionUC.ResolveRequest{
			Identifier: identifier,
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByName", ctx, "test-session").Return(sess, nil)
		mockLogger.On("InfoWithFields", "resolving session", mock.MatchedBy(func(fields map[string]interface{}) bool {
			return fields["identifier"] == "test-session" && fields["identifier_type"] == "name"
		})).Return()
		mockLogger.On("InfoWithFields", "session resolved successfully", mock.MatchedBy(func(fields map[string]interface{}) bool {
			return fields["session_id"] == sess.ID().String() &&
				fields["session_name"] == "test-session" &&
				fields["identifier"] == "test-session" &&
				fields["identifier_type"] == "name"
		})).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle different identifier types correctly", func(t *testing.T) {
		testCases := []struct {
			name          string
			identifierStr string
			expectedType  string
			setupMock     func(*MockSessionRepository, *session.Session)
		}{
			{
				name:          "UUID identifier",
				identifierStr: "550e8400-e29b-41d4-a716-446655440000",
				expectedType:  "id",
				setupMock: func(repo *MockSessionRepository, sess *session.Session) {
					sessionID, _ := session.SessionIDFromString("550e8400-e29b-41d4-a716-446655440000")
					repo.On("GetByID", mock.Anything, sessionID).Return(sess, nil)
				},
			},
			{
				name:          "Name identifier",
				identifierStr: "my-session-name",
				expectedType:  "name",
				setupMock: func(repo *MockSessionRepository, sess *session.Session) {
					repo.On("GetByName", mock.Anything, "my-session-name").Return(sess, nil)
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				mockRepo := new(MockSessionRepository)
				mockLogger := new(MockLogger)

				useCase := sessionUC.NewResolveUseCase(mockRepo, mockLogger)

				sess := session.NewSession("test-session")
				identifier, err := session.NewSessionIdentifier(tc.identifierStr)
				require.NoError(t, err)

				req := sessionUC.ResolveRequest{
					Identifier: identifier,
				}

				ctx := context.Background()

				// Setup mock
				tc.setupMock(mockRepo, sess)
				mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

				// Act
				result, err := useCase.Execute(ctx, req)

				// Assert
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, sess, result.Session)
				assert.Equal(t, tc.expectedType, result.IdentifierType)

				// Verify mocks
				mockRepo.AssertExpectations(t)
				mockLogger.AssertExpectations(t)
			})
		}
	})
}
