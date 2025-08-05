package dto

import (
	"time"

	"wazmeow/internal/domain/session"
)

// SessionResponseBuilder provides a fluent interface for building SessionResponse
type SessionResponseBuilder struct {
	response *SessionResponse
}

// NewSessionResponseBuilder creates a new SessionResponseBuilder
func NewSessionResponseBuilder() *SessionResponseBuilder {
	return &SessionResponseBuilder{
		response: &SessionResponse{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// WithID sets the session ID
func (b *SessionResponseBuilder) WithID(id string) *SessionResponseBuilder {
	b.response.ID = id
	return b
}

// WithName sets the session name
func (b *SessionResponseBuilder) WithName(name string) *SessionResponseBuilder {
	b.response.Name = name
	return b
}

// WithStatus sets the session status
func (b *SessionResponseBuilder) WithStatus(status string) *SessionResponseBuilder {
	b.response.Status = status
	return b
}

// WithWaJID sets the WhatsApp JID
func (b *SessionResponseBuilder) WithWaJID(waJID string) *SessionResponseBuilder {
	b.response.WaJID = waJID
	return b
}

// WithProxyConfig sets the proxy configuration
func (b *SessionResponseBuilder) WithProxyConfig(config *ProxyConfigResponse) *SessionResponseBuilder {
	b.response.ProxyConfig = config
	return b
}

// WithProxy sets proxy configuration using individual parameters
func (b *SessionResponseBuilder) WithProxy(host string, port int, proxyType ProxyType, username, password string) *SessionResponseBuilder {
	b.response.ProxyConfig = NewProxyConfigResponse(host, port, proxyType, username, password)
	return b
}

// WithActive sets the active status
func (b *SessionResponseBuilder) WithActive(isActive bool) *SessionResponseBuilder {
	b.response.IsActive = isActive
	return b
}

// WithTimestamps sets creation and update timestamps
func (b *SessionResponseBuilder) WithTimestamps(createdAt, updatedAt time.Time) *SessionResponseBuilder {
	b.response.CreatedAt = createdAt
	b.response.UpdatedAt = updatedAt
	return b
}

// FromDomainSession builds from a domain session entity
func (b *SessionResponseBuilder) FromDomainSession(sess *session.Session) *SessionResponseBuilder {
	b.response.ID = sess.ID().String()
	b.response.Name = sess.Name()
	b.response.Status = sess.Status().String()
	b.response.WaJID = sess.WaJID()
	b.response.IsActive = sess.IsActive()
	b.response.CreatedAt = sess.CreatedAt()
	b.response.UpdatedAt = sess.UpdatedAt()

	// Add proxy configuration if present
	if sess.HasProxy() {
		proxyType := ProxyType(sess.GetProxyType())
		if !proxyType.IsValid() {
			proxyType = ProxyTypeHTTP
		}

		b.response.ProxyConfig = &ProxyConfigResponse{
			Host: sess.GetProxyHost(),
			Port: parseProxyPort(sess.GetProxyPort()),
			Type: proxyType,
		}

		// Extract username/password from URL if present
		if sess.HasProxyAuth() {
			username, password := extractProxyAuth(sess.ProxyURL())
			b.response.ProxyConfig.Username = username
			b.response.ProxyConfig.Password = password
		}
	}

	return b
}

// Build returns the built SessionResponse
func (b *SessionResponseBuilder) Build() *SessionResponse {
	return b.response
}

// ConnectSessionResponseBuilder provides a fluent interface for building ConnectSessionResponse
type ConnectSessionResponseBuilder struct {
	response *ConnectSessionResponse
}

// NewConnectSessionResponseBuilder creates a new ConnectSessionResponseBuilder
func NewConnectSessionResponseBuilder() *ConnectSessionResponseBuilder {
	return &ConnectSessionResponseBuilder{
		response: &ConnectSessionResponse{},
	}
}

// WithSession sets the session response
func (b *ConnectSessionResponseBuilder) WithSession(session *SessionResponse) *ConnectSessionResponseBuilder {
	b.response.Session = session
	return b
}

// WithQRCode sets the QR code
func (b *ConnectSessionResponseBuilder) WithQRCode(qrCode string) *ConnectSessionResponseBuilder {
	b.response.QRCode = qrCode
	return b
}

// WithNeedsAuth sets the needs authentication flag
func (b *ConnectSessionResponseBuilder) WithNeedsAuth(needsAuth bool) *ConnectSessionResponseBuilder {
	b.response.NeedsAuth = needsAuth
	return b
}

// WithMessage sets the message
func (b *ConnectSessionResponseBuilder) WithMessage(message string) *ConnectSessionResponseBuilder {
	b.response.Message = message
	return b
}

// Build returns the built ConnectSessionResponse
func (b *ConnectSessionResponseBuilder) Build() *ConnectSessionResponse {
	return b.response
}

// ErrorResponseBuilder provides a fluent interface for building ErrorResponse
type ErrorResponseBuilder struct {
	response *ErrorResponse
}

// NewErrorResponseBuilder creates a new ErrorResponseBuilder
func NewErrorResponseBuilder() *ErrorResponseBuilder {
	return &ErrorResponseBuilder{
		response: &ErrorResponse{
			Success:   false,
			Status:    StatusError.String(),
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		},
	}
}

// WithError sets the error message
func (b *ErrorResponseBuilder) WithError(error string) *ErrorResponseBuilder {
	b.response.Error = error
	return b
}

// WithCode sets the error code
func (b *ErrorResponseBuilder) WithCode(code string) *ErrorResponseBuilder {
	b.response.Code = code
	return b
}

// WithDetails sets the error details
func (b *ErrorResponseBuilder) WithDetails(details string) *ErrorResponseBuilder {
	b.response.Details = details
	return b
}

// WithContext sets the error context
func (b *ErrorResponseBuilder) WithContext(context map[string]interface{}) *ErrorResponseBuilder {
	b.response.Context = context
	return b
}

// AddContext adds a key-value pair to the error context
func (b *ErrorResponseBuilder) AddContext(key string, value interface{}) *ErrorResponseBuilder {
	if b.response.Context == nil {
		b.response.Context = make(map[string]interface{})
	}
	b.response.Context[key] = value
	return b
}

// WithTimestamp sets the error timestamp
func (b *ErrorResponseBuilder) WithTimestamp(timestamp time.Time) *ErrorResponseBuilder {
	b.response.Timestamp = timestamp
	return b
}

// Build returns the built ErrorResponse
func (b *ErrorResponseBuilder) Build() *ErrorResponse {
	return b.response
}

// ValidationErrorResponseBuilder provides a fluent interface for building ValidationErrorResponse
type ValidationErrorResponseBuilder struct {
	response *ValidationErrorResponse
}

// NewValidationErrorResponseBuilder creates a new ValidationErrorResponseBuilder
func NewValidationErrorResponseBuilder() *ValidationErrorResponseBuilder {
	return &ValidationErrorResponseBuilder{
		response: &ValidationErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Code:    "VALIDATION_ERROR",
			Fields:  make([]ValidationFieldError, 0),
		},
	}
}

// WithError sets the error message
func (b *ValidationErrorResponseBuilder) WithError(error string) *ValidationErrorResponseBuilder {
	b.response.Error = error
	return b
}

// WithCode sets the error code
func (b *ValidationErrorResponseBuilder) WithCode(code string) *ValidationErrorResponseBuilder {
	b.response.Code = code
	return b
}

// AddField adds a validation field error
func (b *ValidationErrorResponseBuilder) AddField(field, tag, value, message string) *ValidationErrorResponseBuilder {
	b.response.Fields = append(b.response.Fields, ValidationFieldError{
		Field:   field,
		Tag:     tag,
		Value:   value,
		Message: message,
	})
	return b
}

// WithFields sets all validation field errors
func (b *ValidationErrorResponseBuilder) WithFields(fields []ValidationFieldError) *ValidationErrorResponseBuilder {
	b.response.Fields = fields
	return b
}

// Build returns the built ValidationErrorResponse
func (b *ValidationErrorResponseBuilder) Build() *ValidationErrorResponse {
	return b.response
}

// MetricsResponseBuilder provides a fluent interface for building MetricsResponse
type MetricsResponseBuilder struct {
	response *MetricsResponse
}

// NewMetricsResponseBuilder creates a new MetricsResponseBuilder
func NewMetricsResponseBuilder() *MetricsResponseBuilder {
	return &MetricsResponseBuilder{
		response: &MetricsResponse{
			Timestamp: time.Now(),
		},
	}
}

// WithSessionMetrics sets the session metrics
func (b *MetricsResponseBuilder) WithSessionMetrics(metrics SessionMetrics) *MetricsResponseBuilder {
	b.response.Sessions = metrics
	return b
}

// WithWhatsAppMetrics sets the WhatsApp metrics
func (b *MetricsResponseBuilder) WithWhatsAppMetrics(metrics WhatsAppMetrics) *MetricsResponseBuilder {
	b.response.WhatsApp = metrics
	return b
}

// WithSystemMetrics sets the system metrics
func (b *MetricsResponseBuilder) WithSystemMetrics(metrics SystemMetrics) *MetricsResponseBuilder {
	b.response.System = metrics
	return b
}

// WithTimestamp sets the metrics timestamp
func (b *MetricsResponseBuilder) WithTimestamp(timestamp time.Time) *MetricsResponseBuilder {
	b.response.Timestamp = timestamp
	return b
}

// Build returns the built MetricsResponse
func (b *MetricsResponseBuilder) Build() *MetricsResponse {
	return b.response
}
