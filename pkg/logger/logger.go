// Package logger Пакет logger предоставляет структурированный интерфейс логирования, построенный на основе logrus.
// Это позволяет легко логировать сообщения различной важности и с контекстной информацией.
// Интерфейс Logger абстрагирует базовую библиотеку логирования (в данном случае logrus),
// что позволяет легко заменять или модифицировать бэкенд логирования без влияния на остальную часть кодовой базы.
package logger

import (
	"github.com/sirupsen/logrus"
)

// Config представляет конфигурацию для логгера.
type Config struct {
	Level string // Уровень логирования (например, "info", "debug")
	// Дополнительные поля конфигурации могут быть добавлены здесь
}

// Logger реализует интерфейс ILogger и предоставляет методы для структурированного логирования.
type Logger struct {
	logger *logrus.Logger
}

// New создает и возвращает новый экземпляр Logger, настроенный согласно переданной конфигурации.
func New(config Config) *Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		logger.Warn("Уровень логирования не распознан. Используется уровень 'info'.")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Настройка формата логирования и других параметров может быть добавлена здесь

	return &Logger{logger: logger}
}

// Info логирует сообщение с уровнем Info и дополнительными полями.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Info(msg)
}

// Warn логирует сообщение с уровнем Warn и дополнительными полями.
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Warn(msg)
}

// Error логирует сообщение с уровнем Error и дополнительными полями.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Error(msg)
}

// Fatal логирует сообщение с уровнем Fatal и дополнительными полями, после чего вызывает os.Exit(1).
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Fatal(msg)
}

// Debug логирует сообщение с уровнем Debug и дополнительными полями.
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Debug(msg)
}

// GetUnderlyingLogger возвращает базовый логгер (*logrus.Logger) для прямого доступа, если это необходимо.
func (l *Logger) GetUnderlyingLogger() *logrus.Logger {
	return l.logger
}

func (l *Logger) GetLogrusLogger() interface{} {
	return l.logger
}
