package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/mjuopperi/hp/backend/internal/db"
	"github.com/mjuopperi/hp/backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func validateEnvVars() {
	requiredVars := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"PG_HOST",
		"PG_PORT",
	}
	var missingVars []string

	for _, envVar := range requiredVars {
		if value := os.Getenv(envVar); value == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		log.Fatalf("Fatal error: Missing required environment variables: %v", missingVars)
	}
}

func main() {
	_ = godotenv.Load()
	validateEnvVars()

	if err := db.InitDB(db.ConnectionURIFromEnv()); err != nil {
		slog.Error("Failed to initialize database", "err", err)
		return
	}
	defer db.Close()

	r := gin.Default()
	handlers.RegisterRoutes(r)

	r.Run()
}
