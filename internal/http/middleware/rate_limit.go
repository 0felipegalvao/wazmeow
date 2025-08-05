package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"wazmeow/pkg/logger"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	KeyFunc           func(*http.Request) string
}

// DefaultRateLimitConfig returns a default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		KeyFunc: func(r *http.Request) string {
			// Use IP address as default key
			return r.RemoteAddr
		},
	}
}

// rateLimiter implements a token bucket rate limiter
type rateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(requestsPerMinute, burstSize int) *rateLimiter {
	return &rateLimiter{
		tokens:     burstSize,
		maxTokens:  burstSize,
		refillRate: time.Minute / time.Duration(requestsPerMinute),
		lastRefill: time.Now(),
	}
}

// allow checks if a request is allowed
func (rl *rateLimiter) allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Refill tokens based on time elapsed
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	// Check if request is allowed
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware implements rate limiting
func RateLimitMiddleware(config *RateLimitConfig, log logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	limiters := make(map[string]*rateLimiter)
	var mutex sync.RWMutex

	// Cleanup goroutine to remove old limiters
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			// Cleanup old limiters (simplified implementation)
			// In production, you'd implement proper cleanup logic
			// For now, we skip cleanup to avoid empty critical section
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := config.KeyFunc(r)

			// Get or create limiter for this key
			mutex.RLock()
			limiter, exists := limiters[key]
			mutex.RUnlock()

			if !exists {
				mutex.Lock()
				// Double-check after acquiring write lock
				if limiter, exists = limiters[key]; !exists {
					limiter = newRateLimiter(config.RequestsPerMinute, config.BurstSize)
					limiters[key] = limiter
				}
				mutex.Unlock()
			}

			// Check if request is allowed
			if !limiter.allow() {
				log.WarnWithFields("Rate limit exceeded", logger.Fields{
					"key":    key,
					"method": r.Method,
					"path":   r.URL.Path,
				})

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)

				response := `{"success": false, "error": "Rate limit exceeded", "code": "RATE_LIMIT_EXCEEDED"}`
				w.Write([]byte(response))
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))

			next.ServeHTTP(w, r)
		})
	}
}
