package dto

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"wazmeow/internal/domain/session"
)

// ProxyType represents the type of proxy
// @Description Tipo de proxy suportado
// @Enum http socks5
type ProxyType string

const (
	// ProxyTypeHTTP represents HTTP proxy
	ProxyTypeHTTP ProxyType = "http"
	// ProxyTypeSOCKS5 represents SOCKS5 proxy
	ProxyTypeSOCKS5 ProxyType = "socks5"
)

// String returns the string representation of ProxyType
func (pt ProxyType) String() string {
	return string(pt)
}

// IsValid returns true if the proxy type is valid
func (pt ProxyType) IsValid() bool {
	return pt == ProxyTypeHTTP || pt == ProxyTypeSOCKS5
}

// CreateSessionRequest represents the HTTP request to create a session
// @Description Dados para criação de uma nova sessão WhatsApp
type CreateSessionRequest struct {
	Name      string    `json:"name" validate:"required,session_name" example:"minha-sessao" description:"Nome único da sessão (3-50 caracteres, apenas letras, números, hífens e underscores)"`
	ProxyHost string    `json:"proxy_host,omitempty" validate:"omitempty,ip|hostname" example:"78.24.204.134" description:"IP ou hostname do proxy (opcional, requerido se proxy_port for especificado)"`
	ProxyPort int       `json:"proxy_port,omitempty" validate:"omitempty,min=1,max=65535" example:"62122" description:"Porta do proxy (opcional, 1-65535, requerido se proxy_host for especificado)"`
	ProxyType ProxyType `json:"proxy_type,omitempty" validate:"omitempty,oneof=http socks5" example:"http" description:"Tipo do proxy (opcional, padrão: http se proxy configurado)"`
	Username  string    `json:"username,omitempty" validate:"omitempty,min=1,max=255" example:"sgQ4BJZs" description:"Usuário para autenticação do proxy (opcional)"`
	Password  string    `json:"password,omitempty" validate:"omitempty,min=1,max=255" example:"YGFEu7Wx" description:"Senha para autenticação do proxy (opcional, requerido se username for especificado)"`
}

// HasProxy returns true if proxy configuration is provided
func (req *CreateSessionRequest) HasProxy() bool {
	return req.ProxyHost != "" && req.ProxyPort > 0
}

// HasProxyAuth returns true if proxy authentication is provided
func (req *CreateSessionRequest) HasProxyAuth() bool {
	return req.Username != "" && req.Password != ""
}

// BuildProxyURL builds a proxy URL from the request data
func (req *CreateSessionRequest) BuildProxyURL() (string, error) {
	if !req.HasProxy() {
		return "", nil
	}

	if !req.ProxyType.IsValid() {
		return "", fmt.Errorf("invalid proxy type: %s", req.ProxyType)
	}

	var userInfo *url.Userinfo
	if req.HasProxyAuth() {
		userInfo = url.UserPassword(req.Username, req.Password)
	}

	proxyURL := &url.URL{
		Scheme: req.ProxyType.String(),
		User:   userInfo,
		Host:   fmt.Sprintf("%s:%d", req.ProxyHost, req.ProxyPort),
	}

	return proxyURL.String(), nil
}

// Normalize normalizes the request data
func (req *CreateSessionRequest) Normalize() {
	req.Name = strings.TrimSpace(req.Name)
	req.ProxyHost = strings.TrimSpace(req.ProxyHost)
	req.Username = strings.TrimSpace(req.Username)

	// Set default proxy type if proxy is configured but type is not specified
	if req.HasProxy() && req.ProxyType == "" {
		req.ProxyType = ProxyTypeHTTP
	}
}

// ProxyConfigResponse represents the proxy configuration in responses
// @Description Configuração do proxy
type ProxyConfigResponse struct {
	Host     string    `json:"host,omitempty" example:"78.24.204.134" description:"IP ou hostname do proxy"`
	Port     int       `json:"port,omitempty" example:"62122" description:"Porta do proxy"`
	Type     ProxyType `json:"type,omitempty" example:"http" description:"Tipo do proxy: http ou socks5"`
	Username string    `json:"username,omitempty" example:"sgQ4BJZs" description:"Usuário do proxy"`
	Password string    `json:"password,omitempty" example:"YGFEu7Wx" description:"Senha do proxy"`
}

