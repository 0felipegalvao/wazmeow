package app

import (
	"os"
	"testing"
	"time"

	"wazmeow/internal/app"
	"wazmeow/internal/app/container"
)

// setupTestEnv configures environment variables for testing to use tests/unit/data
func setupTestEnv() {
	// Clear any existing environment variables that might interfere
	os.Unsetenv("DB_URL")
	os.Unsetenv("SQLITE_PATH")
	os.Unsetenv("DATABASE_PATH")

	// Set test environment variables to use tests/unit/data
	os.Setenv("DB_URL", "../data/wazmeow.db")
	os.Setenv("SQLITE_PATH", "../data/wazmeow.db")
	os.Setenv("DB_MAX_OPEN_CONNS", "1")
	os.Setenv("DB_MAX_IDLE_CONNS", "1")
}

// cleanupTestEnv cleans up test environment variables
func cleanupTestEnv() {
	os.Unsetenv("DB_URL")
	os.Unsetenv("SQLITE_PATH")
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("DB_MAX_OPEN_CONNS")
	os.Unsetenv("DB_MAX_IDLE_CONNS")
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []container.AppOption
		wantErr bool
	}{
		{
			name:    "successful creation with default options",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "successful creation with custom options",
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
				container.WithDebugMode(true),
				container.WithMetrics(false),
			},
			wantErr: false,
		},
		{
			name: "with auto reconnect disabled",
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
			},
			wantErr: false,
		},
		{
			name: "with custom timeouts",
			opts: []container.AppOption{
				container.WithAutoReconnectTimeout(1 * time.Minute),
				container.WithGracefulShutdownTimeout(10 * time.Second),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment for each subtest
			setupTestEnv()
			defer cleanupTestEnv()

			application, err := app.New(tt.opts...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}

			if application == nil {
				t.Errorf("New() returned nil app")
				return
			}

			// Verify app has container
			if application.GetContainer() == nil {
				t.Errorf("New() app container is nil")
			}

			// Verify app has config
			if application.GetConfig() == nil {
				t.Errorf("New() app config is nil")
			}

			// Clean up
			if err := application.Stop(); err != nil {
				t.Errorf("Failed to stop app: %v", err)
			}
		})
	}
}

func TestApp_Health(t *testing.T) {
	// Setup test environment to use in-memory database
	setupTestEnv()
	defer cleanupTestEnv()

	application, err := app.New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer application.Stop()

	// Start WhatsApp manager for health check
	if err := application.GetContainer().StartWhatsAppManager(); err != nil {
		t.Fatalf("Failed to start WhatsApp manager: %v", err)
	}

	// Test health check
	if err := application.Health(); err != nil {
		t.Errorf("Health() returned error: %v", err)
	}
}

func TestApp_GetMethods(t *testing.T) {
	// Setup test environment to use in-memory database
	setupTestEnv()
	defer cleanupTestEnv()

	application, err := app.New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer application.Stop()

	// Test GetConfig
	if got := application.GetConfig(); got == nil {
		t.Errorf("GetConfig() returned nil")
	}

	// Test GetContainer
	if got := application.GetContainer(); got == nil {
		t.Errorf("GetContainer() returned nil")
	}

	// Test GetServerInfo
	if got := application.GetServerInfo(); got == nil {
		t.Errorf("GetServerInfo() returned nil")
	}

	// Test GetStats
	if got := application.GetStats(); got == nil {
		t.Errorf("GetStats() returned nil")
	}
}

func TestApp_StartStop(t *testing.T) {
	// This test is more complex as it involves starting the actual server
	// For unit tests, we might want to mock the dependencies
	t.Skip("Skipping integration test - requires full infrastructure setup")

	// Setup test environment to use in-memory database
	setupTestEnv()
	defer cleanupTestEnv()

	application, err := app.New(
		container.WithAutoReconnect(false), // Disable auto-reconnect for faster testing
	)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Start app in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- application.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the app
	if err := application.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Wait for start to complete
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Start() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Start() did not complete within timeout")
	}
}

func TestApp_WithOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []container.AppOption
	}{
		{
			name: "with auto reconnect disabled",
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
			},
		},
		{
			name: "with debug mode enabled",
			opts: []container.AppOption{
				container.WithDebugMode(true),
			},
		},
		{
			name: "with custom timeouts",
			opts: []container.AppOption{
				container.WithAutoReconnectTimeout(1 * time.Minute),
				container.WithGracefulShutdownTimeout(10 * time.Second),
			},
		},
		{
			name: "with metrics disabled",
			opts: []container.AppOption{
				container.WithMetrics(false),
			},
		},
		{
			name: "with multiple options",
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
				container.WithDebugMode(true),
				container.WithMetrics(false),
				container.WithLogLevel("debug"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment for each subtest
			setupTestEnv()
			defer cleanupTestEnv()

			application, err := app.New(tt.opts...)
			if err != nil {
				t.Errorf("New() with options failed: %v", err)
				return
			}

			// Verify app was created successfully
			if application == nil {
				t.Errorf("New() returned nil app")
				return
			}

			// Verify container is initialized
			if !application.GetContainer().IsInitialized() {
				t.Errorf("App container not initialized")
			}

			// Clean up
			if err := application.Stop(); err != nil {
				t.Errorf("Failed to stop app: %v", err)
			}
		})
	}
}

func TestApp_Stop(t *testing.T) {
	// Setup test environment
	setupTestEnv()
	defer cleanupTestEnv()

	application, err := app.New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Test Stop
	if err := application.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Test multiple stops (should not error)
	if err := application.Stop(); err != nil {
		t.Errorf("Second Stop() returned error: %v", err)
	}
}
