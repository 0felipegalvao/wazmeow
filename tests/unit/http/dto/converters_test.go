package dto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wazmeow/internal/domain/session"
	"wazmeow/internal/http/dto"
)

func TestSessionConverter(t *testing.T) {
	t.Run("should convert domain session to response", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")
		sess.Connect("123456789@s.whatsapp.net")
		converter := &dto.SessionConverter{}

		// Act
		response := converter.ToSessionResponse(sess)

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, sess.Name(), response.Name)
		assert.Equal(t, sess.Status().String(), response.Status)
		assert.Equal(t, sess.WaJID(), response.WaJID)
		assert.Equal(t, sess.IsActive(), response.IsActive)
		assert.Equal(t, sess.CreatedAt(), response.CreatedAt)
		assert.Equal(t, sess.UpdatedAt(), response.UpdatedAt)
	})

	t.Run("should convert session list to response", func(t *testing.T) {
		// Arrange
		sess1 := session.NewSession("session-1")
		sess2 := session.NewSession("session-2")
		sessions := []*session.Session{sess1, sess2}
		converter := &dto.SessionConverter{}

		// Act
		response := converter.ToSessionListResponse(sessions, 2)

		// Assert
		assert.Len(t, response.Sessions, 2)
		assert.Equal(t, 2, response.Total)
		assert.Equal(t, sess1.ID().String(), response.Sessions[0].ID)
		assert.Equal(t, sess2.ID().String(), response.Sessions[1].ID)
	})

	t.Run("should handle empty session list", func(t *testing.T) {
		// Arrange
		var sessions []*session.Session
		converter := &dto.SessionConverter{}

		// Act
		response := converter.ToSessionListResponse(sessions, 0)

		// Assert
		assert.Empty(t, response.Sessions)
		assert.Equal(t, 0, response.Total)
	})

	t.Run("should convert create session request", func(t *testing.T) {
		// Arrange
		req := &dto.CreateSessionRequest{
			Name:      "  test-session  ",
			ProxyHost: "proxy.example.com",
			ProxyPort: 8080,
			ProxyType: dto.ProxyTypeHTTP,
			Username:  "user",
			Password:  "pass",
		}
		converter := &dto.SessionConverter{}

		// Act
		result, err := converter.ToCreateSessionRequest(req)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "test-session", result.Name) // Should be trimmed
		assert.Equal(t, "proxy.example.com", result.ProxyHost)
		assert.Equal(t, 8080, result.ProxyPort)
		assert.Equal(t, dto.ProxyTypeHTTP, result.ProxyType)
	})
}

func TestProxyConverter(t *testing.T) {
	t.Run("should extract proxy info from URL", func(t *testing.T) {
		// Arrange
		proxyURL := "http://user:pass@proxy.example.com:8080"
		converter := dto.NewProxyConverter()

		// Act
		config, err := converter.ExtractProxyInfo(proxyURL)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "proxy.example.com", config.Host)
		assert.Equal(t, 8080, config.Port)
		assert.Equal(t, dto.ProxyTypeHTTP, config.Type)
		assert.Equal(t, "user", config.Username)
		assert.Equal(t, "pass", config.Password)
	})

	t.Run("should handle URL without credentials", func(t *testing.T) {
		// Arrange
		proxyURL := "socks5://proxy.example.com:1080"
		converter := dto.NewProxyConverter()

		// Act
		config, err := converter.ExtractProxyInfo(proxyURL)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "proxy.example.com", config.Host)
		assert.Equal(t, 1080, config.Port)
		assert.Equal(t, dto.ProxyTypeSOCKS5, config.Type)
		assert.Empty(t, config.Username)
		assert.Empty(t, config.Password)
	})

	t.Run("should handle empty URL", func(t *testing.T) {
		// Arrange
		converter := dto.NewProxyConverter()

		// Act
		config, err := converter.ExtractProxyInfo("")

		// Assert
		require.NoError(t, err)
		assert.Nil(t, config)
	})

	t.Run("should build proxy URL from config", func(t *testing.T) {
		// Arrange
		config := &dto.ProxyConfigResponse{
			Host:     "proxy.example.com",
			Port:     8080,
			Type:     dto.ProxyTypeHTTP,
			Username: "user",
			Password: "pass",
		}
		converter := dto.NewProxyConverter()

		// Act
		proxyURL, err := converter.BuildProxyURL(config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "http://user:pass@proxy.example.com:8080", proxyURL)
	})

	t.Run("should build URL without credentials", func(t *testing.T) {
		// Arrange
		config := &dto.ProxyConfigResponse{
			Host: "proxy.example.com",
			Port: 8080,
			Type: dto.ProxyTypeHTTP,
		}
		converter := dto.NewProxyConverter()

		// Act
		proxyURL, err := converter.BuildProxyURL(config)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "http://proxy.example.com:8080", proxyURL)
	})

	t.Run("should handle nil config", func(t *testing.T) {
		// Arrange
		converter := dto.NewProxyConverter()

		// Act
		proxyURL, err := converter.BuildProxyURL(nil)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, proxyURL)
	})
}

