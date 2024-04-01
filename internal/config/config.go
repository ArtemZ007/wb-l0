package config

import (
	"github.com/spf13/viper"
)

// AppConfig структура для хранения конфигурации приложения
type AppConfig struct {
	Database DatabaseConfig
	NATS     NATSConfig
}

// DatabaseConfig структура для хранения конфигурации базы данных
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NATSConfig структура для хранения конфигурации NATS
type NATSConfig struct {
	URL string
}

// LoadConfig загружает конфигурацию из файла или переменных окружения
func LoadConfig(configPaths ...string) (*AppConfig, error) {
	var config AppConfig

	viper.SetConfigName("app")  // Имя файла конфигурации (без расширения)
	viper.SetConfigType("yaml") // Тип файла конфигурации (yaml, json...)

	// Установка значений по умолчанию
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("nats.url", "nats://localhost:4222")

	// Добавление путей поиска файла конфигурации
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Чтение конфигурации из файла или переменных окружения
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Привязка переменных окружения
	viper.AutomaticEnv()

	// Сопоставление с структурой AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
