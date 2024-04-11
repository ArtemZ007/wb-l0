package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	myhttp "github.com/ArtemZ007/wb-l0/internal/delivery/http"
	"github.com/ArtemZ007/wb-l0/internal/delivery/nats"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/db"
)

func initLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(level)
	}
	log.SetOutput(os.Stdout)
	return log
}

func main() {
	// Загрузка .env файла один раз в начале
	envPath, _ := filepath.Abs("../../.env")
	if err := godotenv.Load(envPath); err != nil {
		logrus.Warn("Файл .env не найден. Продолжение с переменными окружения")
	}

	log := initLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cacheService := cache.NewCacheService(log)

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	dbService, err := db.NewDBService(connectionString, cacheService, log)
	if err != nil {
		log.Fatalf("Ошибка инициализации сервиса базы данных: %v", err)
	}
	defer dbService.Close()

	cacheService.LoadOrdersFromDB(ctx, dbService.DB())

	natsListener, err := nats.NewOrderListener(os.Getenv("NATS_URL"), os.Getenv("NATS_CLUSTER_ID"), os.Getenv("NATS_CLIENT_ID"), cacheService, log)
	if err != nil {
		log.Fatalf("Ошибка инициализации NATS слушателя: %v", err)
	}
	go natsListener.Start(ctx)

	httpHandler := myhttp.NewHandler(cacheService, dbService, log)
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	httpServer := myhttp.NewServer(os.Getenv("SERVER_PORT"), mux, log)
	go httpServer.Start(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Info("Получен сигнал остановки. Завершение работы...")
		cancel()
	}()

	<-ctx.Done()
	log.Info("Приложение завершило работу.")
}
