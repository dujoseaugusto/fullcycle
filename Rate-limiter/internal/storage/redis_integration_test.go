//go:build integration

package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisStorageIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}
	defer container.Terminate(ctx)

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get host: %v", err)
	}

	st, err := NewRedisStorage(host, port.Port(), "")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		if redisSt, ok := st.(*RedisStorage); ok {
			redisSt.Close()
		}
	}()

	t.Run("Increment", func(t *testing.T) {
		testCtx := context.Background()
		key := fmt.Sprintf("test:increment:%d", time.Now().Unix())

		for i := 0; i < 3; i++ {
			count, err := st.Increment(testCtx, key, 10)
			if err != nil {
				t.Fatalf("Increment failed: %v", err)
			}
			if count != int64(i+1) {
				t.Errorf("Expected %d, got %d", i+1, count)
			}
		}
	})

	t.Run("Block", func(t *testing.T) {
		testCtx := context.Background()
		key := fmt.Sprintf("test:block:%d", time.Now().Unix())

		err := st.Block(testCtx, key, 10)
		if err != nil {
			t.Fatalf("Block failed: %v", err)
		}

		isBlocked, err := st.IsBlocked(testCtx, key)
		if err != nil {
			t.Fatalf("IsBlocked failed: %v", err)
		}
		if !isBlocked {
			t.Errorf("Should be blocked")
		}
	})
}
