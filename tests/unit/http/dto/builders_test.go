package dto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/http/dto"
)

func TestSessionResponseBuilder(t *testing.T) {
	t.Run("should build session response with all fields", func(t *testing.T) {
		// Arrange
		builder := dto.NewSessionResponseBuilder()
		now := time.Now()

		// Act
		response := builder.
			WithID("test-id").
			WithName("test-session").
			WithStatus("connected").
			WithWaJID("123456789@s.whatsapp.net").
			WithActive(true).
			WithTimestamps(now, now).
			Build()

		// Assert
		assert.Equal(t, "test-id", response.ID)
		assert.Equal(t, "test-session", response.Name)
		assert.Equal(t, "connected", response.Status)
		assert.Equal(t, "123456789@s.whatsapp.net", response.WaJID)
		assert.True(t, response.IsActive)
		assert.Equal(t, now, response.CreatedAt)
		assert.Equal(t, now, response.UpdatedAt)
	})

	t.Run("should build session response with proxy config", func(t *testing.T) {
		// Arrange
		builder := dto.NewSessionResponseBuilder()

		// Act
		response := builder.
			WithID("test-id").
			WithName("test-session").
			WithProxy("proxy.example.com", 8080, dto.ProxyTypeHTTP, "user", "pass").
			Build()

		// Assert
		require.NotNil(t, response.ProxyConfig)
		assert.Equal(t, "proxy.example.com", response.ProxyConfig.Host)
		assert.Equal(t, 8080, response.ProxyConfig.Port)
		assert.Equal(t, dto.ProxyTypeHTTP, response.ProxyConfig.Type)
		assert.Equal(t, "user", response.ProxyConfig.Username)
		assert.Equal(t, "pass", response.ProxyConfig.Password)
	})

	t.Run("should build from domain session", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")
		sess.Connect("123456789@s.whatsapp.net")
		builder := dto.NewSessionResponseBuilder()

		// Act
		response := builder.FromDomainSession(sess).Build()

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, sess.Name(), response.Name)
		assert.Equal(t, sess.Status().String(), response.Status)
		assert.Equal(t, sess.WaJID(), response.WaJID)
		assert.Equal(t, sess.IsActive(), response.IsActive)
		assert.Equal(t, sess.CreatedAt(), response.CreatedAt)
		assert.Equal(t, sess.UpdatedAt(), response.UpdatedAt)
	})
}

func TestConnectSessionResponseBuilder(t *testing.T) {
	t.Run("should build connect session response", func(t *testing.T) {
		// Arrange
		sessionResponse := &dto.SessionResponse{
			ID:   "test-id",
			Name: "test-session",
		}
		builder := dto.NewConnectSessionResponseBuilder()

		// Act
		response := builder.
			WithSession(sessionResponse).
			WithQRCode("qr-code-data").
			WithNeedsAuth(true).
			WithMessage("QR code generated").
			Build()

		// Assert
		assert.Equal(t, sessionResponse, response.Session)
		assert.Equal(t, "qr-code-data", response.QRCode)
		assert.True(t, response.NeedsAuth)
		assert.Equal(t, "QR code generated", response.Message)
	})
}

func TestErrorResponseBuilder(t *testing.T) {
	t.Run("should build error response with all fields", func(t *testing.T) {
		// Arrange
		builder := dto.NewErrorResponseBuilder()
		now := time.Now()

		// Act
		response := builder.
			WithError("Test error").
			WithCode("TEST_ERROR").
			WithDetails("Error details").
			AddContext("key1", "value1").
			AddContext("key2", 42).
			WithTimestamp(now).
			Build()

		// Assert
		assert.False(t, response.Success)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Test error", response.Error)
		assert.Equal(t, "TEST_ERROR", response.Code)
		assert.Equal(t, "Error details", response.Details)
		assert.Equal(t, "value1", response.Context["key1"])
		assert.Equal(t, 42, response.Context["key2"])
		assert.Equal(t, now, response.Timestamp)
	})

	t.Run("should build error response with context map", func(t *testing.T) {
		// Arrange
		builder := dto.NewErrorResponseBuilder()
		context := map[string]interface{}{
			"request_id": "req-123",
			"user_id":    "user-456",
		}

		// Act
		response := builder.
			WithError("Test error").
			WithContext(context).
			Build()

		// Assert
		assert.Equal(t, context, response.Context)
	})
}

