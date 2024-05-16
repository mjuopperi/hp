package db

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mjuopperi/hp/backend/internal/utils"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn string) error {
	slog.Info("Running database migrations")
	migrationDir := filepath.Join(utils.RootPath, "internal/db/migrations")
	m, err := migrate.New("file://"+migrationDir, dsn)
	if err != nil {
		slog.Error("Failed to load migrations", "err", err, "migrationDir", migrationDir)
		os.Exit(1)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("Failed to run migrations", "err", err)
		os.Exit(1)
	}

	slog.Info("Done with database migrations")
	return nil
}
