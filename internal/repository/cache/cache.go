package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// Cache определяет методы для операций с кэшем.
type Cache interface {
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	AddOrUpdateOrder(order *model.Order) error
	GetData() ([]model.Order, bool)
	UpdateOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, orderUID string) (*model.Order, error)
	InitCacheWithDBOrders(ctx context.Context) error
}

// OrderService определяет методы для операций с заказами.
type OrderService interface {
	ListOrders(ctx context.Context) ([]model.Order, error)
}

// Service представляет собой сервис кэша.
type Service struct {
	client    *redis.Client
	logger    logger.Logger
	dbService OrderService
	orderChan chan *model.Order
	mu        sync.RWMutex
}

// NewService создает и возвращает новый экземпляр Service.
func NewService(redisAddr, redisPassword string, redisDB int, logger logger.Logger) *Service {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return &Service{
		client:    rdb,
		logger:    logger,
		orderChan: make(chan *model.Order),
	}
}

// SetDBService устанавливает сервис базы данных, реализующий интерфейс OrderService.
func (s *Service) SetDBService(dbService OrderService) {
	s.dbService = dbService
}

// InitCacheWithDBOrders инициализирует кэш заказами из базы данных.
func (s *Service) InitCacheWithDBOrders(ctx context.Context) error {
	orders, err := s.dbService.ListOrders(ctx)
	if err != nil {
		s.logger.Error("Ошибка при получении заказов из базы данных", map[string]interface{}{"error": err})
		return err
	}

	for _, order := range orders {
		if err := s.AddOrUpdateOrder(&order); err != nil {
			s.logger.Error("Ошибка при добавлении заказа в кэш", map[string]interface{}{"error": err})
		}
	}
	s.logger.Info("Кэш инициализирован заказами", map[string]interface{}{"count": len(orders)})

	return nil
}

// GetOrder извлекает заказ из кэша по его уникальному идентификатору.
func (s *Service) GetOrder(id string) (*model.Order, bool) {
	orderData, err := s.client.Get(context.Background(), id).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		s.logger.Error("Ошибка при получении заказа из Redis", map[string]interface{}{"error": err})
		return nil, false
	}

	var order model.Order
	if err := json.Unmarshal([]byte(orderData), &order); err != nil {
		s.logger.Error("Ошибка при декодировании заказа из Redis", map[string]interface{}{"error": err})
		return nil, false
	}

	return &order, true
}

// GetAllOrderIDs возвращает все уникальные идентификаторы заказов.
func (s *Service) GetAllOrderIDs() []string {
	keys, err := s.client.Keys(context.Background(), "*").Result()
	if err != nil {
		s.logger.Error("Ошибка при получении всех ключей из Redis", map[string]interface{}{"error": err})
		return nil
	}
	return keys
}

// AddOrUpdateOrder добавляет или обновляет заказ в кэше.
func (s *Service) AddOrUpdateOrder(order *model.Order) error {
	orderData, err := json.Marshal(order)
	if err != nil {
		s.logger.Error("Ошибка при сериализации заказа", map[string]interface{}{"error": err})
		return err
	}

	if err := s.client.Set(context.Background(), order.OrderUID, orderData, 0).Err(); err != nil {
		s.logger.Error("Ошибка при добавлении заказа в Redis", map[string]interface{}{"error": err})
		return err
	}

	return nil
}

// GetData возвращает все заказы из кэша.
func (s *Service) GetData() ([]model.Order, bool) {
	keys, err := s.client.Keys(context.Background(), "*").Result()
	if err != nil {
		s.logger.Error("Ошибка при получении всех ключей из Redis", map[string]interface{}{"error": err})
		return nil, false
	}

	var orders []model.Order
	for _, key := range keys {
		orderData, err := s.client.Get(context.Background(), key).Result()
		if err != nil {
			s.logger.Error("Ошибка при получении заказа из Redis", map[string]interface{}{"error": err})
			continue
		}

		var order model.Order
		if err := json.Unmarshal([]byte(orderData), &order); err != nil {
			s.logger.Error("Ошибка при декодировании заказа из Redis", map[string]interface{}{"error": err})
			continue
		}

		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, false
	}
	return orders, true
}

// UpdateOrder обновляет заказ в кэше.
func (s *Service) UpdateOrder(ctx context.Context, order *model.Order) error {
	return s.AddOrUpdateOrder(order)
}

// GetOrderByID возвращает заказ из кэша по его уникальному идентификатору.
func (s *Service) GetOrderByID(ctx context.Context, orderUID string) (*model.Order, error) {
	order, exists := s.GetOrder(orderUID)
	if !exists {
		return nil, fmt.Errorf("заказ с UID %s не найден в кэше", orderUID)
	}
	return order, nil
}
