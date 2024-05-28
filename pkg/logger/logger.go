package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// ILogger определяет интерфейс для логирования.
type ILogger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debug(args ...interface{})
	WithField(key string, value interface{}) *Logger
	WithFields(fields map[string]interface{}) *Logger
	WithError(err error) *logrus.Entry
}

// Logger реализует ILogger и обеспечивает логирование.
type Logger struct {
	entry *logrus.Entry // Используем *logrus.Entry для поддержки цепочечных вызовов
}

// New создает и возвращает новый экземпляр Logger, настроенный с уровнем логирования.
func New(logLevel string) *Logger {
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

	return &Logger{entry: logrus.NewEntry(logger)} // Инициализируем Logger с logrus.Entry
}

// Info логирует сообщение на уровне Info.
func (l *Logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn логирует сообщение на уровне Warn.
func (l *Logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error логирует сообщение на уровне Error.
func (l *Logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal логирует сообщение на уровне Fatal и завершает выполнение программы.
func (l *Logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Debug логирует сообщение на уровне Debug.
func (l *Logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// WithField добавляет одно поле к записи лога и возвращает Logger для цепочечного вызова.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{entry: l.entry.WithField(key, value)}
}

// WithFields добавляет множество полей к записи лога и возвращает Logger для цепочечного вызова.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{entry: l.entry.WithFields(fields)}
}

// WithError добавляет ошибку к записи лога.
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.entry.WithError(err)
}
