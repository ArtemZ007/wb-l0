package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Cache структура кэша, хранящая заказы.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order
}

// New создает новый экземпляр кэша.
func New() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
	}
}

// LoadOrdersFromDB асинхронно загружает заказы из базы данных в кэш.
func (c *Cache) LoadOrdersFromDB(ctx context.Context, db *sql.DB) {
	go func() {
		rows, err := db.QueryContext(ctx, `SELECT order_uid, order_data FROM orders`)
		if err != nil {
			log.Printf("Ошибка при запросе заказов из базы данных: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var orderUID string
			var orderData []byte
			if err := rows.Scan(&orderUID, &orderData); err != nil {
				log.Printf("Ошибка при чтении строки из базы данных: %v", err)
				continue
			}

			var order model.Order
			if err := json.Unmarshal(orderData, &order); err != nil {
				log.Printf("Ошибка при десериализации заказа с UID %s: %v", orderUID, err)
				continue
			}

			c.mu.Lock()
			c.orders[orderUID] = &order
			c.mu.Unlock()
		}
		if err := rows.Err(); err != nil {
			log.Printf("Ошибка при обработке результатов запроса: %v", err)
		}
		log.Println("Заказы успешно загружены в кэш из базы данных")
	}()
}

// GetOrder возвращает заказ по его ID.
func (c *Cache) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

// UpdateOrder обновляет заказ в кэше и базе данных.
func (c *Cache) UpdateOrder(ctx context.Context, db *sql.DB, id string, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[id]; !exists {
		return errors.New("заказ не найден в кэше")
	}

	c.orders[id] = order

	orderData, err := json.Marshal(order)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "UPDATE orders SET order_data = $1 WHERE order_uid = $2", orderData, id)
	if err != nil {
		return err
	}

	return nil
}

// AddOrder добавляет новый заказ в кэш и базу данных.
func (c *Cache) AddOrder(ctx context.Context, db *sql.DB, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[order.OrderUID]; exists {
		return errors.New("заказ уже существует в кэше")
	}

	c.orders[order.OrderUID] = order

	orderData, err := json.Marshal(order)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "INSERT INTO orders (order_uid, order_data) VALUES ($1, $2)", order.OrderUID, orderData)
	if err != nil {
		return err
	}

	return nil
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
