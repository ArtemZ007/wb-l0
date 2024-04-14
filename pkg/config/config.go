package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path/filepath"
)

// IConfiguration определяет интерфейс для конфигурации приложения.
type IConfiguration interface {
	GetLogLevel() string
	GetDBConnectionString() string
	GetServerPort() int
	GetNATSURL() string
	GetNATSClusterID() string
	GetNATSClientID() string
}

// Config реализует IConfiguration и содержит конфигурационные параметры приложения.
type Config struct {
	LogLevel           string
	DBConnectionString string
	ServerPort         int
	NATSURL            string
	NATSClusterID      string
	NATSClientID       string
}

// LoadConfig загружает конфигурацию из файла .env или переменных окружения.
func LoadConfig() (IConfiguration, error) {
	if err := loadEnv(); err != nil {
		logrus.WithError(err).Warn("Не удалось загрузить конфигурацию из .env")
	}

	viper.AutomaticEnv() // Read from environment variables

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("NATS_URL", "nats://localhost:4222")
	viper.SetDefault("NATS_CLUSTER_ID", "test-cluster")
	viper.SetDefault("NATS_CLIENT_ID", "client-123")

	cfg := &Config{
		LogLevel:           viper.GetString("LOG_LEVEL"),
		DBConnectionString: viper.GetString("DB_CONNECTION_STRING"),
		ServerPort:         viper.GetInt("SERVER_PORT"),
		NATSURL:            viper.GetString("NATS_URL"),
		NATSClusterID:      viper.GetString("NATS_CLUSTER_ID"),
		NATSClientID:       viper.GetString("NATS_CLIENT_ID"),
	}

	return cfg, nil
}

// loadEnv загружает переменные окружения из файла .env.
func loadEnv() error {
	envPath, err := filepath.Abs(".env")
	if err != nil {
		return fmt.Errorf("ошибка определения пути к файлу .env: %w", err)
	}

	if err := godotenv.Load(envPath); err != nil {
		logrus.Warnf("файл .env не найден или не может быть загружен: %v", err)
	}

	return nil
}

// GetDBConnectionString implements IConfiguration.
func (c *Config) GetDBConnectionString() string {
	return c.DBConnectionString
}

// GetLogLevel implements IConfiguration.
func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

// GetNATSClientID implements IConfiguration.
func (c *Config) GetNATSClientID() string {
	return c.NATSClientID
}

// GetNATSClusterID implements IConfiguration.
func (c *Config) GetNATSClusterID() string {
	return c.NATSClusterID
}

// GetNATSURL implements IConfiguration.
func (c *Config) GetNATSURL() string {
	return c.NATSURL
}

// GetServerPort implements IConfiguration.
func (c *Config) GetServerPort() int {
	return c.ServerPort
}
