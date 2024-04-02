package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ArtemZ007/wb-l0/internal/model"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DBConfig структура для конфигурации подключения к базе данных
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewDBConfig создает новый экземпляр конфигурации базы данных
func NewDBConfig(host string, port int, user, password, dbname string) *DBConfig {
	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}

// Connect устанавливает соединение с базой данных
func Connect(cfg *DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка соединения с базой данных
	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the database")
	return db, nil
}

// GetOrderByID извлекает заказ из базы данных по ID
func GetOrderByID(db *sql.DB, id string) (*model.Order, error) {
	var order model.Order
	query := `SELECT order_uid, track_number, entry, delivery, payment, items, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1`
	row := db.QueryRow(query, id)

	var delivery, payment, items string
	if err := row.Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &delivery, &payment, &items, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SMID, &order.DateCreated, &order.OofShard); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no order with id %s", id)
		}
		return nil, err
	}

	// Десериализация JSON-строк в соответствующие поля
	if err := json.Unmarshal([]byte(delivery), &order.Delivery); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(payment), &order.Payment); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(items), &order.Items); err != nil {
		return nil, err
	}

	return &order, nil
}

// Поскольку структура Order содержит сложные типы данных, обновление заказа потребует специфической логики,
// зависящей от того, как именно вы хотите обновлять эти данные в вашей базе данных.
// Пример функции UpdateOrder опустим, так как он требует детального понимания вашей бизнес-логики и структуры базы данных.
