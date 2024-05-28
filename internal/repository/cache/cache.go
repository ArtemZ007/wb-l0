package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/sirupsen/logrus"
)

// IOrderService определяет интерфейс для работы с заказами.
type IOrderService interface {
	GetAllOrders(ctx context.Context) ([]model.Order, error)
}

// Service предоставляет методы для работы с кэшем заказов.
type Service struct {
	mu        sync.RWMutex
	logger    *logrus.Logger
	orders    map[string]*model.Order
	dbService IOrderService
}

// NewService создает новый экземпляр Service.
func NewService(logger *logrus.Logger) *Service {
	return &Service{
		logger: logger,
		orders: make(map[string]*model.Order),
	}
}

// SetDatabaseService устанавливает зависимость от сервиса базы данных.
func (s *Service) SetDatabaseService(dbService IOrderService) {
	s.dbService = dbService
}

// InitCacheWithDBOrders инициализирует кэш заказов из базы данных.
func (s *Service) InitCacheWithDBOrders(ctx context.Context) error {
	orders, err := s.dbService.GetAllOrders(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при получении заказов из базы данных")
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, order := range orders {
		// Создаем копию заказа для безопасного сохранения в кэше.
		orderCopy := order
		s.orders[order.OrderUID] = &orderCopy
	}

	s.logger.Info("Кэш успешно инициализирован из базы данных")
	return nil
}

// GetAllOrders возвращает все заказы из кэша.
func (s *Service) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]model.Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, *order)
	}

	return orders, nil
}

// GetOrder возвращает заказ по его уникальному идентификатору.
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[orderUID]
	if !exists {
		return nil, fmt.Errorf("заказ с UID: %s не найден", orderUID)
	}

	return order, nil
}

// SaveOrder сохраняет заказ в кэше.
func (s *Service) SaveOrder(ctx context.Context, order *model.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Создаем копию заказа для безопасного сохранения в кэше.
	orderCopy := *order
	s.orders[order.OrderUID] = &orderCopy

	s.logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно сохранен в кэше")
	return nil
}

// UpdateOrder обновляет информацию о заказе в кэше.
func (s *Service) UpdateOrder(ctx context.Context, order *model.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Создаем копию заказа для безопасного обновления в кэше.
	orderCopy := *order
	s.orders[order.OrderUID] = &orderCopy

	s.logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно обновлен в кэше")
	return nil
}

// DeleteOrder удаляет заказ из кэша.
func (s *Service) DeleteOrder(ctx context.Context, orderUID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.orders, orderUID)

	s.logger.WithField("order_uid", orderUID).Info("Заказ успешно удален из кэша")
	return nil
}
