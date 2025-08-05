package usecases_session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	sessionUC "wazmeow/internal/usecases/session"
)

// MockWhatsAppManager is a mock implementation of whatsapp.Manager
type MockWhatsAppManager struct {
	mock.Mock
}

func (m *MockWhatsAppManager) CreateClient(sessionID session.SessionID) (whatsapp.Client, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(whatsapp.Client), args.Error(1)
}

func (m *MockWhatsAppManager) GetClient(sessionID session.SessionID) (whatsapp.Client, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(whatsapp.Client), args.Error(1)
}

func (m *MockWhatsAppManager) RemoveClient(sessionID session.SessionID) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockWhatsAppManager) ListClients() []session.SessionID {
	args := m.Called()
	return args.Get(0).([]session.SessionID)
}

func (m *MockWhatsAppManager) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWhatsAppManager) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWhatsAppManager) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhatsAppManager) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

// MockWhatsAppClient is a mock implementation of whatsapp.Client
type MockWhatsAppClient struct {
	mock.Mock
}

func (m *MockWhatsAppClient) Connect(ctx context.Context) (*whatsapp.ConnectionResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*whatsapp.ConnectionResult), args.Error(1)
}

func (m *MockWhatsAppClient) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWhatsAppClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhatsAppClient) GetConnectionStatus() whatsapp.ConnectionStatus {
	args := m.Called()
	return args.Get(0).(whatsapp.ConnectionStatus)
}

func (m *MockWhatsAppClient) GenerateQR(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockWhatsAppClient) PairPhone(ctx context.Context, phoneNumber string) error {
	args := m.Called(ctx, phoneNumber)
	return args.Error(0)
}

func (m *MockWhatsAppClient) IsAuthenticated() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhatsAppClient) GetSessionID() session.SessionID {
	args := m.Called()
	return args.Get(0).(session.SessionID)
}

func (m *MockWhatsAppClient) GetJID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockWhatsAppClient) GetDeviceInfo() *whatsapp.DeviceInfo {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*whatsapp.DeviceInfo)
}

func (m *MockWhatsAppClient) SendMessage(ctx context.Context, to, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

func (m *MockWhatsAppClient) SendImage(ctx context.Context, to, imagePath, caption string) error {
	args := m.Called(ctx, to, imagePath, caption)
	return args.Error(0)
}

func (m *MockWhatsAppClient) SendDocument(ctx context.Context, to, documentPath, filename string) error {
	args := m.Called(ctx, to, documentPath, filename)
	return args.Error(0)
}

func (m *MockWhatsAppClient) SetEventHandler(handler whatsapp.EventHandler) {
	m.Called(handler)
}

func (m *MockWhatsAppClient) RemoveEventHandler() {
	m.Called()
}

func (m *MockWhatsAppClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestConnectUseCase(t *testing.T) {
	t.Run("should connect session successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockSessionRepository)
		mockWAManager := new(MockWhatsAppManager)
		mockLogger := new(MockLogger)
		mockClient := new(MockWhatsAppClient)

		useCase := sessionUC.NewConnectUseCase(mockRepo, mockWAManager, mockLogger)

		sess := session.NewSession("test-session")
		ctx := context.Background()

		req := sessionUC.ConnectRequest{
			SessionID: sess.ID(),
		}

		// Mock expectations
		mockRepo.On("GetByID", ctx, sess.ID()).Return(sess, nil)
		mockWAManager.On("GetClient", sess.ID()).Return(nil, whatsapp.ErrClientNotFound)
		mockWAManager.On("CreateClient", sess.ID()).Return(mockClient, nil)
		mockClient.On("Connect", ctx).Return(&whatsapp.ConnectionResult{
			JID:    "test@s.whatsapp.net",
			Status: whatsapp.StatusConnected,
		}, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*session.Session")).Return(nil)
		mockLogger.On("InfoWithFields", mock.AnythingOfType("string"), mock.AnythingOfType("logger.Fields")).Return()

		// Act
		result, err := useCase.Execute(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Session)
		assert.Equal(t, session.StatusConnected, result.Session.Status())

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockWAManager.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
