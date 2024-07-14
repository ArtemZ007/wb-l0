package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpQS "github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	_ "github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/ArtemZ007/wb-l0/internal/subscription"
	"github.com/ArtemZ007/wb-l0/pkg/config"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Загружаем конфигурацию
	cfg := loadConfig()

	// Инициализируем логгер
	log := logger.New(cfg.GetLogLevel())

	// Запускаем приложение
	if err := runApp(cfg, log); err != nil {
		log.Fatal("Ошибка запуска приложения: ", err)
	}
}

// runApp запускает основное приложение
func runApp(cfg config.IConfiguration, log logger.Logger) error {
	// Подключаемся к базе данных
	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Error("Ошибка подключения к базе данных: ", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Ошибка при закрытии соединения с базой данных: ", err)
		}
	}()

	// Проверяем соединение с базой данных
	if err := db.Ping(); err != nil {
		log.Error("Не удалось подключиться к базе данных: ", err)
		return err
	}
	log.Info("Успешное подключение к базе данных")
	log.Info("Запуск приложения")

	// Инициализируем сервис базы данных
	dbService, err := database.NewService(db, logrus.New())
	if err != nil {
		log.Error("Ошибка создания сервиса базы данных: ", err)
		return err
	}
	log.Info("Сервис базы данных инициализирован")

	// Инициализируем сервис кэша
	cacheService := cache.NewService(cfg.GetRedisAddr(), cfg.GetRedisPassword(), cfg.GetRedisDB(), log)
	if cacheService == nil {
		log.Error("Не удалось создать сервис кэша")
		return errors.New("не удалось создать сервис кэша")
	}
	log.Info("Сервис кэша инициализирован")

	// Устанавливаем сервис базы данных для кэша
	cacheService.SetDBService(dbService)

	// Инициализируем кэш данными из базы данных
	ctx := context.Background()
	if err := cacheService.InitCacheWithDBOrders(ctx); err != nil {
		log.Error("Ошибка инициализации кэша данными из базы данных: ", err)
		return err
	}

	// Инициализируем HTTP обработчик
	logger := logrus.New() // Убедитесь, что здесь используется ваша инициализация логгера
	handler := httpQS.NewHandler(cacheService, logger)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetServerPort()),
		Handler: handler,
	}

	// Запускаем HTTP сервер в отдельной горутине
	go func() {
		log.Info("HTTP сервер запущен на порту ", cfg.GetServerPort())
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("Ошибка HTTP сервера: ", err)
		}
	}()

	// Инициализируем NATS слушатель
	natsListener, err := subscription.NewListener(cfg.GetNATSURL(), cfg.GetNATSClusterID(), cfg.GetNATSClientID(), cacheService, dbService, log)
	if err != nil {
		log.Error("Ошибка инициализации NATS слушателя: ", err)
		return err
	}

	// Запускаем NATS слушатель в отдельной горутине
	go func() {
		if err := natsListener.Start(ctx); err != nil {
			log.Error("Ошибка NATS слушателя: ", err)
		}
	}()

	// Ожидаем сигнал завершения работы
	<-waitForShutdownSignal(log)

	// Завершаем работу HTTP сервера
	if err := server.Shutdown(context.Background()); err != nil {
		log.Error("Ошибка завершения работы HTTP сервера: ", err)
	}

	return nil
}

// waitForShutdownSignal ожидает сигнал завершения работы и возвращает канал, который закрывается при получении сигнала
func waitForShutdownSignal(log logger.Logger) <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		log.Info("Получен сигнал завершения работы")
		close(stop)
	}()
	return stop
}

// loadConfig загружает конфигурацию приложения
func loadConfig() config.IConfiguration {
	return config.NewConfiguration()
}
