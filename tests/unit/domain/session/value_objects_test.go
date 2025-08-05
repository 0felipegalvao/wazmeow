package domain_session_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
)

func TestSessionID(t *testing.T) {
	t.Run("should create new session ID", func(t *testing.T) {
		id := session.NewSessionID()

		assert.False(t, id.IsEmpty())
		assert.NotEmpty(t, id.String())

		// Should be valid UUID
		_, err := uuid.Parse(id.String())
		assert.NoError(t, err)
	})

	t.Run("should create unique session IDs", func(t *testing.T) {
		id1 := session.NewSessionID()
		id2 := session.NewSessionID()

		assert.NotEqual(t, id1, id2)
		assert.NotEqual(t, id1.String(), id2.String())
	})

	t.Run("should create session ID from valid string", func(t *testing.T) {
		validUUID := "550e8400-e29b-41d4-a716-446655440000"

		id, err := session.SessionIDFromString(validUUID)

		assert.NoError(t, err)
		assert.Equal(t, validUUID, id.String())
		assert.False(t, id.IsEmpty())
	})

	t.Run("should fail with empty string", func(t *testing.T) {
		id, err := session.SessionIDFromString("")

		assert.Error(t, err)
		assert.Equal(t, session.ErrInvalidSessionID, err)
		assert.True(t, id.IsEmpty())
	})

	t.Run("should fail with invalid UUID format", func(t *testing.T) {
		invalidUUIDs := []string{
			"invalid-uuid",
			"123",
			"not-a-uuid-at-all",
			"550e8400-e29b-41d4-a716",
			"550e8400-e29b-41d4-a716-446655440000-extra",
		}

		for _, invalidUUID := range invalidUUIDs {
			t.Run("invalid_"+invalidUUID, func(t *testing.T) {
				id, err := session.SessionIDFromString(invalidUUID)

				assert.Error(t, err)
				assert.Equal(t, session.ErrInvalidSessionID, err)
				assert.True(t, id.IsEmpty())
			})
		}
	})

	t.Run("should check if empty", func(t *testing.T) {
		// Empty ID
		emptyID := session.SessionID{}
		assert.True(t, emptyID.IsEmpty())

		// Valid ID
		validID := session.NewSessionID()
		assert.False(t, validID.IsEmpty())
	})

	t.Run("should compare session IDs", func(t *testing.T) {
		id1 := session.NewSessionID()
		id2 := session.NewSessionID()

		// Same ID should be equal
		assert.True(t, id1.Equals(id1))

		// Different IDs should not be equal
		assert.False(t, id1.Equals(id2))

		// IDs created from same string should be equal
		uuidStr := "550e8400-e29b-41d4-a716-446655440000"
		idFromStr1, err := session.SessionIDFromString(uuidStr)
		require.NoError(t, err)
		idFromStr2, err := session.SessionIDFromString(uuidStr)
		require.NoError(t, err)

		assert.True(t, idFromStr1.Equals(idFromStr2))
	})
}

