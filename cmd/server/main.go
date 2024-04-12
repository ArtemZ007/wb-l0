package main

import (
	"context"       // Пакет для управления жизненным циклом контекста
	"fmt"           // Пакет для форматированного ввода и вывода
	"net/http"      // Пакет для работы с HTTP
	"os"            // Пакет для взаимодействия с операционной системой
	"os/signal"     // Пакет для обработки сигналов ОС
	"path/filepath" // Пакет для работы с путями файлов
	"syscall"       // Пакет для работы с системными вызовами

	"github.com/joho/godotenv"   // Пакет для загрузки переменных окружения из файла .env
	"github.com/sirupsen/logrus" // Пакет для логирования

	myhttp "github.com/ArtemZ007/wb-l0/internal/delivery/http" // Пакет для HTTP-сервера
	"github.com/ArtemZ007/wb-l0/internal/delivery/nats"        // Пакет для работы с NATS
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"     // Пакет для работы с кэшем
	"github.com/ArtemZ007/wb-l0/internal/repository/db"        // Пакет для работы с базой данных
)

// initLogger инициализирует и настраивает логгер
func initLogger() *logrus.Logger {
	log := logrus.New() // Создание нового экземпляра логгера
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true, // Включение полной метки времени в логи
	})
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL")) // Получение уровня логирования из переменной окружения
	if err != nil {
		log.SetLevel(logrus.InfoLevel) // Установка уровня логирования по умолчанию, если произошла ошибка
	} else {
		log.SetLevel(level) // Установка уровня логирования из переменной окружения
	}
	log.SetOutput(os.Stdout) // Вывод логов в стандартный поток вывода
	return log
}

func main() {
	// Загрузка .env файла один раз в начале
	envPath, _ := filepath.Abs("../../.env") // Получение абсолютного пути к файлу .env
	if err := godotenv.Load(envPath); err != nil {
		logrus.Warn("Файл .env не найден. Продолжение с переменными окружения") // Предупреждение, если файл .env не найден
	}

	log := initLogger() // Инициализация логгера

	ctx, cancel := context.WithCancel(context.Background()) // Создание контекста с возможностью отмены
	defer cancel()                                          // Отмена контекста при завершении работы функции

	cacheService := cache.NewCacheService(log) // Создание сервиса кэша

	// Формирование строки подключения к базе данных из переменных окружения
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	// Создание экземпляра сервиса для работы с базой данных
	dbService, err := db.NewDBService(connectionString, cacheService, log)
	if err != nil {
		log.Fatalf("Ошибка создания сервиса базы данных: %v", err)
	}

	// Отложенное закрытие подключения к базе данных
	defer func() {
		if err := dbService.Close(); err != nil {
			log.Errorf("Ошибка закрытия подключения к базе данных: %v", err)
		}
	}()

	// Загрузка заказов из базы данных в кэш при старте приложения
	cacheService.LoadOrdersFromDB(ctx, dbService.DB())

	// Инициализация слушателя сообщений NATS. Это позволяет приложению реагировать на сообщения,
	// публикуемые в NATS, например, на новые заказы.
	natsListener, err := nats.NewOrderListener(os.Getenv("NATS_URL"), os.Getenv("NATS_CLUSTER_ID"), os.Getenv("NATS_CLIENT_ID"), cacheService, log)
	if err != nil {
		log.Fatalf("Ошибка инициализации NATS слушателя: %v", err) // Завершение работы приложения в случае ошибки
	}
	go natsListener.Start(ctx) // Запуск слушателя в отдельной горутине

	// Создание HTTP-обработчика, который будет управлять веб-запросами к приложению.
	// Это включает в себя обработку запросов на получение информации о заказах.
	httpHandler := myhttp.NewHandler(cacheService, dbService, log)
	mux := http.NewServeMux()                                          // Создание мультиплексора для маршрутизации запросов
	httpHandler.RegisterRoutes(mux)                                    // Регистрация маршрутов в мультиплексоре
	httpServer := myhttp.NewServer(os.Getenv("SERVER_PORT"), mux, log) // Создание HTTP-сервера
	go httpServer.Start(ctx)                                           // Запуск HTTP-сервера в отдельной горутине

	// Подготовка к корректному завершению работы приложения при получении сигнала остановки (например, SIGINT или SIGTERM)
	signalChan := make(chan os.Signal, 1)                    // Создание канала для приема сигналов
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM) // Настройка приема сигналов прерывания и завершения
	go func() {
		<-signalChan                                               // Ожидание сигнала
		log.Info("Получен сигнал остановки. Завершение работы...") // Логирование получения сигнала
		cancel()                                                   // Отправка сигнала отмены контекста для корректного завершения работы горутин
	}()

	<-ctx.Done()                             // Ожидание сигнала отмены контекста
	log.Info("Приложение завершило работу.") // Логирование завершения работы приложения
}
