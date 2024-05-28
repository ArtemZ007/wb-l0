package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/nats-io/stan.go"
)

func generateRandomOrder() model.Order {
	currencies := []string{"USD", "EUR", "RUB"}
	locales := []string{"en", "ru", "es"}
	names := []string{"John Doe", "Jane Smith", "Test Testov"}
	phones := []string{"+9720000000", "+1234567890", "+9876543210"}
	zips := []string{"2639809", "12345", "67890"}
	cities := []string{"Kiryat Mozkin", "New York", "Los Angeles"}
	addresses := []string{"Ploshad Mira 15", "123 Main St", "456 Elm St"}
	regions := []string{"Kraiot", "NY", "CA"}
	emails := []string{"test@gmail.com", "example@example.com", "user@domain.com"}
	providers := []string{"wbpay", "paypal", "stripe"}
	banks := []string{"alpha", "beta", "gamma"}
	itemNames := []string{"Mascaras", "Lipstick", "Foundation"}
	brands := []string{"Vivienne Sabo", "Maybelline", "L'Oreal"}

	currency := currencies[rand.Intn(len(currencies))]
	locale := locales[rand.Intn(len(locales))]
	amount := rand.Intn(10000) + 100
	paymentDt := int64(time.Now().Unix()) // Convert to int64
	deliveryCost := rand.Intn(1000) + 100
	goodsTotal := rand.Intn(10000) + 500
	customFee := rand.Intn(1000)
	chrtID := rand.Intn(100000)
	price := rand.Intn(10000) + 100
	sale := rand.Intn(100)
	totalPrice := rand.Intn(10000) + 100
	nmID := rand.Intn(100000)
	status := rand.Intn(5) + 1

	entry := "WBIL"
	name := names[rand.Intn(len(names))]
	phone := phones[rand.Intn(len(phones))]
	zip := zips[rand.Intn(len(zips))]
	city := cities[rand.Intn(len(cities))]
	address := addresses[rand.Intn(len(addresses))]
	region := regions[rand.Intn(len(regions))]
	email := emails[rand.Intn(len(emails))]
	transaction := "test-transaction"
	requestID := "test-request"
	provider := providers[rand.Intn(len(providers))]
	bank := banks[rand.Intn(len(banks))]
	itemTrackNumber := "WBILMTESTTRACK"
	rid := "test-rid"
	itemName := itemNames[rand.Intn(len(itemNames))]
	size := "0"
	brand := brands[rand.Intn(len(brands))]
	internalSignature := ""
	customerID := "test"
	deliveryService := "meest"
	shardkey := "9"
	oofShard := "1"
	smID := 99
	trackNumber := "WBILMTESTTRACK"

	return model.Order{
		OrderUID:    "b563feb7b2b84b6test",
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
			PaymentDt:    &paymentDt, // Use the converted int64 value
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
