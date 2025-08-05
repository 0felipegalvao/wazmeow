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

func TestDeleteUseCase(t *testing.T) {
	t.Run("should delete disconnected session successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a disconnected session
		sess := session.NewSession("test-session")

		req := sessionUC.DeleteRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		// No GetClient call for disconnected sessions
		mockRepo.On("Delete", ctx, sess.ID()).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess.ID(), result.SessionID)
		assert.NotEmpty(t, result.Message)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should delete connected session after disconnecting", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DeleteRequest{
			SessionID: sess.ID(),
			Force:     true, // Force delete connected session
		}

		ctx := context.Background()

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(nil)
		mockWAManager.On("RemoveClient", sess.ID()).Return(nil)
		mockRepo.On("Delete", ctx, sess.ID()).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess.ID(), result.SessionID)
		assert.NotEmpty(t, result.Message)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should fail when session not found", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		sessionID := session.NewSessionID()
		req := sessionUC.DeleteRequest{
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
		mockRepo.AssertNotCalled(t, "Delete")
	})

	t.Run("should fail when repository delete fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a disconnected session
		sess := session.NewSession("test-session")

		req := sessionUC.DeleteRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()
		deleteErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		// No GetClient call for disconnected sessions
		mockRepo.On("Delete", ctx, sess.ID()).Return(deleteErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), deleteErr, mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, deleteErr, err)
		assert.Nil(t, result)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle client disconnect error gracefully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DeleteRequest{
			SessionID: sess.ID(),
			Force:     true, // Force delete connected session
		}

		ctx := context.Background()
		disconnectErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(disconnectErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), disconnectErr, mock.AnythingOfType("logger.Fields")).Return()
		mockWAManager.On("RemoveClient", sess.ID()).Return(nil)
		mockRepo.On("Delete", ctx, sess.ID()).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err) // Should succeed despite disconnect error
		assert.NotNil(t, result)
		assert.Equal(t, sess.ID(), result.SessionID)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should handle client removal error gracefully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connected session
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		req := sessionUC.DeleteRequest{
			SessionID: sess.ID(),
			Force:     true, // Force delete connected session
		}

		ctx := context.Background()
		removeErr := assert.AnError

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Disconnect", ctx).Return(nil)
		mockWAManager.On("RemoveClient", sess.ID()).Return(removeErr)
		mockLogger.On("ErrorWithError", mock.AnythingOfType("string"), removeErr, mock.AnythingOfType("logger.Fields")).Return()
		mockRepo.On("Delete", ctx, sess.ID()).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err) // Should succeed despite remove error
		assert.NotNil(t, result)
		assert.Equal(t, sess.ID(), result.SessionID)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should delete session when connecting", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)

		useCase := sessionUC.NewDeleteUseCase(mockRepo, mockWAManager, mockLogger)

		// Create a connecting session
		sess := session.NewSession("test-session")
		sess.SetConnecting()

		req := sessionUC.DeleteRequest{
			SessionID: sess.ID(),
		}

		ctx := context.Background()

		// Mock expectations - connecting session doesn't need disconnection
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockRepo.On("Delete", ctx, sess.ID()).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sess.ID(), result.SessionID)

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
