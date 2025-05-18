package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type (
	Config struct {
		GRPC struct {
			Port        string `env:"GRPC_PORT"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT"`
		}

		PG struct {
			URL      string
			Host     string `env:"POSTGRES_HOST"`
			Port     string `env:"POSTGRES_PORT"`
			DB       string `env:"POSTGRES_DB"`
			User     string `env:"POSTGRES_USER"`
			Password string `env:"POSTGRES_PASSWORD"`
			MaxConn  string `env:"POSTGRES_MAX_CONN"`
		}

		Outbox struct {
			Enabled         bool          `env:"OUTBOX_ENABLED"`
			Workers         int           `env:"OUTBOX_WORKERS"`
			BatchSize       int           `env:"OUTBOX_BATCH_SIZE"`
			WaitTimeMS      time.Duration `env:"OUTBOX_WAIT_TIME_MS"`
			InProgressTTLMS time.Duration `env:"OUTBOX_IN_PROGRESS_TTL_MS"`
			BookSendURL     string        `env:"OUTBOX_BOOK_SEND_URL"`
			AuthorSendURL   string        `env:"OUTBOX_AUTHOR_SEND_URL"`
		}
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	cfg.GRPC.Port = os.Getenv("GRPC_PORT")
	cfg.GRPC.GatewayPort = os.Getenv("GRPC_GATEWAY_PORT")

	cfg.PG.Host = os.Getenv("POSTGRES_HOST")
	cfg.PG.Port = os.Getenv("POSTGRES_PORT")
	cfg.PG.DB = os.Getenv("POSTGRES_DB")
	cfg.PG.User = os.Getenv("POSTGRES_USER")
	cfg.PG.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.PG.MaxConn = os.Getenv("POSTGRES_MAX_CONN")

	hostPort := net.JoinHostPort(cfg.PG.Host, cfg.PG.Port)
	cfg.PG.URL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.PG.User,
		cfg.PG.Password,
		hostPort,
		cfg.PG.DB,
	)

	var err error
	cfg.Outbox.Enabled, err = strconv.ParseBool(os.Getenv("OUTBOX_ENABLED"))

	if err != nil {
		return nil, err
	}

	if cfg.Outbox.Enabled {
		cfg.Outbox.Workers, err = parseInt(os.Getenv("OUTBOX_WORKERS"))

		if err != nil {
			return nil, err
		}

		cfg.Outbox.BatchSize, err = parseInt(os.Getenv("OUTBOX_BATCH_SIZE"))

		if err != nil {
			return nil, err
		}

		cfg.Outbox.WaitTimeMS, err = parseTime(os.Getenv("OUTBOX_WAIT_TIME_MS"))

		if err != nil {
			return nil, err
		}

		cfg.Outbox.InProgressTTLMS, err = parseTime(os.Getenv("OUTBOX_IN_PROGRESS_TTL_MS"))

		if err != nil {
			return nil, err
		}

		cfg.Outbox.BookSendURL = os.Getenv("OUTBOX_BOOK_SEND_URL")
		cfg.Outbox.AuthorSendURL = os.Getenv("OUTBOX_AUTHOR_SEND_URL")
	}

	return cfg, nil
}

func parseTime(s string) (time.Duration, error) {
	t, err := parseInt(s)

	if err != nil {
		return time.Duration(0), err
	}

	return time.Duration(t) * time.Millisecond, nil
}

func parseInt(s string) (int, error) {
	str, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return 0, err
	}

	return int(str), nil
}