func TestBatchConverter(t *testing.T) {
	t.Run("should convert small batch sequentially", func(t *testing.T) {
		// Arrange
		sessions := make([]*session.Session, 5)
		for i := 0; i < 5; i++ {
			sessions[i] = session.NewSession("session-" + string(rune('1'+i)))
		}
		converter := dto.NewBatchConverter()

		// Act
		start := time.Now()
		responses := converter.ConvertSessions(sessions)
		duration := time.Since(start)

		// Assert
		assert.Len(t, responses, 5)
		for i, response := range responses {
			assert.Equal(t, sessions[i].ID().String(), response.ID)
		}
		t.Logf("Sequential conversion took: %v", duration)
	})

	t.Run("should convert large batch in parallel", func(t *testing.T) {
		// Arrange
		sessions := make([]*session.Session, 20)
		for i := 0; i < 20; i++ {
			sessions[i] = session.NewSession("session-" + string(rune('1'+i)))
		}
		converter := dto.NewBatchConverter()

		// Act
		start := time.Now()
		responses := converter.ConvertSessions(sessions)
		duration := time.Since(start)

		// Assert
		assert.Len(t, responses, 20)
		for i, response := range responses {
			assert.Equal(t, sessions[i].ID().String(), response.ID)
		}
		t.Logf("Parallel conversion took: %v", duration)
	})

	t.Run("should handle empty batch", func(t *testing.T) {
		// Arrange
		var sessions []*session.Session
		converter := dto.NewBatchConverter()

		// Act
		responses := converter.ConvertSessions(sessions)

		// Assert
		assert.Nil(t, responses)
	})
}

func TestCachedConverter(t *testing.T) {
	t.Run("should cache converted sessions", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")
		converter := dto.NewCachedConverter()

		// Act - first conversion
		start1 := time.Now()
		response1 := converter.ToSessionResponse(sess)
		duration1 := time.Since(start1)

		// Act - second conversion (should use cache)
		start2 := time.Now()
		response2 := converter.ToSessionResponse(sess)
		duration2 := time.Since(start2)

		// Assert
		assert.Equal(t, response1.ID, response2.ID)
		assert.Equal(t, response1.Name, response2.Name)
		t.Logf("First conversion: %v, Second conversion: %v", duration1, duration2)
		// Second conversion should be faster (cached)
		assert.True(t, duration2 <= duration1)
	})

	t.Run("should invalidate cache when session updated", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")
		converter := dto.NewCachedConverter()

		// Act - first conversion
		response1 := converter.ToSessionResponse(sess)

		// Update session
		time.Sleep(1 * time.Millisecond) // Ensure different timestamp
		sess.Connect("123456789@s.whatsapp.net")

		// Act - second conversion (should not use cache due to updated timestamp)
		response2 := converter.ToSessionResponse(sess)

		// Assert
		assert.Equal(t, response1.ID, response2.ID)
		assert.NotEqual(t, response1.Status, response2.Status)
		assert.NotEqual(t, response1.UpdatedAt, response2.UpdatedAt)
	})

	t.Run("should clear cache", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")
		converter := dto.NewCachedConverter()

		// Act - convert and cache
		converter.ToSessionResponse(sess)

		// Clear cache
		converter.ClearCache()

		// Convert again (should not use cache)
		response := converter.ToSessionResponse(sess)

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
	})

	t.Run("should invalidate specific session cache", func(t *testing.T) {
		// Arrange
		sess1 := session.NewSession("session-1")
		sess2 := session.NewSession("session-2")
		converter := dto.NewCachedConverter()

		// Act - convert both sessions
		converter.ToSessionResponse(sess1)
		converter.ToSessionResponse(sess2)

		// Invalidate only sess1
		converter.InvalidateCache(sess1.ID())

		// Convert again
		response1 := converter.ToSessionResponse(sess1)
		response2 := converter.ToSessionResponse(sess2)

		// Assert
		assert.Equal(t, sess1.ID().String(), response1.ID)
		assert.Equal(t, sess2.ID().String(), response2.ID)
	})
}

func TestStreamingConverter(t *testing.T) {
	t.Run("should convert sessions as stream", func(t *testing.T) {
		// Arrange
		sessions := make([]*session.Session, 10)
		for i := 0; i < 10; i++ {
			sessions[i] = session.NewSession("session-" + string(rune('1'+i)))
		}

		sessionChan := make(chan *session.Session, len(sessions))
		for _, sess := range sessions {
			sessionChan <- sess
		}
		close(sessionChan)

		converter := dto.NewStreamingConverter()

		// Act
		responseChan := converter.ConvertSessionsStream(sessionChan)

		// Collect responses
		var responses []*dto.SessionResponse
		for response := range responseChan {
			responses = append(responses, response)
		}

		// Assert
		assert.Len(t, responses, 10)
		for i, response := range responses {
			assert.Equal(t, sessions[i].ID().String(), response.ID)
		}
	})

	t.Run("should handle empty stream", func(t *testing.T) {
		// Arrange
		sessionChan := make(chan *session.Session)
		close(sessionChan)

		converter := dto.NewStreamingConverter()

		// Act
		responseChan := converter.ConvertSessionsStream(sessionChan)

		// Collect responses
		var responses []*dto.SessionResponse
		for response := range responseChan {
			responses = append(responses, response)
		}

		// Assert
		assert.Empty(t, responses)
	})
}

func TestGlobalConverters(t *testing.T) {
	t.Run("should use global converter functions", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")

		// Act
		response := dto.ConvertSession(sess)

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, sess.Name(), response.Name)
	})

	t.Run("should use global batch converter", func(t *testing.T) {
		// Arrange
		sessions := []*session.Session{
			session.NewSession("session-1"),
			session.NewSession("session-2"),
		}

		// Act
		responses := dto.ConvertSessions(sessions)

		// Assert
		assert.Len(t, responses, 2)
		assert.Equal(t, sessions[0].ID().String(), responses[0].ID)
		assert.Equal(t, sessions[1].ID().String(), responses[1].ID)
	})

	t.Run("should use global cached converter", func(t *testing.T) {
		// Arrange
		sess := session.NewSession("test-session")

		// Act
		response := dto.ConvertSessionCached(sess)

		// Assert
		assert.Equal(t, sess.ID().String(), response.ID)
		assert.Equal(t, sess.Name(), response.Name)
	})
}
