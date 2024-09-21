package config

import (
	"log"
	"os"
	"strconv"
)

// IConfiguration определяет интерфейс для конфигурационных настроек.
type IConfiguration interface {
	GetDBConnectionString() string
	GetRedisAddr() string
	GetRedisPassword() string
	GetRedisDB() int
	GetServerPort() int
	GetLogLevel() string
	GetNATSURL() string
	GetNATSClusterID() string
	GetNATSClientID() string
}

// Configuration содержит конфигурационные настройки.
type Configuration struct {
	DBConnectionString string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	ServerPort         int
	LogLevel           string
	NATSURL            string
	NATSClusterID      string
	NATSClientID       string
}

// NewConfiguration загружает конфигурационные настройки из переменных окружения.
func NewConfiguration() IConfiguration {
	redisDB, err := getEnvAsInt("REDIS_DB", 0)
	if err != nil {
		log.Fatalf("Ошибка преобразования REDIS_DB: %v", err)
	}

	serverPort, err := getEnvAsInt("SERVER_PORT", 8080)
	if err != nil {
		log.Fatalf("Ошибка преобразования SERVER_PORT: %v", err)
	}

	return &Configuration{
		DBConnectionString: getEnv("DB_CONNECTION_STRING", ""),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		ServerPort:         serverPort,
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		NATSURL:            getEnv("NATS_URL", "nats://localhost:4222"),
		NATSClusterID:      getEnv("NATS_CLUSTER_ID", "test-cluster"),
		NATSClientID:       getEnv("NATS_CLIENT_ID", "client-123"),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как целое число или возвращает значение по умолчанию.
func getEnvAsInt(key string, defaultValue int) (int, error) {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// GetDBConnectionString возвращает строку подключения к базе данных.
func (c *Configuration) GetDBConnectionString() string {
	return c.DBConnectionString
}

// GetRedisAddr возвращает адрес Redis.
func (c *Configuration) GetRedisAddr() string {
	return c.RedisAddr
}

// GetRedisPassword возвращает пароль Redis.
func (c *Configuration) GetRedisPassword() string {
	return c.RedisPassword
}

// GetRedisDB возвращает номер базы данных Redis.
func (c *Configuration) GetRedisDB() int {
	return c.RedisDB
}

// GetServerPort возвращает порт сервера.
func (c *Configuration) GetServerPort() int {
	return c.ServerPort
}

// GetLogLevel возвращает уровень логирования.
func (c *Configuration) GetLogLevel() string {
	return c.LogLevel
}

// GetNATSURL возвращает URL NATS.
func (c *Configuration) GetNATSURL() string {
	return c.NATSURL
}

// GetNATSClusterID возвращает идентификатор кластера NATS.
func (c *Configuration) GetNATSClusterID() string {
	return c.NATSClusterID
}

// GetNATSClientID возвращает идентификатор клиента NATS.
func (c *Configuration) GetNATSClientID() string {
	return c.NATSClientID
}
