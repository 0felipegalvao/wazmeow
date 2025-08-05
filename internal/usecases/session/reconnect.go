package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/domain/whatsapp"
	"wazmeow/pkg/logger"
)

// AutoReconnectUseCase handles automatic reconnection of sessions during startup
type AutoReconnectUseCase struct {
	sessionRepo session.Repository
	waManager   whatsapp.Manager
	logger      logger.Logger
}

// NewAutoReconnectUseCase creates a new auto reconnect use case
func NewAutoReconnectUseCase(
	sessionRepo session.Repository,
	waManager whatsapp.Manager,
	logger logger.Logger,
) *AutoReconnectUseCase {
	return &AutoReconnectUseCase{
		sessionRepo: sessionRepo,
		waManager:   waManager,
		logger:      logger,
	}
}

// AutoReconnectRequest represents the request for auto reconnection
type AutoReconnectRequest struct {
	MaxConcurrentReconnections int
	ReconnectionTimeout        time.Duration
}

// AutoReconnectResponse represents the response from auto reconnection
type AutoReconnectResponse struct {
	TotalSessions           int
	EligibleSessions        int
	SuccessfulReconnections int
	FailedReconnections     int
	ReconnectionResults     []SessionReconnectionResult
}

// SessionReconnectionResult represents the result of a single session reconnection
type SessionReconnectionResult struct {
	SessionID   session.SessionID
	SessionName string
	Success     bool
	Error       string
	Duration    time.Duration
}

// Execute performs automatic reconnection of eligible sessions
func (uc *AutoReconnectUseCase) Execute(ctx context.Context, req AutoReconnectRequest) (*AutoReconnectResponse, error) {
	startTime := time.Now()

	uc.logger.Info("🔄 Starting automatic session reconnection process")

	// Set defaults if not provided
	if req.MaxConcurrentReconnections <= 0 {
		req.MaxConcurrentReconnections = 5
	}
	if req.ReconnectionTimeout <= 0 {
		req.ReconnectionTimeout = 30 * time.Second
	}

	// Find eligible sessions for reconnection
	eligibleSessions, err := uc.findEligibleSessions(ctx)
	if err != nil {
		uc.logger.ErrorWithError("failed to find eligible sessions for reconnection", err, nil)
		return nil, fmt.Errorf("failed to find eligible sessions: %w", err)
	}

	totalSessions := len(eligibleSessions)
	uc.logger.InfoWithFields("found eligible sessions for reconnection", logger.Fields{
		"total_sessions": totalSessions,
	})

	if totalSessions == 0 {
		uc.logger.Info("no sessions eligible for reconnection")
		return &AutoReconnectResponse{
			TotalSessions:           0,
			EligibleSessions:        0,
			SuccessfulReconnections: 0,
			FailedReconnections:     0,
			ReconnectionResults:     []SessionReconnectionResult{},
		}, nil
	}

	// Perform reconnections with concurrency control
	results := uc.performReconnections(ctx, eligibleSessions, req.MaxConcurrentReconnections, req.ReconnectionTimeout)

	// Count successful and failed reconnections
	successCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	duration := time.Since(startTime)
	uc.logger.InfoWithFields("automatic reconnection process completed", logger.Fields{
		"total_sessions":           totalSessions,
		"eligible_sessions":        totalSessions,
		"successful_reconnections": successCount,
		"failed_reconnections":     failedCount,
		"duration_ms":              duration.Milliseconds(),
	})

	return &AutoReconnectResponse{
		TotalSessions:           totalSessions,
		EligibleSessions:        totalSessions,
		SuccessfulReconnections: successCount,
		FailedReconnections:     failedCount,
		ReconnectionResults:     results,
	}, nil
}

// findEligibleSessions finds sessions that are eligible for automatic reconnection
func (uc *AutoReconnectUseCase) findEligibleSessions(ctx context.Context) ([]*session.Session, error) {
	var eligibleSessions []*session.Session

	// Find sessions with status "connected" or "connecting"
	connectedSessions, _, err := uc.sessionRepo.GetByStatus(ctx, session.StatusConnected, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected sessions: %w", err)
	}

	connectingSessions, _, err := uc.sessionRepo.GetByStatus(ctx, session.StatusConnecting, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get connecting sessions: %w", err)
	}

	// Combine and filter sessions
	allSessions := append(connectedSessions, connectingSessions...)

	for _, sess := range allSessions {
		// Check eligibility criteria
		if uc.isSessionEligibleForReconnection(sess) {
			eligibleSessions = append(eligibleSessions, sess)

			uc.logger.InfoWithFields("session eligible for reconnection", logger.Fields{
				"session_id":   sess.ID().String(),
				"session_name": sess.Name(),
				"status":       sess.Status().String(),
				"wa_jid":       sess.WaJID(),
				"is_active":    sess.IsActive(),
			})
		} else {
			uc.logger.InfoWithFields("session not eligible for reconnection", logger.Fields{
				"session_id":   sess.ID().String(),
				"session_name": sess.Name(),
				"status":       sess.Status().String(),
				"wa_jid":       sess.WaJID(),
				"is_active":    sess.IsActive(),
			})
		}
	}

	return eligibleSessions, nil
}

