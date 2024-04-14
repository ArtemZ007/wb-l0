// Package cache предоставляет функциональность кэширования для хранения и извлечения заказов.
// Это включает в себя операции, такие как загрузка заказов из базы данных в кэш,
// получение конкретных заказов и управление данными кэша.
package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

// Cache определяет интерфейс для операций с кэшем.
type Cache interface {
	LoadOrdersFromDB(ctx context.Context, db *sql.DB) error
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	DeleteOrder(id string) bool
	AddOrUpdateOrder(order *model.Order) error
	GetData() ([]model.Order, error)
}

// cacheImpl структура, реализующая Cache.
type cacheImpl struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
	logger *logger.Logger
}

// New создает и возвращает новый экземпляр кэша.
func New(logger *logger.Logger) Cache {
	return &cacheImpl{
		orders: make(map[string]*model.Order),
		logger: logger,
	}
}

// LoadOrdersFromDB загружает заказы из базы данных в кэш.
// LoadOrdersFromDB загружает заказы из базы данных в кэш.
func (c *cacheImpl) LoadOrdersFromDB(ctx context.Context, db *sql.DB) error {
	c.logger.Info("Начинаем загрузку заказов из базы данных в кэш", nil)
	query := "SELECT order_uid FROM orders" // Ensure you're selecting order_data as well

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		c.logger.Error("Ошибка выполнения запроса к базе данных", map[string]interface{}{"error": err})
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			c.logger.Error("Ошибка при закрытии строк базы данных", map[string]interface{}{"error": err})
		}
	}()

	for rows.Next() {
		var orderUID string
		var data []byte
		if err := rows.Scan(&orderUID, &data); err != nil {
			c.logger.Error("Ошибка чтения строки из базы данных", map[string]interface{}{"error": err})
			continue
		}

		var order model.Order
		if err := json.Unmarshal(data, &order); err != nil {
			c.logger.Error("Ошибка десериализации данных заказа", map[string]interface{}{"error": err})
			continue
		}

		if err := c.AddOrUpdateOrder(&order); err != nil {
			c.logger.Error("Ошибка добавления или обновления заказа в кэше", map[string]interface{}{"error": err})
			continue
		}
	}

	if err := rows.Err(); err != nil {
		c.logger.Error("Ошибка при обработке результатов запроса", map[string]interface{}{"error": err})
		return err
	}

	c.logger.Info("Заказы успешно загружены в кэш из базы данных", nil)
	return nil
}

// Остальные методы остаются без изменений, но важно следить за тем, чтобы логгирование было информативным и на русском языке.

// GetOrder извлекает заказ по его ID из кэша.
func (c *cacheImpl) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

// GetAllOrderIDs возвращает список всех ID заказов в кэше.
func (c *cacheImpl) GetAllOrderIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.orders))
	for id := range c.orders {
		ids = append(ids, id)
	}
	return ids
}

// DeleteOrder удаляет заказ из кэша по его ID.
func (c *cacheImpl) DeleteOrder(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[id]; exists {
		delete(c.orders, id)
		return true
	}
	return false
}

// AddOrUpdateOrder добавляет новый заказ в кэш или обновляет существующий.
func (c *cacheImpl) AddOrUpdateOrder(order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.OrderUID] = order
	return nil
}

// GetData возвращает все заказы из кэша.
func (c *cacheImpl) GetData() ([]model.Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]model.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, *order)
	}
	return orders, nil
}
