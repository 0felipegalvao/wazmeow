package whats

import (
	"context"
	"fmt"
	"sync"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/internal/infra/config"
	"wazmeow/pkg/logger"

	"go.mau.fi/whatsmeow/store/sqlstore"
)

// SessionEventHandler handles WhatsApp events and updates session state
type SessionEventHandler struct {
	sessionRepo session.Repository
	logger      logger.Logger
}

// OnConnected handles connection events
func (h *SessionEventHandler) OnConnected(sessionID session.SessionID, jid string) {
	h.logger.InfoWithFields("📡 Session connected", logger.Fields{
		"session_id": sessionID.String(),
		"jid":        jid,
	})
}

// OnDisconnected handles disconnection events
func (h *SessionEventHandler) OnDisconnected(sessionID session.SessionID, reason string) {
	h.logger.InfoWithFields("📡 Session disconnected - updating status to disconnected", logger.Fields{
		"session_id": sessionID.String(),
		"reason":     reason,
	})

	ctx := context.Background()

	// Get session from database
	sess, err := h.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithError("Failed to get session for disconnection update", err, logger.Fields{
			"session_id": sessionID.String(),
			"reason":     reason,
		})
		return
	}

	// Update session status to disconnected
	sess.Disconnect()

	// Clear QR code if it exists (since connection failed)
	if sess.QRCode() != "" {
		sess.ClearQRCode()
		h.logger.InfoWithFields("🧹 Clearing QR code due to disconnection", logger.Fields{
			"session_id": sessionID.String(),
			"reason":     reason,
		})
	}

	// Save to database
	if err := h.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithError("Failed to save session disconnection status", err, logger.Fields{
			"session_id": sessionID.String(),
			"reason":     reason,
		})
		return
	}

	h.logger.InfoWithFields("✅ Session status updated to disconnected", logger.Fields{
		"session_id": sessionID.String(),
		"reason":     reason,
		"status":     sess.Status().String(),
	})
}

// OnQRCode handles QR code events
func (h *SessionEventHandler) OnQRCode(sessionID session.SessionID, qrCode string) {
	h.logger.InfoWithFields("📱 QR code generated - saving to database", logger.Fields{
		"session_id": sessionID.String(),
		"qr_length":  len(qrCode),
	})

	ctx := context.Background()

	// Get session from database
	sess, err := h.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithError("Failed to get session for QR code update", err, logger.Fields{
			"session_id": sessionID.String(),
		})
		return
	}

	// Update session with QR code
	sess.SetQRCode(qrCode)

	// Save updated session
	if err := h.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithError("Failed to save QR code to database", err, logger.Fields{
			"session_id": sessionID.String(),
		})
		return
	}

	h.logger.InfoWithFields("✅ QR code saved to database successfully", logger.Fields{
		"session_id": sessionID.String(),
		"qr_length":  len(qrCode),
	})
}

// OnAuthenticated handles successful authentication events
func (h *SessionEventHandler) OnAuthenticated(sessionID session.SessionID, jid string) {
	h.logger.InfoWithFields("🎉 Session authenticated - saving JID and clearing QR code", logger.Fields{
		"session_id": sessionID.String(),
		"jid":        jid,
	})

	ctx := context.Background()

	// Get session from database
	sess, err := h.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		h.logger.ErrorWithError("Failed to get session for JID update", err, logger.Fields{
			"session_id": sessionID.String(),
			"jid":        jid,
		})
		return
	}

	// Update session with JID and connected status
	if err := sess.Connect(jid); err != nil {
		h.logger.ErrorWithError("Failed to connect session with JID", err, logger.Fields{
			"session_id": sessionID.String(),
			"jid":        jid,
		})
		return
	}

	// Clear QR code since authentication is complete
	sess.ClearQRCode()

	// Save to database
	if err := h.sessionRepo.Update(ctx, sess); err != nil {
		h.logger.ErrorWithError("Failed to save session with JID and clear QR code", err, logger.Fields{
			"session_id": sessionID.String(),
			"jid":        jid,
		})
		return
	}

	h.logger.InfoWithFields("✅ Session JID saved and QR code cleared successfully", logger.Fields{
		"session_id": sessionID.String(),
		"jid":        jid,
		"status":     sess.Status().String(),
	})
}

// OnAuthenticationFailed handles authentication failure events
func (h *SessionEventHandler) OnAuthenticationFailed(sessionID session.SessionID, reason string) {
	h.logger.ErrorWithFields("❌ Session authentication failed", logger.Fields{
		"session_id": sessionID.String(),
		"reason":     reason,
	})
}

