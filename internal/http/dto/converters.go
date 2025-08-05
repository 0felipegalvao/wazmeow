package dto

import (
	"net/url"
	"strconv"
	"sync"

	"wazmeow/internal/domain/session"
)

// ConverterPool manages a pool of converters for better performance
type ConverterPool struct {
	sessionConverters sync.Pool
}

// NewConverterPool creates a new converter pool
func NewConverterPool() *ConverterPool {
	return &ConverterPool{
		sessionConverters: sync.Pool{
			New: func() interface{} {
				return &SessionConverter{}
			},
		},
	}
}

// GetSessionConverter gets a session converter from the pool
func (cp *ConverterPool) GetSessionConverter() *SessionConverter {
	return cp.sessionConverters.Get().(*SessionConverter)
}

// PutSessionConverter returns a session converter to the pool
func (cp *ConverterPool) PutSessionConverter(converter *SessionConverter) {
	converter.Reset()
	cp.sessionConverters.Put(converter)
}

// SessionConverter provides optimized conversion between session domain and DTOs
type SessionConverter struct {
	// Reusable buffers to avoid allocations
	proxyConfig *ProxyConfigResponse
	builder     *SessionResponseBuilder
}

// Reset resets the converter for reuse
func (sc *SessionConverter) Reset() {
	sc.proxyConfig = nil
	sc.builder = nil
}

// ToSessionResponse converts a domain session to HTTP response with optimizations
func (sc *SessionConverter) ToSessionResponse(sess *session.Session) *SessionResponse {
	if sc.builder == nil {
		sc.builder = NewSessionResponseBuilder()
	}

	// Use builder for efficient conversion
	response := sc.builder.FromDomainSession(sess).Build()

	// Reset builder for next use
	sc.builder = NewSessionResponseBuilder()

	return response
}

// ToSessionListResponse converts multiple sessions efficiently
func (sc *SessionConverter) ToSessionListResponse(sessions []*session.Session, total int) *SessionListResponse {
	// Pre-allocate slice with known capacity
	sessionResponses := make([]*SessionResponse, 0, len(sessions))

	for _, sess := range sessions {
		sessionResponses = append(sessionResponses, sc.ToSessionResponse(sess))
	}

	return &SessionListResponse{
		Sessions: sessionResponses,
		Total:    total,
	}
}

// ToCreateSessionRequest converts HTTP request to domain-compatible format
func (sc *SessionConverter) ToCreateSessionRequest(req *CreateSessionRequest) (*CreateSessionRequest, error) {
	// Normalize the request
	req.Normalize()

	// Validate proxy URL if provided
	if req.HasProxy() {
		if _, err := req.BuildProxyURL(); err != nil {
			return nil, err
		}
	}

	return req, nil
}

// ProxyConverter handles proxy-related conversions
type ProxyConverter struct{}

// NewProxyConverter creates a new proxy converter
func NewProxyConverter() *ProxyConverter {
	return &ProxyConverter{}
}

// ExtractProxyInfo extracts proxy information from a proxy URL efficiently
func (pc *ProxyConverter) ExtractProxyInfo(proxyURL string) (*ProxyConfigResponse, error) {
	if proxyURL == "" {
		return nil, nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	config := &ProxyConfigResponse{
		Host: parsedURL.Hostname(),
		Type: ProxyType(parsedURL.Scheme),
	}

	// Parse port
	if portStr := parsedURL.Port(); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Port = port
		}
	}

	// Extract credentials
	if parsedURL.User != nil {
		config.Username = parsedURL.User.Username()
		if password, ok := parsedURL.User.Password(); ok {
			config.Password = password
		}
	}

	return config, nil
}

// BuildProxyURL builds a proxy URL from configuration efficiently
func (pc *ProxyConverter) BuildProxyURL(config *ProxyConfigResponse) (string, error) {
	if config == nil || config.Host == "" {
		return "", nil
	}

	// Build URL efficiently
	u := &url.URL{
		Scheme: config.Type.String(),
		Host:   config.Host,
	}

	// Add port if specified
	if config.Port > 0 {
		u.Host = u.Host + ":" + strconv.Itoa(config.Port)
	}

	// Add credentials if provided
	if config.Username != "" {
		if config.Password != "" {
			u.User = url.UserPassword(config.Username, config.Password)
		} else {
			u.User = url.User(config.Username)
		}
	}

	return u.String(), nil
}

// BatchConverter handles batch conversions efficiently
type BatchConverter struct {
	pool *ConverterPool
}

// NewBatchConverter creates a new batch converter
func NewBatchConverter() *BatchConverter {
	return &BatchConverter{
		pool: NewConverterPool(),
	}
}

