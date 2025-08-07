package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// ClientManager manages WhatsApp clients for multiple sessions
type ClientManager struct {
	storeManager *StoreManager
	clients      map[session.SessionID]*whatsmeow.Client
	devices      map[session.SessionID]*store.Device
	mutex        sync.RWMutex
	logger       waLog.Logger
}

// NewClientManager creates a new WhatsApp client manager
func NewClientManager(storeManager *StoreManager, logger waLog.Logger) *ClientManager {
	return &ClientManager{
		storeManager: storeManager,
		clients:      make(map[session.SessionID]*whatsmeow.Client),
		devices:      make(map[session.SessionID]*store.Device),
		logger:       logger,
	}
}

// CreateClient creates a new WhatsApp client for a session
func (cm *ClientManager) CreateClient(ctx context.Context, sessionID session.SessionID) (whatsapp.Client, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if client already exists
	if client, exists := cm.clients[sessionID]; exists {
		return client, nil
	}

	// Create new device
	device := cm.storeManager.GetContainer().NewDevice()

	// Create client
	client := whatsmeow.NewClient(device, cm.logger)

	// Set up event handler
	client.AddEventHandler(cm.createEventHandler(sessionID))

	// Store client and device
	cm.clients[sessionID] = client
	cm.devices[sessionID] = device

	log.Info().
		Str("session_id", sessionID.String()).
		Msg("WhatsApp client created")

	return client, nil
}

// GetClient returns an existing WhatsApp client for a session
func (cm *ClientManager) GetClient(ctx context.Context, sessionID session.SessionID) (*whatsmeow.Client, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	return client, nil
}

// ConnectClient connects a WhatsApp client
func (cm *ClientManager) ConnectClient(ctx context.Context, sessionID session.SessionID) error {
	client, err := cm.GetClient(ctx, sessionID)
	if err != nil {
		return err
	}

	if client.IsConnected() {
		return fmt.Errorf("client already connected for session %s", sessionID)
	}

	err = client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect client: %w", err)
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Msg("WhatsApp client connected")

	return nil
}

// DisconnectClient disconnects a WhatsApp client
func (cm *ClientManager) DisconnectClient(ctx context.Context, sessionID session.SessionID) error {
	client, err := cm.GetClient(ctx, sessionID)
	if err != nil {
		return err
	}

	client.Disconnect()

	log.Info().
		Str("session_id", sessionID.String()).
		Msg("WhatsApp client disconnected")

	return nil
}

// LogoutClient logs out a WhatsApp client
func (cm *ClientManager) LogoutClient(ctx context.Context, sessionID session.SessionID) error {
	client, err := cm.GetClient(ctx, sessionID)
	if err != nil {
		return err
	}

	err = client.Logout(ctx)
	if err != nil {
		return fmt.Errorf("failed to logout client: %w", err)
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Msg("WhatsApp client logged out")

	return nil
}

// RemoveClient removes a WhatsApp client
func (cm *ClientManager) RemoveClient(ctx context.Context, sessionID session.SessionID) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	client, exists := cm.clients[sessionID]
	if exists {
		if client.IsConnected() {
			client.Disconnect()
		}
		delete(cm.clients, sessionID)
	}

	device, exists := cm.devices[sessionID]
	if exists {
		// Delete device from store
		if err := cm.storeManager.GetContainer().DeleteDevice(ctx, device); err != nil {
			log.Warn().Err(err).
				Str("session_id", sessionID.String()).
				Msg("Failed to delete device from store")
		}
		delete(cm.devices, sessionID)
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Msg("WhatsApp client removed")

	return nil
}

// GetQRCode returns the QR code for pairing
func (cm *ClientManager) GetQRCode(ctx context.Context, sessionID session.SessionID) (string, error) {
	client, err := cm.GetClient(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if client.Store.ID != nil {
		return "", fmt.Errorf("client already paired")
	}

	// Generate QR code
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get QR channel: %w", err)
	}

	err = client.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect for QR: %w", err)
	}

	// Wait for QR code
	select {
	case evt := <-qrChan:
		switch evt.Event {
		case "code":
			return evt.Code, nil
		case "timeout":
			return "", fmt.Errorf("QR code timeout")
		case "success":
			return "", fmt.Errorf("already paired")
		default:
			return "", fmt.Errorf("unexpected QR event: %s", evt.Event)
		}
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("QR code timeout")
	}
}

// IsClientConnected checks if a client is connected
func (cm *ClientManager) IsClientConnected(sessionID session.SessionID) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return false
	}

	return client.IsConnected()
}

// createEventHandler creates an event handler for a session
func (cm *ClientManager) createEventHandler(sessionID session.SessionID) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			log.Info().
				Str("session_id", sessionID.String()).
				Str("from", v.Info.Sender.String()).
				Str("message_id", v.Info.ID).
				Msg("Received message")

		case *events.Connected:
			log.Info().
				Str("session_id", sessionID.String()).
				Msg("WhatsApp connected")

		case *events.Disconnected:
			log.Info().
				Str("session_id", sessionID.String()).
				Msg("WhatsApp disconnected")

		case *events.LoggedOut:
			log.Info().
				Str("session_id", sessionID.String()).
				Str("reason", string(v.Reason)).
				Msg("WhatsApp logged out")

		case *events.QR:
			log.Info().
				Str("session_id", sessionID.String()).
				Str("codes", fmt.Sprintf("%v", v.Codes)).
				Msg("QR code received")

		case *events.PairSuccess:
			log.Info().
				Str("session_id", sessionID.String()).
				Str("jid", v.ID.String()).
				Msg("WhatsApp paired successfully")
		}
	}
}

