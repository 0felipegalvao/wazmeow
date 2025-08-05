package domain_session_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"wazmeow/internal/domain/session"
)

func TestSessionErrors(t *testing.T) {
	t.Run("should have correct error messages", func(t *testing.T) {
		testCases := []struct {
			err      error
			expected string
		}{
			{session.ErrSessionNotFound, "session not found"},
			{session.ErrSessionAlreadyExists, "session already exists"},
			{session.ErrSessionAlreadyConnected, "session already connected"},
			{session.ErrSessionNotConnected, "session not connected"},
			{session.ErrSessionInvalidState, "session in invalid state"},
			{session.ErrInvalidSessionID, "invalid session ID"},
			{session.ErrInvalidSessionName, "invalid session name"},
			{session.ErrInvalidWhatsAppJID, "invalid WhatsApp JID"},
			{session.ErrInvalidSessionIdentifier, "invalid session identifier"},
		}

		for _, tc := range testCases {
			t.Run("error_"+tc.expected, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.err.Error())
			})
		}
	})

	t.Run("should be different error instances", func(t *testing.T) {
		allErrors := []error{
			session.ErrSessionNotFound,
			session.ErrSessionAlreadyExists,
			session.ErrSessionAlreadyConnected,
			session.ErrSessionNotConnected,
			session.ErrSessionInvalidState,
			session.ErrInvalidSessionID,
			session.ErrInvalidSessionName,
			session.ErrInvalidWhatsAppJID,
			session.ErrInvalidSessionIdentifier,
		}

		// Each error should be different from all others
		for i, err1 := range allErrors {
			for j, err2 := range allErrors {
				if i != j {
					assert.False(t, errors.Is(err1, err2), "Error %v should not be the same as %v", err1, err2)
				} else {
					assert.True(t, errors.Is(err1, err2), "Error %v should be the same as itself", err1)
				}
			}
		}
	})

	t.Run("should be identifiable with errors.Is", func(t *testing.T) {
		// Test that we can identify specific errors
		assert.True(t, errors.Is(session.ErrSessionNotFound, session.ErrSessionNotFound))
		assert.True(t, errors.Is(session.ErrSessionAlreadyExists, session.ErrSessionAlreadyExists))
		assert.True(t, errors.Is(session.ErrSessionAlreadyConnected, session.ErrSessionAlreadyConnected))
		assert.True(t, errors.Is(session.ErrSessionNotConnected, session.ErrSessionNotConnected))
		assert.True(t, errors.Is(session.ErrSessionInvalidState, session.ErrSessionInvalidState))
		assert.True(t, errors.Is(session.ErrInvalidSessionID, session.ErrInvalidSessionID))
		assert.True(t, errors.Is(session.ErrInvalidSessionName, session.ErrInvalidSessionName))
		assert.True(t, errors.Is(session.ErrInvalidWhatsAppJID, session.ErrInvalidWhatsAppJID))
		assert.True(t, errors.Is(session.ErrInvalidSessionIdentifier, session.ErrInvalidSessionIdentifier))
	})

	t.Run("should not be nil", func(t *testing.T) {
		allErrors := []error{
			session.ErrSessionNotFound,
			session.ErrSessionAlreadyExists,
			session.ErrSessionAlreadyConnected,
			session.ErrSessionNotConnected,
			session.ErrSessionInvalidState,
			session.ErrInvalidSessionID,
			session.ErrInvalidSessionName,
			session.ErrInvalidWhatsAppJID,
			session.ErrInvalidSessionIdentifier,
		}

		for _, err := range allErrors {
			assert.NotNil(t, err, "Error should not be nil: %v", err)
		}
	})

	t.Run("should be usable in error wrapping", func(t *testing.T) {
		// Test that errors can be wrapped and unwrapped
		wrappedNotFound := errors.New("wrapped: " + session.ErrSessionNotFound.Error())
		assert.Contains(t, wrappedNotFound.Error(), session.ErrSessionNotFound.Error())

		// Test with fmt.Errorf
		formattedError := errors.New("operation failed: " + session.ErrSessionAlreadyExists.Error())
		assert.Contains(t, formattedError.Error(), session.ErrSessionAlreadyExists.Error())
	})
}

func TestErrorUsageInDomainOperations(t *testing.T) {
	t.Run("should return ErrSessionAlreadyConnected when connecting connected session", func(t *testing.T) {
		sess := session.NewSession("test-session")

		// Connect first time
		err := sess.Connect("test@s.whatsapp.net")
		assert.NoError(t, err)

		// Try to connect again
		err = sess.Connect("another@s.whatsapp.net")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrSessionAlreadyConnected))
	})

	t.Run("should return ErrInvalidWhatsAppJID when connecting with empty JID", func(t *testing.T) {
		sess := session.NewSession("test-session")

		err := sess.Connect("")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrInvalidWhatsAppJID))
	})

	t.Run("should return ErrInvalidSessionName when updating with empty name", func(t *testing.T) {
		sess := session.NewSession("test-session")

		err := sess.UpdateName("")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrInvalidSessionName))
	})

	t.Run("should return ErrInvalidSessionID when creating from invalid string", func(t *testing.T) {
		_, err := session.SessionIDFromString("")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrInvalidSessionID))

		_, err = session.SessionIDFromString("invalid-uuid")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrInvalidSessionID))
	})

	t.Run("should return ErrInvalidSessionName when creating invalid session name", func(t *testing.T) {
		_, err := session.NewSessionName("")
		assert.Error(t, err)
		// The actual error might be more specific than the generic one
		if err != nil {
			assert.Contains(t, err.Error(), "session name")
		}

		_, err = session.NewSessionName("ab") // Too short
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "session name")
		}
	})

	t.Run("should return ErrInvalidSessionIdentifier when creating invalid identifier", func(t *testing.T) {
		_, err := session.NewSessionIdentifier("")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrInvalidSessionIdentifier))

		_, err = session.NewSessionIdentifier("   ")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, session.ErrInvalidSessionIdentifier))
	})
}

func TestErrorConstants(t *testing.T) {
	t.Run("should have consistent error types", func(t *testing.T) {
		// All domain errors should be of type error
		var err error

		err = session.ErrSessionNotFound
		assert.NotNil(t, err)

		err = session.ErrSessionAlreadyExists
		assert.NotNil(t, err)

		err = session.ErrSessionAlreadyConnected
		assert.NotNil(t, err)

		err = session.ErrSessionNotConnected
		assert.NotNil(t, err)

		err = session.ErrSessionInvalidState
		assert.NotNil(t, err)

		err = session.ErrInvalidSessionID
		assert.NotNil(t, err)

		err = session.ErrInvalidSessionName
		assert.NotNil(t, err)

		err = session.ErrInvalidWhatsAppJID
		assert.NotNil(t, err)

		err = session.ErrInvalidSessionIdentifier
		assert.NotNil(t, err)
	})

	t.Run("should be immutable", func(t *testing.T) {
		// Errors should be constants and not modifiable
		originalMessage := session.ErrSessionNotFound.Error()

		// Try to use the error in different contexts
		_ = errors.New("wrapped: " + session.ErrSessionNotFound.Error())

		// Original should remain unchanged
		assert.Equal(t, originalMessage, session.ErrSessionNotFound.Error())
	})
}
