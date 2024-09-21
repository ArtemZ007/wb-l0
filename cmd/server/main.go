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
	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/ArtemZ007/wb-l0/internal/subscription"
	"github.com/ArtemZ007/wb-l0/pkg/config"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := loadConfig()
	log := logger.New(cfg.GetLogLevel())

	if err := runApp(cfg, log); err != nil {
		log.Fatal("Ошибка запуска приложения: ", err)
	}
}

func runApp(cfg config.IConfiguration, log logger.Logger) error {
	// Подключение к базе данных
	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		log.Error("Ошибка подключения к базе данных: ", err)
		return err
	}
	defer closeDB(db, log)

	// Проверка соединения с базой данных
	if err := db.Ping(); err != nil {
		log.Error("Не удалось подключиться к базе данных: ", err)
		return err
	}
	log.Info("Успешное подключение к базе данных")
	log.Info("Запуск приложения")

	// Инициализация сервиса базы данных
	dbService, err := database.NewService(db, logrus.New())
	if err != nil {
		log.Error("Ошибка создания сервиса базы данных: ", err)
		return err
	}
	log.Info("Сервис базы данных инициализирован")

	// Инициализация сервиса кэша
	cacheService, err := initCacheService(cfg, log)
	if err != nil {
		log.Error("Ошибка инициализации сервиса кэша: ", err)
		return err
	}
	log.Info("Сервис кэша инициализирован")

	// Установка сервиса базы данных в сервис кэша
	cacheService.SetDBService(dbService)

	// Инициализация кэша данными из базы данных
	ctx := context.Background()
	if err := cacheService.InitCacheWithDBOrders(ctx); err != nil {
		log.Error("Ошибка инициализации кэша данными из базы данных: ", err)
		return err
	}

	// Обертка для сервиса кэша
	cacheServiceWrapper := &CacheServiceWrapper{cacheService: cacheService}

	// Инициализация HTTP хендлера
	handler := httpQS.NewHandler(cacheServiceWrapper, log)
	server := initHTTPServer(cfg, handler)

	// Запуск HTTP сервера в отдельной горутине
	go startHTTPServer(server, log)

	// Инициализация NATS слушателя
	natsListener, err := subscription.NewListener(cfg.GetNATSURL(), cfg.GetNATSClusterID(), cfg.GetNATSClientID(), cacheService, dbService, log)
	if err != nil {
		log.Error("Ошибка инициализации NATS слушателя: ", err)
		return err
	}

	// Запуск NATS слушателя в отдельной горутине
	go startNATSListener(natsListener, ctx, log)

	// Ожидание сигнала завершения работы
	<-waitForShutdownSignal(log)

	// Завершение работы HTTP сервера
	if err := server.Shutdown(context.Background()); err != nil {
		log.Error("Ошибка завершения работы HTTP сервера: ", err)
	}

	return nil
}

// closeDB закрывает соединение с базой данных
func closeDB(db *sql.DB, log logger.Logger) {
	if err := db.Close(); err != nil {
		log.Error("Ошибка при закрытии соединения с базой данных: ", err)
	}
}

// initCacheService инициализирует сервис кэша
func initCacheService(cfg config.IConfiguration, log logger.Logger) (*cache.CacheService, error) {
	logrusLogger := logrus.New()
	logWrapper := logger.NewLogrusAdapter(logrusLogger)
	cacheService := cache.NewCacheService(cfg.GetRedisAddr(), cfg.GetRedisPassword(), cfg.GetRedisDB(), logWrapper)
	if cacheService == nil {
		log.Error("Не удалось создать сервис кэша")
		return nil, errors.New("не удалось создать сервис кэша")
	}
	log.Info("Сервис кэша успешно создан")
	return cacheService, nil
}

// initHTTPServer инициализирует HTTP сервер
func initHTTPServer(cfg config.IConfiguration, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetServerPort()),
		Handler: handler,
	}
}

// startHTTPServer запускает HTTP сервер
func startHTTPServer(server *http.Server, log logger.Logger) {
	log.Info("HTTP сервер запущен на порту ", server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error("Ошибка HTTP сервера: ", err)
	}
}

// startNATSListener запускает NATS слушателя
func startNATSListener(listener *subscription.Listener, ctx context.Context, log logger.Logger) {
	if err := listener.Start(ctx); err != nil {
		log.Error("Ошибка NATS слушателя: ", err)
	}
}

// CacheServiceWrapper оборачивает cache.CacheService для реализации интерфейса httpQS.DataService
type CacheServiceWrapper struct {
	cacheService *cache.CacheService
}

// GetData метод для CacheServiceWrapper
func (w *CacheServiceWrapper) GetData() ([]model.Order, bool) {
	return w.cacheService.GetData()
}

// GetOrder метод для CacheServiceWrapper
func (w *CacheServiceWrapper) GetOrder(orderUID string) (*model.Order, error) {
	return w.cacheService.GetOrder(context.Background(), orderUID)
}

// waitForShutdownSignal ожидает сигнала завершения работы
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
