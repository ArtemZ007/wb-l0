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

// ICacheInterface определяет интерфейс для операций с кэшем.
// Использование интерфейсов улучшает тестируемость и гибкость кода.
type ICacheInterface interface {
	LoadOrdersFromDB(ctx context.Context, db *sql.DB) error
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	DeleteOrder(id string) bool
	AddOrUpdateOrder(order *model.Order) error
	GetData() ([]model.Order, error) // Add this line
}

// GetData retrieves all orders from the cache.
func (c *cache) GetData() ([]model.Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orders := make([]model.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, *order)
	}
	return orders, nil
}

// cache структура, реализующая ICacheInterface.
// Используется для хранения заказов в памяти.
type cache struct {
	mu     sync.RWMutex // Защита orders от одновременного доступа.
	orders map[string]*model.Order
	logger logger.ILogger
}

// NewCacheService создает и возвращает новый экземпляр кэша.
// Это функция-конструктор, обеспечивающая инкапсуляцию создания экземпляров.
func NewCacheService(logger logger.ILogger) (ICacheInterface, error) {
	return &cache{
		orders: make(map[string]*model.Order),
		logger: logger,
	}, nil
}

// GetAllOrderIDs возвращает список всех ID заказов, хранящихся в кэше.
// Использование defer для разблокировки мьютекса упрощает чтение кода.
func (c *cache) GetAllOrderIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.orders))
	for id := range c.orders {
		ids = append(ids, id)
	}
	return ids
}

// LoadOrdersFromDB загружает заказы из базы данных в кэш.
// Этот метод демонстрирует использование контекста для управления отменой операций.
func (c *cache) LoadOrdersFromDB(ctx context.Context, db *sql.DB) error {
	c.logger.Info("Загрузка заказов из БД в кэш")
	query := "SELECT order_uid FROM orders" // Ensure you're selecting the 'data' column as well.

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Ошибка при загрузке заказов из БД: %v", err))
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			c.logger.Error(fmt.Sprintf("Ошибка при закрытии строк: %v", err))
		}
	}()

	for rows.Next() {
		var orderUID string
		var data []byte
		if err := rows.Scan(&orderUID, &data); err != nil {
			c.logger.Error(fmt.Sprintf("Ошибка при сканировании заказа из БД: %v", err))
			continue // Продолжаем обработку следующих строк вместо возврата ошибки.
		}

		var order model.Order
		if err := json.Unmarshal(data, &order); err != nil {
			c.logger.Error(fmt.Sprintf("Ошибка при десериализации данных заказа: %v", err))
			continue
		}

		// Handling the error returned by AddOrUpdateOrder
		if err := c.AddOrUpdateOrder(&order); err != nil {
			c.logger.Error(fmt.Sprintf("Ошибка при добавлении или обновлении заказа в кэше: %v", err))
			// Decide whether to continue or not based on the nature of the error.
		}
	}

	if err := rows.Err(); err != nil {
		c.logger.Error(fmt.Sprintf("Ошибка при итерации по строкам базы данных: %v", err))
		return err
	}

	return nil
}

// GetOrder извлекает заказ по его ID из кэша.
func (c *cache) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

// DeleteOrder удаляет заказ из кэша по его ID.
func (c *cache) DeleteOrder(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[id]; exists {
		delete(c.orders, id)
		return true
	}
	return false
}

// AddOrUpdateOrder добавляет новый заказ в кэш или обновляет существующий.
func (c *cache) AddOrUpdateOrder(order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.OrderUID] = order
	return nil
}
