package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ArtemZ007/wb-l0/internal/api"
	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/db"
	"github.com/ArtemZ007/wb-l0/internal/nats"
)

func main() {
	// Загрузка конфигурации из переменных окружения
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	natsURL := os.Getenv("NATS_URL")

	// Подключение к базе данных
	database, err := db.Connect(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer database.Close()

	// Инициализация кэша и его восстановление из базы данных
	c := cache.New()
	err = c.LoadFromDB(database)
	if err != nil {
		log.Fatalf("Failed to load cache from database: %v", err)
	}

	// Подключение к NATS Streaming
	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS Streaming: %v", err)
	}
	defer natsConn.Close()

	// Подписка на канал NATS и обработка сообщений
	err = nats.Subscribe(natsConn, database, c)
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS channel: %v", err)
	}

	// Запуск HTTP-сервера
	router := api.NewRouter(c)
	log.Println("Starting HTTP server on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
