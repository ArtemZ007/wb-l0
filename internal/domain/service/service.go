package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/db"
	"github.com/ArtemZ007/wb-l0/pkg/validator"
	"github.com/nats-io/stan.go"
)

// Service структура сервиса, объединяющая компоненты для обработки сообщений
type Service struct {
	cacheService *cache.Cache         // Сервис кэширования
	db           *sql.DB              // Подключение к базе данных
	validator    *validator.Validator // Валидатор заказов
	stanConn     stan.Conn            // Подключение к NATS Streaming
	subscription stan.Subscription    // Подписка на канал NATS Streaming
}

// NewService создает новый экземпляр Service
func NewService(cacheService *cache.Cache, db *sql.DB, validator *validator.Validator, stanConn stan.Conn) *Service {
	return &Service{
		cacheService: cacheService,
		db:           db,
		validator:    validator,
		stanConn:     stanConn,
	}
}

// Start запускает сервис и его компоненты
func (s *Service) Start(ctx context.Context) error {
	var err error
	// Подписка на канал NATS Streaming
	s.subscription, err = s.stanConn.Subscribe("orders", func(msg *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			log.Printf("Ошибка десериализации сообщения: %v", err)
			return
		}

		// Валидация полученного заказа
		if err := s.validator.ValidateOrder(&order); err != nil {
			log.Printf("Ошибка валидации заказа: %v", err)
			return
		}

		// Кэширование заказа
		s.cacheService.SetOrder(order.OrderUID, &order)

		// Запись заказа в базу данных
		if err := db.SaveOrder(s.db, &order); err != nil {
			log.Printf("Ошибка сохранения заказа в базу данных: %v", err)
			return
		}

		log.Printf("Заказ с ID %s успешно обработан", order.OrderUID)
	}, stan.DurableName("service-durable"), stan.SetManualAckMode())

	if err != nil {
		return err
	}

	log.Println("Сервис успешно запущен")
	return nil
}

// Stop останавливает сервис и его компоненты
func (s *Service) Stop() {
	if err := s.subscription.Unsubscribe(); err != nil {
		log.Printf("Ошибка отписки от канала: %v", err)
	}
	log.Println("Сервис успешно остановлен")
}
