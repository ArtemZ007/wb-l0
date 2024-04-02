package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ArtemZ007/wb-l0/internal/api"
	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/db"
	"github.com/ArtemZ007/wb-l0/internal/model"
	"github.com/ArtemZ007/wb-l0/internal/nats"
	"github.com/nats-io/stan.go"
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
	dbPortInt, err := strconv.Atoi(dbPort)
	if err != nil {
		log.Fatal(err)
	}
	dbConf := db.NewDBConfig(dbHost, dbPortInt, dbUser, dbPassword, dbName)
	database, err := db.Connect(dbConf)
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
	natsConfig := nats.NATSConfig{
		URL:       natsURL,
		ClientID:  os.Getenv("NATS_CLIENT_ID"),
		ClusterID: os.Getenv("NATS_CLUSTER_ID"),
	}
	natsConn, err := nats.Connect(&natsConfig)
	if err != nil {
		log.Fatalf("Failed to connect to NATS Streaming: %v", err)
	}
	defer natsConn.Close()

	// Подписка на канал NATS с именем "orders"
	subscription, err := nats.Subscribe(natsConn, "orders", func(m *stan.Msg) {
		var order model.Order
		err := json.Unmarshal(m.Data, &order)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			return
		}

		// Логика обработки десериализованных данных
		log.Printf("Received order: %+v\n", order)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to NATS channel 'orders': %v", err)
	}
	defer subscription.Unsubscribe()

	// Запуск HTTP-сервера
	router := api.NewRouter(c) // Убедитесь, что функция NewRouter принимает кэш как аргумент
	log.Println("Starting HTTP server on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
