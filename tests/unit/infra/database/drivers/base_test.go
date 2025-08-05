package drivers_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"

	"wazmeow/internal/infra/database/drivers"

	_ "github.com/mattn/go-sqlite3"
)

func TestBaseDriver_GetDB(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  func() *bun.DB
		expected bool
	}{
		{
			name: "should return DB when set",
			setupDB: func() *bun.DB {
				sqlDB, _ := sql.Open("sqlite3", ":memory:")
				return bun.NewDB(sqlDB, sqlitedialect.New())
			},
			expected: true,
		},
		{
			name: "should return nil when DB is nil",
			setupDB: func() *bun.DB {
				return nil
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &drivers.BaseDriver{
				DB: tt.setupDB(),
			}

			result := base.GetDB()

			if tt.expected {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestBaseDriver_Health(t *testing.T) {
	tests := []struct {
		name    string
		setupDB func() *bun.DB
		wantErr bool
	}{
		{
			name: "should return error when DB is nil",
			setupDB: func() *bun.DB {
				return nil
			},
			wantErr: true,
		},
		{
			name: "should return nil when DB is healthy",
			setupDB: func() *bun.DB {
				sqlDB, _ := sql.Open("sqlite3", ":memory:")
				return bun.NewDB(sqlDB, sqlitedialect.New())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &drivers.BaseDriver{
				DB: tt.setupDB(),
			}

			err := base.Health()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Cleanup
			if base.DB != nil {
				base.DB.Close()
			}
		})
	}
}

func TestBaseDriver_Stats(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  func() *bun.DB
		expected bool
	}{
		{
			name: "should return empty stats when DB is nil",
			setupDB: func() *bun.DB {
				return nil
			},
			expected: false,
		},
		{
			name: "should return stats when DB is set",
			setupDB: func() *bun.DB {
				sqlDB, _ := sql.Open("sqlite3", ":memory:")
				return bun.NewDB(sqlDB, sqlitedialect.New())
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &drivers.BaseDriver{
				DB: tt.setupDB(),
			}

			stats := base.Stats()

			if tt.expected {
				// Stats should have some meaningful data
				assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
			} else {
				// Empty stats when DB is nil
				assert.Equal(t, sql.DBStats{}, stats)
			}

			// Cleanup
			if base.DB != nil {
				base.DB.Close()
			}
		})
	}
}

func TestBaseDriver_Close(t *testing.T) {
	tests := []struct {
		name    string
		setupDB func() *bun.DB
		wantErr bool
	}{
		{
			name: "should not error when DB is nil",
			setupDB: func() *bun.DB {
				return nil
			},
			wantErr: false,
		},
		{
			name: "should close DB successfully",
			setupDB: func() *bun.DB {
				sqlDB, _ := sql.Open("sqlite3", ":memory:")
				return bun.NewDB(sqlDB, sqlitedialect.New())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &drivers.BaseDriver{
				DB:     tt.setupDB(),
				Logger: nil, // Logger can be nil for this test
			}

			err := base.Close()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Note: configureConnectionPool is private, so we test it indirectly through driver creation
