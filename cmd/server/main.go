// Package main реализует точку входа для приложения и настраивает веб-сервер и соединение с NATS Streaming.
// Этот файл включает главную функцию, которая инициализирует компоненты приложения,
// такие как HTTP-сервер, соединение с базой данных и подписку на NATS Streaming.
//
// Автор: ArtemZ007
package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httpDelivery "github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/ArtemZ007/wb-l0/internal/subscription"
	"github.com/ArtemZ007/wb-l0/pkg/config"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	_ "github.com/lib/pq"
)

func main() {
	// Загрузка конфигурации приложения из файла .env, файла конфигурации или переменных окружения.
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация логгера с уровнем логирования, указанным в конфигурации.
	appLogger := logger.New(logger.Config{Level: cfg.GetLogLevel()})
	appLogger.Info("Конфигурация успешно загружена", nil)

	// Подключение к базе данных с использованием строки подключения из конфигурации.
	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		appLogger.Fatal("Ошибка подключения к базе данных", map[string]interface{}{"error": err})
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			appLogger.Fatal("Ошибка закрытия подключения к базе данных", map[string]interface{}{"error": err})
			os.Exit(1)
			return
		}
	}(db)

	// Инициализация сервиса кэширования и загрузка данных из базы данных в кэш.
	cacheService := cache.New(appLogger) // Используйте функцию New из cache.go для создания экземпляра кэша

	// Предполагается, что у cacheService есть метод LoadOrdersFromDB, который загружает данные в кэш.
	// Если такого метода нет, вам нужно будет его реализовать в соответствующем месте.
	// Так как в вашем запросе указано, что реализацию кэша, хэндлера, подписчика и БД менять не нужно,
	// предполагается, что метод LoadOrdersFromDB уже реализован в вашем кэше.
	// Если это не так, вам потребуется добавить соответствующую логику в сервис кэширования.
	if err := cacheService.LoadOrdersFromDB(context.Background(), db); err != nil {
		appLogger.Fatal("Ошибка загрузки заказов из базы данных в кэш", map[string]interface{}{"error": err})
	}

	// Инициализация репозитория базы данных.
	// Инициализация и запуск слушателя NATS Streaming.
	// Assuming appLogger.GetLogrusLogger() returns an interface{} that actually is a *logrus.Logger
	logrusLogger, ok := appLogger.GetLogrusLogger().(*logrus.Logger)
	if !ok {
		appLogger.Fatal("Ошибка при получении *logrus.Logger из appLogger", map[string]interface{}{"error": "Type assertion failed"})
	}

	var dbService = database.NewService(db, logrusLogger)             // Now correctly using logrusLogger as *logrus.Logger          // Correctly using logrusLogger
	natsListener := subscription.NewListener(dbService, logrusLogger) // Correctly using logrusLogger
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensures cancel is called when exiting the function

	go func() {
		if err := natsListener.Start(ctx, cfg.GetNATSURL(), cfg.GetNATSClusterID(), cfg.GetNATSClientID(), "orders"); err != nil {
			appLogger.Fatal("Ошибка запуска слушателя NATS Streaming", map[string]interface{}{"error": err})
		}
	}()

	// Ожидание сигнала для остановки приложения.
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	// Остановка слушателя NATS Streaming перед остановкой HTTP сервера и закрытием соединения с базой данных.
	cancel() // Отправляем сигнал для остановки слушателя NATS Streaming через контекст

	// Инициализация и запуск HTTP сервера.
	httpHandler := httpDelivery.NewHandler(cacheService, appLogger)
	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.GetServerPort()),
		Handler:      httpHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		appLogger.Info("HTTP сервер запущен на порту "+strconv.Itoa(cfg.GetServerPort()), nil)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Fatal("Ошибка запуска HTTP сервера", map[string]interface{}{"error": err})
		}
	}()

	// Use the existing stopChan and signal.Notify setup from earlier in the code.
	<-stopChan

	// Остановка HTTP сервера.
	if err := httpServer.Shutdown(context.Background()); err != nil {
		appLogger.Error("Ошибка при остановке HTTP сервера", map[string]interface{}{"error": err})
	}

	appLogger.Info("Сервис успешно остановлен", nil)
}
