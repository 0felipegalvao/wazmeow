package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"wazmeow/internal/handlers"
	"wazmeow/internal/router"

	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	container *Container
	server    *http.Server
	handler   http.Handler
}

// NewServer creates a new HTTP server
func NewServer(container *Container) *Server {
	// Create session handler
	sessionHandler := handlers.NewSessionHandler(
		container.CreateSessionUseCase(),
		container.MultiSessionManager(),
		container.SessionRepository(),
	)

	messageHandler := handlers.NewMessageHandler(
		container.MultiSessionManager(),
	)

	// Setup router
	appRouter := router.NewRouter(sessionHandler, messageHandler)
	handler := appRouter.SetupRoutes()

	server := &Server{
		container: container,
		handler:   handler,
	}

	server.setupHTTPServer()

	return server
}

// setupHTTPServer configures the HTTP server
func (s *Server) setupHTTPServer() {
	cfg := s.container.Config()

	s.server = &http.Server{
		Addr:              cfg.GetServerAddress(),
		Handler:           s.handler,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	log.Info().Msg("HTTP server configured successfully")
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	cfg := s.container.Config()

	// Start server in a goroutine
	go func() {
		var err error
		if cfg.Server.TLS.Enabled {
			log.Info().
				Str("address", cfg.GetServerAddress()).
				Bool("tls", true).
				Msg("Starting HTTPS server")
			err = s.server.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			log.Info().
				Str("address", cfg.GetServerAddress()).
				Bool("tls", false).
				Msg("Starting HTTP server")
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for context cancellation (shutdown signal)
	<-ctx.Done()

	log.Info().Msg("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info().Msg("Server stopped gracefully")
	return nil
}
