package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

type Cache interface {
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	AddOrUpdateOrder(order *model.Order) error
	GetData() ([]model.Order, error)
	ProcessOrder(ctx context.Context, order *model.Order) error
	UpdateOrderInCache(ctx context.Context, order *model.Order) error
	GetOrderFromCache(ctx context.Context, orderUID string) (*model.Order, error)
}

type IOrderService interface {
	ListOrders(ctx context.Context) ([]model.Order, error)
}

type Service struct {
	mu        sync.RWMutex
	orders    map[string]*model.Order
	logger    *logger.Logger
	dbService IOrderService // Используйте интерфейс вместо конкретного типа
	orderChan chan *model.Order
}

// NewCacheService creates and returns a new Cache instance without a direct dependency on a database service.
// The database service can be set later using SetDatabaseService method.
func NewCacheService(logger *logger.Logger) *Service {
	return &Service{
		logger:    logger,
		orders:    make(map[string]*model.Order),
		orderChan: make(chan *model.Order),
	}
}

// SetDatabaseService Adjust the SetDatabaseService method to accept an interface rather than a concrete type.// SetDatabaseService sets the database service that implements the IOrderService interface.
func (c *Service) SetDatabaseService(dbService IOrderService) {
	c.dbService = dbService
}
func (c *Service) InitCacheWithDBOrders(ctx context.Context) error {
	orders, err := c.dbService.ListOrders(ctx)
	if err != nil {
		c.logger.Error("Ошибка при получении заказов из базы данных", map[string]interface{}{"error": err})
		return err // Return the error if there is one
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		orderCopy := order // Создаем копию для безопасного сохранения в кэше
		c.orders[order.OrderUID] = &orderCopy
	}
	c.logger.Info("Кэш инициализирован заказами ", map[string]interface{}{"Значение": len(orders)})

	return nil // Correctly return nil here to indicate success
}

func (c *Service) ProcessOrder(ctx context.Context, order *model.Order) error {
	select {
	case c.orderChan <- order:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Service) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

func (c *Service) GetAllOrderIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.orders))
	for id := range c.orders {
		ids = append(ids, id)
	}
	return ids
}

func (c *Service) AddOrUpdateOrder(order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.OrderUID] = order
	return nil
}

func (c *Service) GetData() ([]model.Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]model.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, *order)
	}
	return orders, nil
}
func (c *Service) UpdateOrderInCache(_ context.Context, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[order.OrderUID]; exists {
		c.logger.Info("Заказ уже существует в кэше и будет обновлен", map[string]interface{}{"orderUID": order.OrderUID})
	} else {
		c.logger.Info("Заказ добавлен в кэш", map[string]interface{}{"orderUID": order.OrderUID})
	}

	c.orders[order.OrderUID] = order
	return nil
}

func (c *Service) GetOrderFromCache(_ context.Context, orderUID string) (*model.Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Пытаемся найти заказ в кэше по его UID
	if order, exists := c.orders[orderUID]; exists {
		return order, nil
	}
	// Если заказ не найден в кэше, возвращаем ошибку
	return nil, fmt.Errorf("заказ с UID %s не найден в кэше", orderUID)
}
