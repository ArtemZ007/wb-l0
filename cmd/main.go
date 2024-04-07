package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/api"
	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/config"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL для использования с database/sql
	"github.com/nats-io/nats.go"
)

func main() {
	// Запуск основной логики приложения и обработка возможной ошибки
	if err := run(); err != nil {
		log.Fatalf("Приложение не удалось запустить: %v", err)
	}
}

func run() error {
	// Загрузка переменных окружения из файла .env, если он существует
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, продолжение с переменными окружения")
	}

	// Загрузка конфигурации приложения
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("не удалось загрузить конфигурацию: %w", err)
	}

	// Настройка подключения к базе данных
	db, err := setupDatabase(cfg)
	if err != nil {
		return err
	}
	defer db.Close() // Отложенное закрытие подключения к базе данных

	// Настройка кэша
	cacheService, err := setupCache(db)
	if err != nil {
		log.Printf("Не удалось настроить кэш: %v", err)
	}

	// Настройка подключения к NATS
	nc, err := setupNATS(cfg)
	if err != nil {
		return err
	}
	defer nc.Close() // Отложенное закрытие подключения к NATS

	// Запуск HTTP-сервера
	return startHTTPServer(cfg, cacheService)
}

func setupDatabase(cfg *config.AppConfig) (*sql.DB, error) {
	// Открытие подключения к базе данных
	db, err := sql.Open("postgres", cfg.Database.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть подключение к базе данных: %w", err)
	}
	// Проверка связи с базой данных
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось выполнить ping базы данных: %w", err)
	}
	log.Println("Подключение к базе данных установлено")
	return db, nil
}

func setupCache(db *sql.DB) (*cache.Cache, error) {
	// Создание экземпляра кэша
	cacheService := cache.New()
	// Загрузка данных в кэш из базы данных
	if err := cacheService.LoadOrdersFromDB(db); err != nil {
		return nil, fmt.Errorf("не удалось загрузить заказы в кэш: %w", err)
	}
	return cacheService, nil
}

func setupNATS(cfg *config.AppConfig) (*nats.Conn, error) {
	// Подключение к NATS
	nc, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к NATS: %w", err)
	}
	log.Println("Подключено к NATS")
	return nc, nil
}

func startHTTPServer(cfg *config.AppConfig, cacheService *cache.Cache) error {
	// Создание обработчика HTTP-запросов
	handler := api.NewHandler(cacheService)

	// Создание нового ServeMux.
	mux := http.NewServeMux()

	// Регистрация маршрутов.
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port, // Адрес и порт сервера
		Handler: mux,                   // Использование ServeMux в качестве обработчика запросов
	}

	// Запуск HTTP-сервера в отдельной горутине
	go func() {
		log.Printf("Запуск сервера на %s", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске HTTP-сервера: %v", err)
		}
	}()

	// Ожидание сигнала ОС для грациозного завершения работы сервера
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	// Грациозное завершение работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("не удалось корректно остановить сервер: %+v", err)
	}
	log.Println("Сервер остановлен корректно")

	return nil
}
