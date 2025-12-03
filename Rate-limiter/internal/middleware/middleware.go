package middleware

import (
	"net/http"
	"strings"

	"github.com/dujoseaugusto/rate-limiter/internal/limiter"
)

const (
	HeaderAPIKey   = "API_KEY"
	BlockedMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"
)

// RateLimiterMiddleware é o middleware HTTP para aplicar rate limiting
type RateLimiterMiddleware struct {
	limiter *limiter.RateLimiter
}

// NewRateLimiterMiddleware cria um novo middleware de rate limiting
func NewRateLimiterMiddleware(rl *limiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: rl,
	}
}

// Handler retorna uma função de middleware HTTP
func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Verifica se há um token no header
		apiKey := r.Header.Get(HeaderAPIKey)

		var identifier string
		var isToken bool

		if apiKey != "" {
			// Token foi fornecido, usa como identificador
			identifier = apiKey
			isToken = true
		} else {
			// Usa IP como identificador
			identifier = ExtractIP(r)
			isToken = false
		}

		// Verifica o limite
		allowed, _, _, err := m.limiter.CheckLimit(ctx, identifier, isToken)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Adiciona headers de rate limit
		w.Header().Set("X-RateLimit-Limit", "")
		if !isToken {
			w.Header().Set("X-RateLimit-Remaining", "")
		}

		if !allowed {
			w.Header().Set("Retry-After", "")
			http.Error(w, BlockedMessage, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ExtractIP extrai o endereço IP da requisição
func ExtractIP(r *http.Request) string {
	// Tenta obter o IP do header X-Forwarded-For (para proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Tenta obter o IP do header X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Retorna o IP remoto direto
	if ip := r.RemoteAddr; ip != "" {
		// Remove a porta se houver
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			return ip[:idx]
		}
		return ip
	}

	return "unknown"
}
