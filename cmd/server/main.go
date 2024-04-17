// Пакет main является точкой входа приложения. Он инициализирует и запускает все необходимые сервисы.
// Автор: ArtemZ007
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpQS "github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"

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

func startApp(cfg config.IConfiguration, appLogger *logger.Logger) {
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

	service, err := database.NewService(db, appLogger)
	if err != nil {
		appLogger.Fatal("Ошибка при создании сервиса: ", err)
	}

	// Since appLogger is already of type *logger.Logger, we can use it directly
	cacheService := cache.NewCacheService(appLogger, service)

	handler := httpQS.NewHandler(cacheService, appLogger)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetServerPort()),
		Handler: handler,
	}

	// The rest of your function remains unchanged...

	// Канал для обработки ошибок
	errChan := make(chan error, 1) // Buffered channel to prevent goroutine leak

	// Запуск HTTP сервера
	go func() {
		appLogger.Info("HTTP сервер запущен на порту ", cfg.GetServerPort())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("ошибка при запуске HTTP сервера: %w", err)
		}
	}()

	// Запуск слушателя NATS
	actualLogger := appLogger
	// Assuming service is an instance that implements IOrderService
	natsListener, err := subscription.NewOrderListener(cfg.GetNATSURL(), cfg.GetNATSClusterID(), cfg.GetNATSClientID(), cacheService, service, actualLogger)
	if err != nil {
		appLogger.Fatal("Ошибка инициализации слушателя NATS: ", err)
	}
	go func() {
		if err := natsListener.Start(ctx); err != nil {
			errChan <- fmt.Errorf("ошибка при запуске слушателя NATS: %w", err)
		}
	}()

	// Ожидание сигнала для завершения работы или ошибки от сервисов
	select {
	case <-waitForShutdownSignal(appLogger):
		appLogger.Info("Остановка приложения")
	case err := <-errChan:
		appLogger.Error("Ошибка во время выполнения: ", err)
	}

	// Остановка сервера
	ctxShutDown, cancelShutDown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutDown()
	if err := server.Shutdown(ctxShutDown); err != nil {
		appLogger.Error("Ошибка при остановке HTTP сервера: ", err)
	}
}

// waitForShutdownSignal needs to be adjusted to match the function signature and usage in startApp.
func waitForShutdownSignal(appLogger *logger.Logger) <-chan struct{} {
	stopChan := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		appLogger.Info("Получен сигнал для завершения работы: ", sig)
		close(stopChan)
		signal.Stop(signals)
	}()

	return stopChan
}