// GetAllClients returns all active clients
func (cm *ClientManager) GetAllClients(ctx context.Context) map[session.SessionID]whatsapp.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[session.SessionID]whatsapp.Client)
	for sessionID, client := range cm.clients {
		result[sessionID] = NewClientWrapper(client, sessionID)
	}
	return result
}

// SetEventHandler sets the event handler for all clients
func (cm *ClientManager) SetEventHandler(handler whatsapp.EventHandler) {
	// TODO: Implement event handler setting
	log.Info().Msg("SetEventHandler called - implementation pending")
}

// ConnectAll connects all sessions marked as active
func (cm *ClientManager) ConnectAll(ctx context.Context) error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for sessionID, client := range cm.clients {
		if !client.IsConnected() {
			if err := client.Connect(); err != nil {
				log.Error().Err(err).
					Str("session_id", sessionID.String()).
					Msg("Failed to connect client in ConnectAll")
				continue
			}
		}
	}
	return nil
}

// DisconnectAll disconnects all active sessions
func (cm *ClientManager) DisconnectAll(ctx context.Context) error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for sessionID, client := range cm.clients {
		if client.IsConnected() {
			client.Disconnect()
			log.Info().
				Str("session_id", sessionID.String()).
				Msg("Client disconnected in DisconnectAll")
		}
	}
	return nil
}

// Connect connects a client (implementing the domain interface)
func (cm *ClientManager) Connect(ctx context.Context, sessionID session.SessionID) error {
	return cm.ConnectClient(ctx, sessionID)
}

// Disconnect disconnects a client (implementing the domain interface)
func (cm *ClientManager) Disconnect(ctx context.Context, sessionID session.SessionID) error {
	return cm.DisconnectClient(ctx, sessionID)
}

// Logout logs out a client (implementing the domain interface)
func (cm *ClientManager) Logout(ctx context.Context, sessionID session.SessionID) error {
	return cm.LogoutClient(ctx, sessionID)
}

// Close closes the client manager
func (cm *ClientManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for sessionID, client := range cm.clients {
		if client.IsConnected() {
			client.Disconnect()
		}
		log.Info().
			Str("session_id", sessionID.String()).
			Msg("WhatsApp client closed")
	}

	return nil
}

// ClientWrapper wraps whatsmeow.Client to implement domain interface
type ClientWrapper struct {
	client    *whatsmeow.Client
	sessionID session.SessionID
}

// NewClientWrapper creates a new ClientWrapper
func NewClientWrapper(client *whatsmeow.Client, sessionID session.SessionID) *ClientWrapper {
	return &ClientWrapper{
		client:    client,
		sessionID: sessionID,
	}
}

// Connect establishes a connection for the given session
func (cw *ClientWrapper) Connect(ctx context.Context, sessionID session.SessionID) error {
	return cw.client.Connect()
}

// Disconnect closes the connection for the given session
func (cw *ClientWrapper) Disconnect(ctx context.Context, sessionID session.SessionID) error {
	cw.client.Disconnect()
	return nil
}

// Logout logs out and clears authentication for the session
func (cw *ClientWrapper) Logout(ctx context.Context, sessionID session.SessionID) error {
	return cw.client.Logout(ctx)
}

// GetQRCode generates a QR code for authentication
func (cw *ClientWrapper) GetQRCode(ctx context.Context, sessionID session.SessionID) (string, error) {
	// Implementation would go here
	return "", fmt.Errorf("not implemented")
}

// PairPhone pairs a phone number for authentication
func (cw *ClientWrapper) PairPhone(ctx context.Context, sessionID session.SessionID, phone string) (string, error) {
	// Implementation would go here
	return "", fmt.Errorf("not implemented")
}

// IsConnected checks if the session is connected
func (cw *ClientWrapper) IsConnected(ctx context.Context, sessionID session.SessionID) bool {
	return cw.client.IsConnected()
}

// IsAuthenticated checks if the session is authenticated
func (cw *ClientWrapper) IsAuthenticated(ctx context.Context, sessionID session.SessionID) bool {
	return cw.client.Store.ID != nil
}

// GetJID returns the WhatsApp JID for the session
func (cw *ClientWrapper) GetJID(ctx context.Context, sessionID session.SessionID) (string, error) {
	if cw.client.Store.ID == nil {
		return "", fmt.Errorf("not authenticated")
	}
	return cw.client.Store.ID.String(), nil
}

// SetProxy configures proxy for the session
func (cw *ClientWrapper) SetProxy(ctx context.Context, sessionID session.SessionID, proxyURL string) error {
	// Implementation would go here
	return fmt.Errorf("not implemented")
}

// GetConnectionStatus returns the current connection status
func (cw *ClientWrapper) GetConnectionStatus(ctx context.Context, sessionID session.SessionID) whatsapp.ConnectionStatus {
	if cw.client.IsConnected() {
		return whatsapp.ConnectionStatusConnected
	}
	return whatsapp.ConnectionStatusDisconnected
}
