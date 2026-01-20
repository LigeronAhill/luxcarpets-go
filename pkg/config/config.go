// Package config предоставляет функционал для загрузки и управления конфигурацией приложения.
// Поддерживает загрузку настроек из файлов (YAML, JSON, TOML и др.) и переменных окружения
// с автоматической подстановкой значений по умолчанию и валидацией обязательных полей.
// Пакет использует result.Result для обработки ошибок, что позволяет строить цепочки вызовов.
//
// Приоритеты конфигурации (от высшего к низшему):
// 1. Переменные окружения (с преобразованием точек в подчеркивания)
// 2. Конфигурационный файл
// 3. Значения по умолчанию
//
// Пример использования:
//
//	configRes := config.New("config.yaml", map[string]any{
//	    "server.port": 8080,
//	    "database.url": "localhost:5432",
//	})
//
//	configRes.Map(func(cfg *viper.Viper) {
//	    port := cfg.GetInt("server.port")
//	    // использование конфигурации
//	})
//
//	// Или с обработкой ошибки:
//	cfg, err := configRes.Unwrap()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Примечание:
// Для работы пакета требуется установленная зависимость:
// go get github.com/spf13/viper
package config

import (
	"fmt"
	"strings"

	"github.com/LigeronAhill/luxcarpets-go/pkg/result"
	"github.com/spf13/viper"
)

// New создает новую конфигурацию с указанными значениями по умолчанию.
// filePath - путь к конфигурационному файлу (может быть пустым)
// defaults - значения по умолчанию для конфигурационных параметров
//
// Возвращает result.Result[*viper.Viper], который содержит:
//   - *viper.Viper: инициализированный объект конфигурации при успехе
//   - error: ошибка при проблемах чтения файла
//
// Пример:
//
//	res := config.New("config.yaml", map[string]any{
//	    "server.host": "localhost",
//	    "server.port": 8080,
//	})
//
//	// Обработка результата
//	if cfg, err := res.Unwrap(); err != nil {
//	    // обработка ошибки
//	} else {
//	    port := cfg.GetInt("server.port")
//	}
func New(filePath string, defaults map[string]any) result.Result[*viper.Viper] {
	return Init(filePath, defaults, nil)
}

// NewWithValidation создает новую конфигурацию с проверкой обязательных полей.
// filePath - путь к конфигурационному файлу (может быть пустым)
// required - список обязательных конфигурационных параметров
//
// Возвращает result.Result[*viper.Viper], который содержит:
//   - *viper.Viper: инициализированный объект конфигурации при успехе
//   - error: ошибка при отсутствии обязательных полей или проблемах чтения файла
//
// Примечание: значения по умолчанию не устанавливаются, только проверяются обязательные поля.
//
// Пример:
//
//	res := config.NewWithValidation("config.yaml", []string{
//	    "database.url",
//	    "api.secret_key",
//	    "server.port",
//	})
//
//	// Использование Map для обработки успешного результата
//	res.Map(func(cfg *viper.Viper) {
//	    // конфигурация гарантированно содержит все обязательные поля
//	    dbURL := cfg.GetString("database.url")
//	})
func NewWithValidation(filePath string, required []string) result.Result[*viper.Viper] {
	return Init(filePath, nil, required)
}

// Init инициализирует конфигурацию с полным набором параметров.
// filePath - путь к конфигурационному файлу (пустая строка если файл не используется)
// defaultValues - значения по умолчанию для конфигурационных параметров
// requiredKeys - список обязательных параметров, которые должны быть установлены
//
// Возвращает result.Result[*viper.Viper], который содержит:
//   - *viper.Viper: инициализированный объект конфигурации при успехе
//   - error: ошибка при валидации или чтении файла
//
// Этот метод является основным и используется New и NewWithValidation.
// Приоритет источников конфигурации:
// 1. Переменные окружения (с префиксом LUXCARPETS_)
// 2. Конфигурационный файл
// 3. Значения по умолчанию
//
// Пример использования:
//
//	configRes := config.Init(
//	    "config.yaml",
//	    map[string]any{
//	        "server.port": 8080,
//	        "log.level": "info",
//	    },
//	    []string{"database.url", "api.key"}
//	)
//
//	// Комбинирование с другими результатами
//	dbRes := database.Connect(cfg)
//	appRes := result.Combine(configRes, dbRes).
//	    Map(func(pair result.Pair[*viper.Viper, *sql.DB]) {
//	        // pair.First - конфигурация
//	        // pair.Second - подключение к БД
//	        return NewApp(pair.First, pair.Second)
//	    })
func Init(filePath string, defaultValues map[string]any, requiredKeys []string) result.Result[*viper.Viper] {
	config := viper.New()

	// Настройка работы с переменными окружения
	config.SetEnvPrefix("luxcarpets")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	// Установка значений по умолчанию
	if len(defaultValues) > 0 {
		for key, value := range defaultValues {
			config.SetDefault(key, value)
		}
	}

	// Загрузка конфигурационного файла если указан
	if filePath != "" {
		config.SetConfigFile(filePath)

		if err := config.ReadInConfig(); err != nil {
			return result.Err[*viper.Viper](fmt.Errorf("ошибка чтения конфигурационного файла: %w", err))
		}
	}

	// Проверка обязательных параметров
	if len(requiredKeys) > 0 {
		if err := validateRequired(config, requiredKeys); err != nil {
			return result.Err[*viper.Viper](err)
		}
	}

	return result.Ok(config)
}

// validateRequired проверяет наличие всех обязательных конфигурационных параметров.
// config - объект конфигурации Viper для проверки
// requiredKeys - список ключей, которые должны быть установлены
// Возвращает ошибку если какие-либо обязательные параметры отсутствуют
//
// Эта функция используется внутри Init для валидации.
// Внешние потребители обычно не вызывают её напрямую.
func validateRequired(config *viper.Viper, requiredKeys []string) error {
	var missing []string
	for _, key := range requiredKeys {
		if !config.IsSet(key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("отсутствуют обязательные параметры конфигурации: %s", strings.Join(missing, "; "))
	}
	return nil
}
