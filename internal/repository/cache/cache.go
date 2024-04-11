package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/sirupsen/logrus"
)

// Cache структура кэша, хранящая заказы.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
	logger *logrus.Logger
}

// NewCacheService создает новый экземпляр кэша.
func NewCacheService(logger *logrus.Logger) *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
		logger: logger,
	}
}

// LoadOrdersFromDB асинхронно загружает заказы из базы данных в кэш.
func (c *Cache) LoadOrdersFromDB(ctx context.Context, db *sql.DB) {
	go func() {
		rows, err := db.QueryContext(ctx, `SELECT order_uid, order_data FROM orders`)
		if err != nil {
			c.logger.Errorf("Ошибка при запросе заказов из базы данных: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var orderUID string
			var orderData []byte
			if err := rows.Scan(&orderUID, &orderData); err != nil {
				c.logger.Errorf("Ошибка при чтении строки из базы данных: %v", err)
				continue
			}

			var order model.Order
			if err := json.Unmarshal(orderData, &order); err != nil {
				c.logger.Errorf("Ошибка при десериализации заказа с UID %s: %v", orderUID, err)
				continue
			}

			c.mu.Lock()
			c.orders[orderUID] = &order
			c.mu.Unlock()
		}
		if err := rows.Err(); err != nil {
			c.logger.Errorf("Ошибка при обработке результатов запроса: %v", err)
		}
		c.logger.Info("Заказы успешно загружены в кэш из базы данных")
	}()
}

// GetOrder возвращает заказ по его ID.
func (c *Cache) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

// AddOrUpdateOrder добавляет новый заказ в кэш или обновляет существующий.
func (c *Cache) AddOrUpdateOrder(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.OrderUID] = order
	c.logger.Infof("Заказ с ID %s добавлен или обновлен в кэше", order.OrderUID)
}

// GetAllOrderIDs возвращает список всех ID заказов в кэше.
func (c *Cache) GetAllOrderIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var ids []string
	for id := range c.orders {
		ids = append(ids, id)
	}
	return ids
}
