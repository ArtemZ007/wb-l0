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

	httpQS "github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/ArtemZ007/wb-l0/internal/subscription"
	"github.com/ArtemZ007/wb-l0/pkg/config"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	_ "github.com/lib/pq"
)

func main() {
	flagCmd := flag.String("cmd", "start", "команда для выполнения: start для запуска, stop для остановки")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при загрузке конфигурации: %v\n", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.GetLogLevel())
	if appLogger == nil {
		fmt.Fprintf(os.Stderr, "Не удалось инициализировать логгер\n")
		os.Exit(1)
	}

	if *flagCmd == "start" {
		appLogger.Info("Запуск приложения")
		startApp(cfg, appLogger)
	} else {
		appLogger.Info("Остановка приложения через CLI не поддерживается")
	}
}

func startApp(cfg config.IConfiguration, appLogger *logger.Logger) {
	// Initialize a connection to the database
	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		appLogger.Fatal("Ошибка при подключении к базе данных: ", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Error("Ошибка при закрытии соединения с базой данных: ", err)
		}
	}()

	// Check the database connection
	if err := db.Ping(); err != nil {
		appLogger.Fatal("Не удалось подключиться к базе данных: ", err)
	}

	// Create the database service with the established connection
	dbService, err := database.NewService(db, appLogger)
	if err != nil {
		appLogger.Fatal("Ошибка при создании сервиса базы данных: ", err)
	}

	// Log dbService to ensure it's not nil
	appLogger.Info("dbService инициализирован: ", dbService)

	// Create the cache service
	cacheService := cache.NewCacheService(appLogger)
	if cacheService == nil {
		appLogger.Fatal("Не удалось создать сервис кэша")
	}

	// Log cacheService to ensure it's not nil
	appLogger.Info("cacheService инициализирован: ", cacheService)

	// Set the database service for the cache service
	cacheService.SetDatabaseService(dbService)

	// Now safe to call InitCacheWithDBOrders
	ctx := context.Background()
	if err := cacheService.InitCacheWithDBOrders(ctx); err != nil {
		appLogger.Fatal("Ошибка при инициализации кэша заказами из базы данных: ", err)
	}

	// Continue with application setup...
	handler := httpQS.NewHandler(cache.Cache, appLogger)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetServerPort()),
		Handler: handler,
	}

	go func() {
		appLogger.Info("HTTP сервер запущен на порту ", cfg.GetServerPort())
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Ошибка при запуске HTTP сервера: ", err)
		}
	}()

	natsListener, err := subscription.NewOrderListener(cfg.GetNATSURL(), cfg.GetNATSClusterID(), cfg.GetNATSClientID(), cacheService, dbService, appLogger)
	if err != nil {
		appLogger.Fatal("Ошибка инициализации слушателя NATS: ", err)
	}

	go func() {
		if err := natsListener.Start(ctx); err != nil {
			appLogger.Fatal("Ошибка при запуске слушателя NATS: ", err)
		}
	}()

	<-waitForShutdownSignal(appLogger)

	if err := server.Shutdown(context.Background()); err != nil {
		appLogger.Error("Ошибка при остановке HTTP сервера: ", err)
	}
}

func waitForShutdownSignal(appLogger *logger.Logger) <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		appLogger.Info("Получен сигнал для завершения работы: ", sig)
	}()

	return signals
}