// NewProxyConfigResponse creates a new proxy config response
func NewProxyConfigResponse(host string, port int, proxyType ProxyType, username, password string) *ProxyConfigResponse {
	return &ProxyConfigResponse{
		Host:     host,
		Port:     port,
		Type:     proxyType,
		Username: username,
		Password: password,
	}
}

// SessionResponse represents the HTTP response for a session
// @Description Dados de uma sessão WhatsApp
type SessionResponse struct {
	ID          string               `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID único da sessão (UUID)"`
	Name        string               `json:"name" example:"minha-sessao" description:"Nome da sessão"`
	Status      string               `json:"status" example:"connected" enums:"disconnected,connecting,connected" description:"Status atual da sessão"`
	WaJID       string               `json:"wa_jid,omitempty" example:"5511999999999@s.whatsapp.net" description:"JID do WhatsApp (quando conectado)"`
	ProxyConfig *ProxyConfigResponse `json:"proxy_config,omitempty" description:"Configuração do proxy"`
	IsActive    bool                 `json:"is_active" example:"true" description:"Indica se a sessão está ativa"`
	CreatedAt   time.Time            `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Data de criação da sessão"`
	UpdatedAt   time.Time            `json:"updated_at" example:"2024-01-01T12:30:00Z" description:"Data da última atualização"`
}

// SessionListResponse represents the HTTP response for listing sessions
// @Description Lista de sessões WhatsApp
type SessionListResponse struct {
	Sessions []*SessionResponse `json:"sessions" description:"Lista de sessões"`
	Total    int                `json:"total" example:"5" description:"Total de sessões encontradas"`
}

// ConnectSessionRequest represents the HTTP request to connect a session
type ConnectSessionRequest struct {
	// No additional fields needed - session ID comes from URL
}

// ConnectSessionResponse represents the HTTP response for connecting a session
// @Description Resposta da operação de conexão de sessão
type ConnectSessionResponse struct {
	Session   *SessionResponse `json:"session" description:"Dados atualizados da sessão"`
	QRCode    string           `json:"qr_code,omitempty" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..." description:"QR Code em base64 (quando necessário)"`
	NeedsAuth bool             `json:"needs_auth" example:"true" description:"Indica se é necessário escanear QR Code"`
	Message   string           `json:"message" example:"QR Code gerado. Escaneie com seu WhatsApp." description:"Mensagem informativa"`
}

// DisconnectSessionRequest represents the HTTP request to disconnect a session
type DisconnectSessionRequest struct {
	// No additional fields needed - session ID comes from URL
}

// DisconnectSessionResponse represents the HTTP response for disconnecting a session
type DisconnectSessionResponse struct {
	Session *SessionResponse `json:"session"`
	Message string           `json:"message"`
}

// DeleteSessionRequest represents the HTTP request to delete a session
type DeleteSessionRequest struct {
	// No fields needed - deletion always forces
}

// DeleteSessionResponse represents the HTTP response for deleting a session
type DeleteSessionResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// QRCodeResponse represents the HTTP response for QR code generation
// @Description Resposta com QR Code para autenticação
type QRCodeResponse struct {
	SessionID string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID da sessão"`
	QRCode    string `json:"qr_code" example:"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..." description:"QR Code em base64"`
	Message   string `json:"message" example:"QR Code gerado com sucesso" description:"Mensagem informativa"`
}

// PairPhoneRequest represents the HTTP request to pair with a phone number
// @Description Dados para emparelhamento com número de telefone
type PairPhoneRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required" example:"5511999999999" description:"Número de telefone para emparelhar"`
}

// PairPhoneResponse represents the HTTP response for phone pairing
// @Description Resposta do emparelhamento com telefone
type PairPhoneResponse struct {
	SessionID   string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID da sessão"`
	PhoneNumber string `json:"phone_number" example:"5511999999999" description:"Número emparelhado"`
	Success     bool   `json:"success" example:"true" description:"Indica se o emparelhamento foi bem-sucedido"`
	Message     string `json:"message" example:"Telefone emparelhado com sucesso" description:"Mensagem informativa"`
}

