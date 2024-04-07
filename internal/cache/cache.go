package cache

import (
	"database/sql"
	"errors"
	"log"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Cache структура для кэша в памяти.
// Использует мьютекс для безопасного доступа в многопоточной среде.
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
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("Начало загрузки заказов из базы данных...")
	rows, err := db.Query("SELECT order_uid, customer_id FROM orders")
	if err != nil {
		log.Printf("Ошибка при загрузке заказов из БД: %v", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.OrderUID, &order.CustomerID); err != nil {
			log.Printf("Ошибка при чтении строки из БД: %v", err)
			continue
		}
		c.orders[order.OrderUID] = &order
		log.Printf("Заказ с ID %s загружен в кэш.", order.OrderUID)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при обработке результатов запроса: %v", err)
		return err
	}

	log.Println("Заказы успешно загружены в кэш из БД.")
	return nil
}

// GetOrder возвращает заказ по ID из кэша.
func (c *Cache) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, found := c.orders[id]
	if found {
		log.Printf("Заказ с ID %s найден в кэше.", id)
	} else {
		log.Printf("Заказ с ID %s не найден в кэше.", id)
	}
	return order, found
}

// UpdateOrder обновляет заказ в кэше.
func (c *Cache) UpdateOrder(id string, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[id]; !exists {
		log.Printf("Заказ с ID %s не найден в кэше для обновления.", id)
		return errors.New("заказ не найден")
	}

	c.orders[id] = order
	log.Printf("Заказ с ID %s успешно обновлен в кэше.", id)
	return nil
}

// AddOrder добавляет новый заказ в кэш.
func (c *Cache) AddOrder(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.orders[order.OrderUID]; exists {
		log.Printf("Заказ с ID %s уже существует в кэше. Будет выполнено обновление.", order.OrderUID)
	} else {
		log.Printf("Добавление нового заказа с ID %s в кэш.", order.OrderUID)
	}

	c.orders[order.OrderUID] = order
	log.Printf("Заказ с ID %s успешно добавлен/обновлен в кэше.", order.OrderUID)
}
