package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	httpQS "github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/subscription"
	"github.com/ArtemZ007/wb-l0/pkg/config"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	_ "github.com/lib/pq"
)

func main() {
	// Парсинг аргументов командной строки.
	var cmd string
	flag.StringVar(&cmd, "cmd", "start", "команда для выполнения: start для запуска, stop для остановки")
	flag.Parse()

	// Загрузка конфигурации.
	cfg, err := config.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Ошибка при загрузке конфигурации: %v\n", err)
		os.Exit(1)
	}

	// Инициализация логгера.
	appLogger := logger.New(cfg.GetLogLevel())

	switch cmd {
	case "start":
		appLogger.Info("Запуск приложения")
		startApp(cfg, appLogger)
	case "stop":
		appLogger.Info("Остановка приложения через CLI не поддерживается")
	default:
		appLogger.Info("Неизвестная команда. Доступные команды: start, stop")
	}
}

func startApp(cfg config.IConfiguration, appLogger logger.ILogger) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		appLogger.Fatal("Ошибка подключения к базе данных: ", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Error("Ошибка при закрытии соединения с базой данных: ", err)
		}
	}()

	cacheService, err := cache.NewCacheService(appLogger)
	if err != nil {
		appLogger.Fatal("Ошибка инициализации сервиса кэширования: ", err)
	}

	// Assuming logger.New returns a *logger.Logger which implements logger.ILogger
	// You need to assert the type to *logger.Logger if httpQS.NewHandler expects it.
	actualLogger, ok := appLogger.(*logger.Logger)
	if !ok {
		appLogger.Fatal("appLogger is not of type *logger.Logger")
	}

	handler := httpQS.NewHandler(cacheService, actualLogger) // Adjusted to pass the asserted type
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetServerPort()),
		Handler: handler,
	}

	go func() {
		appLogger.Info("HTTP сервер запущен на порту ", cfg.GetServerPort())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Ошибка при запуске HTTP сервера: ", err)
		}
	}()

	natsListener, err := subscription.NewOrderListener(cfg.GetNATSURL(), cfg.GetNATSClusterID(), cfg.GetNATSClientID(), cacheService, actualLogger) // Adjusted to pass the asserted type
	if err != nil {
		appLogger.Fatal("Ошибка инициализации слушателя NATS: ", err)
	}
	go func() {
		if err := natsListener.Start(ctx); err != nil {
			appLogger.Error("Ошибка при запуске слушателя NATS: ", err)
		}
	}()
	waitForShutdownSignal(appLogger, func() {
		appLogger.Info("Остановка приложения")
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctxShutDown); err != nil {
			appLogger.Error("Ошибка при остановке HTTP сервера: ", err)
		}
	}, cancel)
}

func waitForShutdownSignal(log logger.ILogger, shutdownFunc func(), cancelFunc context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	log.Info("Получен сигнал для завершения работы: ", sig)

	shutdownFunc()
	cancelFunc()
	signal.Stop(signals)
}
