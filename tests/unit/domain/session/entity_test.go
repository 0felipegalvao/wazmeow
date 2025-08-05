package domain_session_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
)

func TestNewSession(t *testing.T) {
	t.Run("should create session with valid name", func(t *testing.T) {
		name := "test-session"
		sess := session.NewSession(name)

		assert.NotNil(t, sess)
		assert.NotEmpty(t, sess.ID())
		assert.Equal(t, name, sess.Name())
		assert.Equal(t, session.StatusDisconnected, sess.Status())
		assert.Empty(t, sess.WaJID())
		assert.Empty(t, sess.QRCode())
		assert.False(t, sess.IsActive())
		assert.False(t, sess.CreatedAt().IsZero())
		assert.False(t, sess.UpdatedAt().IsZero())
	})

	t.Run("should panic with empty name", func(t *testing.T) {
		assert.Panics(t, func() {
			session.NewSession("")
		})
	})

	t.Run("should have unique IDs for different sessions", func(t *testing.T) {
		sess1 := session.NewSession("session-1")
		sess2 := session.NewSession("session-2")

		assert.NotEqual(t, sess1.ID(), sess2.ID())
	})

	t.Run("should set creation and update timestamps", func(t *testing.T) {
		before := time.Now()
		sess := session.NewSession("test-session")
		after := time.Now()

		assert.True(t, sess.CreatedAt().After(before) || sess.CreatedAt().Equal(before))
		assert.True(t, sess.CreatedAt().Before(after) || sess.CreatedAt().Equal(after))
		assert.True(t, sess.UpdatedAt().After(before) || sess.UpdatedAt().Equal(before))
		assert.True(t, sess.UpdatedAt().Before(after) || sess.UpdatedAt().Equal(after))
	})
}

func TestRestoreSession(t *testing.T) {
	t.Run("should restore session with all fields", func(t *testing.T) {
		id := session.NewSessionID()
		name := "restored-session"
		status := session.StatusConnected
		waJID := "test@s.whatsapp.net"
		qrCode := "test-qr-code"
		isActive := true
		createdAt := time.Now().Add(-1 * time.Hour)
		updatedAt := time.Now()

		sess := session.RestoreSession(id, name, status, waJID, qrCode, "", isActive, createdAt, updatedAt)

		assert.Equal(t, id, sess.ID())
		assert.Equal(t, name, sess.Name())
		assert.Equal(t, status, sess.Status())
		assert.Equal(t, waJID, sess.WaJID())
		assert.Equal(t, qrCode, sess.QRCode())
		assert.Equal(t, isActive, sess.IsActive())
		assert.Equal(t, createdAt, sess.CreatedAt())
		assert.Equal(t, updatedAt, sess.UpdatedAt())
	})

	t.Run("should restore session with minimal fields", func(t *testing.T) {
		id := session.NewSessionID()
		name := "minimal-session"
		status := session.StatusDisconnected
		createdAt := time.Now().Add(-1 * time.Hour)
		updatedAt := time.Now()

		sess := session.RestoreSession(id, name, status, "", "", "", false, createdAt, updatedAt)

		assert.Equal(t, id, sess.ID())
		assert.Equal(t, name, sess.Name())
		assert.Equal(t, status, sess.Status())
		assert.Empty(t, sess.WaJID())
		assert.Empty(t, sess.QRCode())
		assert.False(t, sess.IsActive())
		assert.Equal(t, createdAt, sess.CreatedAt())
		assert.Equal(t, updatedAt, sess.UpdatedAt())
	})
}

