package main

import (
	"context"
	"log/slog"

	"github.com/LigeronAhill/luxcarpets-go/internal/database"
	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	"github.com/LigeronAhill/luxcarpets-go/pkg/config"
	"github.com/LigeronAhill/luxcarpets-go/pkg/logger"
)

func run(ctx context.Context) error {
	cfg, err := config.New(".settings.yml", nil).Unwrap()
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		return err
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
	pool := database.NewPool(ctx, cfg.GetString("database.url"))
	defer pool.Close()
	usersSorage := database.NewUsersStorage(pool)
	usersSorage.List(ctx, types.ListUsersParams{
		Limit:  1,
		Offset: 0,
	})
	return nil
}
