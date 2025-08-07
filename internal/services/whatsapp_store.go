package services

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"wazmeow/internal/domain"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WhatsAppStoreManager manages WhatsApp store containers and devices for multiple sessions
type WhatsAppStoreManager struct {
	// Core components
	container *sqlstore.Container
	logger    waLog.Logger

	// Thread-safe device cache
	devices map[domain.SessionID]*store.Device
	mutex   sync.RWMutex

	// Configuration
	maxDevices int
}

// NewWhatsAppStoreManager creates a new WhatsApp store manager
func NewWhatsAppStoreManager(db *sql.DB, logger waLog.Logger) (*WhatsAppStoreManager, error) {
	// Set up PostgreSQL array wrapper for whatsmeow
	sqlstore.PostgresArrayWrapper = pq.Array

	// Create sqlstore container
	container := sqlstore.NewWithDB(db, "postgres", logger)

	// Upgrade database schema
	ctx := context.Background()
	if err := container.Upgrade(ctx); err != nil {
		return nil, fmt.Errorf("failed to upgrade whatsmeow database schema: %w", err)
	}

	log.Info().Msg("WhatsApp store manager initialized successfully")

	return &WhatsAppStoreManager{
		container:  container,
		logger:     logger,
		devices:    make(map[domain.SessionID]*store.Device),
		maxDevices: 100, // Default limit
	}, nil
}

// GetOrCreateDevice gets an existing device or creates a new one for a session
func (wsm *WhatsAppStoreManager) GetOrCreateDevice(sessionID domain.SessionID, jid string) (*store.Device, error) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	// Check if device already exists in cache
	if device, exists := wsm.devices[sessionID]; exists {
		log.Debug().
			Str("session_id", sessionID.String()).
			Str("jid", jid).
			Msg("Device found in cache")
		return device, nil
	}

	// Check device limit
	if len(wsm.devices) >= wsm.maxDevices {
		return nil, fmt.Errorf("maximum number of devices (%d) reached", wsm.maxDevices)
	}

	var device *store.Device
	var err error

	// Try to restore existing device if JID is provided
	if jid != "" {
		device, err = wsm.restoreDevice(jid)
		if err != nil {
			log.Warn().
				Err(err).
				Str("session_id", sessionID.String()).
				Str("jid", jid).
				Msg("Failed to restore device, creating new one")
		}
	}

	// Create new device if restoration failed or no JID provided
	if device == nil {
		device = wsm.container.NewDevice()
		log.Info().
			Str("session_id", sessionID.String()).
			Msg("Created new WhatsApp device")
	} else {
		log.Info().
			Str("session_id", sessionID.String()).
			Str("jid", jid).
			Msg("Restored existing WhatsApp device")
	}

	// Cache the device
	wsm.devices[sessionID] = device

	return device, nil
}

// restoreDevice attempts to restore a device from the database using JID
func (wsm *WhatsAppStoreManager) restoreDevice(jid string) (*store.Device, error) {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JID %s: %w", jid, err)
	}

	ctx := context.Background()
	device, err := wsm.container.GetDevice(ctx, parsedJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device for JID %s: %w", jid, err)
	}

	return device, nil
}

// GetDevice returns a cached device for a session
func (wsm *WhatsAppStoreManager) GetDevice(sessionID domain.SessionID) (*store.Device, bool) {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()

	device, exists := wsm.devices[sessionID]
	return device, exists
}

// RemoveDevice removes a device from the cache
func (wsm *WhatsAppStoreManager) RemoveDevice(sessionID domain.SessionID) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	if _, exists := wsm.devices[sessionID]; exists {
		delete(wsm.devices, sessionID)
		log.Debug().
			Str("session_id", sessionID.String()).
			Msg("Device removed from cache")
	}
}

// GetContainer returns the sqlstore container
func (wsm *WhatsAppStoreManager) GetContainer() *sqlstore.Container {
	return wsm.container
}

// Close closes the store manager and cleans up resources
func (wsm *WhatsAppStoreManager) Close() error {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	// Clear device cache
	wsm.devices = make(map[domain.SessionID]*store.Device)

	// Close container
	if wsm.container != nil {
		if err := wsm.container.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close WhatsApp store container")
			return err
		}
	}

	log.Info().Msg("WhatsApp store manager closed")
	return nil
}

// GetStats returns statistics about the store manager
func (wsm *WhatsAppStoreManager) GetStats() map[string]interface{} {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()

	return map[string]interface{}{
		"device_count": len(wsm.devices),
		"max_devices":  wsm.maxDevices,
		"device_usage": float64(len(wsm.devices)) / float64(wsm.maxDevices) * 100,
	}
}
