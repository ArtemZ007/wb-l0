package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/sirupsen/logrus"
)

// IDataBase интерфейс для работы с базой данных.
type IDataBase interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// ICache интерфейс для работы с кэшем.
type ICache interface {
	GetAllOrderIDs() []string
	GetOrder(id string) (*model.Order, bool)
}

// Service DBService структура сервиса для работы с базой данных.
type Service struct {
	db     IDataBase
	cache  ICache
	logger *logrus.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDBService создает и инициализирует новый экземпляр DBService.

// Start запускает процесс синхронизации данных между кэшем и базой данных.
func (s *Service) Start() {
	s.wg.Add(1)
	go s.syncCacheToDB()
}

// Stop останавливает процесс синхронизации и ожидает его завершения.
func (s *Service) Stop() {
	s.cancel()
	s.wg.Wait()
}

// syncCacheToDB выполняет периодическую синхронизацию данных из кэша в базу данных.
func (s *Service) syncCacheToDB() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Остановка синхронизации базы данных...")
			return
		case <-ticker.C:
			if err := s.FillTablesFromCache(); err != nil {
				s.logger.Errorf("Ошибка синхронизации кэша с БД: %v", err)
			}
		}
	}
}

// FillTablesFromCache заполняет таблицы базы данных данными из кэша.
func (s *Service) FillTablesFromCache() error {
	s.logger.Info("Начало заполнения таблиц из кэша")

	orderIDs := s.cache.GetAllOrderIDs()
	for _, id := range orderIDs {
		order, found := s.cache.GetOrder(id)
		if !found {
			s.logger.Warnf("Заказ с ID %s не найден в кэше", id)
			continue
		}

		orderData, err := json.Marshal(order)
		if err != nil {
			s.logger.Errorf("Ошибка сериализации заказа с ID %s: %v", id, err)
			continue
		}

		query := `INSERT INTO orders (order_uid) VALUES ($1) ON CONFLICT (order_uid) DO NOTHING`
		if _, err := s.db.ExecContext(s.ctx, query, order.OrderUID, orderData); err != nil {
			s.logger.Errorf("Ошибка вставки заказа с ID %s в базу данных: %v", id, err)
			continue
		}

		s.logger.Infof("Заказ с ID %s успешно добавлен в базу данных из кэша", id)
	}

	s.logger.Info("Завершение заполнения таблиц из кэша")
	return nil
}
