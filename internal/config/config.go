package config

import (
	"fmt"
	"strings"

	"github.com/ArtemZ007/wb-l0/internal/config"
	"github.com/spf13/viper"
)

// AppConfig структура для хранения конфигурации приложения.
type AppConfig struct {
	Database DatabaseConfig
	NATS     NATSConfig
}

// DatabaseConfig структура для хранения конфигурации базы данных.
type DatabaseConfig struct {
	Host             string
	Port             int
	User             string
	Password         string
	DBName           string
	ConnectionString string
}

// NATSConfig структура для хранения конфигурации NATS.
type NATSConfig struct {
	URL         string
	ClusterID   string
	ClientID    string
	QueueGroup  string
	ChannelName string
}

// LoadConfig загружает конфигурацию из файла или переменных окружения.
func LoadConfig(configPaths ...string) (*AppConfig, error) {
	var config AppConfig

	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("nats.url", "nats://localhost:4222")

	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Генерация строки подключения к базе данных на основе отдельных параметров, если она не предоставлена.
	if config.Database.ConnectionString == "" {
		config.Database.ConnectionString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName)
	}

	return &config, nil
}
func setupConfig() (*config.AppConfig, error) {
	// Example configuration structure
	cfg := &config.AppConfig{
		Server: config.ServerConfig{
			Port: ":8080", // Example default port
		},
		// Add other configuration fields here
	}

	// Here you would typically read configuration from environment variables,
	// a configuration file, or command line arguments, and update cfg accordingly.

	// For now, let's just return the example configuration.
	return cfg, nil
}
