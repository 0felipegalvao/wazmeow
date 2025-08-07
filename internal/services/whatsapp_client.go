package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"wazmeow/internal/domain"

	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WhatsAppClientWrapper wraps whatsmeow.Client to implement our domain interface
type WhatsAppClientWrapper struct {
	client    *whatsmeow.Client
	sessionID domain.SessionID
}

// Connect implements the domain interface
func (w *WhatsAppClientWrapper) Connect(ctx context.Context, sessionID domain.SessionID) error {
	return w.client.Connect()
}

// Disconnect implements the domain interface
func (w *WhatsAppClientWrapper) Disconnect(ctx context.Context, sessionID domain.SessionID) error {
	w.client.Disconnect()
	return nil
}

// Logout implements the domain interface
func (w *WhatsAppClientWrapper) Logout(ctx context.Context, sessionID domain.SessionID) error {
	return w.client.Logout(ctx)
}

// IsConnected implements the domain interface
func (w *WhatsAppClientWrapper) IsConnected(ctx context.Context, sessionID domain.SessionID) bool {
	return w.client.IsConnected()
}

// GetQRCode implements the domain interface
func (w *WhatsAppClientWrapper) GetQRCode(ctx context.Context, sessionID domain.SessionID) (string, error) {
	// This would need to be implemented based on whatsmeow's QR code generation
	return "", nil
}

// SendMessage implements the domain interface
func (w *WhatsAppClientWrapper) SendMessage(ctx context.Context, sessionID domain.SessionID, to string, message string) error {
	// This would need to be implemented based on whatsmeow's message sending
	return nil
}

// GetConnectionStatus implements the domain interface
func (w *WhatsAppClientWrapper) GetConnectionStatus(ctx context.Context, sessionID domain.SessionID) domain.ConnectionStatus {
	if w.client.IsConnected() {
		return domain.ConnectionStatusConnected
	}
	return domain.ConnectionStatusDisconnected
}

// GetJID implements the domain interface
func (w *WhatsAppClientWrapper) GetJID(ctx context.Context, sessionID domain.SessionID) (string, error) {
	if w.client.Store != nil && w.client.Store.ID != nil {
		return w.client.Store.ID.String(), nil
	}
	return "", nil
}

// SetProxy implements the domain interface
func (w *WhatsAppClientWrapper) SetProxy(ctx context.Context, sessionID domain.SessionID, proxyURL string) error {
	// This would need to be implemented based on whatsmeow's proxy configuration
	return nil
}

// IsAuthenticated implements the domain interface
func (w *WhatsAppClientWrapper) IsAuthenticated(ctx context.Context, sessionID domain.SessionID) bool {
	return w.client.Store != nil && w.client.Store.ID != nil
}

// PairPhone implements the domain interface
func (w *WhatsAppClientWrapper) PairPhone(ctx context.Context, sessionID domain.SessionID, phoneNumber string) (string, error) {
	// This would need to be implemented based on whatsmeow's phone pairing
	return "", nil
}

// ClientManager manages WhatsApp clients for multiple sessions
type ClientManager struct {
	storeManager *WhatsAppStoreManager
	clients      map[domain.SessionID]*whatsmeow.Client
	devices      map[domain.SessionID]*store.Device
	mutex        sync.RWMutex
	logger       waLog.Logger
}

// NewClientManager creates a new WhatsApp client manager
func NewClientManager(storeManager *WhatsAppStoreManager, logger waLog.Logger) *ClientManager {
	return &ClientManager{
		storeManager: storeManager,
		clients:      make(map[domain.SessionID]*whatsmeow.Client),
		devices:      make(map[domain.SessionID]*store.Device),
		logger:       logger,
	}
}

