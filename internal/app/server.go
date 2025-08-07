package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// Server represents the HTTP server
type Server struct {
	container *Container
	server    *http.Server
	router    *mux.Router
}

// NewServer creates a new HTTP server
func NewServer(container *Container) *Server {
	server := &Server{
		container: container,
		router:    mux.NewRouter(),
	}

	server.setupRoutes()
	server.setupHTTPServer()

	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")

	// API v1 routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Sessions routes
	sessions := api.PathPrefix("/sessions").Subrouter()
	sessions.HandleFunc("", s.createSession).Methods("POST")
	sessions.HandleFunc("", s.listSessions).Methods("GET")
	sessions.HandleFunc("/{sessionID}", s.getSession).Methods("GET")
	sessions.HandleFunc("/{sessionID}", s.deleteSession).Methods("DELETE")

	// WhatsApp operations (to be implemented)
	sessions.HandleFunc("/{sessionID}/connect", s.connectSession).Methods("POST")
	sessions.HandleFunc("/{sessionID}/disconnect", s.disconnectSession).Methods("POST")
	sessions.HandleFunc("/{sessionID}/qr", s.getQRCode).Methods("GET")

	log.Info().Msg("Routes configured successfully")
}

// setupHTTPServer configures the HTTP server
func (s *Server) setupHTTPServer() {
	cfg := s.container.Config()

	s.server = &http.Server{
		Addr:              cfg.GetServerAddress(),
		Handler:           s.router,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}
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

// HTTP Handlers

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"wazmeow","version":"2.0.0"}`)
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session creation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session listing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}

func (s *Server) getSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get session
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}

func (s *Server) deleteSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session deletion
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}

func (s *Server) connectSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session connection
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}

func (s *Server) disconnectSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session disconnection
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}

func (s *Server) getQRCode(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement QR code generation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, `{"error":"not implemented yet"}`)
}
