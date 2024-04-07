package config

import (
	"fmt"
	_ "os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// AppConfig содержит конфигурацию приложения.
type AppConfig struct {
	Server struct {
		Port string
	}
	Database struct {
		Host             string
		Port             int
		User             string
		Password         string
		DBName           string
		ConnectionString string
	}
	NATS struct {
		URL         string
		ChannelName string
	}
}

// LoadConfig загружает конфигурацию приложения.
func LoadConfig() (*AppConfig, error) {
	var config AppConfig

	// Загрузка переменных окружения из файла .env, который находится на уровень выше
	envPath, _ := filepath.Abs("../.env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("Предупреждение: Не удалось загрузить файл .env: %s\n", err)
	}

	// Указание директории, где искать файлы конфигурации
	viper.AddConfigPath("..") // Предполагается, что файлы конфигурации находятся в корневой директории проекта

	// Установка имени и типа файла конфигурации
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")

	// Установка значений по умолчанию
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "user")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "db")
	viper.SetDefault("nats.url", "nats://localhost:4222")
	viper.SetDefault("nats.channelName", "ORDERS")

	// Автоматическая загрузка переменных окружения
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Чтение конфигурации из файла
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла конфигурации: %w", err)
	}

	// Преобразование конфигурации в структуру
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("ошибка при преобразовании конфигурации: %w", err)
	}

	// Генерация строки подключения к базе данных
	config.Database.ConnectionString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName)

	return &config, nil
}
