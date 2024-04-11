package cache_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://user:password@localhost:5432/dbname?sslmode=disable"
)

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	// Убедитесь, что база данных доступна
	err = db.Ping()
	if err != nil {
		t.Fatalf("Не удалось выполнить ping базы данных: %v", err)
	}

	return db
}

func TestIntegration_LoadOrdersFromDB(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	logger := logrus.New()
	cacheService := cache.NewCacheService(logger)

	// Подготовка тестовых данных
	order := model.Order{OrderUID: "test-order-uid", Items: []model.Item{}}
	orderData, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("Ошибка при сериализации заказа: %v", err)
	}

	// Исправлено: Добавлены плейсхолдеры для параметров
	_, err = db.Exec("INSERT INTO orders (order_uid, order_data) VALUES ($1, $2) ON CONFLICT (order_uid) DO NOTHING", order.OrderUID, orderData)
	if err != nil {
		t.Fatalf("Ошибка при вставке тестового заказа в базу данных: %v", err)
	}

	// Выполнение теста
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cacheService.LoadOrdersFromDB(ctx, db)

	// Проверка результатов
	retrievedOrder, found := cacheService.GetOrder("test-order-uid")
	assert.True(t, found, "Заказ должен быть найден в кэше")
	assert.NotNil(t, retrievedOrder, "Полученный заказ не должен быть nil")
	assert.Equal(t, "test-order-uid", retrievedOrder.OrderUID, "UID полученного заказа должен соответствовать ожидаемому")

	// Очистка тестовых данных
	// Исправлено: Добавлен плейсхолдер для параметра
	_, err = db.Exec("DELETE FROM orders WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		log.Printf("Ошибка при очистке тестовых данных: %v", err)
	}
}
