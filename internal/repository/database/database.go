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
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"os"
	"path/filepath"

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

// NewService создает новый экземпляр Service с подключением к базе данных и логгером.
func NewService(db *sql.DB, logger *logger.Logger) (*Service, error) {
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при начале транзакции")
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			s.logger.WithField("panic", p).Error("Паника при сохранении заказа, транзакция отменена")
		}
	}()

	// Сохранение основной информации о заказе
	orderQuery := `INSERT INTO orders (order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING order_uid;`
	var orderId int
	err = tx.QueryRowContext(ctx, orderQuery, order.OrderUID, order.TrackNumber, order.Entry, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard, order.CustomerID).Scan(&orderId)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		s.logger.WithError(err).Error("Ошибка при сохранении заказа")
		return err
	}

	// Сохранение информации о доставке
	deliveryQuery := `INSERT INTO deliveries (id, name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err = tx.ExecContext(ctx, deliveryQuery, orderId, *order.Delivery.Name, *order.Delivery.Phone, *order.Delivery.Zip, *order.Delivery.City, *order.Delivery.Address, *order.Delivery.Region, *order.Delivery.Email); err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		s.logger.WithError(err).Error("Ошибка при сохранении информации о доставке")
		return err
	}

	// Сохранение информации об оплате
	paymentQuery := `INSERT INTO payments (id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err = tx.ExecContext(ctx, paymentQuery, orderId, *order.Payment.Transaction, *order.Payment.RequestID, *order.Payment.Currency, *order.Payment.Provider, *order.Payment.Amount, *order.Payment.PaymentDt, *order.Payment.Bank, *order.Payment.DeliveryCost, *order.Payment.GoodsTotal, *order.Payment.CustomFee); err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
		s.logger.WithError(err).Error("Ошибка при сохранении информации об оплате")
		return err
	}

	// Сохранение информации о товарах
	itemQuery := `INSERT INTO items (id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items {
		if _, err = tx.ExecContext(ctx, itemQuery, orderId, *item.ChrtID, *item.TrackNumber, *item.Price, *item.RID, *item.Name, *item.Sale, *item.Size, *item.TotalPrice, *item.NmID, *item.Brand, *item.Status); err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
			}
			s.logger.WithError(err).Error("Ошибка при сохранении информации о товарах")
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Ошибка при подтверждении транзакции")
		return err
	}

	s.logger.Info("Заказ успешно сохранен", logrus.Fields{"order_uid": order.OrderUID})
	return nil
}

func (s *Service) UpdateOrder(ctx context.Context, order *model.Order) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при сериализации заказа в JSON")
		return err
	}

	query := `UPDATE orders SET order_uid = $1 WHERE order_uid = $2;`
	if _, err = s.db.ExecContext(ctx, query, jsonData, order.OrderUID); err != nil {
		s.logger.WithError(err).WithField("order_uid", order.OrderUID).Error("Ошибка при обновлении заказа в базе данных")
		return err
	}

	s.logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно обновлен в базе данных")
	return nil
}

func (s *Service) ListOrders(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	query := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders;`
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
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	var order model.Order
	query := `SELECT order_uid, track_number, entry, delivery_service, shardkey, sm_id, date_created, oof_shard, customer_id FROM orders WHERE order_uid = $1;`
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
	_, err := s.db.ExecContext(ctx, "DELETE FROM orders WHERE order_uid = $1", orderUID)
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
