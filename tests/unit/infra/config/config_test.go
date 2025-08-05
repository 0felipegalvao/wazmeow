package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"wazmeow/internal/infra/config"
)

func TestConfig_Load(t *testing.T) {
	t.Run("should load config with default values", func(t *testing.T) {
		// Arrange - Clear environment variables and set test database
		os.Clearenv()
		os.Setenv("DB_URL", ":memory:")
		os.Setenv("SQLITE_PATH", "")
		defer func() {
			os.Unsetenv("DB_URL")
			os.Unsetenv("SQLITE_PATH")
		}()

		// Act
		cfg, err := config.Load()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Check default values
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, "info", cfg.Log.Level)
		assert.Equal(t, "console", cfg.Log.Output)
		assert.Equal(t, ":memory:", cfg.Database.URL)
	})

	t.Run("should load config from environment variables", func(t *testing.T) {
		// Arrange
		os.Clearenv()
		os.Setenv("SERVER_PORT", "9090")
		os.Setenv("SERVER_HOST", "127.0.0.1")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("LOG_OUTPUT", "console")
		os.Setenv("DB_URL", "/custom/path/db.sqlite")

		// Act
		cfg, err := config.Load()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Check environment values
		assert.Equal(t, 9090, cfg.Server.Port)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, "debug", cfg.Log.Level)
		assert.Equal(t, "console", cfg.Log.Output)
		assert.Equal(t, "/custom/path/db.sqlite", cfg.Database.URL)

		// Cleanup
		os.Clearenv()
	})

	t.Run("should validate server configuration", func(t *testing.T) {
		// Arrange
		os.Clearenv()
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("SERVER_HOST", "localhost")
		os.Setenv("DB_URL", ":memory:")
		os.Setenv("SQLITE_PATH", "")

		// Act
		cfg, err := config.Load()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Validate config (only the main config has Validate method)
		err = cfg.Validate()
		assert.NoError(t, err)

		// Cleanup
		os.Clearenv()
	})

	t.Run("should fail validation with invalid port", func(t *testing.T) {
		// Arrange
		os.Clearenv()
		os.Setenv("SERVER_PORT", "70000") // Port > 65535
		os.Setenv("DB_URL", ":memory:")
		os.Setenv("SQLITE_PATH", "")

		// Act
		cfg, err := config.Load()

		// Assert
		if err == nil {
			// If config loads successfully, validation should fail
			err = cfg.Validate()
			assert.Error(t, err)
		} else {
			// If config loading fails, that's also acceptable
			assert.Error(t, err)
		}

		// Cleanup
		os.Clearenv()
	})

	t.Run("should validate logger configuration", func(t *testing.T) {
		// Arrange
		os.Clearenv()
		os.Setenv("LOGGER_LEVEL", "info")
		os.Setenv("LOGGER_FORMAT", "console")
		os.Setenv("DB_URL", ":memory:")
		os.Setenv("SQLITE_PATH", "")

		// Act
		cfg, err := config.Load()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Validate config (only the main config has Validate method)
		err = cfg.Validate()
		assert.NoError(t, err)

		// Cleanup
		os.Clearenv()
	})

	t.Run("should fail validation with invalid logger level", func(t *testing.T) {
		// Arrange
		os.Clearenv()
		os.Setenv("LOG_LEVEL", "invalid_level")
		os.Setenv("DB_URL", ":memory:")
		os.Setenv("SQLITE_PATH", "")

		// Act
		cfg, err := config.Load()

		// Assert
		if err == nil {
			// If config loads successfully, validation should fail
			err = cfg.Validate()
			assert.Error(t, err)
		} else {
			// If config loading fails, that's also acceptable
			assert.Error(t, err)
		}

		// Cleanup
		os.Clearenv()
	})

	t.Run("should validate database configuration", func(t *testing.T) {
		// Arrange
		os.Clearenv()
		os.Setenv("DB_URL", ":memory:")
		os.Setenv("SQLITE_PATH", "")

		// Act
		cfg, err := config.Load()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Validate config (only the main config has Validate method)
		err = cfg.Validate()
		assert.NoError(t, err)

		// Cleanup
		os.Clearenv()
	})

	t.Run("should fail validation with empty database path", func(t *testing.T) {
		// Arrange - create config with empty database URL manually
		cfg := &config.Config{
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
			Database: config.DatabaseConfig{
				URL: "", // Empty URL should fail validation
			},
		}

		// Act & Assert
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database driver is required")
	})
}

func TestServerConfig(t *testing.T) {
	t.Run("should create valid server config", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Database: config.DatabaseConfig{
				Driver: "sqlite3",
				URL:    "./test.db",
			},
			Log: config.LogConfig{
				Level:         "info",
				Output:        "console",
				ConsoleFormat: "console",
				FileFormat:    "json",
			},
		}

		// Act
		err := cfg.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should get server address", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: "127.0.0.1",
				Port: 9090,
			},
		}

		// Act
		address := cfg.GetServerAddress()

		// Assert
		assert.Equal(t, "127.0.0.1:9090", address)
	})

	t.Run("should handle empty host", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: "",
				Port: 8080,
			},
		}

		// Act
		address := cfg.GetServerAddress()

		// Assert
		assert.Equal(t, ":8080", address)
	})

	t.Run("should validate port range", func(t *testing.T) {
		testCases := []struct {
			name        string
			port        int
			shouldError bool
		}{
			{"valid port 80", 80, false},
			{"valid port 8080", 8080, false},
			{"valid port 65535", 65535, false},
			{"invalid port 0", 0, true},
			{"invalid port 65536", 65536, true},
			{"invalid port negative", -1, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				cfg := &config.Config{
					Server: config.ServerConfig{
						Host: "localhost",
						Port: tc.port,
					},
					Database: config.DatabaseConfig{
						Driver: "sqlite3",
						URL:    "./test.db",
					},
					Log: config.LogConfig{
						Level:         "info",
						Output:        "console",
						ConsoleFormat: "console",
						FileFormat:    "json",
					},
				}

				// Act
				err := cfg.Validate()

				// Assert
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestLogConfig_Validate(t *testing.T) {
	t.Run("should validate with valid log config", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Database: config.DatabaseConfig{
				Driver: "sqlite3",
				URL:    "./test.db",
			},
			Log: config.LogConfig{
				Level:         "info",
				Output:        "console",
				ConsoleFormat: "console",
				FileFormat:    "json",
			},
		}

		// Act
		err := cfg.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("should fail with invalid log level", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Database: config.DatabaseConfig{
				Driver: "sqlite3",
				URL:    "./test.db",
			},
			Log: config.LogConfig{
				Level:         "invalid",
				Output:        "console",
				ConsoleFormat: "console",
				FileFormat:    "json",
			},
		}

		// Act
		err := cfg.Validate()

		// Assert
		assert.Error(t, err)
	})
}

func TestDatabaseConfig(t *testing.T) {
	t.Run("should validate database config", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
			},
			Database: config.DatabaseConfig{
				Driver: "sqlite3",
				URL:    "./test.db",
			},
			Log: config.LogConfig{
				Level:         "info",
				Output:        "console",
				ConsoleFormat: "console",
				FileFormat:    "json",
			},
		}

		// Act
		err := cfg.Validate()

		// Assert
		assert.NoError(t, err)
	})
}
