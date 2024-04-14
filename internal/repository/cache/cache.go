// Package cache предоставляет функциональность кэширования для хранения и извлечения заказов.
// Это включает в себя операции, такие как загрузка заказов из базы данных в кэш,
// получение конкретных заказов и управление данными кэша.
package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

type Cache interface {
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	AddOrUpdateOrder(order *model.Order) error
	GetData() ([]model.Order, error)
	ProcessOrder(ctx context.Context, order *model.Order) error
}

type Service struct {
	mu        sync.RWMutex
	orders    map[string]*model.Order
	logger    *logger.Logger
	dbService *database.Service
	orderChan chan *model.Order
}

// NewCacheService creates and returns a new Cache instance.
func NewCacheService(logger *logger.Logger, dbService *database.Service) Cache {
	return &Service{
		orders:    make(map[string]*model.Order),
		logger:    logger,
		dbService: dbService,
		orderChan: make(chan *model.Order),
	}
}
func (c *Service) InitCacheWithDBOrders(ctx context.Context) {
	orders, err := c.dbService.ListOrders(ctx)
	if err != nil {
		c.logger.Error("Ошибка при получении заказов из базы данных", map[string]interface{}{"error": err})
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.orders[order.OrderUID] = &order // Предполагается, что order.OrderUID уникальный идентификатор заказа
	}
	c.logger.Info(fmt.Sprintf("Кэш инициализирован %d заказами", len(orders)))
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
