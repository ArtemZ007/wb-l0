package cache_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestNewCacheService проверяет корректность создания нового экземпляра кэш-сервиса.
func TestNewCacheService(t *testing.T) {
	logger := logrus.New()
	cacheService := NewCacheService(logger)

	assert.NotNil(t, cacheService, "Экземпляр кэш-сервиса не должен быть nil")
	assert.NotNil(t, cacheService.orders, "Мапа заказов должна быть инициализирована")
	assert.Equal(t, logger, cacheService.logger, "Логгер должен соответствовать переданному в конструктор")
}

// TestAddOrUpdateOrderAndGetOrder проверяет добавление заказа в кэш и его последующее получение.
func TestAddOrUpdateOrderAndGetOrder(t *testing.T) {
	logger := logrus.New()
	cacheService := NewCacheService(logger)

	order := &model.Order{OrderUID: "order1"}
	cacheService.AddOrUpdateOrder(order)

	retrievedOrder, found := cacheService.GetOrder("order1")
	assert.True(t, found, "Заказ должен быть найден в кэше")
	assert.Equal(t, order, retrievedOrder, "Полученный заказ должен соответствовать добавленному")
}

// TestGetAllOrderIDs проверяет получение всех ID заказов из кэша.
func TestGetAllOrderIDs(t *testing.T) {
	logger := logrus.New()
	cacheService := NewCacheService(logger)

	order1 := &model.Order{OrderUID: "order1"}
	order2 := &model.Order{OrderUID: "order2"}
	cacheService.AddOrUpdateOrder(order1)
	cacheService.AddOrUpdateOrder(order2)

	ids := cacheService.GetAllOrderIDs()
	assert.Contains(t, ids, "order1", "ID первого заказа должен присутствовать в списке")
	assert.Contains(t, ids, "order2", "ID второго заказа должен присутствовать в списке")
	assert.Len(t, ids, 2, "Список ID должен содержать ровно два элемента")
}

// TestLoadOrdersFromDB проверяет асинхронную загрузку заказов из базы данных в кэш.
func TestLoadOrdersFromDB(t *testing.T) {
	logger := logrus.New()
	cacheService := NewCacheService(logger)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Ошибка при создании mock объекта базы данных: %s", err)
	}
	defer db.Close()

	order := &model.Order{OrderUID: "order1"}
	orderData, _ := json.Marshal(order)

	rows := sqlmock.NewRows([]string{"order_uid", "order_data"}).
		AddRow(order.OrderUID, orderData)
	mock.ExpectQuery("^SELECT order_uid, order_data FROM orders$").WillReturnRows(rows)

	ctx := context.Background()
	cacheService.LoadOrdersFromDB(ctx, db)

	// Проверяем, что mock запроса к базе данных был выполнен
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания выполнены: %s", err)
	}

	// Проверяем, что заказ был загружен в кэш
	retrievedOrder, found := cacheService.GetOrder("order1")
	assert.True(t, found, "Заказ должен быть найден в кэше после загрузки из БД")
	assert.NotNil(t, retrievedOrder, "Полученный заказ не должен быть nil")
	assert.Equal(t, "order1", retrievedOrder.OrderUID, "UID полученного заказа должен соответствовать ожидаемому")
}
