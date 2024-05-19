package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Service struct {
	db     *sql.DB
	logger logger.ILogger
	cache  cache.Cache
}

// NewService создает новый экземпляр Service с подключением к базе данных и логгером.
// В database.go
func NewService(db *sql.DB, logger logger.ILogger) (*Service, error) {
	if db == nil {
		return nil, errors.New("db не может быть nil")
	}
	if logger == nil {
		return nil, errors.New("logger не может быть nil")
	}

	var s = &Service{
		db:     db,
		logger: logger,
	}

	// Получение текущего рабочего каталога
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить текущий рабочий каталог: %w", err)
	}

	// Построение абсолютного пути к файлу миграции
	migrationFilePath := filepath.Join(cwd, "../../migrations/setup_db.sql")

	// Выполнение миграции
	if err := s.executeMigration(migrationFilePath); err != nil {
		return nil, fmt.Errorf("ошибка при выполнении миграции: %w", err)
	}

	return s, nil
}

func (s *Service) Set(cacheService cache.Cache) {
	s.cache = cacheService
}

// executeMigration выполняет миграцию базы данных из указанного файла.
// Corrected to include a parameter name and type
func (s *Service) executeMigration(filePath string) error {
	content, err := os.ReadFile(filePath) // Use the filePath parameter
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

func (s *Service) saveOrderMainInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	query := `INSERT INTO ecommerce.orders (order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id, locale) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING order_uid;`
	if err := tx.QueryRowContext(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard, order.CustomerID, order.Locale).Scan(&order.OrderUID); err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении основной информации о заказе")
		return err
	}
	return nil
}

func (s *Service) saveDeliveryInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	// Проверка на nil и корректность значений
	if order.OrderUID == "" || order.Delivery == nil || order.Delivery.Name == nil || order.Delivery.Phone == nil || order.Delivery.Zip == nil || order.Delivery.City == nil || order.Delivery.Address == nil || order.Delivery.Region == nil || order.Delivery.Email == nil {
		s.logger.Error("Некорректные данные для сохранения информации о доставке")
		return fmt.Errorf("некорректные данные для сохранения информации о доставке")
	}

	deliveryQuery := `INSERT INTO ecommerce.deliveries (id, name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := tx.ExecContext(ctx, deliveryQuery, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении информации о доставке")
		return err
	}
	return nil
}

func (s *Service) savePaymentInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	paymentQuery := `INSERT INTO ecommerce.payments (id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	// Prepare the transaction field as sql.NullString
	transaction := sql.NullString{Valid: false}
	if order.Payment.Transaction != nil {
		transaction = sql.NullString{String: *order.Payment.Transaction, Valid: true}
	}

	// Ensure other fields are handled similarly if they can be nil

	// Use the correctly prepared transaction variable in the ExecContext call
	if _, err := tx.ExecContext(ctx, paymentQuery, order.OrderUID, transaction, *order.Payment.RequestID, *order.Payment.Currency, *order.Payment.Provider, *order.Payment.Amount, *order.Payment.PaymentDt, *order.Payment.Bank, *order.Payment.DeliveryCost, *order.Payment.GoodsTotal, *order.Payment.CustomFee); err != nil {
		s.logger.WithError(err).Error("Ошибка при сохранении информации об оплате")
		return err
	}
	return nil
}

