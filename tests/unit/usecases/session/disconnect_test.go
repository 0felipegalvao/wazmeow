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

func TestDisconnectUseCase(t *testing.T) {
	t.Run("should disconnect session successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DisconnectRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*session.Session")).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, session.StatusDisconnected, result.Session.Status())
		assert.False(t, result.Session.IsActive())
		assert.NotEmpty(t, result.Message)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should disconnect session when client doesn't exist", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DisconnectRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(nil, assert.AnError) // Client doesn't exist
		mockRepo.On("Update", ctx, mock.AnythingOfType("*session.Session")).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()
		mockLogger.On("WarnWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, session.StatusDisconnected, result.Session.Status())
		assert.False(t, result.Session.IsActive())

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should disconnect already disconnected session", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a disconnected session
		sess := session.NewSession("test-session")

		req := sessionUC.DisconnectRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()

		// Mock expectations - session already disconnected, returns early
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, session.StatusDisconnected, result.Session.Status())
		assert.False(t, result.Session.IsActive())

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when session not found", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		sessionID := session.NewSessionID()
		req := sessionUC.DisconnectRequest{
			SessionID: sessionID,
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sessionID).Return(nil, session.ErrSessionNotFound)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), session.ErrSessionNotFound, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		mockWAManager.AssertNotCalled(t, "GetClient")
	})

	t.Run("should fail when repository update fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DisconnectRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()
		updateErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*session.Session")).Return(updateErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), updateErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, updateErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle WhatsApp client disconnect error gracefully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DisconnectRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()
		clientErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(clientErr)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*session.Session")).Return(nil)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), mock.AnythingOfType("*errors.errorString"), mock.AnythingOfType("logger.Fields")).Return()
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err) // Should not fail even if client disconnect fails
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, session.StatusDisconnected, result.Session.Status())
		assert.False(t, result.Session.IsActive())

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should remove client after successful disconnect", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDisconnectUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DisconnectRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*session.Session")).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess, result.Session)
		assert.Equal(t, session.StatusDisconnected, result.Session.Status())
		assert.False(t, result.Session.IsActive())

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
