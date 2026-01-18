package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/LigeronAhill/luxcarpets-go/internal/database"
	"github.com/LigeronAhill/luxcarpets-go/pkg/config"
	"github.com/LigeronAhill/luxcarpets-go/pkg/logger"
)

func main() {
	ctx := context.Background()
	cfg, err := config.New(".settings.yml", nil).Unwrap()
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	level := logger.INFO
	environment := cfg.GetString("environment")
	slog.Warn("from .settings.yml", slog.String("environment", environment))
	if environment != "production" {
		level = logger.DEBUG
	}
	customLogger := logger.Init(level)
	slog.Info("Starting server", slog.String("level", level.String()))
	customLogger.Debug("DEBUG")
	err = database.Migrate(ctx, cfg.GetString("database.url"))
	if err != nil {
		slog.Error("Failed to migrate database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	pool := database.NewPool(ctx, cfg.GetString("database.url"))
	defer pool.Close()
}
