// Package database Пакет database предоставляет реализацию взаимодействия с базой данных.
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"testing"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Service struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewService creates a new Service instance.
func NewService(db *sql.DB, logger *logrus.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// SaveOrder saves an order to the database.
func (s *Service) SaveOrder(ctx context.Context, order *model.Order) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.WithError(err).Error("Ошибка при начале транзакции")
		return err
	}

	// Save the delivery information
	deliveryQuery := `INSERT INTO deliveries (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	var deliveryID int
	err = tx.QueryRowContext(ctx, deliveryQuery,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email).Scan(&deliveryID)

	// Save the payment information
	paymentQuery := `INSERT INTO payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;`
	var paymentID int
	err = tx.QueryRowContext(ctx, paymentQuery,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee).Scan(&paymentID)

	// Save the order with references to the delivery and payment information
	orderQuery := `INSERT INTO orders (order_uid, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ON CONFLICT (order_uid) DO UPDATE SET delivery_id = EXCLUDED.delivery_id, payment_id = EXCLUDED.payment_id, locale = EXCLUDED.locale, internal_signature = EXCLUDED.internal_signature, customer_id = EXCLUDED.customer_id, delivery_service = EXCLUDED.delivery_service, shardkey = EXCLUDED.shardkey, sm_id = EXCLUDED.sm_id, date_created = EXCLUDED.date_created, oof_shard = EXCLUDED.oof_shard;`
	_, err = tx.ExecContext(ctx, orderQuery, order.OrderUID, deliveryID, paymentID, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SMID, order.DateCreated, order.OofShard)
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
			s.logger.WithError(err).Error("Ошибка при откате транзакции")
		}
		s.logger.WithError(err).Error("Ошибка при сохранении заказа")
		return err
	}

	// Assuming each order can have multiple items, iterate through each item and save it
	for _, item := range order.Items {
		itemQuery := `INSERT INTO items (id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`
		_, err = tx.ExecContext(ctx, itemQuery, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return err
				s.logger.WithError(err).Error("Ошибка при откате транзакции")
				return err
			}
			s.logger.WithError(err).Error("Ошибка при сохранении информации о товаре")
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		s.logger.WithError(err).Error("Ошибка при подтверждении транзакции")
		return err
	}

	s.logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно сохранен в базу данных")
	return nil
}

// TestSaveOrderFromJSON SaveOrderFromJSON function corrected with the NewService function call.
func _(t *testing.T) {
	// Mock or set up your database connection
	db, err := sql.Open("postgres", "your_connection_string")
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
	}(db)

	// Set up a logger
	logger := logrus.New()

	// Example JSON order
	jsonOrder := []byte(`{"order_uid": "12345", "data": "example data"}`)

	// Context
	ctx := context.Background()

	// Call the corrected function
	err = SaveOrderFromJSON(ctx, db, logger, jsonOrder)
	if err != nil {
		t.Errorf("SaveOrderFromJSON failed: %v", err)
	}
}

func SaveOrderFromJSON(ctx context.Context, db *sql.DB, logger *logrus.Logger, jsonOrder []byte) error {
	// Unmarshal the JSON order
	var order model.Order
	if err := json.Unmarshal(jsonOrder, &order); err != nil {
		logger.WithError(err).Error("Ошибка при десериализации заказа из JSON")
		return err
	}

	// Create a new service instance
	service := NewService(db, logger)

	// Save the order using the service
	if err := service.SaveOrder(ctx, &order); err != nil {
		logger.WithError(err).Error("Ошибка при сохранении заказа в базу данных")
		return err
	}

	logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно сохранен")
	return nil
}

// GetOrder remains unchanged from your original implementation.

// GetOrder метод для получения заказа из базы данных по уникальному идентификатору.
// Принимает контекст и идентификатор заказа.
func (s *Service) GetOrder(ctx context.Context, orderUID string) (*model.Order, error) {
	query := `SELECT order_uid FROM orders WHERE order_uid = $1;`
	row := s.db.QueryRowContext(ctx, query, orderUID)

	var orderData []byte
	if err := row.Scan(&orderData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.WithField("order_uid", orderUID).Info("Заказ не найден в базе данных")
			return nil, nil
		}
		s.logger.WithError(err).WithField("order_uid", orderUID).Error("Ошибка при получении заказа из базы данных")
		return nil, err
	}

	var order model.Order
	if err := json.Unmarshal(orderData, &order); err != nil {
		s.logger.WithError(err).WithField("order_uid", orderUID).Error("Ошибка при десериализации заказа из JSON")
		return nil, err
	}

	s.logger.WithField("order_uid", order.OrderUID).Info("Заказ успешно получен из базы данных")
	return &order, nil
}
