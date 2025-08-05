package session

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"wazmeow/internal/domain/session"
	"wazmeow/pkg/logger"
	"wazmeow/pkg/validator"
)

// SetProxyUseCase handles proxy configuration for sessions
type SetProxyUseCase struct {
	repo      session.Repository
	logger    logger.Logger
	validator validator.Validator
}

// NewSetProxyUseCase creates a new set proxy use case
func NewSetProxyUseCase(repo session.Repository, logger logger.Logger, validator validator.Validator) *SetProxyUseCase {
	return &SetProxyUseCase{
		repo:      repo,
		logger:    logger,
		validator: validator,
	}
}

// SetProxyRequest represents the request to set proxy configuration
type SetProxyRequest struct {
	SessionID session.SessionID `json:"session_id" validate:"required"`
	ProxyHost string            `json:"proxy_host"`
	ProxyPort int               `json:"proxy_port"`
	ProxyType string            `json:"proxy_type"`
	Username  string            `json:"username,omitempty"`
	Password  string            `json:"password,omitempty"`
}

// SetProxyResponse represents the response from setting proxy configuration
type SetProxyResponse struct {
	Session *session.Session `json:"session"`
	Message string           `json:"message"`
}

// Execute sets proxy configuration for a session
func (uc *SetProxyUseCase) Execute(ctx context.Context, req SetProxyRequest) (*SetProxyResponse, error) {
	// Validate request
	if err := uc.validator.Validate(req); err != nil {
		uc.logger.ErrorWithError("validation failed for set proxy", err, logger.Fields{
			"session_id": req.SessionID.String(),
			"proxy_host": req.ProxyHost,
		})
		return nil, err
	}

	// Get session from repository
	sess, err := uc.repo.GetByID(ctx, req.SessionID)
	if err != nil {
		uc.logger.ErrorWithError("failed to get session", err, logger.Fields{
			"session_id": req.SessionID.String(),
		})
		return nil, err
	}

	// Build complete proxy URL with type, credentials and proper format
	proxyURL := uc.buildProxyURL(req.ProxyHost, req.ProxyPort, req.ProxyType, req.Username, req.Password)

	// Validate proxy URL format (only if not empty)
	if proxyURL != "" {
		if err := uc.validateProxyURL(proxyURL); err != nil {
			uc.logger.ErrorWithError("invalid proxy URL", err, logger.Fields{
				"session_id": req.SessionID.String(),
				"proxy_url":  proxyURL,
			})
			return nil, err
		}
	}

	// Set proxy URL on session
	sess.SetProxyURL(proxyURL)

	// Update session in repository
	if err := uc.repo.Update(ctx, sess); err != nil {
		uc.logger.ErrorWithError("failed to update session with proxy", err, logger.Fields{
			"session_id": sess.ID().String(),
			"proxy_url":  proxyURL,
		})
		return nil, err
	}

	uc.logger.InfoWithFields("proxy configured for session", logger.Fields{
		"session_id": sess.ID().String(),
		"proxy_url":  proxyURL,
		"has_auth":   req.Username != "",
	})

	return &SetProxyResponse{
		Session: sess,
		Message: "Proxy configured successfully",
	}, nil
}

// validateProxyURL validates the proxy URL format
func (uc *SetProxyUseCase) validateProxyURL(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

	// Check if scheme is supported
	switch parsedURL.Scheme {
	case "http", "https", "socks5":
		// Valid schemes
	default:
		return session.ErrInvalidProxyURL
	}

	// Check if host is present
	if parsedURL.Host == "" {
		return session.ErrInvalidProxyURL
	}

	return nil
}

// buildProxyURL builds a complete proxy URL with type, credentials and proper format
func (uc *SetProxyUseCase) buildProxyURL(proxyHost string, proxyPort int, proxyType, username, password string) string {
	if proxyHost == "" {
		return ""
	}

	// Default to http if no type specified
	if proxyType == "" {
		proxyType = "http"
	}

	// Normalize proxy type
	switch proxyType {
	case "socks", "socks5":
		proxyType = "socks5"
	case "http", "https":
		proxyType = "http"
	default:
		proxyType = "http"
	}

	// Build host:port if port is specified
	hostPort := proxyHost
	if proxyPort > 0 {
		hostPort = fmt.Sprintf("%s:%d", proxyHost, proxyPort)
	}

	// Check if hostPort already has a scheme
	if !strings.Contains(hostPort, "://") {
		// Add scheme if not present
		hostPort = proxyType + "://" + hostPort
	}

	// Parse URL to add credentials if needed
	parsedURL, err := url.Parse(hostPort)
	if err != nil {
		// If parsing fails, try adding scheme and parse again
		hostPort = proxyType + "://" + strings.TrimPrefix(hostPort, proxyType+"://")
		parsedURL, err = url.Parse(hostPort)
		if err != nil {
			return hostPort
		}
	}

	// Add credentials if provided
	if username != "" && password != "" {
		parsedURL.User = url.UserPassword(username, password)
	}

	return parsedURL.String()
}
