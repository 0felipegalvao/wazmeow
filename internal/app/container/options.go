package container

import (
	"time"
)

// AppOption defines a functional option for configuring AppContainer
type AppOption func(*AppOptions)

// AppOptions holds configuration options for AppContainer
type AppOptions struct {
	// Startup options
	EnableAutoReconnect     bool
	AutoReconnectTimeout    time.Duration
	MaxConcurrentReconnects int

	// Server options
	StartServerAsync        bool
	GracefulShutdownTimeout time.Duration

	// Logging options
	LogLevel                string
	EnableStructuredLogging bool

	// Development options
	EnableDebugMode bool
	EnableMetrics   bool
	EnableProfiling bool
}

// DefaultAppOptions returns default configuration options
func DefaultAppOptions() *AppOptions {
	return &AppOptions{
		EnableAutoReconnect:     true,
		AutoReconnectTimeout:    2 * time.Minute,
		MaxConcurrentReconnects: 5,
		StartServerAsync:        true,
		GracefulShutdownTimeout: 30 * time.Second,
		LogLevel:                "info",
		EnableStructuredLogging: true,
		EnableDebugMode:         false,
		EnableMetrics:           true,
		EnableProfiling:         false,
	}
}

// WithAutoReconnect enables/disables automatic session reconnection
func WithAutoReconnect(enabled bool) AppOption {
	return func(opts *AppOptions) {
		opts.EnableAutoReconnect = enabled
	}
}

// WithAutoReconnectTimeout sets the timeout for automatic reconnection
func WithAutoReconnectTimeout(timeout time.Duration) AppOption {
	return func(opts *AppOptions) {
		opts.AutoReconnectTimeout = timeout
	}
}

// WithMaxConcurrentReconnects sets the maximum number of concurrent reconnections
func WithMaxConcurrentReconnects(max int) AppOption {
	return func(opts *AppOptions) {
		opts.MaxConcurrentReconnects = max
	}
}

// WithServerAsync enables/disables asynchronous server startup
func WithServerAsync(async bool) AppOption {
	return func(opts *AppOptions) {
		opts.StartServerAsync = async
	}
}

// WithGracefulShutdownTimeout sets the graceful shutdown timeout
func WithGracefulShutdownTimeout(timeout time.Duration) AppOption {
	return func(opts *AppOptions) {
		opts.GracefulShutdownTimeout = timeout
	}
}

// WithLogLevel sets the logging level
func WithLogLevel(level string) AppOption {
	return func(opts *AppOptions) {
		opts.LogLevel = level
	}
}

// WithStructuredLogging enables/disables structured logging
func WithStructuredLogging(enabled bool) AppOption {
	return func(opts *AppOptions) {
		opts.EnableStructuredLogging = enabled
	}
}

// WithDebugMode enables/disables debug mode
func WithDebugMode(enabled bool) AppOption {
	return func(opts *AppOptions) {
		opts.EnableDebugMode = enabled
	}
}

// WithMetrics enables/disables metrics collection
func WithMetrics(enabled bool) AppOption {
	return func(opts *AppOptions) {
		opts.EnableMetrics = enabled
	}
}

// WithProfiling enables/disables profiling
func WithProfiling(enabled bool) AppOption {
	return func(opts *AppOptions) {
		opts.EnableProfiling = enabled
	}
}

// UseCaseOption defines a functional option for configuring UseCaseContainer
type UseCaseOption func(*UseCaseOptions)

// UseCaseOptions holds configuration options for UseCaseContainer
type UseCaseOptions struct {
	EnableCaching    bool
	CacheTTL         time.Duration
	EnableValidation bool
	EnableMetrics    bool
	MaxRetries       int
	RetryDelay       time.Duration
}

// DefaultUseCaseOptions returns default use case options
func DefaultUseCaseOptions() *UseCaseOptions {
	return &UseCaseOptions{
		EnableCaching:    false,
		CacheTTL:         5 * time.Minute,
		EnableValidation: true,
		EnableMetrics:    true,
		MaxRetries:       3,
		RetryDelay:       1 * time.Second,
	}
}

// WithCaching enables/disables caching for use cases
func WithCaching(enabled bool) UseCaseOption {
	return func(opts *UseCaseOptions) {
		opts.EnableCaching = enabled
	}
}

// WithCacheTTL sets the cache TTL
func WithCacheTTL(ttl time.Duration) UseCaseOption {
	return func(opts *UseCaseOptions) {
		opts.CacheTTL = ttl
	}
}

// WithValidation enables/disables validation
func WithValidation(enabled bool) UseCaseOption {
	return func(opts *UseCaseOptions) {
		opts.EnableValidation = enabled
	}
}

// WithUseCaseMetrics enables/disables metrics for use cases
func WithUseCaseMetrics(enabled bool) UseCaseOption {
	return func(opts *UseCaseOptions) {
		opts.EnableMetrics = enabled
	}
}

// WithRetries sets retry configuration
func WithRetries(maxRetries int, delay time.Duration) UseCaseOption {
	return func(opts *UseCaseOptions) {
		opts.MaxRetries = maxRetries
		opts.RetryDelay = delay
	}
}

// HTTPOption defines a functional option for configuring HTTPContainer
type HTTPOption func(*HTTPOptions)

// HTTPOptions holds configuration options for HTTPContainer
type HTTPOptions struct {
	EnableCORS            bool
	EnableRateLimit       bool
	RateLimitRPS          int
	EnableRequestLogging  bool
	EnableResponseLogging bool
	EnableMetrics         bool
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
}

// DefaultHTTPOptions returns default HTTP options
func DefaultHTTPOptions() *HTTPOptions {
	return &HTTPOptions{
		EnableCORS:            true,
		EnableRateLimit:       true,
		RateLimitRPS:          100,
		EnableRequestLogging:  true,
		EnableResponseLogging: false,
		EnableMetrics:         true,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
	}
}

// WithCORS enables/disables CORS
func WithCORS(enabled bool) HTTPOption {
	return func(opts *HTTPOptions) {
		opts.EnableCORS = enabled
	}
}

// WithRateLimit configures rate limiting
func WithRateLimit(enabled bool, rps int) HTTPOption {
	return func(opts *HTTPOptions) {
		opts.EnableRateLimit = enabled
		opts.RateLimitRPS = rps
	}
}

// WithRequestLogging enables/disables request logging
func WithRequestLogging(enabled bool) HTTPOption {
	return func(opts *HTTPOptions) {
		opts.EnableRequestLogging = enabled
	}
}

// WithResponseLogging enables/disables response logging
func WithResponseLogging(enabled bool) HTTPOption {
	return func(opts *HTTPOptions) {
		opts.EnableResponseLogging = enabled
	}
}

// WithHTTPMetrics enables/disables HTTP metrics
func WithHTTPMetrics(enabled bool) HTTPOption {
	return func(opts *HTTPOptions) {
		opts.EnableMetrics = enabled
	}
}

// WithTimeouts sets HTTP server timeouts
func WithTimeouts(read, write, idle time.Duration) HTTPOption {
	return func(opts *HTTPOptions) {
		opts.ReadTimeout = read
		opts.WriteTimeout = write
		opts.IdleTimeout = idle
	}
}
