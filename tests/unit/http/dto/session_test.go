package dto_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/http/dto"
)

func TestCreateSessionRequest(t *testing.T) {
	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		req := dto.CreateSessionRequest{
			Name: "test-session",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(req)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.CreateSessionRequest
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, req.Name, unmarshaled.Name)
	})

	t.Run("should handle empty name", func(t *testing.T) {
		// Arrange
		jsonData := `{"name": ""}`

		// Act
		var req dto.CreateSessionRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		// Assert
		assert.Empty(t, req.Name)
	})

	t.Run("should handle missing name field", func(t *testing.T) {
		// Arrange
		jsonData := `{}`

		// Act
		var req dto.CreateSessionRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		// Assert
		assert.Empty(t, req.Name)
	})

	t.Run("should validate JSON tags", func(t *testing.T) {
		// Arrange
		jsonData := `{"name": "test-session"}`

		// Act
		var req dto.CreateSessionRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, "test-session", req.Name)
	})
}

func TestSessionResponse(t *testing.T) {
	t.Run("should create from domain session", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")

		// Act
		response := dto.ToSessionResponse(sess)

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, sess.Name(), response.Name)
		assert.Equal(t, sess.Status().String(), response.Status)
		assert.Equal(t, sess.WaJID(), response.WaJID)
		assert.Equal(t, sess.IsActive(), response.IsActive)
		assert.Equal(t, sess.CreatedAt(), response.CreatedAt)
		assert.Equal(t, sess.UpdatedAt(), response.UpdatedAt)
	})

	t.Run("should create from connected session", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("connected-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		// Act
		response := dto.ToSessionResponse(sess)

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, "connected-session", response.Name)
		assert.Equal(t, "connected", response.Status)
		assert.Equal(t, "test@s.whatsapp.net", response.WaJID)
		assert.True(t, response.IsActive)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("marshal-test")
		response := dto.ToSessionResponse(sess)

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.SessionResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.ID, unmarshaled.ID)
		assert.Equal(t, response.Name, unmarshaled.Name)
		assert.Equal(t, response.Status, unmarshaled.Status)
		assert.Equal(t, response.WaJID, unmarshaled.WaJID)
		// QRCode field doesn't exist in SessionResponse
		assert.Equal(t, response.IsActive, unmarshaled.IsActive)
		// Compare timestamps using Unix time to avoid timezone issues
		assert.Equal(t, response.CreatedAt.Unix(), unmarshaled.CreatedAt.Unix())
		assert.Equal(t, response.UpdatedAt.Unix(), unmarshaled.UpdatedAt.Unix())
	})

	t.Run("should handle time formatting correctly", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("time-test")
		response := dto.ToSessionResponse(sess)

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Assert - Check that timestamps are in RFC3339 format
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		createdAt, ok := jsonMap["created_at"].(string)
		require.True(t, ok)
		updatedAt, ok := jsonMap["updated_at"].(string)
		require.True(t, ok)

		// Verify RFC3339 format
		_, err = time.Parse(time.RFC3339, createdAt)
		assert.NoError(t, err)
		_, err = time.Parse(time.RFC3339, updatedAt)
		assert.NoError(t, err)
	})
}

func TestListSessionsResponse(t *testing.T) {
	t.Run("should create from session list", func(t *testing.T) {
		// Arrange
		sessions := []*session.Session{
			session.NewSession("session-1"),
			session.NewSession("session-2"),
			session.NewSession("session-3"),
		}
		total := 3

		// Act
		response := dto.ToSessionListResponse(sessions, total)

		// Assert
		assert.Len(t, response.Sessions, 3)
		assert.Equal(t, total, response.Total)

		// Check individual sessions
		for i, sess := range sessions {
			assert.Equal(t, sess.ID().String(), response.Sessions[i].ID)
			assert.Equal(t, sess.Name(), response.Sessions[i].Name)
		}
	})

	t.Run("should handle empty session list", func(t *testing.T) {
		// Arrange
		sessions := []*session.Session{}
		total := 0

		// Act
		response := dto.ToSessionListResponse(sessions, total)

		// Assert
		assert.Empty(t, response.Sessions)
		assert.Equal(t, 0, response.Total)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sessions := []*session.Session{
			session.NewSession("list-test-1"),
			session.NewSession("list-test-2"),
		}
		response := dto.ToSessionListResponse(sessions, 2)

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.SessionListResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Len(t, unmarshaled.Sessions, 2)
		assert.Equal(t, response.Total, unmarshaled.Total)

		for i := range response.Sessions {
			assert.Equal(t, response.Sessions[i].ID, unmarshaled.Sessions[i].ID)
			assert.Equal(t, response.Sessions[i].Name, unmarshaled.Sessions[i].Name)
		}
	})
}

