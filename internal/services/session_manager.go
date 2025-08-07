// Package services provides business logic implementations for WazMeow.
// This file contains the MultiSessionManager which handles multiple WhatsApp sessions
// concurrently, providing session lifecycle management, QR code generation,
// phone pairing, and event handling.
package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"wazmeow/internal/domain"

	"github.com/mdp/qrterminal/v3"
	"github.com/rs/zerolog/log"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types/events"
)

// ConnectionStatus represents the connection status of a WhatsApp session
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusError        ConnectionStatus = "error"
)

// SessionClient holds all components for a single WhatsApp session
type SessionClient struct {
	Client      *whatsmeow.Client
	Device      *store.Device
	KillChannel chan bool
	Status      ConnectionStatus
	LastSeen    time.Time
}

// MultiSessionManager manages multiple WhatsApp sessions concurrently
type MultiSessionManager struct {
	// Thread-safe maps for multiple sessions
	sessions map[domain.SessionID]*SessionClient

	// Components
	storeManager *WhatsAppStoreManager
	sessionRepo  domain.Repository

	// Concurrency control
	mutex sync.RWMutex

	// Configuration
	maxSessions int
}

// NewMultiSessionManager creates a new multi-session manager
func NewMultiSessionManager(
	storeManager *WhatsAppStoreManager,
	sessionRepo domain.Repository,
) *MultiSessionManager {
	msm := &MultiSessionManager{
		sessions:     make(map[domain.SessionID]*SessionClient),
		storeManager: storeManager,
		sessionRepo:  sessionRepo,
		maxSessions:  50, // Default limit
	}

	// Start automatic reconnection of previously connected sessions
	go msm.connectOnStartup()

	return msm
}

// connectOnStartup connects to WhatsApp sessions that were previously connected
func (msm *MultiSessionManager) connectOnStartup() {
	// Wait a bit for the system to fully initialize
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get all sessions that were previously connected
	sessions, err := msm.sessionRepo.GetConnectedSessions(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sessions for startup reconnection")
		return
	}

	connectedCount := 0
	for _, session := range sessions {
		// Only reconnect sessions that were connected and have a JID
		if session.Status == domain.StatusConnected && session.WAJID != "" {
			log.Info().
				Str("session_id", session.ID.String()).
				Str("session_name", session.Name).
				Str("wa_jid", session.WAJID).
				Msg("Reconnecting session on startup")

			// Start the session asynchronously
			go func(sessionID domain.SessionID) {
				if err := msm.StartSession(context.Background(), sessionID); err != nil {
					log.Error().
						Err(err).
						Str("session_id", sessionID.String()).
						Msg("Failed to reconnect session on startup")
				}
			}(session.ID)

			connectedCount++

			// Add a small delay between connections to avoid overwhelming the system
			time.Sleep(500 * time.Millisecond)
		}
	}

	if connectedCount > 0 {
		log.Info().
			Int("reconnected_sessions", connectedCount).
			Msg("Startup reconnection completed")
	} else {
		log.Info().Msg("No sessions to reconnect on startup")
	}
}

// StartSession starts a WhatsApp session for the given session ID
func (msm *MultiSessionManager) StartSession(ctx context.Context, sessionID domain.SessionID) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()

	// Check if session already exists
	if client, exists := msm.sessions[sessionID]; exists {
		if client.Status == StatusConnected {
			log.Info().Str("session_id", sessionID.String()).Msg("Session already connected")
			return nil
		}
		// Clean up existing session if it's in error state
		msm.cleanupSessionUnsafe(sessionID)
	}

	// Check session limit
	if len(msm.sessions) >= msm.maxSessions {
		return fmt.Errorf("maximum number of sessions (%d) reached", msm.maxSessions)
	}

	// Get session from database
	session, err := msm.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session from database: %w", err)
	}

	// Get or create device
	device, err := msm.storeManager.GetOrCreateDevice(sessionID, session.WAJID)
	if err != nil {
		return fmt.Errorf("failed to get or create device: %w", err)
	}

	// Create WhatsApp client
	client := whatsmeow.NewClient(device, nil)

	// Create session client
	sessionClient := &SessionClient{
		Client:      client,
		Device:      device,
		KillChannel: make(chan bool, 1),
		Status:      StatusDisconnected,
		LastSeen:    time.Now(),
	}

	// Store session client
	msm.sessions[sessionID] = sessionClient

	// Start connection in goroutine with background context
	go msm.handleSessionConnection(context.Background(), sessionID, sessionClient)

	log.Info().Str("session_id", sessionID.String()).Msg("Session started")
	return nil
}

