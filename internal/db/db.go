package db

import (
	"database/sql"
	"fmt"
	"log"

	model "github.com/ArtemZ007/wb-l0/internal/model"
	// PostgreSQL driver
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
func GetOrderByID(db *sql.DB, id int) (*model.Order, error) {
	var order model.Order
	query := `SELECT id, product, quantity FROM orders WHERE id = $1`
	row := db.QueryRow(query, id)

	if err := row.Scan(&order.ID, &order.Product, &order.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no order with id %d", id)
		}
		return nil, err
	}

	return &order, nil
}

// UpdateOrder обновляет информацию о заказе в базе данных
func UpdateOrder(db *sql.DB, order *model.Order) error {
	query := `UPDATE orders SET product = $1, quantity = $2 WHERE id = $3`
	_, err := db.Exec(query, order.Product, order.Quantity, order.ID)
	return err
}