func TestSessionConnect(t *testing.T) {
	t.Run("should connect disconnected session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		waJID := "test@s.whatsapp.net"
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		err := sess.Connect(waJID)

		assert.NoError(t, err)
		assert.Equal(t, session.StatusConnected, sess.Status())
		assert.Equal(t, waJID, sess.WaJID())
		assert.True(t, sess.IsActive())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should connect session in connecting state", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.SetConnecting()
		waJID := "test@s.whatsapp.net"

		err := sess.Connect(waJID)

		assert.NoError(t, err)
		assert.Equal(t, session.StatusConnected, sess.Status())
		assert.Equal(t, waJID, sess.WaJID())
		assert.True(t, sess.IsActive())
	})

	t.Run("should fail to connect already connected session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		waJID1 := "test1@s.whatsapp.net"
		waJID2 := "test2@s.whatsapp.net"

		// First connection should succeed
		err := sess.Connect(waJID1)
		require.NoError(t, err)

		// Second connection should fail
		err = sess.Connect(waJID2)
		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionAlreadyConnected, err)
		assert.Equal(t, waJID1, sess.WaJID()) // Should keep original JID
		assert.Equal(t, session.StatusConnected, sess.Status())
	})

	t.Run("should fail with empty waJID", func(t *testing.T) {
		sess := session.NewSession("test-session")

		err := sess.Connect("")
		assert.Error(t, err)
		assert.Equal(t, session.ErrInvalidWhatsAppJID, err)
		assert.Equal(t, session.StatusDisconnected, sess.Status())
		assert.False(t, sess.IsActive())
		assert.Empty(t, sess.WaJID())
	})

	t.Run("should handle various valid waJID formats", func(t *testing.T) {
		testCases := []string{
			"5511999999999@s.whatsapp.net",
			"123456789@c.us",
			"test@g.us",
		}

		for _, waJID := range testCases {
			t.Run("waJID_"+waJID, func(t *testing.T) {
				sess := session.NewSession("test-session")

				err := sess.Connect(waJID)

				assert.NoError(t, err)
				assert.Equal(t, waJID, sess.WaJID())
				assert.Equal(t, session.StatusConnected, sess.Status())
				assert.True(t, sess.IsActive())
			})
		}
	})
}

func TestSessionDisconnect(t *testing.T) {
	t.Run("should disconnect connected session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		waJID := "test@s.whatsapp.net"

		// Connect first
		err := sess.Connect(waJID)
		require.NoError(t, err)
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		// Disconnect
		sess.Disconnect()

		assert.Equal(t, session.StatusDisconnected, sess.Status())
		assert.False(t, sess.IsActive())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
		// WaJID should remain (for reconnection purposes)
		assert.Equal(t, waJID, sess.WaJID())
	})

	t.Run("should disconnect already disconnected session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		// Should not panic or error
		sess.Disconnect()

		assert.Equal(t, session.StatusDisconnected, sess.Status())
		assert.False(t, sess.IsActive())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should disconnect connecting session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.SetConnecting()

		sess.Disconnect()

		assert.Equal(t, session.StatusDisconnected, sess.Status())
		assert.False(t, sess.IsActive())
	})
}

func TestSessionSetConnecting(t *testing.T) {
	t.Run("should set session to connecting state", func(t *testing.T) {
		sess := session.NewSession("test-session")
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		sess.SetConnecting()

		assert.Equal(t, session.StatusConnecting, sess.Status())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
		assert.False(t, sess.IsActive()) // Should not be active when just connecting
	})

	t.Run("should set connected session to connecting state", func(t *testing.T) {
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		sess.SetConnecting()

		assert.Equal(t, session.StatusConnecting, sess.Status())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
		// Should remain active when going from connected to connecting
		assert.True(t, sess.IsActive())
	})

	t.Run("should set disconnected session to connecting state", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.Disconnect() // Ensure it's disconnected

		sess.SetConnecting()

		assert.Equal(t, session.StatusConnecting, sess.Status())
		assert.False(t, sess.IsActive())
	})
}

func TestSessionQRCode(t *testing.T) {
	t.Run("should set QR code", func(t *testing.T) {
		sess := session.NewSession("test-session")
		qrCode := "test-qr-code-data"
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		sess.SetQRCode(qrCode)

		assert.Equal(t, qrCode, sess.QRCode())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should clear QR code", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.SetQRCode("test-qr-code")
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		sess.ClearQRCode()

		assert.Empty(t, sess.QRCode())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should handle empty QR code", func(t *testing.T) {
		sess := session.NewSession("test-session")
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		sess.SetQRCode("")

		assert.Empty(t, sess.QRCode())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should update QR code multiple times", func(t *testing.T) {
		sess := session.NewSession("test-session")

		sess.SetQRCode("first-qr")
		assert.Equal(t, "first-qr", sess.QRCode())

		sess.SetQRCode("second-qr")
		assert.Equal(t, "second-qr", sess.QRCode())

		sess.ClearQRCode()
		assert.Empty(t, sess.QRCode())
	})
}