// OnMessage handles message events
func (h *SessionEventHandler) OnMessage(sessionID session.SessionID, message *whatsapp.Message) {
	h.logger.InfoWithFields("📨 Message received", logger.Fields{
		"session_id": sessionID.String(),
		"message_id": message.ID,
	})
}

// OnError handles error events
func (h *SessionEventHandler) OnError(sessionID session.SessionID, err error) {
	h.logger.ErrorWithError("💥 Session error", err, logger.Fields{
		"session_id": sessionID.String(),
	})
}

// Manager implements whatsapp.Manager with whatsmeow integration
type Manager struct {
	config       *config.WhatsAppConfig
	logger       logger.Logger
	container    *sqlstore.Container
	sessionRepo  session.Repository
	clients      map[session.SessionID]whatsapp.Client
	clientsMutex sync.RWMutex
	isRunning    bool
	eventHandler whatsapp.EventHandler
}

// NewManager creates a new WhatsApp manager
func NewManager(cfg *config.WhatsAppConfig, container *sqlstore.Container, sessionRepo session.Repository, log logger.Logger) whatsapp.Manager {
	manager := &Manager{
		config:      cfg,
		logger:      log,
		container:   container,
		sessionRepo: sessionRepo,
		clients:     make(map[session.SessionID]whatsapp.Client),
	}

	// Configure global event handler to save JID on authentication
	manager.eventHandler = &SessionEventHandler{
		sessionRepo: sessionRepo,
		logger:      log,
	}

	return manager
}

// Start initializes the manager
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("starting WhatsApp manager (simple implementation)")

	m.isRunning = true
	m.logger.Info("WhatsApp manager started successfully")

	return nil
}

// Stop shuts down the manager
func (m *Manager) Stop() error {
	m.logger.Info("stopping WhatsApp manager")

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Close all clients
	for sessionID, client := range m.clients {
		if err := client.Close(); err != nil {
			m.logger.ErrorWithError("failed to close WhatsApp client", err, logger.Fields{
				"session_id": sessionID.String(),
			})
		}
	}

	// Clear clients map
	m.clients = make(map[session.SessionID]whatsapp.Client)
	m.isRunning = false

	m.logger.Info("WhatsApp manager stopped")
	return nil
}

// IsRunning returns true if the manager is running
func (m *Manager) IsRunning() bool {
	return m.isRunning
}

// CreateClient creates a new WhatsApp client for a session
func (m *Manager) CreateClient(sessionID session.SessionID) (whatsapp.Client, error) {
	return m.createClientWithContext(context.Background(), sessionID)
}

// createClientWithContext creates a new WhatsApp client for a session with context
func (m *Manager) createClientWithContext(ctx context.Context, sessionID session.SessionID) (whatsapp.Client, error) {
	if !m.isRunning {
		return nil, fmt.Errorf("manager not running")
	}

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check if client already exists
	if client, exists := m.clients[sessionID]; exists {
		return client, nil
	}

	// Get saved JID and proxy URL from database for proper device management
	savedJID := ""
	proxyURL := ""
	if sess, err := m.sessionRepo.GetByID(ctx, sessionID); err == nil {
		savedJID = sess.WaJID()
		proxyURL = sess.ProxyURL()
		m.logger.InfoWithFields("Retrieved session data for client creation", logger.Fields{
			"session_id": sessionID.String(),
			"jid":        savedJID,
			"has_proxy":  proxyURL != "",
		})
	} else {
		m.logger.InfoWithFields("No session data found", logger.Fields{
			"session_id": sessionID.String(),
			"error":      err.Error(),
		})
	}

	// Create new client using whatsmeow with proper device management and proxy
	client, err := NewClient(sessionID, m.container, savedJID, proxyURL, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create whatsmeow client: %w", err)
	}

	// Set global event handler if available
	if m.eventHandler != nil {
		client.SetEventHandler(m.eventHandler)
	}

	// Store client
	m.clients[sessionID] = client

	m.logger.InfoWithFields("WhatsApp client created", logger.Fields{
		"session_id": sessionID.String(),
	})

	return client, nil
}

// GetClient retrieves an existing WhatsApp client
func (m *Manager) GetClient(sessionID session.SessionID) (whatsapp.Client, error) {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	client, exists := m.clients[sessionID]
	if !exists {
		return nil, whatsapp.ErrClientNotFound
	}

	return client, nil
}

