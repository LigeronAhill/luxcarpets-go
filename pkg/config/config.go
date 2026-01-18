// Package config предоставляет функционал для загрузки и управления конфигурацией приложения.
// Поддерживает загрузку настроек из файлов (YAML, JSON, TOML и др.) и переменных окружения
// с автоматической подстановкой значений по умолчанию и валидацией обязательных полей.
//
// Приоритеты конфигурации (от высшего к низшему):
// 1. Переменные окружения (с преобразованием точек в подчеркивания)
// 2. Конфигурационный файл
// 3. Значения по умолчанию
// Примечание:
// Для работы пакета требуется установленная зависимость:
// go get github.com/spf13/viper
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// New создает новую конфигурацию с указанными значениями по умолчанию.
// filePath - путь к конфигурационному файлу (может быть пустым)
// defaults - значения по умолчанию для конфигурационных параметров
// Возвращает инициализированный объект Viper или ошибку при проблемах чтения файла
func New(filePath string, defaults map[string]any) (*viper.Viper, error) {
	return Init(filePath, defaults, nil)
}

// NewWithValidation создает новую конфигурацию с проверкой обязательных полей.
// filePath - путь к конфигурационному файлу (может быть пустым)
// required - список обязательных конфигурационных параметров
// Возвращает инициализированный объект Viper или ошибку при отсутствии обязательных полей
// или проблемах чтения файла
func NewWithValidation(filePath string, required []string) (*viper.Viper, error) {
	return Init(filePath, nil, required)
}

// Init инициализирует конфигурацию с полным набором параметров.
// filePath - путь к конфигурационному файлу (пустая строка если файл не используется)
// defaultValues - значения по умолчанию для конфигурационных параметров
// requiredKeys - список обязательных параметров, которые должны быть установлены
// Возвращает инициализированный объект Viper или ошибку при валидации или чтении файла
//
// Пример использования:
//
//	config, err := config.Init(
//	    "config.yaml",
//	    map[string]any{"server.port": 8080},
//	    []string{"database.url", "api.key"}
//	)
func Init(filePath string, defaultValues map[string]any, requiredKeys []string) (*viper.Viper, error) {
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
			return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %w", err)
		}
	}

	// Проверка обязательных параметров
	if len(requiredKeys) > 0 {
		if err := validateRequired(config, requiredKeys); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// validateRequired проверяет наличие всех обязательных конфигурационных параметров.
// config - объект конфигурации Viper для проверки
// requiredKeys - список ключей, которые должны быть установлены
// Возвращает ошибку если какие-либо обязательные параметры отсутствуют
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
