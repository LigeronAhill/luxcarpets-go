package main

import (
	"log"
	"log/slog"

	"github.com/LigeronAhill/luxcarpets-go/pkg/config"
	"github.com/LigeronAhill/luxcarpets-go/pkg/logger"
)

func main() {
	cfg, err := config.New(".settings.yml", nil)
	if err != nil {
		log.Fatal(err)
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
}
