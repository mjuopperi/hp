package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB(dbURL string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		slog.Error("Unable to connect to database", "err", err)
		return err
	}

	err = RunMigrations(dbURL)
	if err != nil {
		slog.Error("Unable to run migrations", "err", err)
		return err
	}

	return nil
}

func ConnectionURIFromEnv() string {
	host := os.Getenv("PG_HOST")
	port, _ := strconv.Atoi(os.Getenv("PG_PORT"))
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	database := os.Getenv("POSTGRES_DB")
	return ConnectionURI(host, port, user, password, database)
}

func ConnectionURI(host string, port int, user string, password string, database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database)
}

func Close() {
	slog.Info("Closing database connection")
	Pool.Close()
}