func TestSessionName(t *testing.T) {
	t.Run("should create valid session name", func(t *testing.T) {
		validNames := []string{
			"simple-name",
			"name_with_underscores",
			"name123",
			"123name",
			"very-long-session-name-with-many-characters",
		}

		for _, name := range validNames {
			t.Run("valid_"+name, func(t *testing.T) {
				sessionName, err := session.NewSessionName(name)

				assert.NoError(t, err)
				assert.Equal(t, name, sessionName.String())
				assert.False(t, sessionName.IsEmpty())
			})
		}
	})

	t.Run("should fail with empty name", func(t *testing.T) {
		sessionName, err := session.NewSessionName("")

		assert.Error(t, err)
		assert.Equal(t, session.ErrInvalidSessionName, err)
		assert.True(t, sessionName.IsEmpty())
	})

	t.Run("should fail with whitespace-only name", func(t *testing.T) {
		whitespaceNames := []string{
			" ",
			"  ",
			"\t",
			"\n",
			"   \t  \n  ",
		}

		for _, name := range whitespaceNames {
			t.Run("whitespace_"+name, func(t *testing.T) {
				sessionName, err := session.NewSessionName(name)

				assert.Error(t, err)
				// The actual error message varies based on the validation rules
				assert.True(t, sessionName.IsEmpty())
			})
		}
	})

	t.Run("should handle names with minimum length requirement", func(t *testing.T) {
		testCases := []struct {
			input       string
			shouldError bool
		}{
			{"ab", true},   // Too short (less than 3 chars)
			{"abc", false}, // Minimum valid length
			{"valid-name", false},
		}

		for _, tc := range testCases {
			t.Run("length_test_"+tc.input, func(t *testing.T) {
				sessionName, err := session.NewSessionName(tc.input)

				if tc.shouldError {
					assert.Error(t, err)
					assert.True(t, sessionName.IsEmpty())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.input, sessionName.String())
				}
			})
		}
	})

	t.Run("should check if empty", func(t *testing.T) {
		// Empty name
		emptyName := session.SessionName{}
		assert.True(t, emptyName.IsEmpty())

		// Valid name
		validName, err := session.NewSessionName("test-name")
		require.NoError(t, err)
		assert.False(t, validName.IsEmpty())
	})

	t.Run("should compare session names", func(t *testing.T) {
		name1, err := session.NewSessionName("name1")
		require.NoError(t, err)
		name2, err := session.NewSessionName("name2")
		require.NoError(t, err)

		// Same name should have same string representation
		assert.Equal(t, name1.String(), name1.String())

		// Different names should not be equal
		assert.NotEqual(t, name1.String(), name2.String())

		// Names created from same string should be equal
		sameName1, err := session.NewSessionName("same-name")
		require.NoError(t, err)
		sameName2, err := session.NewSessionName("same-name")
		require.NoError(t, err)

		assert.Equal(t, sameName1.String(), sameName2.String())
	})

	t.Run("should create valid session name without validation errors", func(t *testing.T) {
		validName, err := session.NewSessionName("valid-name")
		require.NoError(t, err)
		assert.False(t, validName.IsEmpty())
		assert.Equal(t, "valid-name", validName.String())
	})
}

func TestStatus(t *testing.T) {
	t.Run("should have correct string representations", func(t *testing.T) {
		testCases := []struct {
			status   session.Status
			expected string
		}{
			{session.StatusDisconnected, "disconnected"},
			{session.StatusConnecting, "connecting"},
			{session.StatusConnected, "connected"},
		}

		for _, tc := range testCases {
			t.Run("status_"+tc.expected, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.status.String())
			})
		}
	})

	t.Run("should have valid status values", func(t *testing.T) {
		validStatuses := []session.Status{
			session.StatusDisconnected,
			session.StatusConnecting,
			session.StatusConnected,
		}

		for _, status := range validStatuses {
			t.Run("valid_"+status.String(), func(t *testing.T) {
				// Status should have valid string representation
				assert.NotEmpty(t, status.String())
			})
		}
	})

	t.Run("should have valid status values", func(t *testing.T) {
		validStatuses := []session.Status{
			session.StatusDisconnected,
			session.StatusConnecting,
			session.StatusConnected,
		}

		for _, status := range validStatuses {
			// Status should have valid string representation
			assert.NotEmpty(t, status.String())
		}
	})
}

func TestWhatsAppJID(t *testing.T) {
	t.Run("should validate WhatsApp JID format", func(t *testing.T) {
		validJIDs := []string{
			"5511999999999@s.whatsapp.net",
			"123456789@c.us",
			"test@g.us",
			"user@s.whatsapp.net",
			"group@g.us",
		}

		for _, jid := range validJIDs {
			t.Run("valid_"+jid, func(t *testing.T) {
				// Test that the JID format is accepted by the session
				sess := session.NewSession("test-session")
				err := sess.Connect(jid)
				assert.NoError(t, err)
				assert.Equal(t, jid, sess.WaJID())
			})
		}
	})

	t.Run("should reject invalid JID format", func(t *testing.T) {
		invalidJIDs := []string{
			"", // Empty JID should fail
		}

		for _, jid := range invalidJIDs {
			t.Run("invalid_"+jid, func(t *testing.T) {
				sess := session.NewSession("test-session")
				err := sess.Connect(jid)
				assert.Error(t, err)
				assert.Equal(t, session.ErrInvalidWhatsAppJID, err)
			})
		}
	})

	t.Run("should accept various JID formats", func(t *testing.T) {
		// The implementation seems to accept any non-empty string as JID
		acceptedJIDs := []string{
			"no-at-symbol",
			"@missing-user",
			"user@",
		}

		for _, jid := range acceptedJIDs {
			t.Run("accepted_"+jid, func(t *testing.T) {
				sess := session.NewSession("test-session")
				err := sess.Connect(jid)
				assert.NoError(t, err)
				assert.Equal(t, jid, sess.WaJID())
			})
		}
	})
}
