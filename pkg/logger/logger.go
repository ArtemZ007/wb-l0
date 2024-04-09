package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// log представляет собой экземпляр логгера.
var log = logrus.New()

func init() {
	// Установка формата логирования.
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Установка уровня логирования.
	log.SetLevel(logrus.InfoLevel)

	// Установка вывода логов.
	log.SetOutput(os.Stdout)
}

// LogInfo используется для логирования информационных сообщений.
func LogInfo(message string) {
	log.Info(message)
}

// LogError используется для логирования сообщений об ошибках.
func LogError(message string) {
	log.Error(message)
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
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}
