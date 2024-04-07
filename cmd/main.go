package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/api"
	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/config"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Настройка логгера
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	if err := run(); err != nil {
		logrus.Fatalf("Ошибка запуска приложения: %v", err)
	}
}

func run() error {
	if err := setupEnvironment(); err != nil {
		return err
	}

	cfg, err := setupConfig()
	if err != nil {
		return err
	}

	database, err := setupDatabase(cfg)
	if err != nil {
		return err
	}

	cacheService, err := setupCache(database)
	if err != nil {
		logrus.Warnf("Ошибка настройки кэша: %v", err)
	}

	if err := setupNATS(cfg); err != nil {
		return err
	}

	return startHTTPServer(cfg, cacheService)
}

func setupConfig() {
	panic("unimplemented")
}

func setupEnvironment() error {
	if err := godotenv.Load(); err != nil {
		logrus.Warn("Файл .env не найден. Используются переменные окружения по умолчанию.")
	}
	return nil
}

// Остальные функции setupConfig, setupDatabase, setupCache, setupNATS остаются без изменений

func startHTTPServer(cfg *config.AppConfig, cacheService *cache.Cache) error {
	handler := api.NewHandler(cacheService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: mux,
	}

	go func() {
		logrus.Println("Запуск сервера на порту:", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Ошибка при запуске HTTP-сервера: %v", err)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	logrus.Println("Сервер остановлен")
	return nil
}