// GetClient returns the WhatsApp client for a session
func (msm *MultiSessionManager) GetClient(sessionID domain.SessionID) (*whatsmeow.Client, error) {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()

	sessionClient, exists := msm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if sessionClient.Client == nil {
		return nil, fmt.Errorf("session client not initialized: %s", sessionID)
	}

	return sessionClient.Client, nil
}

// StopSession stops a WhatsApp session
func (msm *MultiSessionManager) StopSession(ctx context.Context, sessionID domain.SessionID) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()

	return msm.cleanupSessionUnsafe(sessionID)
}

// GetSessionStatus returns the current status of a session
func (msm *MultiSessionManager) GetSessionStatus(sessionID domain.SessionID) ConnectionStatus {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()

	if client, exists := msm.sessions[sessionID]; exists {
		return client.Status
	}
	return StatusDisconnected
}

// GetSessionClient returns the WhatsApp client for a session (thread-safe)
func (msm *MultiSessionManager) GetSessionClient(sessionID domain.SessionID) (*whatsmeow.Client, bool) {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()

	if sessionClient, exists := msm.sessions[sessionID]; exists && sessionClient.Status == StatusConnected {
		return sessionClient.Client, true
	}
	return nil, false
}

// GetActiveSessions returns a list of all active session IDs
func (msm *MultiSessionManager) GetActiveSessions() []domain.SessionID {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()

	var activeSessions []domain.SessionID
	for sessionID, client := range msm.sessions {
		if client.Status == StatusConnected {
			activeSessions = append(activeSessions, sessionID)
		}
	}
	return activeSessions
}

// GetSessionCount returns the total number of managed sessions
func (msm *MultiSessionManager) GetSessionCount() int {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()

	return len(msm.sessions)
}

// GenerateQRCode generates a QR code for session authentication
func (msm *MultiSessionManager) GenerateQRCode(ctx context.Context, sessionID domain.SessionID) (string, error) {
	msm.mutex.RLock()
	sessionClient, exists := msm.sessions[sessionID]
	msm.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if sessionClient.Status == StatusConnected {
		return "", fmt.Errorf("session %s is already connected", sessionID)
	}

	// Check if session already has a QR code stored
	session, err := msm.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session from database: %w", err)
	}

	// If QR code exists and is not empty, return it
	if session.QRCode != "" {
		log.Info().
			Str("session_id", sessionID.String()).
			Msg("Returning existing QR code")
		return session.QRCode, nil
	}

	// Start QR code generation process asynchronously with background context
	go msm.handleQRCodeGeneration(context.Background(), sessionID, sessionClient)

	// Wait a bit for QR code to be generated
	time.Sleep(2 * time.Second)

	// Try to get QR code again
	session, err = msm.sessionRepo.GetByID(ctx, sessionID)
	if err == nil && session.QRCode != "" {
		log.Info().
			Str("session_id", sessionID.String()).
			Msg("QR code generated and retrieved")
		return session.QRCode, nil
	}

	// Return a message indicating QR code is being generated
	return "", fmt.Errorf("QR code is being generated, please try again in a few seconds")
}

// handleQRCodeGeneration handles the asynchronous QR code generation
func (msm *MultiSessionManager) handleQRCodeGeneration(ctx context.Context, sessionID domain.SessionID, sessionClient *SessionClient) {
	// Get QR code channel BEFORE connecting (this is the correct order)
	qrChan, err := sessionClient.Client.GetQRChannel(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID.String()).
			Msg("Failed to get QR channel")
		return
	}

	// Connect client AFTER getting QR channel
	if err := sessionClient.Client.Connect(); err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionID.String()).
			Msg("Failed to connect client for QR generation")
		return
	}

	// Wait for QR code events
	for evt := range qrChan {
		switch evt.Event {
		case "code":
			// Generate base64 QR code image
			qrCodeBase64, err := msm.generateQRCodeImage(evt.Code)
			if err != nil {
				log.Error().
					Err(err).
					Str("session_id", sessionID.String()).
					Msg("Failed to generate QR code image")
				continue
			}

			// Store QR code in database
			log.Info().
				Str("session_id", sessionID.String()).
				Str("qr_code_length", fmt.Sprintf("%d", len(qrCodeBase64))).
				Msg("Attempting to store QR code in database")

			if err := msm.sessionRepo.SetQRCode(ctx, sessionID, qrCodeBase64); err != nil {
				log.Error().
					Err(err).
					Str("session_id", sessionID.String()).
					Str("qr_code", qrCodeBase64).
					Msg("Failed to store QR code in database")
			} else {
				log.Info().
					Str("session_id", sessionID.String()).
					Msg("QR code generated and stored successfully")
			}

		case "success":
			log.Info().
				Str("session_id", sessionID.String()).
				Msg("QR code pairing successful")

			// Clear QR code from database
			if err := msm.sessionRepo.SetQRCode(ctx, sessionID, ""); err != nil {
				log.Error().
					Err(err).
					Str("session_id", sessionID.String()).
					Msg("Failed to clear QR code from database")
			}
			return

		case "timeout":
			log.Warn().
				Str("session_id", sessionID.String()).
				Msg("QR code timeout")

			// Clear QR code from database
			if err := msm.sessionRepo.SetQRCode(ctx, sessionID, ""); err != nil {
				log.Error().
					Err(err).
					Str("session_id", sessionID.String()).
					Msg("Failed to clear QR code from database")
			}
			return

		default:
			log.Info().
				Str("session_id", sessionID.String()).
				Str("event", evt.Event).
				Msg("QR code event")
		}
	}
}