// ProxySetRequest represents the HTTP request to set proxy configuration
// @Description Configuração de proxy para a sessão
type ProxySetRequest struct {
	ProxyHost string    `json:"proxy_host" validate:"required" example:"78.24.204.134" description:"IP ou hostname do proxy"`
	ProxyPort int       `json:"proxy_port" validate:"required,min=1,max=65535" example:"62122" description:"Porta do proxy"`
	ProxyType ProxyType `json:"proxy_type" validate:"required,oneof=http socks5" example:"http" description:"Tipo do proxy: http ou socks5"`
	Username  string    `json:"username,omitempty" example:"sgQ4BJZs" description:"Usuário do proxy (opcional)"`
	Password  string    `json:"password,omitempty" example:"YGFEu7Wx" description:"Senha do proxy (opcional)"`
}

// HasProxy returns true if proxy configuration is provided
func (req *ProxySetRequest) HasProxy() bool {
	return req.ProxyHost != "" && req.ProxyPort > 0
}

// HasProxyAuth returns true if proxy authentication is provided
func (req *ProxySetRequest) HasProxyAuth() bool {
	return req.Username != "" && req.Password != ""
}

// Normalize normalizes the request data
func (req *ProxySetRequest) Normalize() {
	req.ProxyHost = strings.TrimSpace(req.ProxyHost)
	req.Username = strings.TrimSpace(req.Username)

	// Set default proxy type if not specified
	if req.HasProxy() && req.ProxyType == "" {
		req.ProxyType = ProxyTypeHTTP
	}
}

// ProxySetResponse represents the HTTP response for proxy configuration
// @Description Resposta da configuração de proxy
type ProxySetResponse struct {
	SessionID string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID da sessão"`
	ProxyURL  string `json:"proxy_url" example:"http://proxy.example.com:8080" description:"URL do proxy configurado"`
	Success   bool   `json:"success" example:"true" description:"Indica se a configuração foi bem-sucedida"`
	Message   string `json:"message" example:"Proxy configurado com sucesso" description:"Mensagem informativa"`
}

// ToSessionResponse converts a domain session to HTTP response using optimized converter
func ToSessionResponse(sess *session.Session) *SessionResponse {
	return ConvertSession(sess)
}

// ToSessionListResponse converts a list of domain sessions to HTTP response using batch converter
func ToSessionListResponse(sessions []*session.Session, total int) *SessionListResponse {
	sessionResponses := ConvertSessions(sessions)
	return &SessionListResponse{
		Sessions: sessionResponses,
		Total:    total,
	}
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

// Factory Methods for Session DTOs

// CreateSessionResponse creates a session response from domain session
func CreateSessionResponse(sess *session.Session) *SessionResponse {
	return ToSessionResponse(sess)
}

// CreateSessionListResponse creates a session list response
func CreateSessionListResponse(sessions []*session.Session, total int) *SessionListResponse {
	return ToSessionListResponse(sessions, total)
}

// CreateConnectSessionResponse creates a connect session response
func CreateConnectSessionResponse(sess *session.Session, qrCode string, needsAuth bool, message string) *ConnectSessionResponse {
	return NewConnectSessionResponseBuilder().
		WithSession(ToSessionResponse(sess)).
		WithQRCode(qrCode).
		WithNeedsAuth(needsAuth).
		WithMessage(message).
		Build()
}

// CreateQRCodeResponse creates a QR code response
func CreateQRCodeResponse(sessionID, qrCode, message string) *QRCodeResponse {
	return &QRCodeResponse{
		SessionID: sessionID,
		QRCode:    qrCode,
		Message:   message,
	}
}

// CreatePairPhoneResponse creates a pair phone response
func CreatePairPhoneResponse(sessionID, phoneNumber string, success bool, message string) *PairPhoneResponse {
	return &PairPhoneResponse{
		SessionID:   sessionID,
		PhoneNumber: phoneNumber,
		Success:     success,
		Message:     message,
	}
}

// CreateProxySetResponse creates a proxy set response
func CreateProxySetResponse(sessionID, proxyURL string, success bool, message string) *ProxySetResponse {
	return &ProxySetResponse{
		SessionID: sessionID,
		ProxyURL:  proxyURL,
		Success:   success,
		Message:   message,
	}
}

// CreateDeleteSessionResponse creates a delete session response
func CreateDeleteSessionResponse(sessionID, message string) *DeleteSessionResponse {
	return &DeleteSessionResponse{
		SessionID: sessionID,
		Message:   message,
	}
}

// CreateDisconnectSessionResponse creates a disconnect session response
func CreateDisconnectSessionResponse(sess *session.Session, message string) *DisconnectSessionResponse {
	return &DisconnectSessionResponse{
		Session: ToSessionResponse(sess),
		Message: message,
	}
}
