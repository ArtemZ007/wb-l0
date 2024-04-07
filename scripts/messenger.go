package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Определение структуры данных, аналогичной модели в вашем проекте
type OrderMessage struct {
	OrderID     string `json:"order_id"`
	TrackNumber string `json:"track_number"`
	// Дополните структуру данными, соответствующими вашей модели
}

// Функция для генерации рандомного заказа
func generateRandomOrder() OrderMessage {
	return OrderMessage{
		OrderID:     uuid.New().String(),
		TrackNumber: uuid.New().String(),
		// Заполните остальные поля рандомными данными
	}
}

func main() {
	// Подключение к NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Ошибка подключения к NATS:", err)
	}
	defer nc.Close()

	// Генерация и отправка сообщений
	for {
		order := generateRandomOrder()
		data, err := json.Marshal(order)
		if err != nil {
			log.Println("Ошибка при маршалинге заказа:", err)
			continue
		}

		// Отправка сообщения
		if err := nc.Publish("orders", data); err != nil {
			log.Println("Ошибка при публикации сообщения:", err)
			continue
		}

		// Логирование отправленного сообщения
		log.Printf("Отправлен новый заказ: %s\n", order.OrderID)

		// Таймаут в одну секунду между сообщениями
		time.Sleep(1 * time.Second)
	}
}