// CreateClient creates a new WhatsApp client for a session
func (cm *ClientManager) CreateClient(ctx context.Context, sessionID domain.SessionID) (domain.Client, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if client already exists
	if client, exists := cm.clients[sessionID]; exists {
		wrapper := &WhatsAppClientWrapper{
			client:    client,
			sessionID: sessionID,
		}
		return wrapper, nil
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

	wrapper := &WhatsAppClientWrapper{
		client:    client,
		sessionID: sessionID,
	}
	return wrapper, nil
}

// GetClient returns an existing WhatsApp client for a session
func (cm *ClientManager) GetClient(ctx context.Context, sessionID domain.SessionID) (*whatsmeow.Client, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	return client, nil
}

// ConnectClient connects a WhatsApp client
func (cm *ClientManager) ConnectClient(ctx context.Context, sessionID domain.SessionID) error {
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
func (cm *ClientManager) DisconnectClient(ctx context.Context, sessionID domain.SessionID) error {
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
func (cm *ClientManager) LogoutClient(ctx context.Context, sessionID domain.SessionID) error {
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
func (cm *ClientManager) RemoveClient(ctx context.Context, sessionID domain.SessionID) error {
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
func (cm *ClientManager) GetQRCode(ctx context.Context, sessionID domain.SessionID) (string, error) {
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
func (cm *ClientManager) IsClientConnected(sessionID domain.SessionID) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return false
	}

	return client.IsConnected()
}

// createEventHandler creates an event handler for a session
func (cm *ClientManager) createEventHandler(sessionID domain.SessionID) func(interface{}) {
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
				Int("reason", int(v.Reason)).
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
func (cm *ClientManager) GetAllClients(ctx context.Context) map[domain.SessionID]domain.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	result := make(map[domain.SessionID]domain.Client)
	for sessionID, client := range cm.clients {
		result[sessionID] = NewClientWrapper(client, sessionID)
	}
	return result
}

// SetEventHandler sets the event handler for all clients
func (cm *ClientManager) SetEventHandler(handler domain.EventHandler) {
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
func (cm *ClientManager) Connect(ctx context.Context, sessionID domain.SessionID) error {
	return cm.ConnectClient(ctx, sessionID)
}

// Disconnect disconnects a client (implementing the domain interface)
func (cm *ClientManager) Disconnect(ctx context.Context, sessionID domain.SessionID) error {
	return cm.DisconnectClient(ctx, sessionID)
}

// Logout logs out a client (implementing the domain interface)
func (cm *ClientManager) Logout(ctx context.Context, sessionID domain.SessionID) error {
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
	sessionID domain.SessionID
}

// NewClientWrapper creates a new ClientWrapper
func NewClientWrapper(client *whatsmeow.Client, sessionID domain.SessionID) *ClientWrapper {
	return &ClientWrapper{
		client:    client,
		sessionID: sessionID,
	}
}

// Connect establishes a connection for the given session
func (cw *ClientWrapper) Connect(ctx context.Context, sessionID domain.SessionID) error {
	return cw.client.Connect()
}

// Disconnect closes the connection for the given session
func (cw *ClientWrapper) Disconnect(ctx context.Context, sessionID domain.SessionID) error {
	cw.client.Disconnect()
	return nil
}

// Logout logs out and clears authentication for the session
func (cw *ClientWrapper) Logout(ctx context.Context, sessionID domain.SessionID) error {
	return cw.client.Logout(ctx)
}

// GetQRCode generates a QR code for authentication
func (cw *ClientWrapper) GetQRCode(ctx context.Context, sessionID domain.SessionID) (string, error) {
	// Implementation would go here
	return "", fmt.Errorf("not implemented")
}

// PairPhone pairs a phone number for authentication
func (cw *ClientWrapper) PairPhone(ctx context.Context, sessionID domain.SessionID, phone string) (string, error) {
	// Implementation would go here
	return "", fmt.Errorf("not implemented")
}

// IsConnected checks if the session is connected
func (cw *ClientWrapper) IsConnected(ctx context.Context, sessionID domain.SessionID) bool {
	return cw.client.IsConnected()
}

// IsAuthenticated checks if the session is authenticated
func (cw *ClientWrapper) IsAuthenticated(ctx context.Context, sessionID domain.SessionID) bool {
	return cw.client.Store.ID != nil
}

// GetJID returns the WhatsApp JID for the session
func (cw *ClientWrapper) GetJID(ctx context.Context, sessionID domain.SessionID) (string, error) {
	if cw.client.Store.ID == nil {
		return "", fmt.Errorf("not authenticated")
	}
	return cw.client.Store.ID.String(), nil
}

// SetProxy configures proxy for the session
func (cw *ClientWrapper) SetProxy(ctx context.Context, sessionID domain.SessionID, proxyURL string) error {
	// Implementation would go here
	return fmt.Errorf("not implemented")
}

// GetConnectionStatus returns the current connection status
func (cw *ClientWrapper) GetConnectionStatus(ctx context.Context, sessionID domain.SessionID) domain.ConnectionStatus {
	if cw.client.IsConnected() {
		return domain.ConnectionStatusConnected
	}
	return domain.ConnectionStatusDisconnected
}
