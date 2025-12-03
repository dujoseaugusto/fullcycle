package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dujoseaugusto/rate-limiter/internal/limiter"
	mw "github.com/dujoseaugusto/rate-limiter/internal/middleware"
	"github.com/dujoseaugusto/rate-limiter/internal/storage"
)

func main() {
	// Carrega configurações das variáveis de ambiente
	ipLimit, err := strconv.ParseInt(os.Getenv("RATE_LIMITER_IP_LIMIT"), 10, 64)
	if err != nil {
		ipLimit = 10
	}

	ipBlockDuration, err := strconv.ParseInt(os.Getenv("RATE_LIMITER_IP_BLOCK_DURATION"), 10, 64)
	if err != nil {
		ipBlockDuration = 300
	}

	tokenEnabled := os.Getenv("RATE_LIMITER_TOKEN_ENABLED") == "true"
	tokensStr := os.Getenv("RATE_LIMITER_TOKENS")

	// Faz parse das configurações de tokens
	tokens, err := limiter.ParseTokenConfig(tokensStr)
	if err != nil {
		log.Fatalf("Failed to parse token config: %v", err)
	}

	// Cria conexão com Redis
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Conecta ao Redis com retry
	var st storage.Storage
	var lastErr error
	for i := 0; i < 10; i++ {
		st, err = storage.NewRedisStorage(redisHost, redisPort, redisPassword)
		if err == nil {
			break
		}
		lastErr = err
		log.Printf("Failed to connect to Redis (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if st == nil {
		log.Fatalf("Failed to connect to Redis after 10 attempts: %v", lastErr)
	}

	defer func() {
		if redisSt, ok := st.(*storage.RedisStorage); ok {
			redisSt.Close()
		}
	}()

	// Cria o rate limiter
	config := &limiter.Config{
		IPLimit:         ipLimit,
		IPBlockDuration: ipBlockDuration,
		TokenEnabled:    tokenEnabled,
		Tokens:          tokens,
	}

	rateLimiter := limiter.NewRateLimiter(config, st)

	// Cria o middleware
	rateLimiterMiddleware := mw.NewRateLimiterMiddleware(rateLimiter)

	// Cria o mux para rotas protegidas
	protectedMux := http.NewServeMux()

	// Rota de teste (protegida por rate limiting)
	protectedMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to Rate Limiter! Your request was allowed."))
	})

	// Rota para debug (mostra informações de rate limiting)
	protectedMux.HandleFunc("GET /debug", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(mw.HeaderAPIKey)
		ip := mw.ExtractIP(r)

		var identifier string
		if apiKey != "" {
			identifier = apiKey
		} else {
			identifier = ip
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		count, _ := st.GetCount(ctx, identifier)
		isBlocked, _ := st.IsBlocked(ctx, identifier)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"identifier":"%s","count":%d,"blocked":%v}`, identifier, count, isBlocked)
	})

	// Aplica o middleware de rate limiting apenas às rotas protegidas
	protectedHandler := rateLimiterMiddleware.Handler(protectedMux)

	// Cria o mux principal
	mainMux := http.NewServeMux()

	// Rota de health check (NÃO protegida por rate limiting)
	mainMux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Todas as outras rotas passam pelo middleware
	mainMux.Handle("/", protectedHandler)

	// Inicia o servidor
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Rate Limiter Configuration:")
	log.Printf("  IP Limit: %d req/s", ipLimit)
	log.Printf("  IP Block Duration: %d seconds", ipBlockDuration)
	log.Printf("  Token Enabled: %v", tokenEnabled)
	log.Printf("  Tokens: %d configured", len(tokens))

	if err := http.ListenAndServe(addr, mainMux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
