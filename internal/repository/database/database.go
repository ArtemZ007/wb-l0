package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/sirupsen/logrus"
)

// IOrderService определяет интерфейс для работы с заказами.
type IOrderService interface {
	GetOrder(ctx context.Context, orderUID string) (*model.Order, error)
	SaveOrder(ctx context.Context, order *model.Order) error
	UpdateOrder(ctx context.Context, order *model.Order) error
	DeleteOrder(ctx context.Context, orderUID string) error
	ListOrders(ctx context.Context) ([]model.Order, error)
	Start(ctx context.Context) error
}

// Service представляет собой реализацию IOrderService.
type Service struct {
	db     *sql.DB
	cache  cache.Cache
	logger *logrus.Logger
}

// NewService создает новый экземпляр Service.
func NewService(db *sql.DB, logger *logrus.Logger) (*Service, error) {
	s := &Service{
		db:     db,
		logger: logger,
	}

	// Инициализация базы данных
	if err := s.initDB(); err != nil {
		return nil, err
	}

	return s, nil
}

// SetCache устанавливает кэш для сервиса.
func (s *Service) SetCache(cacheService cache.Cache) {
	s.cache = cacheService
}

// initDB инициализирует базу данных, выполняя миграции.
func (s *Service) initDB() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("не удалось получить текущий рабочий каталог: %w", err)
	}

	// Построение абсолютного пути к файлу миграции
	migrationFilePath := filepath.Join(cwd, "../../migrations/setup_db.sql")

	// Выполнение миграции
	if err := s.executeMigration(migrationFilePath); err != nil {
		return fmt.Errorf("ошибка при выполнении миграции: %w", err)
	}

	return nil
}

// executeMigration выполняет миграцию базы данных из указанного файла.
func (s *Service) executeMigration(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при чтении файла миграции")
		return err
	}

	if _, err = s.db.Exec(string(content)); err != nil {
		s.logger.WithError(err).Error("Ошибка при выполнении миграции")
		return err
	}

	s.logger.Info("Миграция успешно выполнена")
	return nil
}

// GetOrder возвращает заказ по его уникальному идентификатору.
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	// Проверка наличия заказа в кэше
	if s.cache != nil {
		if cachedOrder, found := s.cache.Get(orderUID); found {
			s.logger.Info("Заказ найден в кэше", orderUID)
			return cachedOrder, nil
		}
	}

	query := "SELECT * FROM orders WHERE order_uid = $1"
	row := s.db.QueryRowContext(ctx, query, orderUID)

	var order model.Order
	if err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard, &order.CustomerID, &order.Locale); err != nil {
		if err == sql.ErrNoRows {
			s.logger.WithError(err).Error("Заказ не найден")
			return nil, nil
		}
		s.logger.WithError(err).Error("Ошибка при получении заказа")
		return nil, err
	}

	// Сохранение заказа в кэше
	if s.cache != nil {
		s.cache.Set(order.OrderUID, &order)
	}

	s.logger.Info("Заказ успешно получен", order.OrderUID)
	return &order, nil
}

// SaveOrder сохраняет заказ в базе данных.
func (s *Service) SaveOrder(ctx context.Context, order *model.Order) error {
	query := "INSERT INTO orders (order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id, locale) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	_, err := s.db.ExecContext(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard, order.CustomerID, order.Locale)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении заказа")
		return err
	}

	// Сохранение заказа в кэше
	if s.cache != nil {
		s.cache.Set(order.OrderUID, order)
	}

	s.logger.Info("Заказ успешно сохранен", order.OrderUID)
	return nil
}

// UpdateOrder обновляет информацию о заказе.
func (s *Service) UpdateOrder(ctx context.Context, order *model.Order) error {
	query := "UPDATE orders SET track_number = $2, entry = $3, delivery_service = $4, shardkey = $5, sm_id = $6, date_created = $7, oof_shard = $8, customer_id = $9, locale = $10 WHERE order_uid = $1"
	_, err := s.db.ExecContext(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard, order.CustomerID, order.Locale)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при обновлении заказа")
		return err
	}

	// Обновление заказа в кэше
	if s.cache != nil {
		s.cache.Set(order.OrderUID, order)
	}

	s.logger.Info("Заказ успешно обновлен", order.OrderUID)
	return nil
}

// DeleteOrder удаляет заказ по его уникальному идентификатору.
func (s *Service) DeleteOrder(ctx context.Context, orderUID string) error {
	query := "DELETE FROM orders WHERE order_uid = $1"
	_, err := s.db.ExecContext(ctx, query, orderUID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при удалении заказа")
		return err
	}

	s.logger.Info("Заказ успешно удален", orderUID)
	return nil
}

// ListOrders возвращает список всех заказов из базы данных.
func (s *Service) ListOrders(ctx context.Context) ([]model.Order, error) {
	query := `
        SELECT
            order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id, locale
        FROM
            orders`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при получении списка заказов")
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard, &order.CustomerID, &order.Locale); err != nil {
			s.logger.WithError(err).Error("Ошибка при сканировании заказа")
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		s.logger.WithError(err).Error("Ошибка при итерации по строкам")
		return nil, err
	}

	s.logger.Info("Список заказов успешно получен")
	return orders, nil
}

// Start запускает основную логику сервиса в фоновом режиме.
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Сервис успешно запущен")
	return nil
}
