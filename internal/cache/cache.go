package cache

import (
	"database/sql"
	"errors"
	"log"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Cache структура для кэша в памяти
type Cache struct {
	mu     sync.RWMutex
	orders map[int]*model.Order
}

// New создает новый экземпляр кэша
func New() *Cache {
	return &Cache{
		orders: make(map[int]*model.Order),
	}
}

// LoadFromDB загружает данные в кэш из базы данных
func (c *Cache) LoadFromDB(db *sql.DB) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rows, err := db.Query("SELECT id, customer_name, price FROM orders")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.ID, &order.CustomerName, &order.Price); err != nil {
			log.Printf("Failed to load order from DB: %v", err)
			continue // или return err, если хотите прервать загрузку при первой ошибке
		}
		c.orders[order.ID] = &order
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// GetOrder возвращает заказ по ID из кэша
func (c *Cache) GetOrder(id int) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

// UpdateOrder обновляет заказ в кэше
func (c *Cache) UpdateOrder(id int, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[id]; !exists {
		return errors.New("order not found")
	}

	c.orders[id] = order
	return nil
}

// AddOrder добавляет новый заказ в кэш
func (c *Cache) AddOrder(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[order.ID] = order
}
