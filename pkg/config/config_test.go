// Пакет config_test содержит тесты для пакета config
package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/LigeronAhill/luxcarpets-go/pkg/config"
	"github.com/LigeronAhill/luxcarpets-go/pkg/result"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_Success тестирует успешное создание конфигурации
func TestNew_Success(t *testing.T) {
	t.Parallel()

	// Создаем временный файл конфигурации
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	err := os.WriteFile(configFile, []byte(`
server:
  port: 9090
  host: "test-host"
database:
  url: "postgres://test:5432/db"
`), 0644)
	require.NoError(t, err)

	defaults := map[string]any{
		"server.port":   8080,
		"log.level":     "info",
		"database.pool": 10,
	}

	// Тестируем создание конфигурации
	res := config.New(configFile, defaults)

	// Проверяем что результат успешен
	assert.True(t, res.IsOk())

	// Извлекаем конфигурацию
	cfg, err := res.Unwrap()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Проверяем приоритеты: значение из файла должно переопределить значение по умолчанию
	assert.Equal(t, 9090, cfg.GetInt("server.port"))

	// Проверяем значение из файла
	assert.Equal(t, "test-host", cfg.GetString("server.host"))

	// Проверяем значение по умолчанию (которого нет в файле)
	assert.Equal(t, "info", cfg.GetString("log.level"))

	// Проверяем значение из файла (которого нет в defaults)
	assert.Equal(t, "postgres://test:5432/db", cfg.GetString("database.url"))
}

// TestNew_FileNotFound тестирует обработку отсутствующего файла
func TestNew_FileNotFound(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"server.port": 8080,
	}

	// Пытаемся загрузить несуществующий файл
	res := config.New("/non/existent/file.yaml", defaults)

	// Проверяем что результат содержит ошибку
	assert.True(t, res.IsErr())

	_, err := res.Unwrap()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка чтения конфигурационного файла")
}

// TestNew_EmptyFilePath тестирует создание конфигурации без файла
func TestNew_EmptyFilePath(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"server.port": 8080,
		"log.level":   "debug",
	}

	// Создаем конфигурацию без файла
	res := config.New("", defaults)

	// Проверяем что результат успешен
	assert.True(t, res.IsOk())

	cfg, err := res.Unwrap()
	require.NoError(t, err)

	// Проверяем значения по умолчанию
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
	assert.Equal(t, "debug", cfg.GetString("log.level"))
}

// TestNewWithValidation_Success тестирует успешную валидацию
func TestNewWithValidation_Success(t *testing.T) {
	dbURL := os.Getenv("LUXCARPETS_DATABASE_URL")
	os.Unsetenv("LUXCARPETS_DATABASE_URL")
	defer func() {
		os.Setenv("LUXCARPETS_DATABASE_URL", dbURL)
	}()
	t.Parallel()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "validation_config.yaml")

	err := os.WriteFile(configFile, []byte(`
database:
  url: "postgres://localhost:5432/test"
api:
  key: "secret-key"
server:
  port: 8080
`), 0644)
	require.NoError(t, err)

	required := []string{
		"database.url",
		"api.key",
		"server.port",
	}

	res := config.NewWithValidation(configFile, required)

	// Проверяем что результат успешен (все обязательные поля присутствуют)
	assert.True(t, res.IsOk())

	cfg, err := res.Unwrap()
	require.NoError(t, err)

	// Проверяем что все обязательные поля доступны
	assert.Equal(t, "postgres://localhost:5432/test", cfg.GetString("database.url"))
	assert.Equal(t, "secret-key", cfg.GetString("api.key"))
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
}

// TestNewWithValidation_MissingRequired тестирует ошибку валидации при отсутствии обязательных полей
func TestNewWithValidation_MissingRequired(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "missing_config.yaml")

	// Создаем файл без некоторых обязательных полей
	err := os.WriteFile(configFile, []byte(`
database:
  url: "postgres://localhost:5432/test"
# api.key отсутствует
server:
  port: 8080
`), 0644)
	require.NoError(t, err)

	required := []string{
		"database.url",
		"api.key", // Отсутствует в файле
		"server.port",
	}

	res := config.NewWithValidation(configFile, required)

	// Проверяем что результат содержит ошибку
	assert.True(t, res.IsErr())

	_, err = res.Unwrap()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "отсутствуют обязательные параметры конфигурации")
	assert.Contains(t, err.Error(), "api.key")
}

// TestInit_EnvironmentVariables тестирует приоритет переменных окружения
func TestInit_EnvironmentVariables(t *testing.T) {
	// Убираем t.Parallel() из-за использования t.Setenv
	// Устанавливаем переменные окружения ДО параллельного выполнения
	t.Setenv("LUXCARPETS_SERVER_PORT", "7070")
	t.Setenv("LUXCARPETS_DATABASE_URL", "env://database")

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "env_config.yaml")

	err := os.WriteFile(configFile, []byte(`
server:
  port: 9090
database:
  url: "file:///db.sqlite"
`), 0644)
	require.NoError(t, err)

	defaults := map[string]any{
		"server.port": 8080,
		"log.level":   "info",
	}

	res := config.Init(configFile, defaults, nil)

	require.True(t, res.IsOk())
	cfg, err := res.Unwrap()
	require.NoError(t, err)

	// Проверяем приоритеты: окружение > файл > defaults
	assert.Equal(t, 7070, cfg.GetInt("server.port"))                 // Из окружения
	assert.Equal(t, "env://database", cfg.GetString("database.url")) // Из окружения
	assert.Equal(t, "info", cfg.GetString("log.level"))              // Из defaults
}

