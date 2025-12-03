package limiter

import (
	"context"
	"testing"
)

// MockStorage implementa a interface Storage para testes
type MockStorage struct {
	counters map[string]int64
	blocked  map[string]bool
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		counters: make(map[string]int64),
		blocked:  make(map[string]bool),
	}
}

func (m *MockStorage) Increment(ctx context.Context, key string, ttl int64) (int64, error) {
	m.counters[key]++
	return m.counters[key], nil
}

func (m *MockStorage) GetCount(ctx context.Context, key string) (int64, error) {
	return m.counters[key], nil
}

func (m *MockStorage) Reset(ctx context.Context, key string) error {
	delete(m.counters, key)
	return nil
}

func (m *MockStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	return m.blocked[key], nil
}

func (m *MockStorage) Block(ctx context.Context, key string, duration int64) error {
	m.blocked[key] = true
	m.counters[key] = 0
	return nil
}

func (m *MockStorage) Unblock(ctx context.Context, key string) error {
	delete(m.blocked, key)
	return nil
}

func (m *MockStorage) Health(ctx context.Context) error {
	return nil
}

// TestCheckLimitWithIP testa o rate limiting por IP
func TestCheckLimitWithIP(t *testing.T) {
	st := NewMockStorage()
	config := &Config{
		IPLimit:         5,
		IPBlockDuration: 10,
		TokenEnabled:    false,
		Tokens:          make(map[string]*TokenConfig),
	}

	limiter := NewRateLimiter(config, st)
	ctx := context.Background()

	// Deve permitir as primeiras 5 requisições
	for i := 1; i <= 5; i++ {
		allowed, remaining, _, err := limiter.CheckLimit(ctx, "192.168.1.1", false)
		if err != nil {
			t.Fatalf("CheckLimit failed: %v", err)
		}
		if !allowed {
			t.Fatalf("Request %d should be allowed", i)
		}
		if remaining != int64(5-i) {
			t.Fatalf("Expected %d remaining requests, got %d", 5-i, remaining)
		}
	}

	// A 6ª requisição deve ser bloqueada
	allowed, _, _, err := limiter.CheckLimit(ctx, "192.168.1.1", false)
	if err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}
	if allowed {
		t.Fatalf("Request 6 should be blocked")
	}

	// IP deve estar bloqueado agora
	blocked, err := st.IsBlocked(ctx, "192.168.1.1")
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if !blocked {
		t.Fatalf("IP should be blocked")
	}
}

// TestCheckLimitWithToken testa o rate limiting por token
func TestCheckLimitWithToken(t *testing.T) {
	st := NewMockStorage()
	config := &Config{
		IPLimit:         5,
		IPBlockDuration: 10,
		TokenEnabled:    true,
		Tokens: map[string]*TokenConfig{
			"abc123": {
				Token:         "abc123",
				RequestLimit:  10,
				BlockDuration: 20,
			},
		},
	}

	limiter := NewRateLimiter(config, st)
	ctx := context.Background()

	// Token deve ter limite de 10
	for i := 1; i <= 10; i++ {
		allowed, _, _, err := limiter.CheckLimit(ctx, "abc123", true)
		if err != nil {
			t.Fatalf("CheckLimit failed: %v", err)
		}
		if !allowed {
			t.Fatalf("Request %d should be allowed for token", i)
		}
	}

	// 11ª requisição deve ser bloqueada
	allowed, _, blockDuration, err := limiter.CheckLimit(ctx, "abc123", true)
	if err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}
	if allowed {
		t.Fatalf("Request 11 should be blocked for token")
	}
	if blockDuration != 20 {
		t.Fatalf("Expected block duration 20, got %d", blockDuration)
	}
}

// TestTokenOverridesIP testa que token override o limite de IP
func TestTokenOverridesIP(t *testing.T) {
	st := NewMockStorage()
	config := &Config{
		IPLimit:         5,
		IPBlockDuration: 10,
		TokenEnabled:    true,
		Tokens: map[string]*TokenConfig{
			"premium": {
				Token:         "premium",
				RequestLimit:  100,
				BlockDuration: 300,
			},
		},
	}

	limiter := NewRateLimiter(config, st)
	ctx := context.Background()

	// Token premium deve suportar 100 requisições
	for i := 1; i <= 100; i++ {
		allowed, _, _, err := limiter.CheckLimit(ctx, "premium", true)
		if err != nil {
			t.Fatalf("CheckLimit failed: %v", err)
		}
		if !allowed {
			t.Fatalf("Request %d should be allowed for premium token", i)
		}
	}

	// 101ª requisição deve ser bloqueada
	allowed, _, _, err := limiter.CheckLimit(ctx, "premium", true)
	if err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}
	if allowed {
		t.Fatalf("Request 101 should be blocked for premium token")
	}
}

// TestParseTokenConfig testa o parse de configuração de tokens
func TestParseTokenConfig(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int
		shouldErr bool
	}{
		{
			name:     "Empty config",
			input:    "",
			expected: 0,
		},
		{
			name:     "Single token",
			input:    "token1:10:300",
			expected: 1,
		},
		{
			name:     "Multiple tokens",
			input:    "token1:10:300;token2:20:600",
			expected: 2,
		},
		{
			name:      "Invalid format",
			input:     "invalid",
			shouldErr: true,
		},
		{
			name:      "Invalid limit",
			input:     "token:notanumber:300",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTokenConfig(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(result) != tt.expected {
				t.Fatalf("Expected %d tokens, got %d", tt.expected, len(result))
			}
		})
	}
}

// TestBlockedIdentifierIsRejected testa que identificadores bloqueados são rejeitados
func TestBlockedIdentifierIsRejected(t *testing.T) {
	st := NewMockStorage()
	config := &Config{
		IPLimit:         2,
		IPBlockDuration: 10,
		TokenEnabled:    false,
		Tokens:          make(map[string]*TokenConfig),
	}

	limiter := NewRateLimiter(config, st)
	ctx := context.Background()

	// Realiza 2 requisições (limite atingido)
	for i := 0; i < 2; i++ {
		limiter.CheckLimit(ctx, "10.0.0.1", false)
	}

	// 3ª requisição bloqueia o IP
	limiter.CheckLimit(ctx, "10.0.0.1", false)

	// Próximas requisições devem ser rejeitadas
	allowed, _, _, err := limiter.CheckLimit(ctx, "10.0.0.1", false)
	if err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}
	if allowed {
		t.Fatalf("Blocked IP should not be allowed")
	}
}
