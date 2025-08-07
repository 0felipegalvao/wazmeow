package router

import (
	"net/http"

	"wazmeow/internal/handlers"
	"wazmeow/internal/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Router holds all the route handlers
type Router struct {
	sessionHandler *handlers.SessionHandler
	messageHandler *handlers.MessageHandler
}

// NewRouter creates a new router instance
func NewRouter(sessionHandler *handlers.SessionHandler, messageHandler *handlers.MessageHandler) *Router {
	return &Router{
		sessionHandler: sessionHandler,
		messageHandler: messageHandler,
	}
}

// SetupRoutes configures all the HTTP routes
func (rt *Router) SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Add Chi built-in middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Compress(5))

	// Add custom middleware
	r.Use(middleware.LoggingMiddleware)

	// Setup CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure this properly for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", rt.healthCheck)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		rt.setupSessionRoutes(r)
		rt.setupMessageRoutes(r)
	})

	return r
}

// setupSessionRoutes configures session-related routes
func (rt *Router) setupSessionRoutes(r chi.Router) {
	r.Route("/sessions", func(r chi.Router) {
		// Session management
		r.Post("/add", rt.sessionHandler.CreateSession)
		r.Get("/list", rt.sessionHandler.ListSessions)

		// Session-specific routes
		r.Route("/{sessionID}", func(r chi.Router) {
			r.Get("/info", rt.sessionHandler.GetSessionInfo)
			r.Delete("/", rt.sessionHandler.DeleteSession)

			// Session operations
			r.Post("/connect", rt.sessionHandler.ConnectSession)
			r.Post("/logout", rt.sessionHandler.LogoutSession)
			r.Get("/qr", rt.sessionHandler.GetQRCode)
			r.Post("/pairphone", rt.sessionHandler.PairPhone)
			r.Post("/proxy/set", rt.sessionHandler.SetProxy)
		})
	})
}

// setupMessageRoutes configures message-related routes
func (rt *Router) setupMessageRoutes(r chi.Router) {
	r.Route("/message/{sessionId}", func(r chi.Router) {
		// Text messages
		r.Post("/send/text", rt.messageHandler.SendTextMessage)

		// Media messages (not implemented yet)
		r.Post("/send/image", rt.messageHandler.SendImageMessage)
		r.Post("/send/audio", rt.messageHandler.SendAudioMessage)
		r.Post("/send/video", rt.messageHandler.SendVideoMessage)
		r.Post("/send/document", rt.messageHandler.SendDocumentMessage)

		// Special messages (not implemented yet)
		r.Post("/send/location", rt.messageHandler.SendLocationMessage)
		r.Post("/send/contact", rt.messageHandler.SendContactMessage)
	})
}

// healthCheck provides a simple health check endpoint
func (rt *Router) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"wazmeow"}`))
}
