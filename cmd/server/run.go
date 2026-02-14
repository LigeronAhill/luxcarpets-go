package main

import (
	"context"
	"log/slog"

	"github.com/LigeronAhill/luxcarpets-go/internal/database"
	"github.com/LigeronAhill/luxcarpets-go/internal/service"
	"github.com/LigeronAhill/luxcarpets-go/pkg/config"
	"github.com/LigeronAhill/luxcarpets-go/pkg/logger"
)

func run(ctx context.Context) error {
	cfg, err := config.Init("")
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		return err
	}
	level := logger.INFO
	environment := cfg.Environment
	if environment != "production" {
		level = logger.DEBUG
	}
	customLogger := logger.Init(level)
	slog.Info("Starting server", slog.String("level", level.String()))
	customLogger.Debug("DEBUG")
	pool := database.NewPool(ctx, cfg.DatabaseSettings.URL)
	defer pool.Close()
	usersSorage := database.NewUsersStorage(pool)
	service.NewUsersService(usersSorage)
	return nil
}
