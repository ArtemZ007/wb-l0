package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger определяет интерфейс для логирования.
type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debug(args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) *logrus.Entry
}

// LogrusAdapter адаптирует logrus.Logger к интерфейсу Logger.
type LogrusAdapter struct {
	logger *logrus.Logger
}

// New создает и возвращает новый экземпляр LogrusAdapter, настроенный с уровнем логирования.
func New(logLevel string) *LogrusAdapter {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logger.Out = os.Stdout

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Warn("Неверно указан уровень логирования, используется уровень по умолчанию: Info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return &LogrusAdapter{logger: logger}
}

// Info логирует сообщение на уровне Info.
func (l *LogrusAdapter) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Warn логирует сообщение на уровне Warn.
func (l *LogrusAdapter) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Error логирует сообщение на уровне Error.
func (l *LogrusAdapter) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Fatal логирует сообщение на уровне Fatal и завершает выполнение программы.
func (l *LogrusAdapter) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Debug логирует сообщение на уровне Debug.
func (l *LogrusAdapter) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// WithField добавляет одно поле к записи лога и возвращает Logger для цепочечного вызова.
func (l *LogrusAdapter) WithField(key string, value interface{}) Logger {
	return &LogrusAdapter{logger: l.logger.WithField(key, value).Logger}
}

// WithFields добавляет множество полей к записи лога и возвращает Logger для цепочечного вызова.
func (l *LogrusAdapter) WithFields(fields map[string]interface{}) Logger {
	return &LogrusAdapter{logger: l.logger.WithFields(fields).Logger}
}

// WithError добавляет ошибку к записи лога.
func (l *LogrusAdapter) WithError(err error) *logrus.Entry {
	return l.logger.WithError(err)
}