func (s *Service) saveItemsInfo(ctx context.Context, tx *sql.Tx, order *model.Order) error {
	itemQuery := `INSERT INTO ecommerce.items (id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
	for _, item := range order.Items {
		// First, check if item.Status is not nil to avoid runtime panic
		if item.Status == nil {
			s.logger.Error("Статус товара не может быть nil")
			continue // Skip this item or handle as needed
		}
		// Validate the status value by dereferencing item.Status
		if *item.Status < 1 || *item.Status > 5 {
			s.logger.WithField("status", *item.Status).Error("Статус товара вне допустимого диапазона")
			continue // Example: skipping this item
		}

		// Generate a new UUID for the item ID
		itemID, err := uuid.NewUUID()
		if err != nil {
			s.logger.WithError(err).Error("Ошибка при генерации UUID для товара")
			return err
		}

		if _, err := tx.ExecContext(ctx, itemQuery, itemID, *item.ChrtID, *item.TrackNumber, *item.Price, *item.RID, *item.Name, *item.Sale, *item.Size, *item.TotalPrice, *item.NmID, *item.Brand, *item.Status); err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				s.logger.WithError(txErr).Error("Ошибка при откате транзакции")
			}
			s.logger.WithError(err).Error("Ошибка при сохранении информации о товарах")
			return err
		}
	}
	return nil
}

func (s *Service) UpdateOrder(ctx context.Context, order *model.Order) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сериализации заказа в JSON")
		return err
	}

	query := `UPDATE ecommerce.orders SET order_uid = $1 WHERE order_uid = $2;`
	if _, err = s.db.ExecContext(ctx, query, jsonData, order.OrderUID); err != nil {
		s.logger.WithError(err).WithField("order_uid", order.OrderUID).Error("Ошибка при обновлении заказа в базе данных")
		return err
	}

	s.logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно обновлен в базе данных")
	return nil
}

func (s *Service) ListOrders(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	query := `SELECT order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM ecommerce.orders;`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при получении списка заказов из базы данных")
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			s.logger.WithError(err).Error("Ошибка при закрытии результата запроса")
		}
	}(rows)

	for rows.Next() {
		var order model.Order
		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Delivery.ID, &order.Payment.ID, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard); err != nil {
			s.logger.WithError(err).Error("Ошибка при чтении данных заказа")
			continue
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		s.logger.WithError(err).Error("Ошибка при обработке результатов запроса")
		return nil, err
	}

	return orders, nil
}
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	var order model.Order
	query := `SELECT order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id FROM ecommerce.orders WHERE order_uid = $1;`
	err := s.db.QueryRowContext(ctx, query, orderUID).Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard, &order.CustomerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("заказ с UID: %s не найден", orderUID)
		}
		s.logger.WithError(err).Error("Ошибка при запросе заказа по UID")
		return nil, err
	}
	// Предполагая, что у вас есть метод для загрузки связанных сущностей, таких как элементы, информация о доставке и т. д.
	// Вы бы вызвали его здесь, чтобы полностью заполнить объект заказа перед его возвращением.
	return &order, nil
}

func (s *Service) DeleteOrder(ctx context.Context, orderUID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM ecommerce.orders WHERE order_uid = $1", orderUID)
	if err != nil {
		s.logger.WithError(err).WithField("order_uid", orderUID).Error("Ошибка при удалении заказа из базы данных")
		return err
	}
	return nil
}

func (s *Service) Start(_ context.Context) error {
	// Реализуйте любую логику запуска, которая требуется вашему сервису, например, инициализацию соединений,
	// подготовку кэшей или что-либо еще, что требуется перед тем, как сервис сможет работать в обычном режиме.
	// Это реализация-заглушка и может быть не нужна, если ваш сервис не требует
	// какой-либо специальной логики запуска.
	return nil
}
func (s *Service) GetFullOrderDetails(ctx context.Context, orderUID string) (*model.Order, error) {
	var order model.Order
	// Обновленный SQL-запрос с учетом структуры таблиц
	query := `
    SELECT o.order_uid, o.track_number, o.entry, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard, o.customer_id,
           d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
           p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
           i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
    FROM ecommerce.orders o
    LEFT JOIN ecommerce.deliveries d ON o.delivery_id = d.phone
    LEFT JOIN ecommerce.payments p ON o.payment_id = p.transaction
    LEFT JOIN ecommerce.items i ON o.order_uid = i.id
    WHERE o.order_uid = $1;
    `

	if err := s.cache.UpdateOrderInCache(ctx, &order); err != nil {
		s.logger.WithError(err).Error("Ошибка при обновлении кэша")
		// Decide whether to return the error or just log it
	}
	rows, err := s.db.QueryContext(ctx, query, orderUID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при запросе полных данных заказа")
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			s.logger.WithError(err).Error("Ошибка при закрытии результата запроса")
		}
	}()

	// Later in your code, when processing items:
	itemsMap := make(map[string]*model.Item)
	for rows.Next() {
		var delivery model.Delivery
		var payment model.Payment
		var item model.Item
		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard, &order.CustomerID,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			s.logger.WithError(err).Error("Ошибка при чтении данных заказа")
			continue
		}
		// Convert ChrtID to string to use as map key
		if _, exists := itemsMap[strconv.Itoa(*item.ChrtID)]; !exists {
			order.Items = append(order.Items, item)
			itemsMap[strconv.Itoa(*item.ChrtID)] = &item
		}
	}

	if err := rows.Err(); err != nil {
		s.logger.WithError(err).Error("Ошибка при обработке результатов запроса")
		return nil, err
	}

	if order.OrderUID == "" {
		return nil, fmt.Errorf("заказ с UID: %s не найден", orderUID)
	}

	return &order, nil
}
