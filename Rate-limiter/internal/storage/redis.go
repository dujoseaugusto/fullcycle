package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implementa a interface Storage usando Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage cria uma nova instância de RedisStorage
func NewRedisStorage(host string, port string, password string) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// Increment incrementa o contador para uma chave
func (rs *RedisStorage) Increment(ctx context.Context, key string, ttl int64) (int64, error) {
	pipe := rs.client.Pipeline()

	// Incrementa o valor
	incrCmd := pipe.Incr(ctx, key)

	// Define TTL se a chave ainda não tiver
	pipe.Expire(ctx, key, time.Duration(ttl)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return incrCmd.Val(), nil
}

// GetCount obtém o contador atual para uma chave
func (rs *RedisStorage) GetCount(ctx context.Context, key string) (int64, error) {
	val, err := rs.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	return val, nil
}

// Reset reseta o contador para uma chave
func (rs *RedisStorage) Reset(ctx context.Context, key string) error {
	if err := rs.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}
	return nil
}

// IsBlocked verifica se uma chave está bloqueada
func (rs *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockedKey := fmt.Sprintf("blocked:%s", key)
	val, err := rs.client.Get(ctx, blockedKey).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check if blocked: %w", err)
	}
	return val == "true", nil
}

// Block bloqueia uma chave por um tempo específico
func (rs *RedisStorage) Block(ctx context.Context, key string, duration int64) error {
	blockedKey := fmt.Sprintf("blocked:%s", key)
	if err := rs.client.Set(ctx, blockedKey, "true", time.Duration(duration)*time.Second).Err(); err != nil {
		return fmt.Errorf("failed to block: %w", err)
	}
	// Reseta o contador
	if err := rs.Reset(ctx, key); err != nil {
		return fmt.Errorf("failed to reset counter after blocking: %w", err)
	}
	return nil
}

// Unblock desbloqueia uma chave
func (rs *RedisStorage) Unblock(ctx context.Context, key string) error {
	blockedKey := fmt.Sprintf("blocked:%s", key)
	if err := rs.client.Del(ctx, blockedKey).Err(); err != nil {
		return fmt.Errorf("failed to unblock: %w", err)
	}
	return nil
}

// Health verifica a saúde da conexão com o storage
func (rs *RedisStorage) Health(ctx context.Context) error {
	return rs.client.Ping(ctx).Err()
}

// Close fecha a conexão com Redis
func (rs *RedisStorage) Close() error {
	return rs.client.Close()
}