// generateQRCodeImage generates a base64 encoded QR code image
func (msm *MultiSessionManager) generateQRCodeImage(code string) (string, error) {
	// Display QR code in terminal for easy scanning
	fmt.Println("\n=== QR CODE FOR WHATSAPP ===")
	qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)
	fmt.Printf("QR Code String: %s\n", code)
	fmt.Println("=============================")

	// Generate QR code image as PNG
	image, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Encode as base64
	base64qrcode := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)
	return base64qrcode, nil
}

// IsSessionConnected checks if a session is connected
func (msm *MultiSessionManager) IsSessionConnected(sessionID domain.SessionID) bool {
	return msm.GetSessionStatus(sessionID) == StatusConnected
}

// PairPhone initiates phone pairing for a session
func (msm *MultiSessionManager) PairPhone(ctx context.Context, sessionID domain.SessionID, phoneNumber string) (string, error) {
	msm.mutex.RLock()
	sessionClient, exists := msm.sessions[sessionID]
	msm.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("session %s not found", sessionID)
	}

	if sessionClient.Status == StatusConnected {
		return "", fmt.Errorf("session %s is already connected", sessionID)
	}

	// Use whatsmeow's PairPhone method
	linkingCode, err := sessionClient.Client.PairPhone(ctx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return "", fmt.Errorf("failed to initiate phone pairing: %w", err)
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Str("phone_number", phoneNumber).
		Str("linking_code", linkingCode).
		Msg("Phone pairing initiated")

	return linkingCode, nil
}

// GetSessionInfo returns detailed information about a session
func (msm *MultiSessionManager) GetSessionInfo(sessionID domain.SessionID) map[string]any {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()

	sessionClient, exists := msm.sessions[sessionID]
	if !exists {
		return map[string]any{
			"session_id": sessionID.String(),
			"exists":     false,
		}
	}

	info := map[string]any{
		"session_id": sessionID.String(),
		"exists":     true,
		"status":     string(sessionClient.Status),
		"last_seen":  sessionClient.LastSeen,
	}

	// Add device info if available
	if sessionClient.Device != nil && sessionClient.Device.ID != nil {
		info["jid"] = sessionClient.Device.ID.String()
		info["push_name"] = sessionClient.Device.PushName
		info["platform"] = sessionClient.Device.Platform
	}

	// Add client info if connected
	if sessionClient.Client != nil && sessionClient.Status == StatusConnected {
		store := sessionClient.Client.Store
		if store.ID != nil {
			info["connected_jid"] = store.ID.String()
		}
	}

	return info
}

// cleanupSessionUnsafe cleans up a session (must be called with mutex locked)
func (msm *MultiSessionManager) cleanupSessionUnsafe(sessionID domain.SessionID) error {
	sessionClient, exists := msm.sessions[sessionID]
	if !exists {
		return nil
	}

	// Send kill signal
	select {
	case sessionClient.KillChannel <- true:
	default:
		// Channel might be full or closed
	}

	// Disconnect client
	if sessionClient.Client != nil {
		sessionClient.Client.Disconnect()
	}

	// Remove from sessions map
	delete(msm.sessions, sessionID)

	log.Info().Str("session_id", sessionID.String()).Msg("Session cleaned up")
	return nil
}

