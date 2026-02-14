package logger

import (
	"log/slog"
	"os"

	prettylogger "github.com/jacute/prettylogger"
)

const (
	DEBUG = slog.LevelDebug // Уровень отладки - все сообщения, включая отладочные
	INFO  = slog.LevelInfo  // Уровень информации - INFO, WARN, ERROR (без DEBUG)
	WARN  = slog.LevelWarn  // Уровень предупреждений - WARN и ERROR (без DEBUG, INFO)
	ERROR = slog.LevelError // Уровень ошибок - только ERROR сообщения
)

func Init(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(prettylogger.NewColoredHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	slog.Info("Logger initialized", "level", level.String())
	return logger
}
