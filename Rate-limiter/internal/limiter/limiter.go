package limiter

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dujoseaugusto/rate-limiter/internal/storage"
)

// TokenConfig representa a configuração de um token específico
type TokenConfig struct {
	Token         string
	RequestLimit  int64
	BlockDuration int64
}

// Config representa a configuração do rate limiter
type Config struct {
	IPLimit         int64
	IPBlockDuration int64
	TokenEnabled    bool
	Tokens          map[string]*TokenConfig
}

// RateLimiter é a estrutura principal do rate limiter
type RateLimiter struct {
	config  *Config
	storage storage.Storage
}

// NewRateLimiter cria uma nova instância de RateLimiter
func NewRateLimiter(cfg *Config, st storage.Storage) *RateLimiter {
	return &RateLimiter{
		config:  cfg,
		storage: st,
	}
}

// CheckLimit verifica se uma requisição deve ser permitida
// Retorna (allowed, remainingRequests, retryAfter, error)
func (rl *RateLimiter) CheckLimit(ctx context.Context, identifier string, isToken bool) (bool, int64, int64, error) {
	// Verifica se o identificador está bloqueado
	isBlocked, err := rl.storage.IsBlocked(ctx, identifier)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to check if blocked: %w", err)
	}

	if isBlocked {
		return false, 0, 0, nil
	}

	// Determina o limite e a duração do bloqueio
	var limit int64
	var blockDuration int64

	if isToken && rl.config.TokenEnabled {
		tokenCfg, exists := rl.config.Tokens[identifier]
		if exists {
			limit = tokenCfg.RequestLimit
			blockDuration = tokenCfg.BlockDuration
		} else {
			// Se o token não está configurado, usa o limite de IP
			limit = rl.config.IPLimit
			blockDuration = rl.config.IPBlockDuration
		}
	} else {
		limit = rl.config.IPLimit
		blockDuration = rl.config.IPBlockDuration
	}

	// Incrementa o contador
	count, err := rl.storage.Increment(ctx, identifier, blockDuration)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Verifica se o limite foi excedido
	if count > limit {
		// Bloqueia o identificador
		if err := rl.storage.Block(ctx, identifier, blockDuration); err != nil {
			return false, 0, 0, fmt.Errorf("failed to block: %w", err)
		}
		return false, 0, blockDuration, nil
	}

	remainingRequests := limit - count
	return true, remainingRequests, 0, nil
}

// ParseTokenConfig faz parse da configuração de tokens a partir de uma string
// Formato: "token1:limit:duration;token2:limit:duration"
func ParseTokenConfig(tokenStr string) (map[string]*TokenConfig, error) {
	tokens := make(map[string]*TokenConfig)

	if tokenStr == "" {
		return tokens, nil
	}

	for _, tokenPair := range strings.Split(tokenStr, ";") {
		parts := strings.Split(strings.TrimSpace(tokenPair), ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid token config format: %s", tokenPair)
		}

		token := strings.TrimSpace(parts[0])
		limit, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid limit for token %s: %w", token, err)
		}

		duration, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid duration for token %s: %w", token, err)
		}

		tokens[token] = &TokenConfig{
			Token:         token,
			RequestLimit:  limit,
			BlockDuration: duration,
		}
	}

	return tokens, nil
}

// Reset reseta o contador de um identificador
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	return rl.storage.Reset(ctx, identifier)
}

// Unblock desbloqueia um identificador
func (rl *RateLimiter) Unblock(ctx context.Context, identifier string) error {
	return rl.storage.Unblock(ctx, identifier)
}
