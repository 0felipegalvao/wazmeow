package container

import (
	"wazmeow/internal/infra/config"
)

// createTestConfig creates a valid test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Driver:       "sqlite3",
			URL:          "file:testdb?mode=memory&cache=shared&_foreign_keys=1",
			AutoMigrate:  true, // Enable auto-migrate for tests
			MaxOpenConns: 1,    // Force single connection to avoid SQLite memory issues
			MaxIdleConns: 1,    // Keep only one idle connection
			SQLite: config.SQLiteConfig{
				Path:        "",   // Use URL instead of Path
				ForeignKeys: true, // Enable foreign keys
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
			Level:         "error", // Reduce log noise in tests
			Output:        "console",
			ConsoleFormat: "console",
			FileFormat:    "json",
		},
		WhatsApp: config.WhatsAppConfig{},
		Security: config.SecurityConfig{},
		Features: config.FeaturesConfig{},
		Auth:     config.AuthConfig{},
		Proxy:    config.ProxyConfig{},
	}
}