func TestConnectSessionRequest(t *testing.T) {
	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		req := dto.ConnectSessionRequest{}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(req)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.ConnectSessionRequest
		err = json.Unmarshal([]byte(jsonData), &unmarshaled)
		require.NoError(t, err)

		// Assert - ConnectSessionRequest has no fields
		assert.NotNil(t, jsonData)
	})

	t.Run("should handle empty request", func(t *testing.T) {
		// Arrange
		jsonData := `{}`

		// Act
		var req dto.ConnectSessionRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		// Assert - ConnectSessionRequest has no fields
		assert.NotNil(t, req)
	})
}

func TestConnectSessionResponse(t *testing.T) {
	t.Run("should create response with QR code", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("qr-test")
		qrCode := "test-qr-code"
		message := "QR code generated"

		// Act
		response := dto.ConnectSessionResponse{
			Session: &dto.SessionResponse{
				ID:        sess.ID().String(),
				Name:      sess.Name(),
				Status:    sess.Status().String(),
				IsActive:  sess.IsActive(),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			},
			QRCode:    qrCode,
			NeedsAuth: true,
			Message:   message,
		}

		// Assert
		assert.Equal(t, sess.ID().String(), response.Session.ID)
		assert.Equal(t, qrCode, response.QRCode)
		assert.True(t, response.NeedsAuth)
		assert.Equal(t, message, response.Message)
		assert.NotNil(t, response.Session)
	})

	t.Run("should create response without QR code", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("connected-test")
		message := "Session connected"

		// Act
		response := dto.ConnectSessionResponse{
			Session: &dto.SessionResponse{
				ID:        sess.ID().String(),
				Name:      sess.Name(),
				Status:    sess.Status().String(),
				IsActive:  sess.IsActive(),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			},
			QRCode:    "",
			NeedsAuth: false,
			Message:   message,
		}

		// Assert
		assert.Equal(t, sess.ID().String(), response.Session.ID)
		assert.Empty(t, response.QRCode)
		assert.False(t, response.NeedsAuth)
		assert.Equal(t, message, response.Message)
		assert.NotNil(t, response.Session)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("marshal-connect-test")
		response := dto.ConnectSessionResponse{
			Session: &dto.SessionResponse{
				ID:        sess.ID().String(),
				Name:      sess.Name(),
				Status:    sess.Status().String(),
				IsActive:  sess.IsActive(),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			},
			QRCode:    "qr-data",
			NeedsAuth: true,
			Message:   "test message",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.ConnectSessionResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.Session.ID, unmarshaled.Session.ID)
		assert.Equal(t, response.QRCode, unmarshaled.QRCode)
		assert.Equal(t, response.NeedsAuth, unmarshaled.NeedsAuth)
		assert.Equal(t, response.Message, unmarshaled.Message)
		assert.NotNil(t, unmarshaled.Session)
	})
}

func TestDisconnectSessionResponse(t *testing.T) {
	t.Run("should create response correctly", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("disconnect-test")
		message := "Session disconnected successfully"

		// Act
		response := dto.DisconnectSessionResponse{
			Session: &dto.SessionResponse{
				ID:        sess.ID().String(),
				Name:      sess.Name(),
				Status:    sess.Status().String(),
				IsActive:  sess.IsActive(),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			},
			Message: message,
		}

		// Assert
		assert.Equal(t, sess.ID().String(), response.Session.ID)
		assert.Equal(t, message, response.Message)
		assert.NotNil(t, response.Session)
		assert.Equal(t, sess.Name(), response.Session.Name)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("marshal-disconnect-test")
		response := dto.DisconnectSessionResponse{
			Session: &dto.SessionResponse{
				ID:        sess.ID().String(),
				Name:      sess.Name(),
				Status:    sess.Status().String(),
				IsActive:  sess.IsActive(),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			},
			Message: "test disconnect message",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.DisconnectSessionResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.Session.ID, unmarshaled.Session.ID)
		assert.Equal(t, response.Message, unmarshaled.Message)
		assert.NotNil(t, unmarshaled.Session)
	})
}

