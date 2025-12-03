package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujoseaugusto/stress-test/internal/limiter"
)

// MockStorageForMiddleware implementa Storage para testes do middleware
type MockStorageForMiddleware struct {
	counters map[string]int64
	blocked  map[string]bool
}

func NewMockStorageForMiddleware() *MockStorageForMiddleware {
	return &MockStorageForMiddleware{
		counters: make(map[string]int64),
		blocked:  make(map[string]bool),
	}
}

func (m *MockStorageForMiddleware) Increment(ctx context.Context, key string, ttl int64) (int64, error) {
	m.counters[key]++
	return m.counters[key], nil
}

func (m *MockStorageForMiddleware) GetCount(ctx context.Context, key string) (int64, error) {
	return m.counters[key], nil
}

func (m *MockStorageForMiddleware) Reset(ctx context.Context, key string) error {
	delete(m.counters, key)
	return nil
}

func (m *MockStorageForMiddleware) IsBlocked(ctx context.Context, key string) (bool, error) {
	return m.blocked[key], nil
}

func (m *MockStorageForMiddleware) Block(ctx context.Context, key string, duration int64) error {
	m.blocked[key] = true
	m.counters[key] = 0
	return nil
}

func (m *MockStorageForMiddleware) Unblock(ctx context.Context, key string) error {
	delete(m.blocked, key)
	return nil
}

func (m *MockStorageForMiddleware) Health(ctx context.Context) error {
	return nil
}

// TestMiddlewareAllowsRequests testa que o middleware permite requisições dentro do limite
func TestMiddlewareAllowsRequests(t *testing.T) {
	st := NewMockStorageForMiddleware()
	config := &limiter.Config{
		IPLimit:         5,
		IPBlockDuration: 10,
		TokenEnabled:    false,
		Tokens:          make(map[string]*limiter.TokenConfig),
	}

	rateLimiter := limiter.NewRateLimiter(config, st)
	middleware := NewRateLimiterMiddleware(rateLimiter)

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Faz 5 requisições (limite)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Request %d should be allowed (got %d)", i+1, w.Code)
		}
	}

	// 6ª requisição deve ser rejeitada
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Request 6 should be blocked (got %d)", w.Code)
	}

	if w.Body.String() != BlockedMessage+"\n" {
		t.Fatalf("Expected blocked message, got: %s", w.Body.String())
	}
}

// TestMiddlewareWithToken testa que o middleware respeita limites de token
func TestMiddlewareWithToken(t *testing.T) {
	st := NewMockStorageForMiddleware()
	config := &limiter.Config{
		IPLimit:         5,
		IPBlockDuration: 10,
		TokenEnabled:    true,
		Tokens: map[string]*limiter.TokenConfig{
			"mytoken": {
				Token:         "mytoken",
				RequestLimit:  3,
				BlockDuration: 10,
			},
		},
	}

	rateLimiter := limiter.NewRateLimiter(config, st)
	middleware := NewRateLimiterMiddleware(rateLimiter)

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Faz 3 requisições com token (limite)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "mytoken")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Request %d with token should be allowed (got %d)", i+1, w.Code)
		}
	}

	// 4ª requisição com token deve ser rejeitada
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", "mytoken")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Request 4 with token should be blocked (got %d)", w.Code)
	}
}

// TestExtractIP testa a extração de IP
func TestExtractIP(t *testing.T) {
	tests := []struct {
		name     string
		remoteIP string
		headers  map[string]string
		expected string
	}{
		{
			name:     "Direct IP",
			remoteIP: "192.168.1.1:8080",
			expected: "192.168.1.1",
		},
		{
			name:     "X-Forwarded-For",
			remoteIP: "127.0.0.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1, 10.0.0.2",
			},
			expected: "10.0.0.1",
		},
		{
			name:     "X-Real-IP",
			remoteIP: "127.0.0.1:8080",
			headers: map[string]string{
				"X-Real-IP": "172.16.0.1",
			},
			expected: "172.16.0.1",
		},
		{
			name:     "IP without port",
			remoteIP: "192.168.1.100",
			expected: "192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteIP

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := ExtractIP(req)
			if ip != tt.expected {
				t.Fatalf("Expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

// TestMiddlewareWith429Response testa que a resposta 429 é devolvida
func TestMiddlewareWith429Response(t *testing.T) {
	st := NewMockStorageForMiddleware()
	config := &limiter.Config{
		IPLimit:         1,
		IPBlockDuration: 10,
		TokenEnabled:    false,
		Tokens:          make(map[string]*limiter.TokenConfig),
	}

	rateLimiter := limiter.NewRateLimiter(config, st)
	middleware := NewRateLimiterMiddleware(rateLimiter)

	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 1ª requisição
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 2ª requisição (deve ser bloqueada)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected 429, got %d", w.Code)
	}
}
