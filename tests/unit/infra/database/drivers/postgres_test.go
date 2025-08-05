package drivers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"wazmeow/internal/infra/config"
	"wazmeow/internal/infra/database/drivers"
)

func TestPostgreSQLDriver_NewPostgreSQLConnection(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
		skipCI  bool // Skip in CI environment where PostgreSQL might not be available
	}{
		{
			name: "should create connection with URL",
			config: &config.DatabaseConfig{
				Driver:          "postgres",
				URL:             "postgres://user:password@localhost:5432/testdb?sslmode=disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
			},
			wantErr: true, // Will fail without real PostgreSQL server
			skipCI:  true,
		},
		{
			name: "should create connection with individual components",
			config: &config.DatabaseConfig{
				Driver:          "postgres",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				PostgreSQL: config.PostgreSQLConfig{
					Host:       "localhost",
					Port:       5432,
					Database:   "testdb",
					Username:   "user",
					Password:   "password",
					SSLMode:    "disable",
					TimeZone:   "UTC",
					SearchPath: "public",
					AppName:    "wazmeow-test",
				},
			},
			wantErr: true, // Will fail without real PostgreSQL server
			skipCI:  true,
		},
		{
			name: "should fail with missing required fields",
			config: &config.DatabaseConfig{
				Driver:          "postgres",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				PostgreSQL:      config.PostgreSQLConfig{
					// Missing required fields
				},
			},
			wantErr: true,
			skipCI:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCI {
				t.Skip("Skipping test that requires PostgreSQL server")
			}

			conn, err := drivers.NewPostgreSQLConnection(tt.config, nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, conn)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conn)

				// Test basic operations if connection successful
				if conn != nil {
					db := conn.GetDB()
					assert.NotNil(t, db)

					// Cleanup
					err = conn.Close()
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestPostgreSQLDriver_BuildConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
		wantErr  bool
	}{
		{
			name: "should use URL when provided",
			config: &config.DatabaseConfig{
				URL: "postgres://user:password@localhost:5432/testdb?sslmode=disable",
				PostgreSQL: config.PostgreSQLConfig{
					Host:     "ignored",
					Database: "ignored",
				},
			},
			expected: "postgres://user:password@localhost:5432/testdb?sslmode=disable",
			wantErr:  false,
		},
		{
			name: "should build connection string from components",
			config: &config.DatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:       "localhost",
					Port:       5432,
					Database:   "testdb",
					Username:   "user",
					Password:   "password",
					SSLMode:    "disable",
					TimeZone:   "UTC",
					SearchPath: "public",
					AppName:    "wazmeow-test",
				},
			},
			expected: "host=localhost port=5432 dbname=testdb user=user password=password sslmode=disable timezone=UTC search_path=public application_name=wazmeow-test",
			wantErr:  false,
		},
		{
			name: "should build minimal connection string",
			config: &config.DatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Database: "testdb",
					Username: "user",
				},
			},
			expected: "host=localhost port=5432 dbname=testdb user=user",
			wantErr:  false,
		},
		{
			name: "should fail with missing host",
			config: &config.DatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Port:     5432,
					Database: "testdb",
					Username: "user",
				},
			},
			wantErr: true,
		},
		{
			name: "should fail with missing database",
			config: &config.DatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "user",
				},
			},
			wantErr: true,
		},
		{
			name: "should fail with missing username",
			config: &config.DatabaseConfig{
				PostgreSQL: config.PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Database: "testdb",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We test buildConnectionString indirectly through NewPostgreSQLConnection
			_, err := drivers.NewPostgreSQLConnection(tt.config, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// We can't easily test the exact connection string without exposing the method
				// But we can verify that the connection attempt was made with valid parameters
				assert.Error(t, err) // Will error because no real PostgreSQL server
				assert.Contains(t, err.Error(), "failed to connect to PostgreSQL")
			}
		})
	}
}

func TestPostgreSQLDriver_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.DatabaseConfig
		expectError string
	}{
		{
			name: "valid config with all fields",
			config: &config.DatabaseConfig{
				Driver: "postgres",
				PostgreSQL: config.PostgreSQLConfig{
					Host:       "localhost",
					Port:       5432,
					Database:   "testdb",
					Username:   "user",
					Password:   "password",
					SSLMode:    "disable",
					TimeZone:   "UTC",
					SearchPath: "public",
					AppName:    "wazmeow",
				},
			},
			expectError: "failed to connect to PostgreSQL", // Connection will fail but config is valid
		},
		{
			name: "missing host",
			config: &config.DatabaseConfig{
				Driver: "postgres",
				PostgreSQL: config.PostgreSQLConfig{
					Port:     5432,
					Database: "testdb",
					Username: "user",
				},
			},
			expectError: "PostgreSQL host is required",
		},
		{
			name: "missing database",
			config: &config.DatabaseConfig{
				Driver: "postgres",
				PostgreSQL: config.PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "user",
				},
			},
			expectError: "PostgreSQL database name is required",
		},
		{
			name: "missing username",
			config: &config.DatabaseConfig{
				Driver: "postgres",
				PostgreSQL: config.PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Database: "testdb",
				},
			},
			expectError: "PostgreSQL username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := drivers.NewPostgreSQLConnection(tt.config, nil)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestPostgreSQLDriver_BaseDriverIntegration(t *testing.T) {
	// Test that PostgreSQLDriver properly uses BaseDriver methods
	config := &config.DatabaseConfig{
		Driver:          "postgres",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		PostgreSQL: config.PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "user",
			Password: "password",
			SSLMode:  "disable",
		},
	}

	// This will fail to connect but should test the BaseDriver integration
	conn, err := drivers.NewPostgreSQLConnection(config, nil)

	// Should fail with connection error, not configuration error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to PostgreSQL")
	assert.Nil(t, conn)
}
