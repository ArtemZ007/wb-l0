package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/google/uuid"
	"github.com/nats-io/stan.go"
)

func generateRandomOrder() model.Order {
	currency := "RUB"
	locale := "ru"
	amount := rand.Intn(10000) + 100
	paymentDt := int(time.Now().Unix())
	deliveryCost := rand.Intn(1000) + 100
	goodsTotal := rand.Intn(10000) + 500
	customFee := rand.Intn(1000)
	chrtID := rand.Intn(100000)
	price := rand.Intn(10000) + 100
	sale := rand.Intn(100)
	totalPrice := rand.Intn(10000) + 100
	nmID := rand.Intn(100000)
	status := rand.Intn(5) + 1 // Corrected to match the expected type

	entry := "entry_" + uuid.New().String()
	name := "Name_" + uuid.New().String()
	phone := fmt.Sprintf("+7%010d", rand.Intn(10000000000)) // Генерация номера телефона в формате +7XXXXXXXXXX
	zip := "Zip_" + uuid.New().String()
	city := "City_" + uuid.New().String()
	address := "Address_" + uuid.New().String()
	region := "Region_" + uuid.New().String()
	email := fmt.Sprintf("test%d@example.com", rand.Intn(100000)) // Простая генерация email
	transaction := uuid.New().String()
	requestID := uuid.New().String()
	provider := "Provider_" + uuid.New().String()
	bank := "Bank_" + uuid.New().String()
	itemTrackNumber := uuid.New().String()
	rid := uuid.New().String()
	itemName := "ItemName_" + uuid.New().String()
	size := "Size_" + uuid.New().String()
	brand := "Brand_" + uuid.New().String()
	internalSignature := uuid.New().String()
	customerID := uuid.New().String()
	deliveryService := "DeliveryService_" + uuid.New().String()
	shardkey := uuid.New().String()
	oofShard := uuid.New().String()
	smID := rand.Int()                 // This generates a random int
	trackNumber := uuid.New().String() // Corrected: Definition of trackNumber

	return model.Order{
		OrderUID:    uuid.New().String(),
		TrackNumber: &trackNumber,
		Entry:       &entry,
		Delivery: &model.Delivery{
			Name:    &name,
			Phone:   &phone,
			Zip:     &zip,
			City:    &city,
			Address: &address,
			Region:  &region,
			Email:   &email,
		},
		Payment: &model.Payment{
			Transaction:  &transaction,
			RequestID:    &requestID,
			Currency:     &currency,
			Provider:     &provider,
			Amount:       &amount,
			PaymentDt:    &paymentDt,
			Bank:         &bank,
			DeliveryCost: &deliveryCost,
			GoodsTotal:   &goodsTotal,
			CustomFee:    &customFee,
		},
		Items: []model.Item{
			{
				ChrtID:      &chrtID,
				TrackNumber: &itemTrackNumber,
				Price:       &price,
				RID:         &rid,
				Name:        &itemName,
				Sale:        &sale,
				Size:        &size,
				TotalPrice:  &totalPrice,
				NmID:        &nmID,
				Brand:       &brand,
				Status:      &status,
			},
		},
		Locale:            &locale,
		InternalSignature: &internalSignature,
		CustomerID:        &customerID,
		DeliveryService:   &deliveryService,
		Shardkey:          &shardkey,
		SMID:              &smID,
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          &oofShard,
	}
}

func main() {
	// Подключение к NATS Streaming
	sc, err := stan.Connect("test-cluster", "publisher", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("Ошибка подключения к NATS Streaming:", err)
	}
	defer func() {
		if closeErr := sc.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии соединения с NATS Streaming: %v", closeErr)
		}
	}()

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
