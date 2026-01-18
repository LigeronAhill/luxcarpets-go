// Package logger предоставляет функционал для инициализации и настройки структурированного логгера.
// Пакет использует стандартную библиотеку slog с кастомным обработчиком для цветного вывода в консоль.
//
// Особенности:
//   - Цветной вывод для разных уровней логирования
//   - Поддержка уровней: DEBUG, INFO, WARN, ERROR
//   - Структурированные логи с временными метками
//   - Использует prettylogger для красивого форматирования
//   - Гибкая настройка уровня логирования через параметры
//   - Предопределенные константы уровней логирования для удобства
//
// Типичное использование в приложении:
//
//	func main() {
//	    env := "local"
//	    level := logger.DEBUG
//	    if prod := os.Getenv("ENVIRONMENT"); prod != "" {
//	        if strings.ToLower(prod) == "production" {
//	            env = "production"
//	            level = logger.INFO
//	        }
//	    }
//	    logger.Init(level)
//	    // ... дальнейшая инициализация приложения
//	}
package logger

import (
	"log/slog"
	"os"

	prettylogger "github.com/jacute/prettylogger"
)

// Уровни логирования для удобного использования.
// Используйте эти константы вместо прямого использования slog.Level*.
const (
	DEBUG = slog.LevelDebug // Уровень отладки - все сообщения, включая отладочные
	INFO  = slog.LevelInfo  // Уровень информации - INFO, WARN, ERROR (без DEBUG)
	WARN  = slog.LevelWarn  // Уровень предупреждений - WARN и ERROR (без DEBUG, INFO)
	ERROR = slog.LevelError // Уровень ошибок - только ERROR сообщения
)

// Init инициализирует и настраивает глобальный логгер приложения с указанным уровнем логирования.
//
// Функция создает логгер с цветным выводом в стандартный вывод (stdout)
// и устанавливает его в качестве логгера по умолчанию для всего приложения.
// Рекомендуется вызывать эту функцию в самом начале main() перед любой другой логикой.
//
// Параметры:
//   - level: уровень логирования. Рекомендуется использовать предопределенные константы:
//     logger.DEBUG, logger.INFO, logger.WARN, logger.ERROR
//
// Рекомендуемые уровни логирования по окружениям:
//   - Разработка (local/development): logger.DEBUG - все сообщения
//   - Продакшен (production): logger.INFO - только информационные сообщения и выше
//   - Тестирование: logger.WARN - только предупреждения и ошибки
//   - Критичные окружения: logger.ERROR - только ошибки
//
// После инициализации автоматически выводятся тестовые сообщения
// всех уровней для демонстрации работы логгера и сообщение о успешной инициализации.
//
// Примеры использования с константами пакета:
//
//	// Для разработки - все сообщения
//	logger.Init(logger.DEBUG)
//
//	// Для продакшена - только INFO и выше
//	logger.Init(logger.INFO)
//
//	// Только предупреждения и ошибки
//	logger.Init(logger.WARN)
//
//	// Только критические ошибки
//	logger.Init(logger.ERROR)
//
// Примечание:
// Для работы пакета требуется установленная зависимость:
// go get github.com/jacute/prettylogger
func Init(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(prettylogger.NewColoredHandler(os.Stdout, opts))
	slog.SetDefault(logger)
	slog.Info("Будут выведены сообщения", slog.String("Уровень", slog.LevelDebug.String()))
	slog.Info("Будут выведены сообщения", slog.String("Уровень", slog.LevelInfo.String()))
	slog.Info("Будут выведены сообщения", slog.String("Уровень", slog.LevelWarn.String()))
	slog.Info("Будут выведены сообщения", slog.String("Уровень", slog.LevelError.String()))
	slog.Info("Журнал создан", "Уровень", level.String())
	return logger
}