// RemoveClient removes a WhatsApp client
func (m *Manager) RemoveClient(sessionID session.SessionID) error {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	client, exists := m.clients[sessionID]
	if !exists {
		return whatsapp.ErrClientNotFound
	}

	// Close the client
	if err := client.Close(); err != nil {
		m.logger.ErrorWithError("failed to close WhatsApp client", err, logger.Fields{
			"session_id": sessionID.String(),
		})
	}

	// Remove from map
	delete(m.clients, sessionID)

	m.logger.InfoWithFields("WhatsApp client removed", logger.Fields{
		"session_id": sessionID.String(),
	})

	return nil
}

// ListClients returns a list of all session IDs with active clients
func (m *Manager) ListClients() []session.SessionID {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	sessionIDs := make([]session.SessionID, 0, len(m.clients))
	for sessionID := range m.clients {
		sessionIDs = append(sessionIDs, sessionID)
	}

	return sessionIDs
}

// HealthCheck performs a health check on the manager
func (m *Manager) HealthCheck() error {
	if !m.isRunning {
		return fmt.Errorf("manager not running")
	}

	return nil
}

// SetGlobalEventHandler sets a global event handler for all clients
func (m *Manager) SetGlobalEventHandler(handler whatsapp.EventHandler) {
	m.eventHandler = handler

	// Apply to existing clients
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	for _, client := range m.clients {
		client.SetEventHandler(handler)
	}
}

// RemoveGlobalEventHandler removes the global event handler
func (m *Manager) RemoveGlobalEventHandler() {
	m.eventHandler = nil

	// Remove from existing clients
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	for _, client := range m.clients {
		client.RemoveEventHandler()
	}
}

// GetStats returns manager statistics
func (m *Manager) GetStats() *whatsapp.ManagerStats {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	stats := &whatsapp.ManagerStats{
		TotalClients: len(m.clients),
	}

	for _, client := range m.clients {
		if client.IsConnected() {
			stats.ConnectedClients++
		}
		if client.IsAuthenticated() {
			stats.AuthenticatedClients++
		}
		if client.GetConnectionStatus() == whatsapp.StatusError {
			stats.ErrorClients++
		}
	}

	return stats
}

// GetClientStats returns statistics for a specific client
func (m *Manager) GetClientStats(sessionID session.SessionID) (*whatsapp.ClientStats, error) {
	client, err := m.GetClient(sessionID)
	if err != nil {
		return nil, err
	}

	return &whatsapp.ClientStats{
		SessionID: sessionID,
		Status:    client.GetConnectionStatus(),
		JID:       client.GetJID(),
	}, nil
}

// ConnectAll connects all clients
func (m *Manager) ConnectAll(ctx context.Context) error {
	m.clientsMutex.RLock()
	clients := make([]whatsapp.Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientsMutex.RUnlock()

	for _, client := range clients {
		if !client.IsConnected() {
			if _, err := client.Connect(ctx); err != nil {
				m.logger.ErrorWithError("failed to connect client", err, logger.Fields{
					"session_id": client.GetSessionID().String(),
				})
			}
		}
	}

	return nil
}

// DisconnectAll disconnects all clients
func (m *Manager) DisconnectAll(ctx context.Context) error {
	m.clientsMutex.RLock()
	clients := make([]whatsapp.Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientsMutex.RUnlock()

	for _, client := range clients {
		if client.IsConnected() {
			if err := client.Disconnect(ctx); err != nil {
				m.logger.ErrorWithError("failed to disconnect client", err, logger.Fields{
					"session_id": client.GetSessionID().String(),
				})
			}
		}
	}

	return nil
}

// RestartClient restarts a specific client
func (m *Manager) RestartClient(sessionID session.SessionID) error {
	client, err := m.GetClient(sessionID)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Disconnect first
	if client.IsConnected() {
		if err := client.Disconnect(ctx); err != nil {
			m.logger.ErrorWithError("failed to disconnect client for restart", err, logger.Fields{
				"session_id": sessionID.String(),
			})
		}
	}

	// Reconnect
	if _, err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to reconnect client: %w", err)
	}

	return nil
}

// GetHealthStatus returns health status for all clients
func (m *Manager) GetHealthStatus() map[session.SessionID]bool {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	status := make(map[session.SessionID]bool)
	for sessionID, client := range m.clients {
		status[sessionID] = client.IsConnected()
	}

	return status
}

// RestartUnhealthyClients restarts clients that are not healthy
func (m *Manager) RestartUnhealthyClients(ctx context.Context) error {
	healthStatus := m.GetHealthStatus()

	for sessionID, isHealthy := range healthStatus {
		if !isHealthy {
			if err := m.RestartClient(sessionID); err != nil {
				m.logger.ErrorWithError("failed to restart unhealthy client", err, logger.Fields{
					"session_id": sessionID.String(),
				})
			}
		}
	}

	return nil
}