func TestValidationErrorResponseBuilder(t *testing.T) {
	t.Run("should build validation error response", func(t *testing.T) {
		// Arrange
		builder := dto.NewValidationErrorResponseBuilder()

		// Act
		response := builder.
			WithError("Custom validation error").
			WithCode("CUSTOM_VALIDATION").
			AddField("name", "required", "", "Name is required").
			AddField("email", "email", "invalid-email", "Invalid email format").
			Build()

		// Assert
		assert.False(t, response.Success)
		assert.Equal(t, "Custom validation error", response.Error)
		assert.Equal(t, "CUSTOM_VALIDATION", response.Code)
		assert.Len(t, response.Fields, 2)

		// Check first field
		assert.Equal(t, "name", response.Fields[0].Field)
		assert.Equal(t, "required", response.Fields[0].Tag)
		assert.Equal(t, "", response.Fields[0].Value)
		assert.Equal(t, "Name is required", response.Fields[0].Message)

		// Check second field
		assert.Equal(t, "email", response.Fields[1].Field)
		assert.Equal(t, "email", response.Fields[1].Tag)
		assert.Equal(t, "invalid-email", response.Fields[1].Value)
		assert.Equal(t, "Invalid email format", response.Fields[1].Message)
	})

	t.Run("should build with fields array", func(t *testing.T) {
		// Arrange
		builder := dto.NewValidationErrorResponseBuilder()
		fields := []dto.ValidationFieldError{
			{Field: "field1", Tag: "tag1", Value: "value1", Message: "message1"},
			{Field: "field2", Tag: "tag2", Value: "value2", Message: "message2"},
		}

		// Act
		response := builder.WithFields(fields).Build()

		// Assert
		assert.Equal(t, fields, response.Fields)
	})
}

func TestMetricsResponseBuilder(t *testing.T) {
	t.Run("should build metrics response", func(t *testing.T) {
		// Arrange
		builder := dto.NewMetricsResponseBuilder()
		now := time.Now()

		sessionMetrics := dto.SessionMetrics{
			Total:        10,
			Connected:    5,
			Disconnected: 3,
			Error:        1,
			Active:       4,
		}

		whatsappMetrics := dto.WhatsAppMetrics{
			TotalClients:         5,
			ConnectedClients:     3,
			AuthenticatedClients: 2,
			ErrorClients:         1,
			MessagesSent:         150,
			MessagesReceived:     75,
		}

		systemMetrics := dto.SystemMetrics{
			Uptime:              "2h30m45s",
			MemoryUsage:         "256MB",
			CPUUsage:            "15%",
			DatabaseStatus:      "healthy",
			DatabaseConnections: 5,
		}

		// Act
		response := builder.
			WithSessionMetrics(sessionMetrics).
			WithWhatsAppMetrics(whatsappMetrics).
			WithSystemMetrics(systemMetrics).
			WithTimestamp(now).
			Build()

		// Assert
		assert.Equal(t, sessionMetrics, response.Sessions)
		assert.Equal(t, whatsappMetrics, response.WhatsApp)
		assert.Equal(t, systemMetrics, response.System)
		assert.Equal(t, now, response.Timestamp)
	})
}

func TestBuilderChaining(t *testing.T) {
	t.Run("should support method chaining", func(t *testing.T) {
		// Arrange & Act
		response := dto.NewSessionResponseBuilder().
			WithID("test-id").
			WithName("test-session").
			WithStatus("connected").
			WithActive(true).
			Build()

		// Assert
		assert.Equal(t, "test-id", response.ID)
		assert.Equal(t, "test-session", response.Name)
		assert.Equal(t, "connected", response.Status)
		assert.True(t, response.IsActive)
	})

	t.Run("should support partial building", func(t *testing.T) {
		// Arrange
		builder1 := dto.NewSessionResponseBuilder().
			WithID("test-id").
			WithName("test-session")

		// Act - build partial response
		response1 := builder1.Build()

		// Create new builder and continue building
		builder2 := dto.NewSessionResponseBuilder().
			WithID("test-id").
			WithName("test-session").
			WithStatus("connected").
			WithActive(true)

		response2 := builder2.Build()

		// Assert
		assert.Equal(t, "test-id", response1.ID)
		assert.Equal(t, "test-session", response1.Name)
		assert.Equal(t, "", response1.Status)
		assert.False(t, response1.IsActive)

		assert.Equal(t, "test-id", response2.ID)
		assert.Equal(t, "test-session", response2.Name)
		assert.Equal(t, "connected", response2.Status)
		assert.True(t, response2.IsActive)
	})
}
