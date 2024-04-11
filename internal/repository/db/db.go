package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type DBService struct {
	db     *sql.DB
	cache  *cache.Cache
	logger *logrus.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (s *DBService) DB() *sql.DB {
	return s.db
}

// NewDBService создает новый сервис для работы с базой данных и кэшем.
func NewDBService(connectionString string, cache *cache.Cache, logger *logrus.Logger) (*DBService, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Errorf("Ошибка подключения к базе данных: %v", err)
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}
	if err = db.Ping(); err != nil {
		logger.Errorf("Ошибка проверки соединения с базой данных: %v", err)
		return nil, fmt.Errorf("ошибка проверки соединения с базой данных: %w", err)
	}
	logger.Info("Соединение с базой данных успешно установлено")

	ctx, cancel := context.WithCancel(context.Background())
	service := &DBService{
		db:     db,
		cache:  cache,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	return service, nil
}

// Start запускает процесс синхронизации данных между кэшем и базой данных.
func (s *DBService) Start() {
	s.wg.Add(1)
	go s.syncCacheToDB()
}

// Stop останавливает процесс синхронизации и ожидает его завершения.
func (s *DBService) Stop() {
	s.cancel()
	s.wg.Wait()
}

// syncCacheToDB выполняет периодическую синхронизацию данных из кэша в базу данных.
func (s *DBService) syncCacheToDB() {
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

// GetOrderByID извлекает заказ по его ID из базы данных.
func (service *DBService) GetOrderByID(ctx context.Context, orderID string) (*model.Order, error) {
	var order model.Order
	// Предполагается, что у вас есть SQL запрос для получения заказа по ID. Настройте запрос в соответствии с вашей схемой.
	query := `SELECT order_uid, track_number, date_created FROM orders WHERE order_uid = $1`
	err := service.db.QueryRowContext(ctx, query, orderID).Scan(&order.OrderUID, &order.TrackNumber, &order.DateCreated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Или ваша собственная ошибка, указывающая на отсутствие результата
		}
		return nil, err
	}
	// Возможно, вам потребуется отдельно загрузить связанные сущности (например, товары, доставку, оплату) или настроить запрос для их включения.
	return &order, nil
}

// FillTablesFromCache заполняет таблицы базы данных данными из кэша.
func (s *DBService) FillTablesFromCache() error {
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

		query := `INSERT INTO orders (order_uid, order_data) VALUES ($1, $2) ON CONFLICT (order_uid) DO NOTHING`
		if _, err := s.db.ExecContext(s.ctx, query, order.OrderUID, orderData); err != nil {
			s.logger.Errorf("Ошибка вставки заказа с ID %s в базу данных: %v", id, err)
			continue
		}

		s.logger.Infof("Заказ с ID %s успешно добавлен в базу данных из кэша", id)
	}

	s.logger.Info("Завершение заполнения таблиц из кэша")
	return nil
}
func (service *DBService) Close() error {
	if service.db != nil {
		return service.db.Close()
	}
	return nil
}
