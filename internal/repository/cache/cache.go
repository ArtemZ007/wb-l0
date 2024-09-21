package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// Cache определяет интерфейс для сервиса кэша.
type Cache interface {
	Get(key string) (*model.Order, bool)
	Set(key string, value *model.Order)
	Delete(key string)
}

// OrderService определяет методы для операций с заказами.
type OrderService interface {
	ListOrders(ctx context.Context) ([]model.Order, error)
}

// CacheService представляет собой сервис кэша.
type CacheService struct {
	client    *redis.Client
	logger    logger.Logger
	dbService OrderService
}

// NewCacheService создает и возвращает новый экземпляр CacheService.
func NewCacheService(redisAddr, redisPassword string, redisDB int, logger logger.Logger) *CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return &CacheService{
		client: rdb,
		logger: logger,
	}
}

// SetDBService устанавливает сервис базы данных, реализующий интерфейс OrderService.
func (s *CacheService) SetDBService(dbService OrderService) {
	s.dbService = dbService
}

// InitCacheWithDBOrders инициализирует кэш заказами из базы данных.
func (s *CacheService) InitCacheWithDBOrders(ctx context.Context) error {
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
func (s *CacheService) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	orderData, err := s.client.Get(ctx, orderUID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("заказ с UID %s не найден в кэше", orderUID)
	} else if err != nil {
		s.logger.Error("Ошибка при получении заказа из Redis", map[string]interface{}{"error": err})
		return nil, err
	}

	var order model.Order
	if err := json.Unmarshal([]byte(orderData), &order); err != nil {
		s.logger.Error("Ошибка при декодировании заказа из Redis", map[string]interface{}{"error": err})
		return nil, err
	}

	return &order, nil
}

// GetAllOrderIDs возвращает все уникальные идентификаторы заказов.
func (s *CacheService) GetAllOrderIDs(ctx context.Context) ([]string, error) {
	keys, err := s.client.Keys(ctx, "*").Result()
	if err != nil {
		s.logger.Error("Ошибка при получении всех ключей из Redis", map[string]interface{}{"error": err})
		return nil, err
	}
	return keys, nil
}

// AddOrUpdateOrder добавляет или обновляет заказ в кэше.
func (s *CacheService) AddOrUpdateOrder(order *model.Order) error {
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
func (s *CacheService) GetData() ([]model.Order, bool) {
	ctx := context.Background()
	keys, err := s.client.Keys(ctx, "*").Result()
	if err != nil {
		s.logger.Error("Ошибка при получении всех ключей из Redis", map[string]interface{}{"error": err})
		return nil, false
	}

	var orders []model.Order
	for _, key := range keys {
		orderData, err := s.client.Get(ctx, key).Result()
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
