package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew_Default(t *testing.T) {
	t.Setenv("GRPC_PORT", "9090")
	t.Setenv("GRPC_GATEWAY_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "127.0.0.1")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "Library")
	t.Setenv("POSTGRES_USER", "redkin")
	t.Setenv("POSTGRES_PASSWORD", "00000000")
	t.Setenv("POSTGRES_MAX_CONN", "10")
	t.Setenv("OUTBOX_ENABLED", "false")

	cfg, err := New()
	require.NoError(t, err)

	require.Equal(t, "9090", cfg.GRPC.Port)
	require.Equal(t, "8080", cfg.GRPC.GatewayPort)

	require.Equal(t, "127.0.0.1", cfg.PG.Host)
	require.Equal(t, "5432", cfg.PG.Port)
	require.Equal(t, "Library", cfg.PG.DB)
	require.Equal(t, "redkin", cfg.PG.User)
	require.Equal(t, "00000000", cfg.PG.Password)
	require.Equal(t, "10", cfg.PG.MaxConn)

	expectedURL := "postgres://redkin:00000000@127.0.0.1:5432/Library?sslmode=disable"
	require.Equal(t, expectedURL, cfg.PG.URL)

	require.False(t, cfg.Outbox.Enabled)
	require.Zero(t, cfg.Outbox.Workers)
	require.Zero(t, cfg.Outbox.BatchSize)
	require.Zero(t, cfg.Outbox.WaitTimeMS)
	require.Zero(t, cfg.Outbox.InProgressTTLMS)
	require.Empty(t, cfg.Outbox.BookSendURL)
	require.Empty(t, cfg.Outbox.AuthorSendURL)
}

func TestNew_OutboxEnabled(t *testing.T) {
	t.Setenv("GRPC_PORT", "10000")
	t.Setenv("GRPC_GATEWAY_PORT", "11000")
	t.Setenv("POSTGRES_HOST", "192.168.1.1")
	t.Setenv("POSTGRES_PORT", "6543")
	t.Setenv("POSTGRES_DB", "TestDB")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_MAX_CONN", "5")

	t.Setenv("OUTBOX_ENABLED", "true")
	t.Setenv("OUTBOX_WORKERS", "4")
	t.Setenv("OUTBOX_BATCH_SIZE", "20")
	t.Setenv("OUTBOX_WAIT_TIME_MS", "150")
	t.Setenv("OUTBOX_IN_PROGRESS_TTL_MS", "300")
	t.Setenv("OUTBOX_BOOK_SEND_URL", "book-url")
	t.Setenv("OUTBOX_AUTHOR_SEND_URL", "author-url")

	cfg, err := New()
	require.NoError(t, err)

	require.Equal(t, "192.168.1.1", cfg.PG.Host)
	require.Equal(t, "6543", cfg.PG.Port)
	require.Equal(t, "TestDB", cfg.PG.DB)
	require.Equal(t, "testuser", cfg.PG.User)
	require.Equal(t, "testpass", cfg.PG.Password)
	require.Equal(t, "5", cfg.PG.MaxConn)

	expectedURL := "postgres://testuser:testpass@192.168.1.1:6543/TestDB?sslmode=disable"
	require.Equal(t, expectedURL, cfg.PG.URL)

	require.True(t, cfg.Outbox.Enabled)
	require.Equal(t, 4, cfg.Outbox.Workers)
	require.Equal(t, 20, cfg.Outbox.BatchSize)
	require.Equal(t, 150*time.Millisecond, cfg.Outbox.WaitTimeMS)
	require.Equal(t, 300*time.Millisecond, cfg.Outbox.InProgressTTLMS)
	require.Equal(t, "book-url", cfg.Outbox.BookSendURL)
	require.Equal(t, "author-url", cfg.Outbox.AuthorSendURL)
}
