package database

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"wazmeow/internal/domain/session"

	"github.com/uptrace/bun"
)

// ProxyConfig represents the proxy configuration stored as JSON
type ProxyConfig struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// WazMeowSessionModel represents the database model for sessions
type WazMeowSessionModel struct {
	bun.BaseModel `bun:"table:wazmeow_sessions"`

	ID          string       `bun:"id,pk,type:varchar(36)" json:"id"`
	Name        string       `bun:"name,unique,notnull,type:varchar(50)" json:"name"`
	Status      string       `bun:"status,notnull,type:varchar(20),default:'disconnected'" json:"status"`
	WaJID       string       `bun:"wa_jid,type:varchar(100)" json:"wa_jid,omitempty"`
	QRCode      string       `bun:"qr_code,type:text" json:"qr_code,omitempty"`
	ProxyConfig *ProxyConfig `bun:"proxy_config,type:text" json:"proxy_config,omitempty"`
	IsActive    bool         `bun:"is_active,notnull,default:false" json:"is_active"`
	CreatedAt   time.Time    `bun:"created_at,notnull,default:current_timestamp,type:datetime" json:"created_at"`
	UpdatedAt   time.Time    `bun:"updated_at,notnull,default:current_timestamp,type:datetime" json:"updated_at"`
}

// ToWazMeowSessionModel converts a domain session to database model
func ToWazMeowSessionModel(sess *session.Session) *WazMeowSessionModel {
	var proxyConfig *ProxyConfig
	if sess.HasProxy() {
		proxyConfig = &ProxyConfig{
			Host: sess.GetProxyHost(),
			Port: parseProxyPort(sess.GetProxyPort()),
			Type: sess.GetProxyType(),
		}

		// Extract username/password from URL if present
		if sess.HasProxyAuth() {
			username, password := extractProxyAuth(sess.ProxyURL())
			proxyConfig.Username = username
			proxyConfig.Password = password
		}
	}

	return &WazMeowSessionModel{
		ID:          sess.ID().String(),
		Name:        sess.Name(),
		Status:      sess.Status().String(),
		WaJID:       sess.WaJID(),
		QRCode:      sess.QRCode(),
		ProxyConfig: proxyConfig,
		IsActive:    sess.IsActive(),
		CreatedAt:   sess.CreatedAt(),
		UpdatedAt:   sess.UpdatedAt(),
	}
}

// FromWazMeowSessionModel converts a database model to domain session
func FromWazMeowSessionModel(model *WazMeowSessionModel) (*session.Session, error) {
	status, err := session.StatusFromString(model.Status)
	if err != nil {
		return nil, err
	}

	sessionID, err := session.SessionIDFromString(model.ID)
	if err != nil {
		return nil, err
	}

	// Convert ProxyConfig back to URL string for domain entity
	proxyURL := ""
	if model.ProxyConfig != nil {
		proxyURL = buildProxyURL(model.ProxyConfig)
	}

	return session.RestoreSession(
		sessionID,
		model.Name,
		status,
		model.WaJID,
		model.QRCode,
		proxyURL,
		model.IsActive,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

// parseProxyPort converts string port to int
func parseProxyPort(portStr string) int {
	if portStr == "" {
		return 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}
	return port
}

// extractProxyAuth extracts username and password from proxy URL
func extractProxyAuth(proxyURL string) (string, string) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil || parsedURL.User == nil {
		return "", ""
	}

	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()
	return username, password
}

// buildProxyURL builds a complete proxy URL from ProxyConfig
func buildProxyURL(config *ProxyConfig) string {
	if config.Host == "" || config.Port == 0 {
		return ""
	}

	// Build base URL
	proxyURL := fmt.Sprintf("%s://%s:%d", config.Type, config.Host, config.Port)

	// Add authentication if present
	if config.Username != "" && config.Password != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return proxyURL
		}
		parsedURL.User = url.UserPassword(config.Username, config.Password)
		return parsedURL.String()
	}

	return proxyURL
}
