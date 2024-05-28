package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

// Cache interface defines the methods for cache operations.
type Cache interface {
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	AddOrUpdateOrder(order *model.Order) error
	GetData() ([]model.Order, error)
	ProcessOrder(ctx context.Context, order *model.Order) error
	UpdateOrderInCache(ctx context.Context, order *model.Order) error
	GetOrderFromCache(ctx context.Context, orderUID string) (*model.Order, error)
}

// IOrderService interface defines the methods for order operations.
type IOrderService interface {
	ListOrders(ctx context.Context) ([]model.Order, error)
}

// Service represents the cache service.
type Service struct {
	cache     map[string]*model.Order
	mu        sync.RWMutex
	orders    map[string]*model.Order
	logger    *logger.Logger
	dbService IOrderService
	orderChan chan *model.Order
}

// NewCacheService creates and returns a new Cache instance.
func NewCacheService(logger *logger.Logger) *Service {
	return &Service{
		cache:     make(map[string]*model.Order),
		logger:    logger,
		orders:    make(map[string]*model.Order),
		orderChan: make(chan *model.Order),
	}
}

// SetDatabaseService sets the database service that implements the IOrderService interface.
func (c *Service) SetDatabaseService(dbService IOrderService) {
	c.dbService = dbService
}

// InitCacheWithDBOrders initializes the cache with orders from the database.
func (c *Service) InitCacheWithDBOrders(ctx context.Context) error {
	orders, err := c.dbService.ListOrders(ctx)
	if err != nil {
		c.logger.Error("Ошибка при получении заказов из базы данных", map[string]interface{}{"error": err})
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		orderCopy := order
		c.orders[order.OrderUID] = &orderCopy
	}
	c.logger.Info("Кэш инициализирован заказами ", map[string]interface{}{"Значение": len(orders)})

	return nil
}

// ProcessOrder processes an order by sending it to the order channel.
func (c *Service) ProcessOrder(ctx context.Context, order *model.Order) error {
	select {
	case c.orderChan <- order:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetOrder returns an order by its unique identifier.
func (s *Service) GetOrder(id string) (*model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.cache[id]
	if !exists {
		return nil, fmt.Errorf("order with ID %s not found", id)
	}
	return order, nil
}

// GetAllOrderIDs returns all unique order IDs.
func (s *Service) GetAllOrderIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.cache))
	for id := range s.cache {
		ids = append(ids, id)
	}
	return ids
}

// AddOrUpdateOrder adds or updates an order in the cache.
func (s *Service) AddOrUpdateOrder(order *model.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[order.OrderUID] = order
	return nil
}

// GetData returns all orders from the cache.
func (c *Service) GetData() ([]model.Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]model.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, *order)
	}
	return orders, nil
}

// UpdateOrderInCache updates an order in the cache.
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

// GetOrderFromCache returns an order from the cache by its unique identifier.
func (c *Service) GetOrderFromCache(_ context.Context, orderUID string) (*model.Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if order, exists := c.orders[orderUID]; exists {
		return order, nil
	}
	return nil, fmt.Errorf("заказ с UID %s не найден в кэше", orderUID)
}

// LoadOrdersFromDB loads orders from the database and populates the cache.
func (s *Service) LoadOrdersFromDB(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id, order_data FROM orders")
	if err != nil {
		return err
	}
	defer rows.Close()

	s.mu.Lock()
	defer s.mu.Unlock()

	for rows.Next() {
		var id string
		var orderData []byte
		if err := rows.Scan(&id, &orderData); err != nil {
			return err
		}

		var order model.Order
		if err := json.Unmarshal(orderData, &order); err != nil {
			return err
		}

		s.cache[id] = &order
	}

	if err := rows.Err(); err != nil {
		return err
	}

	s.logger.Info("Orders loaded from database into cache")
	return nil
}
