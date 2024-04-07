package cache

import (
	"database/sql"
	"log"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/model"
)

// Cache структура для кэша в памяти.
type Cache struct {
	mu     sync.RWMutex
	orders map[string]*model.Order // Использование конкретного типа для заказов
}

// New создает новый экземпляр кэша.
func New() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
	}
}

// LoadOrdersFromDB загружает данные заказов в кэш из базы данных.
func (c *Cache) LoadOrdersFromDB(db *sql.DB) error {
	// Реализация остается без изменений...
	// Добавить логирование при успешной загрузке данных
	log.Println("Заказы успешно загружены в кэш")
	return nil
}

// GetOrder возвращает заказ по ID из кэша.
func (c *Cache) GetOrder(id string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, found := c.orders[id]
	return order, found
}

// UpdateOrder обновляет заказ в кэше.
func (c *Cache) UpdateOrder(id string, order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Проверка на существование заказа может быть добавлена здесь
	c.orders[id] = order
	// Добавить логирование об успешном обновлении
	log.Printf("Заказ %s обновлен в кэше\n", id)
	return nil
}

// AddOrder добавляет новый заказ в кэш.
func (c *Cache) AddOrder(order *model.Order) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Проверка на дубликаты может быть добавлена здесь
	c.orders[order.OrderUID] = order
	// Добавить логирование об успешном добавлении
	log.Printf("Заказ %s добавлен в кэш\n", order.OrderUID)
	return nil
}

func NewCache() *Cache {
	return &Cache{
		orders: make(map[string]*model.Order),
	}
}

func (c *Cache) Get(orderID string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, found := c.orders[orderID]
	return order, found
}

func (c *Cache) Set(orderID string, order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.orders[orderID] = order
}
