package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"wazmeow/internal/domain"
	"wazmeow/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// SessionHandler handles HTTP requests for session operations
type SessionHandler struct {
	createSessionUC     *services.CreateSessionUseCase
	multiSessionManager *services.MultiSessionManager
	sessionRepo         domain.Repository
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	createSessionUC *services.CreateSessionUseCase,
	multiSessionManager *services.MultiSessionManager,
	sessionRepo domain.Repository,
) *SessionHandler {
	return &SessionHandler{
		createSessionUC:     createSessionUC,
		multiSessionManager: multiSessionManager,
		sessionRepo:         sessionRepo,
	}
}

// CreateSession handles POST /sessions/add
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req services.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode create session request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.createSessionUC.Execute(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create session")

		// Handle different error types
		switch err.(type) {
		case *domain.ValidationError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case *domain.AlreadyExistsError:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListSessions handles GET /sessions/list
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.sessionRepo.List(r.Context(), nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list sessions")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response []map[string]any
	for _, session := range sessions {
		response = append(response, map[string]any{
			"id":         session.ID.String(),
			"name":       session.Name,
			"status":     string(session.Status),
			"created_at": session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"updated_at": session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"sessions": response,
		"total":    len(response),
	})
}

// GetSessionInfo handles GET /sessions/{sessionID}/info
func (h *SessionHandler) GetSessionInfo(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := h.sessionRepo.GetByID(r.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to get session")

		switch err.(type) {
		case *domain.NotFoundError:
			http.Error(w, "Session not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]any{
		"id":         session.ID.String(),
		"name":       session.Name,
		"status":     string(session.Status),
		"created_at": session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updated_at": session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteSession handles DELETE /sessions/{sessionID}
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	err := h.sessionRepo.Delete(r.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionIDStr).Msg("Failed to delete session")

		switch err.(type) {
		case *domain.NotFoundError:
			http.Error(w, "Session not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ConnectSession handles POST /sessions/{sessionID}/connect
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Log connection attempt
	log.Info().
		Str("session_id", sessionIDStr).
		Str("remote_addr", r.RemoteAddr).
		Msg("Session connection requested")

	// Start session using MultiSessionManager
	err := h.multiSessionManager.StartSession(r.Context(), sessionID)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Msg("Failed to start session")

		// Handle different error types
		switch err.(type) {
		case *domain.NotFoundError:
			http.Error(w, "Session not found", http.StatusNotFound)
		case *domain.BusinessError:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Failed to connect session", http.StatusInternalServerError)
		}
		return
	}

	// Get session status
	status := h.multiSessionManager.GetSessionStatus(sessionID)

	response := map[string]any{
		"session_id": sessionIDStr,
		"status":     string(status),
		"message":    "Session connection initiated successfully",
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("status", string(status)).
		Msg("Session connection initiated")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutSession handles POST /sessions/{sessionID}/logout
func (h *SessionHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Log logout attempt
	log.Info().
		Str("session_id", sessionIDStr).
		Str("remote_addr", r.RemoteAddr).
		Msg("Session logout requested")

	// Stop session using MultiSessionManager
	err := h.multiSessionManager.StopSession(r.Context(), sessionID)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Msg("Failed to stop session")

		// Handle different error types
		switch err.(type) {
		case *domain.NotFoundError:
			http.Error(w, "Session not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to logout session", http.StatusInternalServerError)
		}
		return
	}

	// Update session status in database to disconnected
	session, err := h.sessionRepo.GetByID(r.Context(), sessionID)
	if err == nil {
		session.Status = domain.StatusDisconnected
		session.QRCode = "" // Clear QR code
		if updateErr := h.sessionRepo.Update(r.Context(), session); updateErr != nil {
			log.Error().
				Err(updateErr).
				Str("session_id", sessionIDStr).
				Msg("Failed to update session status in database")
		}
	}

	response := map[string]any{
		"session_id": sessionIDStr,
		"status":     "disconnected",
		"message":    "Session logged out successfully",
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Msg("Session logged out successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetQRCode handles GET /sessions/{sessionID}/qr
func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Log QR code request
	log.Info().
		Str("session_id", sessionIDStr).
		Str("remote_addr", r.RemoteAddr).
		Msg("QR code generation requested")

	// Create context with timeout for QR code generation
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Generate QR code using MultiSessionManager
	qrCode, err := h.multiSessionManager.GenerateQRCode(ctx, sessionID)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Msg("Failed to generate QR code")

		// Handle different error types
		switch {
		case strings.Contains(err.Error(), "not found"):
			http.Error(w, "Session not found", http.StatusNotFound)
		case strings.Contains(err.Error(), "already connected"):
			http.Error(w, "Session is already connected", http.StatusConflict)
		case ctx.Err() == context.DeadlineExceeded:
			http.Error(w, "QR code generation timeout", http.StatusRequestTimeout)
		default:
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		}
		return
	}

	// Calculate expiration time (QR codes typically expire in 20 seconds)
	expiresAt := time.Now().Add(20 * time.Second)

	response := map[string]any{
		"session_id": sessionIDStr,
		"qr_code":    qrCode,
		"expires_at": expiresAt.Format(time.RFC3339),
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Time("expires_at", expiresAt).
		Msg("QR code generated successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PairPhone handles POST /sessions/{sessionID}/pairphone
func (h *SessionHandler) PairPhone(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate phone number format (basic validation)
	if req.PhoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	// Remove non-numeric characters and validate
	phoneNumber := strings.ReplaceAll(req.PhoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")

	if len(phoneNumber) < 10 || len(phoneNumber) > 15 {
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Log pairing attempt
	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone_number", phoneNumber).
		Str("remote_addr", r.RemoteAddr).
		Msg("Phone pairing requested")

	// Create context with timeout for pairing
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Initiate phone pairing using MultiSessionManager
	linkingCode, err := h.multiSessionManager.PairPhone(ctx, sessionID, phoneNumber)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", sessionIDStr).
			Str("phone_number", phoneNumber).
			Msg("Failed to initiate phone pairing")

		// Handle different error types
		switch {
		case strings.Contains(err.Error(), "not found"):
			http.Error(w, "Session not found", http.StatusNotFound)
		case strings.Contains(err.Error(), "already connected"):
			http.Error(w, "Session is already connected", http.StatusConflict)
		case ctx.Err() == context.DeadlineExceeded:
			http.Error(w, "Phone pairing timeout", http.StatusRequestTimeout)
		default:
			http.Error(w, "Failed to initiate phone pairing", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]any{
		"session_id":   sessionIDStr,
		"phone_number": phoneNumber,
		"linking_code": linkingCode,
		"status":       "pairing_initiated",
		"message":      "Enter the linking code on your phone to complete pairing",
	}

	log.Info().
		Str("session_id", sessionIDStr).
		Str("phone_number", phoneNumber).
		Str("linking_code", linkingCode).
		Msg("Phone pairing initiated successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetProxy handles POST /sessions/{sessionID}/proxy/set
func (h *SessionHandler) SetProxy(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionID")

	sessionID := domain.SessionID(sessionIDStr)
	if !sessionID.IsValid() {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ProxyURL string `json:"proxy_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Proxy configuration is not yet implemented
	// This endpoint is reserved for future proxy functionality
	response := map[string]any{
		"session_id": sessionIDStr,
		"proxy_url":  req.ProxyURL,
		"status":     "proxy_not_implemented",
		"message":    "Proxy configuration is not yet implemented",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
