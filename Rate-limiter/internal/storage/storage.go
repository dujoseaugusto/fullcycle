package storage

import "context"

// Storage é a interface que define os métodos para persistência de dados do rate limiter
type Storage interface {
	// Increment incrementa o contador para uma chave e retorna o novo valor
	Increment(ctx context.Context, key string, ttl int64) (int64, error)
	
	// GetCount obtém o contador atual para uma chave
	GetCount(ctx context.Context, key string) (int64, error)
	
	// Reset reseta o contador para uma chave
	Reset(ctx context.Context, key string) error
	
	// IsBlocked verifica se uma chave está bloqueada
	IsBlocked(ctx context.Context, key string) (bool, error)
	
	// Block bloqueia uma chave por um tempo específico
	Block(ctx context.Context, key string, duration int64) error
	
	// Unblock desbloqueia uma chave
	Unblock(ctx context.Context, key string) error
	
	// Health verifica a saúde da conexão com o storage
	Health(ctx context.Context) error
}