// isSessionEligibleForReconnection checks if a session meets the criteria for automatic reconnection
func (uc *AutoReconnectUseCase) isSessionEligibleForReconnection(sess *session.Session) bool {
	// Criteria for reconnection:
	// 1. Status is "connected" or "connecting"
	// 2. Has WhatsApp JID (wa_jid is not empty) - indicates previous successful authentication
	// 3. is_active is true

	hasValidStatus := sess.Status() == session.StatusConnected || sess.Status() == session.StatusConnecting
	hasWaJID := sess.WaJID() != ""
	isActive := sess.IsActive()

	return hasValidStatus && hasWaJID && isActive
}

// performReconnections performs the actual reconnection process with concurrency control
func (uc *AutoReconnectUseCase) performReconnections(
	ctx context.Context,
	sessions []*session.Session,
	maxConcurrent int,
	timeout time.Duration,
) []SessionReconnectionResult {
	results := make([]SessionReconnectionResult, len(sessions))

	// Create a semaphore to limit concurrent reconnections
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, sess := range sessions {
		wg.Add(1)
		go func(index int, session *session.Session) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Perform reconnection with timeout
			result := uc.reconnectSession(ctx, session, timeout)
			results[index] = result
		}(i, sess)
	}

	wg.Wait()
	return results
}

// reconnectSession attempts to reconnect a single session
func (uc *AutoReconnectUseCase) reconnectSession(
	ctx context.Context,
	sess *session.Session,
	timeout time.Duration,
) SessionReconnectionResult {
	startTime := time.Now()
	sessionID := sess.ID()
	sessionName := sess.Name()

	uc.logger.InfoWithFields("🔌 attempting to reconnect session", logger.Fields{
		"session_id":   sessionID.String(),
		"session_name": sessionName,
		"wa_jid":       sess.WaJID(),
	})

	// Create context with timeout
	reconnectCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Try to get existing client first
	waClient, err := uc.waManager.GetClient(sessionID)
	if err != nil {
		// Create new client if it doesn't exist
		uc.logger.InfoWithFields("creating new WhatsApp client for reconnection", logger.Fields{
			"session_id": sessionID.String(),
		})

		waClient, err = uc.waManager.CreateClient(sessionID)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to create WhatsApp client: %v", err)
			uc.logger.ErrorWithError("failed to create WhatsApp client for reconnection", err, logger.Fields{
				"session_id": sessionID.String(),
			})

			// Update session status to disconnected
			sess.Disconnect()
			uc.sessionRepo.Update(ctx, sess)

			return SessionReconnectionResult{
				SessionID:   sessionID,
				SessionName: sessionName,
				Success:     false,
				Error:       errorMsg,
				Duration:    time.Since(startTime),
			}
		}
	}

	// Attempt to connect
	uc.logger.InfoWithFields("connecting WhatsApp client", logger.Fields{
		"session_id": sessionID.String(),
	})

	connectionResult, err := waClient.Connect(reconnectCtx)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to connect WhatsApp client: %v", err)
		uc.logger.ErrorWithError("failed to connect WhatsApp client during reconnection", err, logger.Fields{
			"session_id": sessionID.String(),
		})

		// Update session status to disconnected
		sess.Disconnect()
		uc.sessionRepo.Update(ctx, sess)

		return SessionReconnectionResult{
			SessionID:   sessionID,
			SessionName: sessionName,
			Success:     false,
			Error:       errorMsg,
			Duration:    time.Since(startTime),
		}
	}

	// Check if connection was successful (connected or authenticated)
	if connectionResult.Status == whatsapp.StatusConnected || connectionResult.Status == whatsapp.StatusAuthenticated {
		uc.logger.InfoWithFields("✅ session reconnected successfully", logger.Fields{
			"session_id":   sessionID.String(),
			"session_name": sessionName,
			"jid":          connectionResult.JID,
			"duration_ms":  time.Since(startTime).Milliseconds(),
		})

		// Update session with new connection info if JID changed
		if connectionResult.JID != "" && connectionResult.JID != sess.WaJID() {
			if err := sess.Connect(connectionResult.JID); err == nil {
				uc.sessionRepo.Update(ctx, sess)
			}
		}

		return SessionReconnectionResult{
			SessionID:   sessionID,
			SessionName: sessionName,
			Success:     true,
			Error:       "",
			Duration:    time.Since(startTime),
		}
	} else {
		errorMsg := fmt.Sprintf("connection failed with status: %s", connectionResult.Status)
		uc.logger.WarnWithFields("session reconnection failed", logger.Fields{
			"session_id":   sessionID.String(),
			"session_name": sessionName,
			"status":       connectionResult.Status.String(),
		})

		return SessionReconnectionResult{
			SessionID:   sessionID,
			SessionName: sessionName,
			Success:     false,
			Error:       errorMsg,
			Duration:    time.Since(startTime),
		}
	}
}