// TestInit_DefaultValues тестирует установку значений по умолчанию
func TestInit_DefaultValues(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"server.port":   8080,
		"log.level":     "warn",
		"database.pool": 20,
	}

	res := config.Init("", defaults, nil)

	require.True(t, res.IsOk())
	cfg, err := res.Unwrap()
	require.NoError(t, err)

	// Проверяем значения по умолстанию
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
	assert.Equal(t, "warn", cfg.GetString("log.level"))
	assert.Equal(t, 20, cfg.GetInt("database.pool"))
}

// TestInit_ValidationWithDefaults тестирует комбинацию валидации и значений по умолчанию
func TestInit_ValidationWithDefaults(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"server.port":   8080,
		"database.pool": 10,
	}

	required := []string{
		"server.port",   // Есть в defaults
		"database.pool", // Есть в defaults
		"api.secret",    // Нет в defaults - должна вызвать ошибку
	}

	res := config.Init("", defaults, required)

	// Должна быть ошибка так как api.secret отсутствует
	assert.True(t, res.IsErr())

	_, err := res.Unwrap()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api.secret")
}

// TestConfig_Integration тестирует интеграцию с result.Result методами
func TestConfig_Integration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "integration_config.yaml")

	err := os.WriteFile(configFile, []byte(`
app:
  name: "TestApp"
  version: "1.0.0"
`), 0644)
	require.NoError(t, err)

	defaults := map[string]any{
		"server.port": 8080,
	}

	// Тестируем получение значения из конфигурации
	res := config.New(configFile, defaults)

	// Используем Match для обработки результата
	appName := result.Match(res,
		func(cfg *viper.Viper) string {
			return cfg.GetString("app.name")
		},
		func(err error) string {
			return "DefaultApp"
		},
	)

	assert.Equal(t, "TestApp", appName)

	// Тестируем обработку ошибок
	invalidRes := config.New("/invalid/path.yaml", defaults)

	handledRes := invalidRes.MapErr(func(err error) error {
		return fmt.Errorf("контекст: %w", err)
	})

	assert.True(t, handledRes.IsErr())
	_, err = handledRes.Unwrap()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "контекст: ошибка чтения конфигурационного файла")
}

// TestConfig_AndThen тестирует цепочку вызовов с AndThen
func TestConfig_AndThen(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "andthen_config.yaml")

	err := os.WriteFile(configFile, []byte(`
port: "8080"
timeout: "30s"
`), 0644)
	require.NoError(t, err)

	defaults := map[string]any{
		"port":    "3000",
		"timeout": "10s",
	}

	// Получаем Result[*viper.Viper]
	configRes := config.New(configFile, defaults)

	// Строим цепочку преобразований используя функцию AndThen с двумя типами
	res := result.AndThen(configRes, func(cfg *viper.Viper) result.Result[int] {
		portStr := cfg.GetString("port")
		t.Logf("Получен порт из конфигурации: %s", portStr)

		// Преобразуем строку в число (здесь могла бы быть более сложная логика)
		// В реальном коде использовали бы strconv.Atoi с обработкой ошибки
		return result.Ok(8080)
	})

	assert.True(t, res.IsOk())
	port, err := res.Unwrap()
	require.NoError(t, err)
	assert.Equal(t, 8080, port)
}

// TestConfig_OrElse тестирует значения по умолчанию для Result
func TestConfig_OrElse(t *testing.T) {
	t.Parallel()

	// Случай с ошибкой
	invalidRes := config.New("/invalid/path.yaml", nil)
	defaultCfg := viper.New()
	defaultCfg.Set("default", "value")

	resultCfg := invalidRes.OrElse(defaultCfg)
	assert.Equal(t, "value", resultCfg.GetString("default"))

	// Успешный случай
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "orelse_config.yaml")
	err := os.WriteFile(configFile, []byte(`key: "actual"`), 0644)
	require.NoError(t, err)

	validRes := config.New(configFile, nil)
	actualCfg := validRes.OrElse(defaultCfg)
	assert.Equal(t, "actual", actualCfg.GetString("key"))
}

// TestConfig_EdgeCases тестирует граничные случаи
func TestConfig_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("пустой файл конфигурации", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "empty_config.yaml")

		err := os.WriteFile(configFile, []byte(``), 0644)
		require.NoError(t, err)

		res := config.New(configFile, map[string]any{"key": "value"})
		assert.True(t, res.IsOk())
	})

	t.Run("конфигурация с спецсимволами в ключах", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "special_config.yaml")

		err := os.WriteFile(configFile, []byte(`
"complex-key.with-special_chars":
  value: "test"
`), 0644)
		require.NoError(t, err)

		res := config.New(configFile, nil)
		assert.True(t, res.IsOk())
	})

	t.Run("nil defaults", func(t *testing.T) {
		res := config.New("", nil)
		assert.True(t, res.IsOk())
	})

	t.Run("пустой required список", func(t *testing.T) {
		res := config.NewWithValidation("", []string{})
		assert.True(t, res.IsOk())
	})
}

// TestValidateRequired_Integration тестирует валидацию через публичный API
func TestValidateRequired_Integration(t *testing.T) {
	t.Parallel()

	cfg := viper.New()
	cfg.Set("existing.key", "value")
	cfg.Set("another.key", 123)

	t.Run("успешная валидация через Init", func(t *testing.T) {
		res := config.Init("", map[string]any{
			"existing.key": "default",
			"another.key":  456,
		}, []string{"existing.key", "another.key"})

		assert.True(t, res.IsOk())
	})

	t.Run("неуспешная валидация через Init", func(t *testing.T) {
		res := config.Init("", nil, []string{"missing.key"})

		assert.True(t, res.IsErr())
		_, err := res.Unwrap()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing.key")
	})
}
