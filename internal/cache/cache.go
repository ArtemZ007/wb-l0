package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Cache структура для кэша в памяти.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order // Использование конкретного типа для заказов.
}

// New создает новый экземпляр кэша.
func New() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
	}
}

// LoadOrdersFromDB загружает данные заказов в кэш из базы данных.
func (c *Cache) LoadOrdersFromDB(db *sql.DB) error {
	// Предполагается, что здесь будет реализация загрузки заказов из базы данных в кэш.
	return nil // Заглушка, чтобы соответствовать требованиям компилятора.
}

// GetOrder возвращает заказ по ID из кэша.
func (c *Cache) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.orders[id]
	return order, exists
}

// UpdateOrder обновляет заказ в кэше и базе данных.
func (c *Cache) UpdateOrder(db *sql.DB, id string, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[id]; !exists {
		return errors.New("заказ не найден в кэше")
	}

	c.orders[id] = order

	// Здесь должен быть код для обновления заказа в базе данных.
	// Примерный SQL запрос: UPDATE orders SET order_data = $1 WHERE order_uid = $2
	_, err := db.Exec("UPDATE orders SET order_data = $1 WHERE order_uid = $2", order, id)
	if err != nil {
		log.Printf("Ошибка при обновлении заказа с ID %s в базе данных: %v", id, err)
		return err
	}

	log.Printf("Заказ с ID %s успешно обновлен в кэше и базе данных.", id)
	return nil
}

// AddOrder добавляет новый заказ в кэш и базу данных.
func (c *Cache) AddOrder(db *sql.DB, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[order.OrderUID]; exists {
		return errors.New("заказ уже существует в кэше")
	}

	c.orders[order.OrderUID] = order

	// Здесь должен быть код для добавления заказа в базу данных.
	// Примерный SQL запрос: INSERT INTO orders (order_uid, order_data) VALUES ($1, $2)
	_, err := db.Exec("INSERT INTO orders (order_uid, order_data) VALUES ($1, $2)", order.OrderUID, order)
	if err != nil {
		log.Printf("Ошибка при добавлении заказа с ID %s в базу данных: %v", order.OrderUID, err)
		return err
	}

	log.Printf("Заказ с ID %s успешно добавлен в кэш и базу данных.", order.OrderUID)
	return nil
}

// GetOrderJSON возвращает заказ в формате JSON по ID из кэша.
func (c *Cache) GetOrderJSON(id string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, found := c.orders[id]
	if !found {
		log.Printf("Заказ с ID %s не найден в кэше.", id)
		return nil, errors.New("заказ не найден")
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("Ошибка при сериализации заказа с ID %s в JSON: %v", id, err)
		return nil, fmt.Errorf("ошибка при сериализации заказа: %w", err)
	}

	log.Printf("Заказ с ID %s найден в кэше и сериализован в JSON.", id)
	return orderJSON, nil
}