// handleSessionConnection handles the connection lifecycle for a session
func (msm *MultiSessionManager) handleSessionConnection(ctx context.Context, sessionID domain.SessionID, sessionClient *SessionClient) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Str("session_id", sessionID.String()).
				Interface("panic", r).
				Msg("Panic in session connection handler")
		}
	}()

	// Update status to connecting
	msm.updateSessionStatus(sessionID, StatusConnecting)

	// Set up event handlers
	msm.setupEventHandlers(sessionID, sessionClient)

	// Check if device has stored ID (already logged in)
	if sessionClient.Device.ID == nil {
		// No ID stored, new login - need QR code
		log.Info().
			Str("session_id", sessionID.String()).
			Msg("New device, QR code authentication required")

		// Start QR code process
		msm.handleQRCodeGeneration(ctx, sessionID, sessionClient)
	} else {
		// Device already has ID, try to connect directly
		log.Info().
			Str("session_id", sessionID.String()).
			Msg("Device has stored ID, attempting direct connection")

		if err := sessionClient.Client.Connect(); err != nil {
			log.Error().
				Err(err).
				Str("session_id", sessionID.String()).
				Msg("Failed to connect to WhatsApp")

			msm.updateSessionStatus(sessionID, StatusError)
			return
		}
	}

	// Wait for kill signal or context cancellation
	select {
	case <-sessionClient.KillChannel:
		log.Info().Str("session_id", sessionID.String()).Msg("Session received kill signal")
	case <-ctx.Done():
		log.Info().Str("session_id", sessionID.String()).Msg("Session context cancelled")
	}

	// Cleanup
	sessionClient.Client.Disconnect()
	msm.updateSessionStatus(sessionID, StatusDisconnected)
}

// updateSessionStatus updates the status of a session both in memory and database
func (msm *MultiSessionManager) updateSessionStatus(sessionID domain.SessionID, status ConnectionStatus) {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()

	if sessionClient, exists := msm.sessions[sessionID]; exists {
		sessionClient.Status = status
		sessionClient.LastSeen = time.Now()

		// Update status in database asynchronously to avoid blocking
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Convert ConnectionStatus to domain.Status
			var domainStatus domain.Status
			switch status {
			case StatusDisconnected:
				domainStatus = domain.StatusDisconnected
			case StatusConnecting:
				domainStatus = domain.StatusConnecting
			case StatusConnected:
				domainStatus = domain.StatusConnected
			case StatusError:
				domainStatus = domain.StatusError
			default:
				domainStatus = domain.StatusDisconnected
			}

			if err := msm.sessionRepo.UpdateStatus(ctx, sessionID, domainStatus); err != nil {
				log.Error().
					Err(err).
					Str("session_id", sessionID.String()).
					Str("status", string(status)).
					Msg("Failed to update session status in database")
			} else {
				log.Info().
					Str("session_id", sessionID.String()).
					Str("status", string(status)).
					Msg("Session status updated in database")
			}
		}()
	}
}

// setupEventHandlers sets up event handlers for a WhatsApp client
func (msm *MultiSessionManager) setupEventHandlers(sessionID domain.SessionID, sessionClient *SessionClient) {
	// This will be expanded in the next phase with proper event handling
	// For now, we'll add basic connection status handling

	sessionClient.Client.AddEventHandler(func(evt any) {
		switch v := evt.(type) {
		case *events.Connected:
			log.Info().Str("session_id", sessionID.String()).Msg("WhatsApp connected")
			msm.updateSessionStatus(sessionID, StatusConnected)

		case *events.Disconnected:
			log.Info().Str("session_id", sessionID.String()).Msg("WhatsApp disconnected")
			msm.updateSessionStatus(sessionID, StatusDisconnected)

		case *events.PairSuccess:
			jid := v.ID.String()
			log.Info().
				Str("session_id", sessionID.String()).
				Str("jid", jid).
				Msg("WhatsApp pairing successful")

			// Update session with JID in database
			ctx := context.Background()
			if err := msm.sessionRepo.SetWAJID(ctx, sessionID, jid); err != nil {
				log.Error().
					Err(err).
					Str("session_id", sessionID.String()).
					Str("jid", jid).
					Msg("Failed to update session JID in database")
			} else {
				log.Info().
					Str("session_id", sessionID.String()).
					Str("jid", jid).
					Msg("Session JID updated in database")
			}

		default:
			// Handle other events as needed
			_ = v
		}
	})
}

// Shutdown gracefully shuts down all sessions
func (msm *MultiSessionManager) Shutdown(ctx context.Context) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()

	log.Info().Int("session_count", len(msm.sessions)).Msg("Shutting down all sessions")

	for sessionID := range msm.sessions {
		if err := msm.cleanupSessionUnsafe(sessionID); err != nil {
			log.Error().
				Err(err).
				Str("session_id", sessionID.String()).
				Msg("Error cleaning up session during shutdown")
		}
	}

	return nil
}
