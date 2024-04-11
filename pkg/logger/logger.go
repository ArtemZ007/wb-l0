package logger

// LoggerInterface определяет интерфейс для логгера.
type LoggerInterface interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// Log - глобальный экземпляр логгера, доступный для использования в других пакетах.
var Log LoggerInterface

func SetGlobalLogger(logger LoggerInterface) {
	Log = logger
}
