package migrations

import (
	"embed"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed *.sql
var embedMigrations embed.FS

var gooseMu sync.Mutex

func SetupPostgres(pool *pgxpool.Pool, logger *zap.Logger) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("cannot set dialect in goose: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("cannot apply migrations: %w", err)
	}

	logger.Info("Postgres migrations applied successfully")
	return nil
}
