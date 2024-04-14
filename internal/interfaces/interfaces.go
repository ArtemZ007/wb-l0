package interfaces

import (
	"context"
	"database/sql"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
)

// IService объединяет все сервисы в одном интерфейсе для удобства инъекции зависимостей и тестирования.
type IService interface {
	ICacheService
	ILoggerService
	IOrderService
}

// ICacheService определяет интерфейс для работы с кэшем.
type ICacheService interface {
	LoadOrdersFromDB(ctx context.Context, db *sql.DB) error
	CacheOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	AddOrUpdateOrder(order *model.Order) error
}

// ILoggerService определяет интерфейс для логирования.
type ILoggerService interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debug(args ...interface{})
	WithField(key string, value interface{}) *logger.Logger
	WithFields(fields map[string]interface{}) *logger.Logger
}

// IOrderService определяет интерфейс для работы с заказами.
type IOrderService interface {
	// GetOrder возвращает заказ по его уникальному идентификатору.
	GetOrder(ctx context.Context, orderUID string) (*model.Order, error)

	// SaveOrder сохраняет заказ в базе данных.
	SaveOrder(ctx context.Context, order *model.Order) error

	// UpdateOrder обновляет информацию о заказе.
	UpdateOrder(ctx context.Context, order *model.Order) error

	// DeleteOrder удаляет заказ по его уникальному идентификатору.
	DeleteOrder(ctx context.Context, orderUID string) error

	// ListOrders возвращает список всех заказов.
	ListOrders(ctx context.Context) ([]model.Order, error)

	// Start запускает основную логику сервиса в фоновом режиме.
	Start(ctx context.Context) error
}
