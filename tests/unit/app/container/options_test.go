package container

import (
	"testing"
	"time"

	"wazmeow/internal/app/container"
)

func TestDefaultAppOptions(t *testing.T) {
	opts := container.DefaultAppOptions()
	
	if opts == nil {
		t.Fatal("DefaultAppOptions() returned nil")
	}
	
	// Test default values
	if !opts.EnableAutoReconnect {
		t.Errorf("Expected EnableAutoReconnect to be true, got false")
	}
	
	if opts.AutoReconnectTimeout != 2*time.Minute {
		t.Errorf("Expected AutoReconnectTimeout to be 2m, got %v", opts.AutoReconnectTimeout)
	}
	
	if opts.MaxConcurrentReconnects != 5 {
		t.Errorf("Expected MaxConcurrentReconnects to be 5, got %d", opts.MaxConcurrentReconnects)
	}
	
	if !opts.StartServerAsync {
		t.Errorf("Expected StartServerAsync to be true, got false")
	}
	
	if opts.GracefulShutdownTimeout != 30*time.Second {
		t.Errorf("Expected GracefulShutdownTimeout to be 30s, got %v", opts.GracefulShutdownTimeout)
	}
	
	if opts.LogLevel != "info" {
		t.Errorf("Expected LogLevel to be 'info', got %s", opts.LogLevel)
	}
	
	if !opts.EnableStructuredLogging {
		t.Errorf("Expected EnableStructuredLogging to be true, got false")
	}
	
	if opts.EnableDebugMode {
		t.Errorf("Expected EnableDebugMode to be false, got true")
	}
	
	if !opts.EnableMetrics {
		t.Errorf("Expected EnableMetrics to be true, got false")
	}
	
	if opts.EnableProfiling {
		t.Errorf("Expected EnableProfiling to be false, got true")
	}
}

func TestAppOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []container.AppOption
		validate func(*container.AppOptions) error
	}{
		{
			name: "WithAutoReconnect false",
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.EnableAutoReconnect {
					t.Errorf("Expected EnableAutoReconnect to be false, got true")
				}
				return nil
			},
		},
		{
			name: "WithAutoReconnectTimeout",
			opts: []container.AppOption{
				container.WithAutoReconnectTimeout(1 * time.Minute),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.AutoReconnectTimeout != 1*time.Minute {
					t.Errorf("Expected AutoReconnectTimeout to be 1m, got %v", opts.AutoReconnectTimeout)
				}
				return nil
			},
		},
		{
			name: "WithMaxConcurrentReconnects",
			opts: []container.AppOption{
				container.WithMaxConcurrentReconnects(10),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.MaxConcurrentReconnects != 10 {
					t.Errorf("Expected MaxConcurrentReconnects to be 10, got %d", opts.MaxConcurrentReconnects)
				}
				return nil
			},
		},
		{
			name: "WithServerAsync false",
			opts: []container.AppOption{
				container.WithServerAsync(false),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.StartServerAsync {
					t.Errorf("Expected StartServerAsync to be false, got true")
				}
				return nil
			},
		},
		{
			name: "WithGracefulShutdownTimeout",
			opts: []container.AppOption{
				container.WithGracefulShutdownTimeout(10 * time.Second),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.GracefulShutdownTimeout != 10*time.Second {
					t.Errorf("Expected GracefulShutdownTimeout to be 10s, got %v", opts.GracefulShutdownTimeout)
				}
				return nil
			},
		},
		{
			name: "WithLogLevel",
			opts: []container.AppOption{
				container.WithLogLevel("debug"),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.LogLevel != "debug" {
					t.Errorf("Expected LogLevel to be 'debug', got %s", opts.LogLevel)
				}
				return nil
			},
		},
		{
			name: "WithStructuredLogging false",
			opts: []container.AppOption{
				container.WithStructuredLogging(false),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.EnableStructuredLogging {
					t.Errorf("Expected EnableStructuredLogging to be false, got true")
				}
				return nil
			},
		},
		{
			name: "WithDebugMode true",
			opts: []container.AppOption{
				container.WithDebugMode(true),
			},
			validate: func(opts *container.AppOptions) error {
				if !opts.EnableDebugMode {
					t.Errorf("Expected EnableDebugMode to be true, got false")
				}
				return nil
			},
		},
		{
			name: "WithMetrics false",
			opts: []container.AppOption{
				container.WithMetrics(false),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.EnableMetrics {
					t.Errorf("Expected EnableMetrics to be false, got true")
				}
				return nil
			},
		},
		{
			name: "WithProfiling true",
			opts: []container.AppOption{
				container.WithProfiling(true),
			},
			validate: func(opts *container.AppOptions) error {
				if !opts.EnableProfiling {
					t.Errorf("Expected EnableProfiling to be true, got false")
				}
				return nil
			},
		},
		{
			name: "Multiple options",
			opts: []container.AppOption{
				container.WithAutoReconnect(false),
				container.WithDebugMode(true),
				container.WithMetrics(false),
				container.WithLogLevel("debug"),
			},
			validate: func(opts *container.AppOptions) error {
				if opts.EnableAutoReconnect {
					t.Errorf("Expected EnableAutoReconnect to be false, got true")
				}
				if !opts.EnableDebugMode {
					t.Errorf("Expected EnableDebugMode to be true, got false")
				}
				if opts.EnableMetrics {
					t.Errorf("Expected EnableMetrics to be false, got true")
				}
				if opts.LogLevel != "debug" {
					t.Errorf("Expected LogLevel to be 'debug', got %s", opts.LogLevel)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := container.DefaultAppOptions()
			
			for _, opt := range tt.opts {
				opt(options)
			}
			
			if err := tt.validate(options); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

func TestUseCaseOptions(t *testing.T) {
	opts := container.DefaultUseCaseOptions()
	
	if opts == nil {
		t.Fatal("DefaultUseCaseOptions() returned nil")
	}
	
	// Test default values
	if opts.EnableCaching {
		t.Errorf("Expected EnableCaching to be false, got true")
	}
	
	if opts.CacheTTL != 5*time.Minute {
		t.Errorf("Expected CacheTTL to be 5m, got %v", opts.CacheTTL)
	}
	
	if !opts.EnableValidation {
		t.Errorf("Expected EnableValidation to be true, got false")
	}
	
	if !opts.EnableMetrics {
		t.Errorf("Expected EnableMetrics to be true, got false")
	}
	
	if opts.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", opts.MaxRetries)
	}
	
	if opts.RetryDelay != 1*time.Second {
		t.Errorf("Expected RetryDelay to be 1s, got %v", opts.RetryDelay)
	}
}

func TestHTTPOptions(t *testing.T) {
	opts := container.DefaultHTTPOptions()
	
	if opts == nil {
		t.Fatal("DefaultHTTPOptions() returned nil")
	}
	
	// Test default values
	if !opts.EnableCORS {
		t.Errorf("Expected EnableCORS to be true, got false")
	}
	
	if !opts.EnableRateLimit {
		t.Errorf("Expected EnableRateLimit to be true, got false")
	}
	
	if opts.RateLimitRPS != 100 {
		t.Errorf("Expected RateLimitRPS to be 100, got %d", opts.RateLimitRPS)
	}
	
	if !opts.EnableRequestLogging {
		t.Errorf("Expected EnableRequestLogging to be true, got false")
	}
	
	if opts.EnableResponseLogging {
		t.Errorf("Expected EnableResponseLogging to be false, got true")
	}
	
	if !opts.EnableMetrics {
		t.Errorf("Expected EnableMetrics to be true, got false")
	}
	
	if opts.ReadTimeout != 30*time.Second {
		t.Errorf("Expected ReadTimeout to be 30s, got %v", opts.ReadTimeout)
	}
	
	if opts.WriteTimeout != 30*time.Second {
		t.Errorf("Expected WriteTimeout to be 30s, got %v", opts.WriteTimeout)
	}
	
	if opts.IdleTimeout != 120*time.Second {
		t.Errorf("Expected IdleTimeout to be 120s, got %v", opts.IdleTimeout)
	}
}