func TestSessionUpdateName(t *testing.T) {
	t.Run("should update session name", func(t *testing.T) {
		sess := session.NewSession("old-name")
		newName := "new-name"
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		err := sess.UpdateName(newName)

		assert.NoError(t, err)
		assert.Equal(t, newName, sess.Name())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		sess := session.NewSession("original-name")
		originalName := sess.Name()
		initialUpdatedAt := sess.UpdatedAt()

		err := sess.UpdateName("")

		assert.Error(t, err)
		assert.Equal(t, session.ErrInvalidSessionName, err)
		assert.Equal(t, originalName, sess.Name())          // Should remain unchanged
		assert.Equal(t, initialUpdatedAt, sess.UpdatedAt()) // Should not update timestamp on error
	})

	t.Run("should update to same name", func(t *testing.T) {
		sess := session.NewSession("same-name")
		initialUpdatedAt := sess.UpdatedAt()

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		err := sess.UpdateName("same-name")

		assert.NoError(t, err)
		assert.Equal(t, "same-name", sess.Name())
		assert.True(t, sess.UpdatedAt().After(initialUpdatedAt))
	})

	t.Run("should handle special characters in name", func(t *testing.T) {
		sess := session.NewSession("old-name")
		specialNames := []string{
			"name-with-hyphens",
			"name_with_underscores",
			"name123",
			"123name",
		}

		for _, name := range specialNames {
			err := sess.UpdateName(name)
			assert.NoError(t, err)
			assert.Equal(t, name, sess.Name())
		}
	})
}

func TestCanConnect(t *testing.T) {
	testCases := []struct {
		name     string
		status   session.Status
		expected bool
	}{
		{"disconnected", session.StatusDisconnected, true},
		{"connecting", session.StatusConnecting, true},
		{"connected", session.StatusConnected, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sess := session.RestoreSession(
				session.NewSessionID(),
				"test-session",
				tc.status,
				"",
				"",
				"",
				false,
				time.Now(),
				time.Now(),
			)

			result := sess.CanConnect()
			assert.Equal(t, tc.expected, result)
		})
	}

	t.Run("should allow connection after disconnect", func(t *testing.T) {
		sess := session.NewSession("test-session")

		// Connect
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)
		assert.False(t, sess.CanConnect())

		// Disconnect
		sess.Disconnect()
		assert.True(t, sess.CanConnect())
	})
}

func TestIsConnected(t *testing.T) {
	t.Run("should return true for connected and active session", func(t *testing.T) {
		sess := session.RestoreSession(
			session.NewSessionID(),
			"test-session",
			session.StatusConnected,
			"test@s.whatsapp.net",
			"",
			"",
			true,
			time.Now(),
			time.Now(),
		)

		assert.True(t, sess.IsConnected())
	})

	t.Run("should return false for connected but inactive session", func(t *testing.T) {
		sess := session.RestoreSession(
			session.NewSessionID(),
			"test-session",
			session.StatusConnected,
			"test@s.whatsapp.net",
			"",
			"",
			false,
			time.Now(),
			time.Now(),
		)

		assert.False(t, sess.IsConnected())
	})

	t.Run("should return false for disconnected session", func(t *testing.T) {
		sess := session.NewSession("test-session")

		assert.False(t, sess.IsConnected())
	})

	t.Run("should return false for connecting session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.SetConnecting()

		assert.False(t, sess.IsConnected())
	})

	t.Run("should return true after successful connection", func(t *testing.T) {
		sess := session.NewSession("test-session")

		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		assert.True(t, sess.IsConnected())
	})
}

func TestIsConnecting(t *testing.T) {
	t.Run("should return true for connecting session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.SetConnecting()

		assert.True(t, sess.IsConnecting())
	})

	t.Run("should return false for disconnected session", func(t *testing.T) {
		sess := session.NewSession("test-session")

		assert.False(t, sess.IsConnecting())
	})

	t.Run("should return false for connected session", func(t *testing.T) {
		sess := session.NewSession("test-session")
		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)

		assert.False(t, sess.IsConnecting())
	})

	t.Run("should return false after disconnect", func(t *testing.T) {
		sess := session.NewSession("test-session")
		sess.SetConnecting()
		assert.True(t, sess.IsConnecting())

		sess.Disconnect()
		assert.False(t, sess.IsConnecting())
	})
}

