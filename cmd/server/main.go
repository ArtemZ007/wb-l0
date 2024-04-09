package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ArtemZ007/wb-l0/internal/config"
	"github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"github.com/ArtemZ007/wb-l0/internal/delivery/nats"
	"github.com/ArtemZ007/wb-l0/internal/domain/service"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Инициализация логгера
	logger.Init()

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("Файл .env не найден. Продолжение с переменными окружения")
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создание контекста для управления жизненным циклом приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов ОС для грациозного завершения
	go handleShutdown(cancel)

	// Инициализация и запуск сервисов
	if err := startServices(ctx, cfg); err != nil {
		logger.Log.Fatalf("Ошибка запуска сервисов: %v", err)
	}
}

func handleShutdown(cancel context.CancelFunc) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	logger.Log.Info("Получен сигнал остановки. Завершение работы...")
	cancel()
}

func startServices(ctx context.Context, cfg *config.AppConfig) error {
	// Инициализация и запуск сервиса базы данных
	dbService, err := service.NewDatabaseService(cfg)
	if err != nil {
		return err
	}
	go dbService.Start(ctx)

	// Инициализация и запуск сервиса кэша
	cacheService, err := service.NewCacheService(dbService.GetDB())
	if err != nil {
		return err
	}
	go cacheService.Start(ctx)

	// Инициализация и запуск NATS
	natsService, err := nats.NewNatsService(cfg, dbService.GetDB(), cacheService.GetCache())
	if err != nil {
		return err
	}
	go natsService.Start(ctx)

	// Инициализация и запуск HTTP сервера
	httpService := api.NewHTTPService(cfg, cacheService.GetCache(), dbService.GetDB())
	go httpService.Start(ctx)

	// Блокировка основного потока до получения сигнала остановки
	<-ctx.Done()

	// Остановка сервисов
	natsService.Stop()
	httpService.Stop()
	cacheService.Stop()
	dbService.Stop()

	return nil
}
