package container

import (
	"testing"

	"wazmeow/internal/app/container"
	"wazmeow/internal/infra/config"
)

func TestNewAppContainer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		opts    []container.AppOption
		wantErr bool
	}{
		{
			name:    "successful creation with default options",
			config:  createTestConfig(),
			opts:    nil,
			wantErr: false,
		},
		{
			name: "successful creation with custom options",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Driver:       "sqlite3",
					URL:          "file:testdb2?mode=memory&cache=shared&_foreign_keys=1",
					AutoMigrate:  true,
					MaxOpenConns: 1,
					MaxIdleConns: 1,
					SQLite: config.SQLiteConfig{
						Path:        "",
						ForeignKeys: true,
						JournalMode: "MEMORY",
						Synchronous: "OFF",
						CacheSize:   1000,
						TempStore:   "memory",
						MmapSize:    0,
					},
				},
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Log: config.LogConfig{
					Level:         "info",
					Output:        "console",
					ConsoleFormat: "console",
					FileFormat:    "json",
				},
				WhatsApp: config.WhatsAppConfig{},
				Security: config.SecurityConfig{},
				Features: config.FeaturesConfig{},
				Auth:     config.AuthConfig{},
				Proxy:    config.ProxyConfig{},
			},
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
				container.WithDebugMode(true),
				container.WithMetrics(false),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appContainer, err := container.NewAppContainer(tt.config, tt.opts...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAppContainer() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewAppContainer() unexpected error: %v", err)
				return
			}

			if appContainer == nil {
				t.Errorf("NewAppContainer() returned nil container")
				return
			}

			// Verify container is initialized
			if !appContainer.IsInitialized() {
				t.Errorf("NewAppContainer() container not initialized")
			}

			// Verify config is set
			if appContainer.GetConfig() != tt.config {
				t.Errorf("NewAppContainer() config not set correctly")
			}

			// Verify logger is available
			if appContainer.GetLogger() == nil {
				t.Errorf("NewAppContainer() logger not available")
			}

			// Clean up
			if err := appContainer.Close(); err != nil {
				t.Errorf("Failed to close container: %v", err)
			}
		})
	}
}

func TestAppContainer_Health(t *testing.T) {
	config := createTestConfig()

	appContainer, err := container.NewAppContainer(config)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer appContainer.Close()

	// Start WhatsApp manager for health check
	if err := appContainer.StartWhatsAppManager(); err != nil {
		t.Fatalf("Failed to start WhatsApp manager: %v", err)
	}

	// Test health check
	if err := appContainer.Health(); err != nil {
		t.Errorf("Health() returned error: %v", err)
	}
}

func TestAppContainer_GetMethods(t *testing.T) {
	config := createTestConfig()

	appContainer, err := container.NewAppContainer(config)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer appContainer.Close()

	// Test GetConfig
	if got := appContainer.GetConfig(); got != config {
		t.Errorf("GetConfig() = %v, want %v", got, config)
	}

	// Test GetLogger
	if got := appContainer.GetLogger(); got == nil {
		t.Errorf("GetLogger() returned nil")
	}

	// Test GetServerManager
	if got := appContainer.GetServerManager(); got == nil {
		t.Errorf("GetServerManager() returned nil")
	}

	// Test GetInfraContainer
	if got := appContainer.GetInfraContainer(); got == nil {
		t.Errorf("GetInfraContainer() returned nil")
	}

	// Test IsInitialized
	if !appContainer.IsInitialized() {
		t.Errorf("IsInitialized() returned false for initialized container")
	}
}

func TestAppContainer_StartWhatsAppManager(t *testing.T) {
	config := createTestConfig()

	appContainer, err := container.NewAppContainer(config)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer appContainer.Close()

	// Test StartWhatsAppManager
	if err := appContainer.StartWhatsAppManager(); err != nil {
		t.Errorf("StartWhatsAppManager() returned error: %v", err)
	}
}

func TestAppContainer_GetStats(t *testing.T) {
	config := createTestConfig()

	appContainer, err := container.NewAppContainer(config)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	defer appContainer.Close()

	// Test GetDatabaseStats
	if got := appContainer.GetDatabaseStats(); got == nil {
		t.Errorf("GetDatabaseStats() returned nil")
	}

	// Test GetWhatsAppStats
	if got := appContainer.GetWhatsAppStats(); got == nil {
		t.Errorf("GetWhatsAppStats() returned nil")
	}

	// Test GetServerInfo
	serverInfo := appContainer.GetServerInfo()
	if serverInfo.Address == "" {
		t.Errorf("GetServerInfo() returned empty address")
	}
}