func TestSessionValidation(t *testing.T) {
	t.Run("should validate correct session", func(t *testing.T) {
		sess := session.NewSession("valid-session")
		err := sess.Validate()
		assert.NoError(t, err)
	})

	t.Run("should reject empty session name", func(t *testing.T) {
		sess := session.RestoreSession(
			session.NewSessionID(),
			"",
			session.StatusDisconnected,
			"",
			"",
			"",
			false,
			time.Now(),
			time.Now(),
		)
		err := sess.Validate()
		assert.Error(t, err)
	})

	t.Run("should validate session with all fields", func(t *testing.T) {
		sess := session.RestoreSession(
			session.NewSessionID(),
			"complete-session",
			session.StatusConnected,
			"test@s.whatsapp.net",
			"qr-code-data",
			"",
			true,
			time.Now(),
			time.Now(),
		)
		err := sess.Validate()
		assert.NoError(t, err)
	})

	t.Run("should validate session in different states", func(t *testing.T) {
		statuses := []session.Status{
			session.StatusDisconnected,
			session.StatusConnecting,
			session.StatusConnected,
		}

		for _, status := range statuses {
			t.Run("status_"+status.String(), func(t *testing.T) {
				sess := session.RestoreSession(
					session.NewSessionID(),
					"test-session",
					status,
					"",
					"",
					"",
					false,
					time.Now(),
					time.Now(),
				)
				err := sess.Validate()
				assert.NoError(t, err)
			})
		}
	})
}

func TestSessionGetters(t *testing.T) {
	t.Run("should return correct ID", func(t *testing.T) {
		sess := session.NewSession("test-session")
		id := sess.ID()

		assert.False(t, id.IsEmpty())
		assert.NotEmpty(t, id.String())
	})

	t.Run("should return correct name", func(t *testing.T) {
		name := "test-session-name"
		sess := session.NewSession(name)

		assert.Equal(t, name, sess.Name())
	})

	t.Run("should return correct status", func(t *testing.T) {
		sess := session.NewSession("test-session")

		assert.Equal(t, session.StatusDisconnected, sess.Status())

		sess.SetConnecting()
		assert.Equal(t, session.StatusConnecting, sess.Status())

		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)
		assert.Equal(t, session.StatusConnected, sess.Status())
	})

	t.Run("should return correct WaJID", func(t *testing.T) {
		sess := session.NewSession("test-session")
		assert.Empty(t, sess.WaJID())

		waJID := "test@s.whatsapp.net"
		err := sess.Connect(waJID)
		require.NoError(t, err)
		assert.Equal(t, waJID, sess.WaJID())
	})

	t.Run("should return correct QRCode", func(t *testing.T) {
		sess := session.NewSession("test-session")
		assert.Empty(t, sess.QRCode())

		qrCode := "test-qr-code"
		sess.SetQRCode(qrCode)
		assert.Equal(t, qrCode, sess.QRCode())
	})

	t.Run("should return correct IsActive", func(t *testing.T) {
		sess := session.NewSession("test-session")
		assert.False(t, sess.IsActive())

		err := sess.Connect("test@s.whatsapp.net")
		require.NoError(t, err)
		assert.True(t, sess.IsActive())

		sess.Disconnect()
		assert.False(t, sess.IsActive())
	})

	t.Run("should return correct timestamps", func(t *testing.T) {
		before := time.Now()
		sess := session.NewSession("test-session")
		after := time.Now()

		createdAt := sess.CreatedAt()
		updatedAt := sess.UpdatedAt()

		assert.False(t, createdAt.IsZero())
		assert.False(t, updatedAt.IsZero())
		assert.True(t, createdAt.After(before) || createdAt.Equal(before))
		assert.True(t, createdAt.Before(after) || createdAt.Equal(after))
		assert.True(t, updatedAt.After(before) || updatedAt.Equal(before))
		assert.True(t, updatedAt.Before(after) || updatedAt.Equal(after))
	})
}
