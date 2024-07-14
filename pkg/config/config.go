package config

import (
	"os"
	"strconv"
)

// IConfiguration defines the interface for configuration settings.
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

// Configuration holds the configuration settings.
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

// NewConfiguration loads the configuration settings from environment variables.
func NewConfiguration() IConfiguration {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	serverPort, _ := strconv.Atoi(os.Getenv("SERVER_PORT"))

	return &Configuration{
		DBConnectionString: os.Getenv("DB_CONNECTION_STRING"),
		RedisAddr:          os.Getenv("REDIS_ADDR"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:            redisDB,
		ServerPort:         serverPort,
		LogLevel:           os.Getenv("LOG_LEVEL"),
		NATSURL:            os.Getenv("NATS_URL"),
		NATSClusterID:      os.Getenv("NATS_CLUSTER_ID"),
		NATSClientID:       os.Getenv("NATS_CLIENT_ID"),
	}
}

func (c *Configuration) GetDBConnectionString() string {
	return c.DBConnectionString
}

func (c *Configuration) GetRedisAddr() string {
	return c.RedisAddr
}

func (c *Configuration) GetRedisPassword() string {
	return c.RedisPassword
}

func (c *Configuration) GetRedisDB() int {
	return c.RedisDB
}

func (c *Configuration) GetServerPort() int {
	return c.ServerPort
}

func (c *Configuration) GetLogLevel() string {
	return c.LogLevel
}

func (c *Configuration) GetNATSURL() string {
	return c.NATSURL
}

func (c *Configuration) GetNATSClusterID() string {
	return c.NATSClusterID
}

func (c *Configuration) GetNATSClientID() string {
	return c.NATSClientID
}
