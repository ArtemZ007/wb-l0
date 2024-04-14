package cache_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStanConn реализует интерфейс stan.Conn для мокинга соединения с NATS.
type MockStanConn struct {
	mock.Mock
}

func (m *MockStanConn) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	args := m.Called(subject, cb, opts)
	return args.Get(0).(stan.Subscription), args.Error(1)
}

func (m *MockStanConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockCache реализует интерфейс cache.Cache для мокинга кэша.
type MockCache struct {
	mock.Mock
}

func (m *MockCache) AddOrUpdateOrder(order *model.Order) {
	m.Called(order)
}

func TestOrderListener(t *testing.T) {
	logger, _ := test.NewNullLogger()
	mockStanConn := new(MockStanConn)
	mockCache := new(MockCache)

	// Создание экземпляра OrderListener с мокированными зависимостями
	listener, err := nats.NewOrderListener("nats://test", "test-cluster", "test-client", mockCache, logger)
	assert.NoError(t, err)

	// Подготовка данных заказа для тестирования
	order := model.Order{OrderUID: "test-order-uid"}
	orderData, _ := json.Marshal(order)

	// Настройка ожиданий для моков
	mockStanConn.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(&stan.Subscription{}, nil)
	mockCache.On("AddOrUpdateOrder", &order).Once()

	// Эмуляция получения сообщения
	go func() {
		time.Sleep(1 * time.Second) // Даем время на инициализацию подписки
		cb := mockStanConn.Calls[0].Arguments.Get(1).(stan.MsgHandler)
		cb(&stan.Msg{Data: orderData})
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск слушателя
	go listener.Start(ctx)

	// Ожидание, чтобы дать время на обработку сообщения
	time.Sleep(2 * time.Second)

	// Проверка, что все ожидания моков были выполнены
	mockStanConn.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
