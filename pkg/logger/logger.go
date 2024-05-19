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
	entry *logrus.Entry // Changed from *logrus.Logger to *logrus.Entry
}

// WithError adds an error to the log entry.
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.entry.WithError(err)
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

	return &Logger{entry: logrus.NewEntry(logger)} // Initialize Logger with a logrus.Entry
}

// Info logs a message at level Info on the standard logger.
func (l *Logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l *Logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func (l *Logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l *Logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Debug logs a message at level Debug on the standard logger.
func (l *Logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// WithField добавляет одно поле к записи лога и возвращает ILogger для цепочечного вызова.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{entry: l.entry.WithField(key, value)} // Correctly return a Logger with an updated *logrus.Entry
}

// WithFields добавляет множество полей к записи лога и возвращает ILogger для цепочечного вызова.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{entry: l.entry.WithFields(fields)} // Correctly return a Logger with an updated *logrus.Entry
}
