package interfaces

import (
	"context"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
)

// IOrderCache определяет интерфейс для кэширования заказов.
// Этот интерфейс абстрагирует логику кэширования, позволяя использовать различные реализации.
type IOrderCache interface {
	// LoadOrdersFromDB загружает заказы из базы данных в кэш.
	// Принимает контекст для управления временем выполнения.
	LoadOrdersFromDB(ctx context.Context) error

	// GetOrder извлекает заказ по его идентификатору.
	// Возвращает указатель на заказ и флаг, указывающий на его наличие в кэше.
	GetOrder(id string) (*model.Order, bool)

	// GetAllOrderIDs возвращает список всех идентификаторов заказов, хранящихся в кэше.
	GetAllOrderIDs() []string

	// DeleteOrder удаляет заказ из кэша по его идентификатору.
	// Возвращает флаг, указывающий на успешность удаления.
	DeleteOrder(id string) bool

	// AddOrUpdateOrder добавляет новый заказ в кэш или обновляет существующий.
	// Принимает указатель на заказ. Возвращает ошибку в случае неудачи.
	AddOrUpdateOrder(order *model.Order) error
}
