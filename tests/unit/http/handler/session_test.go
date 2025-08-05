package http_handler_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/http/dto"
)

// TestSessionHandler_RequestResponseStructures tests the HTTP request/response structures
// without testing the full handler integration (which requires complex mocking)
// This approach focuses on validating the data structures and JSON serialization
func TestSessionHandler_RequestResponseStructures(t *testing.T) {
	t.Run("should handle CreateSessionRequest correctly", func(t *testing.T) {
		// Test request structure
		reqBody := dto.CreateSessionRequest{
			Name: "test-session",
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		var parsedReq dto.CreateSessionRequest
		err = json.Unmarshal(jsonBody, &parsedReq)
		require.NoError(t, err)
		assert.Equal(t, "test-session", parsedReq.Name)
	})

	t.Run("should handle SessionResponse correctly", func(t *testing.T) {
		// Test response structure
		sess := session.NewSession("test-session")
		response := dto.SuccessResponse{
			Success: true,
			Message: "Session created successfully",
			Data: &dto.SessionResponse{
				ID:        sess.ID().String(),
				Name:      sess.Name(),
				Status:    sess.Status().String(),
				IsActive:  sess.IsActive(),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			},
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.SuccessResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.True(t, parsedResponse.Success)
		assert.Equal(t, "Session created successfully", parsedResponse.Message)
		assert.NotNil(t, parsedResponse.Data)
	})

	t.Run("should handle ConnectSessionResponse correctly", func(t *testing.T) {
		sess := session.NewSession("connect-test")
		response := dto.ConnectSessionResponse{
			Session: dto.ToSessionResponse(sess),
			QRCode:    "test-qr-code",
			NeedsAuth: true,
			Message:   "QR Code generated",
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.ConnectSessionResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.Equal(t, sess.ID().String(), parsedResponse.Session.ID)
		assert.Equal(t, "test-qr-code", parsedResponse.QRCode)
		assert.True(t, parsedResponse.NeedsAuth)
		assert.Equal(t, "QR Code generated", parsedResponse.Message)
	})

	t.Run("should handle ErrorResponse correctly", func(t *testing.T) {
		response := dto.ErrorResponse{
			Success: false,
			Error:   "Session not found",
			Code:    "NOT_FOUND",
			Details: "The requested session does not exist",
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.ErrorResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.False(t, parsedResponse.Success)
		assert.Equal(t, "Session not found", parsedResponse.Error)
		assert.Equal(t, "NOT_FOUND", parsedResponse.Code)
		assert.Equal(t, "The requested session does not exist", parsedResponse.Details)
	})

	t.Run("should handle DisconnectSessionResponse correctly", func(t *testing.T) {
		sess := session.NewSession("disconnect-test")
		response := dto.DisconnectSessionResponse{
			Session: dto.ToSessionResponse(sess),
			Message: "Session disconnected successfully",
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.DisconnectSessionResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.Equal(t, sess.ID().String(), parsedResponse.Session.ID)
		assert.Equal(t, "Session disconnected successfully", parsedResponse.Message)
	})

	t.Run("should handle DeleteSessionResponse correctly", func(t *testing.T) {
		sessionID := session.NewSessionID()
		response := dto.DeleteSessionResponse{
			SessionID: sessionID.String(),
			Message:   "Session deleted successfully",
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.DeleteSessionResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.Equal(t, sessionID.String(), parsedResponse.SessionID)
		assert.Equal(t, "Session deleted successfully", parsedResponse.Message)
	})

	t.Run("should handle QRCodeResponse correctly", func(t *testing.T) {
		sessionID := session.NewSessionID()
		response := dto.QRCodeResponse{
			SessionID: sessionID.String(),
			QRCode:    "test-qr-code-data",
			Message:   "QR Code generated successfully",
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.QRCodeResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.Equal(t, sessionID.String(), parsedResponse.SessionID)
		assert.Equal(t, "test-qr-code-data", parsedResponse.QRCode)
		assert.Equal(t, "QR Code generated successfully", parsedResponse.Message)
	})

	t.Run("should handle PairPhoneRequest correctly", func(t *testing.T) {
		request := dto.PairPhoneRequest{
			PhoneNumber: "+5511999999999",
		}

		jsonRequest, err := json.Marshal(request)
		require.NoError(t, err)

		var parsedRequest dto.PairPhoneRequest
		err = json.Unmarshal(jsonRequest, &parsedRequest)
		require.NoError(t, err)

		assert.Equal(t, "+5511999999999", parsedRequest.PhoneNumber)
	})

	t.Run("should handle PairPhoneResponse correctly", func(t *testing.T) {
		sessionID := session.NewSessionID()
		response := dto.PairPhoneResponse{
			SessionID:   sessionID.String(),
			PhoneNumber: "+5511999999999",
			Success:     true,
			Message:     "Phone paired successfully",
		}

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.PairPhoneResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.Equal(t, sessionID.String(), parsedResponse.SessionID)
		assert.Equal(t, "+5511999999999", parsedResponse.PhoneNumber)
		assert.True(t, parsedResponse.Success)
		assert.Equal(t, "Phone paired successfully", parsedResponse.Message)
	})

	t.Run("should handle SessionListResponse correctly", func(t *testing.T) {
		sess1 := session.NewSession("session-1")
		sess2 := session.NewSession("session-2")
		sessions := []*session.Session{sess1, sess2}

		response := dto.ToSessionListResponse(sessions, 2)

		jsonResponse, err := json.Marshal(response)
		require.NoError(t, err)

		var parsedResponse dto.SessionListResponse
		err = json.Unmarshal(jsonResponse, &parsedResponse)
		require.NoError(t, err)

		assert.Equal(t, 2, parsedResponse.Total)
		assert.Len(t, parsedResponse.Sessions, 2)
		assert.Equal(t, sess1.ID().String(), parsedResponse.Sessions[0].ID)
		assert.Equal(t, sess2.ID().String(), parsedResponse.Sessions[1].ID)
	})
}

// TestSessionHandler_DTOConversions tests the DTO conversion functions
func TestSessionHandler_DTOConversions(t *testing.T) {
	t.Run("should convert session to SessionResponse correctly", func(t *testing.T) {
		sess := session.NewSession("test-conversion")
		
		response := dto.ToSessionResponse(sess)
		
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, sess.Name(), response.Name)
		assert.Equal(t, sess.Status().String(), response.Status)
		assert.Equal(t, sess.WaJID(), response.WaJID)
		assert.Equal(t, sess.IsActive(), response.IsActive)
		assert.Equal(t, sess.CreatedAt(), response.CreatedAt)
		assert.Equal(t, sess.UpdatedAt(), response.UpdatedAt)
	})

	t.Run("should convert session list to SessionListResponse correctly", func(t *testing.T) {
		sessions := []*session.Session{
			session.NewSession("session-1"),
			session.NewSession("session-2"),
			session.NewSession("session-3"),
		}
		
		response := dto.ToSessionListResponse(sessions, 3)
		
		assert.Equal(t, 3, response.Total)
		assert.Len(t, response.Sessions, 3)
		
		for i, sess := range sessions {
			assert.Equal(t, sess.ID().String(), response.Sessions[i].ID)
			assert.Equal(t, sess.Name(), response.Sessions[i].Name)
		}
	})
}