func TestDeleteSessionResponse(t *testing.T) {
	t.Run("should create response correctly", func(t *testing.T) {
		// Arrange
		sessionID := session.NewSessionID()
		message := "Session deleted successfully"

		// Act
		response := dto.DeleteSessionResponse{
			SessionID: sessionID.String(),
			Message:   message,
		}

		// Assert
		assert.Equal(t, sessionID.String(), response.SessionID)
		assert.Equal(t, message, response.Message)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sessionID := session.NewSessionID()
		response := dto.DeleteSessionResponse{
			SessionID: sessionID.String(),
			Message:   "test delete message",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.DeleteSessionResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.SessionID, unmarshaled.SessionID)
		assert.Equal(t, response.Message, unmarshaled.Message)
	})
}

func TestQRCodeResponse(t *testing.T) {
	t.Run("should create QR code response correctly", func(t *testing.T) {
		// Arrange
		sessionID := session.NewSessionID()
		qrCode := "test-qr-code-data"
		message := "QR code generated successfully"

		// Act
		response := dto.QRCodeResponse{
			SessionID: sessionID.String(),
			QRCode:    qrCode,
			Message:   message,
		}

		// Assert
		assert.Equal(t, sessionID.String(), response.SessionID)
		assert.Equal(t, qrCode, response.QRCode)
		assert.Equal(t, message, response.Message)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sessionID := session.NewSessionID()
		response := dto.QRCodeResponse{
			SessionID: sessionID.String(),
			QRCode:    "qr-data",
			Message:   "test message",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Unmarshal back
		var unmarshaled dto.QRCodeResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.SessionID, unmarshaled.SessionID)
		assert.Equal(t, response.QRCode, unmarshaled.QRCode)
		assert.Equal(t, response.Message, unmarshaled.Message)
	})
}

func TestPairPhoneRequest(t *testing.T) {
	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		req := dto.PairPhoneRequest{
			PhoneNumber: "+5511999999999",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(req)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.PairPhoneRequest
		err = json.Unmarshal([]byte(jsonData), &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, req.PhoneNumber, unmarshaled.PhoneNumber)
	})

	t.Run("should handle empty fields", func(t *testing.T) {
		// Arrange
		jsonData := `{"session_id": "", "phone_number": ""}`

		// Act
		var req dto.PairPhoneRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		// Assert
		assert.Empty(t, req.PhoneNumber)
	})

	t.Run("should handle missing fields", func(t *testing.T) {
		// Arrange
		jsonData := `{}`

		// Act
		var req dto.PairPhoneRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		// Assert
		assert.Empty(t, req.PhoneNumber)
	})
}

func TestPairPhoneResponse(t *testing.T) {
	t.Run("should create response correctly", func(t *testing.T) {
		// Arrange
		sessionID := session.NewSessionID()
		phoneNumber := "+5511999999999"
		message := "Phone paired successfully"

		// Act
		response := dto.PairPhoneResponse{
			SessionID:   sessionID.String(),
			PhoneNumber: phoneNumber,
			Success:     true,
			Message:     message,
		}

		// Assert
		assert.Equal(t, sessionID.String(), response.SessionID)
		assert.Equal(t, phoneNumber, response.PhoneNumber)
		assert.True(t, response.Success)
		assert.Equal(t, message, response.Message)
	})

	t.Run("should marshal and unmarshal correctly", func(t *testing.T) {
		// Arrange
		sessionID := session.NewSessionID()
		response := dto.PairPhoneResponse{
			SessionID:   sessionID.String(),
			PhoneNumber: "+5511999999999",
			Success:     true,
			Message:     "test message",
		}

		// Act - Marshal to JSON
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Act - Unmarshal from JSON
		var unmarshaled dto.PairPhoneResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)

		// Assert
		assert.Equal(t, response.SessionID, unmarshaled.SessionID)
		assert.Equal(t, response.PhoneNumber, unmarshaled.PhoneNumber)
		assert.Equal(t, response.Success, unmarshaled.Success)
		assert.Equal(t, response.Message, unmarshaled.Message)
	})
}
