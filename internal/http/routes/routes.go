package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"wazmeow/internal/http/handler"
	"wazmeow/internal/http/middleware"
	"wazmeow/internal/infra/config"
	"wazmeow/pkg/logger"

	// Import generated docs
	_ "wazmeow/docs"
)

// Router holds all route handlers and dependencies
type Router struct {
	sessionHandler *handler.SessionHandler
	healthHandler  *handler.HealthHandler
	config         *config.Config
	logger         logger.Logger
}

// NewRouter creates a new router with all handlers
func NewRouter(
	sessionHandler *handler.SessionHandler,
	healthHandler *handler.HealthHandler,
	config *config.Config,
	logger logger.Logger,
) *Router {
	return &Router{
		sessionHandler: sessionHandler,
		healthHandler:  healthHandler,
		config:         config,
		logger:         logger,
	}
}

// SetupRoutes configures all routes and middleware
func (rt *Router) SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	rt.setupGlobalMiddleware(r)

	// Health and metrics routes (no auth required)
	rt.setupHealthRoutes(r)

	// Swagger documentation route (no auth required)
	rt.setupSwaggerRoute(r)

	// API routes with authentication
	rt.setupAPIRoutes(r)

	return r
}

// setupGlobalMiddleware configures global middleware
func (rt *Router) setupGlobalMiddleware(r *chi.Mux) {
	// Recovery middleware (should be first)
	r.Use(middleware.RecoveryMiddleware(rt.logger))

	// Request ID middleware
	r.Use(middleware.RequestIDMiddleware())

	// Security headers
	r.Use(middleware.SecurityHeadersMiddleware())

	// CORS middleware
	corsConfig := &middleware.CORSConfig{
		AllowedOrigins:   rt.config.Server.CORS.AllowedOrigins,
		AllowedMethods:   rt.config.Server.CORS.AllowedMethods,
		AllowedHeaders:   rt.config.Server.CORS.AllowedHeaders,
		AllowCredentials: rt.config.Server.CORS.AllowCredentials,
		MaxAge:           rt.config.Server.CORS.MaxAge,
	}
	r.Use(middleware.CORSMiddleware(corsConfig))

	// Logging middleware
	r.Use(middleware.LoggingMiddleware(rt.logger))

	// Rate limiting middleware
	rateLimitConfig := &middleware.RateLimitConfig{
		RequestsPerMinute: rt.config.Server.RateLimit.RequestsPerMinute,
		BurstSize:         rt.config.Server.RateLimit.BurstSize,
		KeyFunc: func(r *http.Request) string {
			return r.RemoteAddr
		},
	}
	r.Use(middleware.RateLimitMiddleware(rateLimitConfig, rt.logger))

	// Content validation middleware
	r.Use(middleware.ValidationMiddleware(rt.logger))
}

// setupHealthRoutes configures health and metrics routes
func (rt *Router) setupHealthRoutes(r *chi.Mux) {
	r.Get("/health", rt.healthHandler.Health)
	r.Get("/metrics", rt.healthHandler.Metrics)
}

// setupAPIRoutes configures API routes with authentication
func (rt *Router) setupAPIRoutes(r *chi.Mux) {
	// Authentication middleware for API routes
	if rt.config.Auth.Enabled {
		switch rt.config.Auth.Type {
		case "api_key":
			authConfig := &middleware.AuthConfig{
				APIKeys:    rt.config.Auth.APIKeys,
				SkipPaths:  []string{"/health", "/metrics"},
				HeaderName: rt.config.Auth.HeaderName,
			}
			r.Use(middleware.AuthMiddleware(authConfig, rt.logger))
		case "basic":
			r.Use(middleware.BasicAuthMiddleware(
				rt.config.Auth.BasicAuth.Username,
				rt.config.Auth.BasicAuth.Password,
				rt.logger,
			))
		}
	}

	// Session routes
	rt.setupSessionRoutes(r)

}

// setupSessionRoutes configures session-related routes
func (rt *Router) setupSessionRoutes(r chi.Router) {
	r.Route("/sessions", func(r chi.Router) {
		// Session CRUD operations
		r.Post("/add", rt.sessionHandler.CreateSession)
		r.Get("/list", rt.sessionHandler.ListSessions)

		// Individual session operations
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/info", rt.sessionHandler.GetSession)
			r.Delete("/", rt.sessionHandler.DeleteSession)

			// Session state operations
			r.Post("/connect", rt.sessionHandler.ConnectSession)
			r.Post("/logout", rt.sessionHandler.LogoutSession)

			// WhatsApp operations for specific session
			r.Get("/qr", rt.sessionHandler.GenerateQR)
			r.Post("/pairphone", rt.sessionHandler.PairPhone)
			r.Post("/proxy/set", rt.sessionHandler.SetProxy)
		})
	})
}

// setupSwaggerRoute configures the Swagger documentation route
func (rt *Router) setupSwaggerRoute(r *chi.Mux) {
	// Swagger documentation route - accessible without authentication
	r.Get("/swagger/*", httpSwagger.WrapHandler)
}
