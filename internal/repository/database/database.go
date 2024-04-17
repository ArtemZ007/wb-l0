// Package database Пакет предоставляет операции с базой данных, необходимые для управления заказами в системе.
// Это включает создание, обновление и получение заказов из базы данных.
//
// Автор: ArtemZ007
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
	"time"

	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/google/uuid"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/interfaces"
	_ "github.com/ArtemZ007/wb-l0/pkg/logger"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Service struct {
	db     *sql.DB
	logger logger.ILogger // Change this to use ILogger
}

// Убедимся, что Service реализует интерфейс OrderService.
var _ interfaces.IOrderService = &Service{}

// В начале файла, где определены импорты
var _ uuid.UUID

// NewService создает новый экземпляр Service с подключением к базе данных и логгером.
func NewService(db *sql.DB, logger logger.ILogger) (*Service, error) {
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
	// Использование контекста с таймаутом для операций
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при начале транзакции")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				s.logger.WithError(txErr).Error("Ошибка при откате транзакции")
			}
			s.logger.WithField("panic", p).Error("Паника при сохранении заказа, транзакция отменена")
		}
	}()

	// Проверка на nil перед использованием указателей
	if order.Delivery == nil || order.Payment == nil {
		s.logger.Error("Доставка или оплата не могут быть nil")
		return errors.New("доставка или оплата не могут быть nil")
	}

	// Сохранение основной информации о заказе
	orderQuery := `INSERT INTO ecommerce.orders (order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id, locale) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING order_uid;`
	var orderId uuid.UUID
	if err := tx.QueryRowContext(ctx, orderQuery, order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard, order.CustomerID, *order.Locale).Scan(&orderId); err != nil {
		s.logger.WithFields(logrus.Fields{"query": orderQuery, "orderUID": order.OrderUID}).WithError(err).Error("Ошибка при сохранении заказа")
		return tx.Rollback()
	}

	// Сохранение информации о доставке
	deliveryQuery := `INSERT INTO ecommerce.deliveries (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := tx.ExecContext(ctx, deliveryQuery, *order.Delivery.Name, *order.Delivery.Phone, *order.Delivery.Zip, *order.Delivery.City, *order.Delivery.Address, *order.Delivery.Region, *order.Delivery.Email); err != nil {
		s.logger.WithFields(logrus.Fields{"query": deliveryQuery}).WithError(err).Error("Ошибка при сохранении информации о доставке")
		return tx.Rollback()
	}

	// Сохранение информации об оплате
	paymentQuery := `INSERT INTO ecommerce.payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	if _, err := tx.ExecContext(ctx, paymentQuery, *order.Payment.Transaction, *order.Payment.RequestID, *order.Payment.Currency, *order.Payment.Provider, *order.Payment.Amount, *order.Payment.PaymentDt, *order.Payment.Bank, *order.Payment.DeliveryCost, *order.Payment.GoodsTotal, *order.Payment.CustomFee); err != nil {
		s.logger.WithFields(logrus.Fields{"query": paymentQuery}).WithError(err).Error("Ошибка при сохранении информации об оплате")
		return tx.Rollback()
	}

	// Сохранение информации о товарах
	itemQuery := `INSERT INTO ecommerce.items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	for _, item := range order.Items {
		if _, err := tx.ExecContext(ctx, itemQuery, *item.ChrtID, *item.TrackNumber, *item.Price, *item.RID, *item.Name, *item.Sale, *item.Size, *item.TotalPrice, *item.NmID, *item.Brand, *item.Status); err != nil {
			s.logger.WithFields(logrus.Fields{"query": itemQuery, "itemID": item.ChrtID}).WithError(err).Error("Ошибка при сохранении информации о товарах")
			return tx.Rollback()
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Ошибка при подтверждении транзакции")
		return err
	}

	s.logger.WithFields(logrus.Fields{"order_uid": order.OrderUID}).Info("Заказ успешно сохранен")
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
	query := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM ecommerce.orders;`
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
		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard); err != nil {
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

// GetOrder возвращает заказ по его уникальному идентификатору, включая детали доставки, оплаты и товары.
func (s *Service) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	var order *model.Order
	itemsMap := make(map[string]*model.Item)

	query := `SELECT o.order_uid, o.track_number, o.date_created, d.address, p.payment_dt, i.name, i.price
FROM ecommerce.orders o
LEFT JOIN ecommerce.deliveries d ON o.delivery_id = d.id
LEFT JOIN ecommerce.payments p ON o.payment_id = p.id
LEFT JOIN ecommerce.items i ON o.order_uid = i.order_uid
WHERE o.order_uid = $1;`

	rows, err := s.db.QueryContext(ctx, query, orderID)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при выполнении запроса к базе данных")
		return nil, err
	}
	defer rows.Close()

	var deliveryAddress, paymentDt, itemName, itemPrice sql.NullString

	for rows.Next() {
		if order == nil {
			order = new(model.Order)
		}

		if err := rows.Scan(&order.OrderUID, &order.TrackNumber, &order.DateCreated, &deliveryAddress, &paymentDt, &itemName, &itemPrice); err != nil {
			s.logger.WithError(err).Error("Ошибка при считывании данных из базы")
			continue
		}

		if deliveryAddress.Valid && order.Delivery == nil {
			order.Delivery = &model.Delivery{Address: &deliveryAddress.String}
		}
		if paymentDt.Valid && order.Payment == nil {
			paymentDtInt, err := strconv.Atoi(paymentDt.String)
			if err != nil {
				s.logger.WithError(err).Error("Ошибка при конвертации paymentDt в int")
				return nil, err
			}
			order.Payment = &model.Payment{PaymentDt: &paymentDtInt}
		}

		if itemName.Valid && itemPrice.Valid {
    		itemID := // ваш код для получения уникального идентификатора товара
    		if _, exists := itemsMap[itemID]; !exists {
       			itemPriceInt, err := strconv.Atoi(itemPrice.String)
        		if err != nil {
            		s.logger.WithError(err).Error("Ошибка при конвертации itemPrice в int")
            		return nil, err
       			}
       			item := model.Item{Name: &itemName.String, Price: &itemPriceInt}
       			itemsMap[itemID] = &item
       			order.Items = append(order.Items, item)
   			}
		}

	if err := rows.Err(); err != nil {
		s.logger.WithError(err).Error("Ошибка при обработке результатов запроса")
		return nil, err
	}

	if order == nil || order.OrderUID == "" {
		return nil, fmt.Errorf("заказ с UID %s не найден", orderID)
	}

	s.logger.Info(fmt.Sprintf("Заказ с UID: %s успешно получен", orderID))
	return order, nil
}

func (s *Service) Start(ctx context.Context) error {
	// Реализуйте любую логику запуска, которая требуется вашему сервису, например, инициализацию соединений,
	// подготовку кэшей или что-либо еще, что требуется перед тем, как сервис сможет работать в обычном режиме.
	// Это реализация-заглушка и может быть не нужна, если ваш сервис не требует
	// какой-либо специальной логики запуска.

	// Пример реализации:
	// Проверка соединения с базой данных
	if err := s.db.PingContext(ctx); err != nil {
		s.logger.WithError(err).Error("Не удалось подключиться к базе данных")
		return err
	}

	// Инициализация кэша, подключений к внешним сервисам и т.д.
	// s.initializeCache()
	// s.connectToExternalServices()

	s.logger.Info("Сервис успешно запущен и готов к работе")
	return nil
}
