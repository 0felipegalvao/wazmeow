package whatsapp

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// StoreManager manages WhatsApp store containers
type StoreManager struct {
	container *sqlstore.Container
	logger    waLog.Logger
}

// NewStoreManager creates a new WhatsApp store manager
func NewStoreManager(db *sql.DB, logger waLog.Logger) (*StoreManager, error) {
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

	return &StoreManager{
		container: container,
		logger:    logger,
	}, nil
}

// GetContainer returns the sqlstore container
func (sm *StoreManager) GetContainer() *sqlstore.Container {
	return sm.container
}

// Close closes the store manager
func (sm *StoreManager) Close() error {
	if sm.container != nil {
		return sm.container.Close()
	}
	return nil
}

// GetAllDevices returns all devices in the store
func (sm *StoreManager) GetAllDevices(ctx context.Context) ([]DeviceInfo, error) {
	devices, err := sm.container.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	deviceInfos := make([]DeviceInfo, len(devices))
	for i, device := range devices {
		deviceInfos[i] = DeviceInfo{
			JID:       device.ID.String(),
			PushName:  device.PushName,
			Platform:  device.Platform,
			Connected: false, // Will be updated by client manager
		}
	}

	return deviceInfos, nil
}

// DeviceInfo represents device information
type DeviceInfo struct {
	JID       string `json:"jid"`
	PushName  string `json:"push_name"`
	Platform  string `json:"platform"`
	Connected bool   `json:"connected"`
}
