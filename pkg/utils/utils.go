package utils

import (
	"log"
	"os"
)

// CheckError проверяет наличие ошибки и логирует ее
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// GetEnv получает переменную окружения с возможностью установки значения по умолчанию
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// LogInfo используется для логирования информационных сообщений
func LogInfo(message string) {
	log.Println("[INFO] " + message)
}

// LogError используется для логирования сообщений об ошибках
func LogError(message string) {
	log.Println("[ERROR] " + message)
}
