package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Инициализация логгера logrus
var log = logrus.New()

func init() {
	// Установка формата логирования
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	// Установка уровня логирования
	log.SetLevel(logrus.InfoLevel)
}

// CheckError проверяет наличие ошибки и логирует ее.
// Возвращает ошибку, чтобы вызывающий код мог ее обработать.
func CheckError(err error) error {
	if err != nil {
		log.WithError(err).Error("Произошла ошибка")
		return err
	}
	return nil
}

// GetEnv получает переменную окружения с возможностью установки значения по умолчанию.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// LogInfo используется для логирования информационных сообщений.
func LogInfo(message string) {
	log.Info(message)
}

// LogError используется для логирования сообщений об ошибках.
func LogError(message string) {
	log.Error(message)
}
