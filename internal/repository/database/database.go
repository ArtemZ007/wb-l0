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

// SetCache устанавливает кэш для сервиса.
func (s *Service) SetCache(cacheService cache.Cache) {
	s.cache = cacheService
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

// ListOrders возвращает список всех заказов из базы данных.
func (s *Service) ListOrders(ctx context.Context) ([]model.Order, error) {
	query := `
        SELECT
            o.order_uid, o.track_number, o.entry, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard, o.customer_id, o.locale,
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
            p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
            i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
        FROM
            ecommerce.orders o
        LEFT JOIN
            ecommerce.deliveries d ON o.order_uid = d.id
        LEFT JOIN
            ecommerce.payments p ON o.order_uid = p.id
        LEFT JOIN
            ecommerce.items i ON o.order_uid = i.id
    `

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при выполнении запроса к базе данных")
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*model.Order)
	for rows.Next() {
		var order model.Order
		var delivery model.Delivery
		var payment model.Payment
		var item model.Item

		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard, &order.CustomerID, &order.Locale,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			s.logger.WithError(err).Error("Ошибка при сканировании строки")
			return nil, err
		}

		if existingOrder, exists := orderMap[order.OrderUID]; exists {
			existingOrder.Items = append(existingOrder.Items, item)
		} else {
			order.Delivery = &delivery
			order.Payment = &payment
			order.Items = []model.Item{item}
			orderMap[order.OrderUID] = &order
		}
	}

	var orders []model.Order
	for _, order := range orderMap {
		orders = append(orders, *order)
	}

	s.logger.Info("Заказы успешно получены из базы данных")
	return orders, nil
}

// SaveOrder сохраняет заказ в базе данных.
func (s *Service) SaveOrder(ctx context.Context, order *model.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при начале транзакции")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			s.logger.WithField("panic", p).Error("Паника при сохранении заказа, транзакция отменена")
		}
	}()

	if err := s.saveOrderMainInfo(ctx, tx, order); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := s.saveDeliveryInfo(ctx, tx, order); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := s.savePaymentInfo(ctx, tx, order); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := s.saveItemsInfo(ctx, tx, order); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Ошибка при подтверждении транзакции")
		return err
	}

	return nil
}

// saveOrderMainInfo сохраняет основную информацию о заказе.
func (s *Service) saveOrderMainInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO ecommerce.orders (order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id, locale)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := tx.ExecContext(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard, order.CustomerID, order.Locale)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении основной информации о заказе")
		return err
	}
	return nil
}

// saveDeliveryInfo сохраняет информацию о доставке.
func (s *Service) saveDeliveryInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO ecommerce.deliveries (id, name, phone, zip, city, address, region, email)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := tx.ExecContext(ctx, query, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении информации о доставке")
		return err
	}
	return nil
}

// savePaymentInfo сохраняет информацию об оплате.
func (s *Service) savePaymentInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO ecommerce.payments (id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err := tx.ExecContext(ctx, query, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении информации об оплате")
		return err
	}
	return nil
}

// saveItemsInfo сохраняет информацию о товарах в заказе.
func (s *Service) saveItemsInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `
        INSERT INTO ecommerce.items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, query, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			s.logger.WithError(err).Error("Ошибка при сохранении информации о товаре")
			return err
		}
	}
	return nil
}