// ConvertSessions converts multiple sessions in parallel for better performance
func (bc *BatchConverter) ConvertSessions(sessions []*session.Session) []*SessionResponse {
	if len(sessions) == 0 {
		return nil
	}

	// For small batches, use sequential processing
	if len(sessions) <= 10 {
		return bc.convertSessionsSequential(sessions)
	}

	// For larger batches, use parallel processing
	return bc.convertSessionsParallel(sessions)
}

// convertSessionsSequential converts sessions sequentially
func (bc *BatchConverter) convertSessionsSequential(sessions []*session.Session) []*SessionResponse {
	converter := bc.pool.GetSessionConverter()
	defer bc.pool.PutSessionConverter(converter)

	responses := make([]*SessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		responses = append(responses, converter.ToSessionResponse(sess))
	}

	return responses
}

// convertSessionsParallel converts sessions in parallel
func (bc *BatchConverter) convertSessionsParallel(sessions []*session.Session) []*SessionResponse {
	responses := make([]*SessionResponse, len(sessions))

	// Use worker pool pattern
	const numWorkers = 4
	jobs := make(chan int, len(sessions))

	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			converter := bc.pool.GetSessionConverter()
			defer bc.pool.PutSessionConverter(converter)

			for i := range jobs {
				responses[i] = converter.ToSessionResponse(sessions[i])
			}
		}()
	}

	// Send jobs
	for i := range sessions {
		jobs <- i
	}
	close(jobs)

	// Wait for completion
	wg.Wait()

	return responses
}

// CachedConverter provides caching for frequently converted objects
type CachedConverter struct {
	cache sync.Map // session.SessionID -> *SessionResponse
	pool  *ConverterPool
}

// NewCachedConverter creates a new cached converter
func NewCachedConverter() *CachedConverter {
	return &CachedConverter{
		pool: NewConverterPool(),
	}
}

// ToSessionResponse converts with caching
func (cc *CachedConverter) ToSessionResponse(sess *session.Session) *SessionResponse {
	// Check cache first
	if cached, ok := cc.cache.Load(sess.ID()); ok {
		if response, ok := cached.(*SessionResponse); ok {
			// Verify cache is still valid (check UpdatedAt)
			if response.UpdatedAt.Equal(sess.UpdatedAt()) {
				return response
			}
		}
	}

	// Convert and cache
	converter := cc.pool.GetSessionConverter()
	defer cc.pool.PutSessionConverter(converter)

	response := converter.ToSessionResponse(sess)
	cc.cache.Store(sess.ID(), response)

	return response
}

// InvalidateCache invalidates cache for a specific session
func (cc *CachedConverter) InvalidateCache(sessionID session.SessionID) {
	cc.cache.Delete(sessionID)
}

// ClearCache clears all cached entries
func (cc *CachedConverter) ClearCache() {
	cc.cache.Range(func(key, value interface{}) bool {
		cc.cache.Delete(key)
		return true
	})
}

// StreamingConverter handles streaming conversions for large datasets
type StreamingConverter struct {
	pool *ConverterPool
}

// NewStreamingConverter creates a new streaming converter
func NewStreamingConverter() *StreamingConverter {
	return &StreamingConverter{
		pool: NewConverterPool(),
	}
}

// ConvertSessionsStream converts sessions as a stream
func (sc *StreamingConverter) ConvertSessionsStream(sessions <-chan *session.Session) <-chan *SessionResponse {
	responses := make(chan *SessionResponse)

	go func() {
		defer close(responses)
		converter := sc.pool.GetSessionConverter()
		defer sc.pool.PutSessionConverter(converter)

		for sess := range sessions {
			responses <- converter.ToSessionResponse(sess)
		}
	}()

	return responses
}

// Global converter instances for convenience
var (
	defaultConverter = NewBatchConverter()
	cachedConverter  = NewCachedConverter()
	streamConverter  = NewStreamingConverter()
)

// Convenience functions using global converters

// ConvertSession converts a single session using the default converter
func ConvertSession(sess *session.Session) *SessionResponse {
	converter := defaultConverter.pool.GetSessionConverter()
	defer defaultConverter.pool.PutSessionConverter(converter)
	return converter.ToSessionResponse(sess)
}

// ConvertSessions converts multiple sessions using the batch converter
func ConvertSessions(sessions []*session.Session) []*SessionResponse {
	return defaultConverter.ConvertSessions(sessions)
}

// ConvertSessionCached converts a session using the cached converter
func ConvertSessionCached(sess *session.Session) *SessionResponse {
	return cachedConverter.ToSessionResponse(sess)
}

// ConvertSessionsStream converts sessions as a stream
func ConvertSessionsStream(sessions <-chan *session.Session) <-chan *SessionResponse {
	return streamConverter.ConvertSessionsStream(sessions)
}
