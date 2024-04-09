package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/model"
	"github.com/google/uuid"
	"github.com/nats-io/stan.go"
)

func generateRandomOrder() model.Order {
	// Генерация случайных значений для полей заказа
	return model.Order{
		OrderUID:    uuid.New().String(),
		TrackNumber: uuid.New().String(),
		Entry:       "entry_" + uuid.New().String(),
		Delivery: &model.Delivery{
			Name:    "Name_" + uuid.New().String(),
			Phone:   "+7900" + uuid.New().String(),
			Zip:     "Zip_" + uuid.New().String(),
			City:    "City_" + uuid.New().String(),
			Address: "Address_" + uuid.New().String(),
			Region:  "Region_" + uuid.New().String(),
			Email:   "Email_" + uuid.New().String(),
		},
		Payment: &model.Payment{
			Transaction:  uuid.New().String(),
			RequestID:    uuid.New().String(),
			Currency:     "RUB",
			Provider:     "Provider_" + uuid.New().String(),
			Amount:       rand.Intn(10000) + 100,
			PaymentDt:    int(time.Now().Unix()),
			Bank:         "Bank_" + uuid.New().String(),
			DeliveryCost: rand.Intn(1000) + 100,
			GoodsTotal:   rand.Intn(10000) + 500,
			CustomFee:    rand.Intn(1000),
		},
		Items: []model.Item{
			{
				ChrtID:      rand.Intn(100000),
				TrackNumber: uuid.New().String(),
				Price:       rand.Intn(10000) + 100,
				RID:         uuid.New().String(),
				Name:        "ItemName_" + uuid.New().String(),
				Sale:        rand.Intn(100),
				Size:        "Size_" + uuid.New().String(),
				TotalPrice:  rand.Intn(10000) + 100,
				NmID:        rand.Intn(100000),
				Brand:       "Brand_" + uuid.New().String(),
				Status:      rand.Intn(100),
			},
		},
		Locale:            "ru",
		InternalSignature: uuid.New().String(),
		CustomerID:        uuid.New().String(),
		DeliveryService:   "DeliveryService_" + uuid.New().String(),
		Shardkey:          uuid.New().String(),
		SMID:              rand.Intn(100),
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          uuid.New().String(),
	}
}

func main() {
	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "publisher", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("Ошибка подключения к NATS Streaming:", err)
	}
	defer sc.Close()

	// Отправка 20 сообщений
	for i := 0; i < 20; i++ {
		order := generateRandomOrder()
		data, err := json.Marshal(order)
		if err != nil {
			log.Println("Ошибка при маршалинге заказа:", err)
			continue
		}

		// Отправка сообщения
		if err := sc.Publish("orders", data); err != nil {
			log.Println("Ошибка при публикации сообщения:", err)
			continue
		}

		// Логирование отправленного сообщения
		log.Printf("Отправлен новый заказ: %s", order.OrderUID)
		time.Sleep(1 * time.Second) // Таймаут в 1 секунду перед отправкой следующего заказа
	}
}
