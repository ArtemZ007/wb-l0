package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ArtemZ007/wb-l0/internal/model"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DBConfig структура для конфигурации подключения к базе данных.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// Connect устанавливает соединение с базой данных и возвращает *sql.DB.
func Connect(cfg *DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось установить соединение с базой данных: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось выполнить ping базы данных: %w", err)
	}

	log.Println("Успешное подключение к базе данных")
	return db, nil
}

// GetOrderByID извлекает заказ из базы данных по ID и возвращает его.
func GetOrderByID(db *sql.DB, id string) (*model.Order, error) {
	var order model.Order
	query := `SELECT order_uid, track_number, entry, delivery, payment, items, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1`
	row := db.QueryRow(query, id)

	var delivery, payment, items string
	if err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &delivery, &payment, &items, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("заказ с ID %s не найден", id)
		}
		return nil, fmt.Errorf("ошибка при извлечении заказа из базы данных: %w", err)
	}

	// Десериализация JSON-строк в соответствующие поля структуры Order.
	if err := json.Unmarshal([]byte(delivery), &order.Delivery); err != nil {
		return nil, fmt.Errorf("ошибка при десериализации данных доставки: %w", err)
	}
	if err := json.Unmarshal([]byte(payment), &order.Payment); err != nil {
		return nil, fmt.Errorf("ошибка при десериализации данных оплаты: %w", err)
	}
	if err := json.Unmarshal([]byte(items), &order.Items); err != nil {
		return nil, fmt.Errorf("ошибка при десериализации данных товаров: %w", err)
	}

	return &order, nil
}

// InsertData вставляет данные в указанную таблицу базы данных.
func InsertData(db *sql.DB, tableName string, dataColumn string, dataValue string) error {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1)", tableName, dataColumn)
	_, err := db.Exec(query, dataValue)
	if err != nil {
		return fmt.Errorf("ошибка при вставке данных в таблицу %s: %w", tableName, err)
	}
	return nil
}
