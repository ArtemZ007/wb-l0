// Package config Пакет config предоставляет конфигурацию, необходимую для приложения.
// Включает функциональность для загрузки конфигурации из переменных окружения,
// файлов .env и JSON файлов конфигурации.
//
// Автор: ArtemZ007
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config представляет структуру конфигурации приложения.
type Config struct {
	DBConnectionString string // Строка подключения к базе данных
	LogLevel           string // Уровень логирования
	ServerPort         int    // Порт сервера
	NATSURL            string // URL для подключения к NATS
	NATSClusterID      string // ID кластера NATS
	NATSClientID       string // ID клиента NATS
}

// IConfiguration определяет интерфейс для конфигурации приложения.
type IConfiguration interface {
	GetDBConnectionString() string
	GetLogLevel() string
	GetServerPort() int
	GetNATSURL() string
	GetNATSClusterID() string
	GetNATSClientID() string
}

// LoadConfig загружает конфигурацию из файла .env, файла конфигурации или переменных окружения.
// Путь к файлу .env должен быть на два уровня выше текущей директории.
func LoadConfig() (IConfiguration, error) {
	// Путь к файлу .env на два уровня выше
	envPath, err := filepath.Abs("../../.env")
	if err != nil {
		logrus.WithError(err).Error("Ошибка определения пути к файлу .env")
		return nil, err
	}

	// Загрузка переменных из файла .env
	if err := godotenv.Load(envPath); err != nil {
		logrus.WithError(err).Warn("Не удалось загрузить конфигурацию из .env файла")
		// Возвращаем ошибку, если файл .env обязателен
		return nil, err
	}

	// Создание конфигурации из переменных окружения
	cfg := &Config{
		LogLevel:           os.Getenv("LOG_LEVEL"),
		DBConnectionString: os.Getenv("DB_CONNECTION_STRING"),
		ServerPort:         parseEnvAsInt("SERVER_PORT", 8080),
		NATSURL:            os.Getenv("NATS_URL"),
		NATSClusterID:      os.Getenv("NATS_CLUSTER_ID"),
		NATSClientID:       os.Getenv("NATS_CLIENT_ID"),
	}

	// Валидация загруженных конфигураций
	if err := validateConfig(cfg); err != nil {
		logrus.WithError(err).Error("Ошибка валидации конфигурации")
		return nil, err
	}

	return cfg, nil
}

// parseEnvAsInt помогает преобразовать переменную окружения в целое число с значением по умолчанию
func parseEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// validateConfig проверяет корректность конфигурационных параметров.
func validateConfig(cfg *Config) error {
	// Пример валидации: проверка наличия обязательных параметров
	if cfg.DBConnectionString == "" {
		return fmt.Errorf("отсутствует обязательный параметр: DB_CONNECTION_STRING")
	}
	// Добавьте дополнительные проверки по мере необходимости
	return nil
}

// validateConfig проверяет корректность конфигурационных параметров.

// GetDBConnectionString Реализация интерфейса IConfiguration
func (c *Config) GetDBConnectionString() string { return c.DBConnectionString }
func (c *Config) GetLogLevel() string           { return c.LogLevel }
func (c *Config) GetServerPort() int            { return c.ServerPort }
func (c *Config) GetNATSURL() string            { return c.NATSURL }
func (c *Config) GetNATSClusterID() string      { return c.NATSClusterID }
func (c *Config) GetNATSClientID() string       { return c.NATSClientID }
