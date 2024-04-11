package nats

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

// OrderListener слушатель сообщений о заказах через NATS Streaming.
type OrderListener struct {
	sc           stan.Conn
	cacheService *cache.Cache
	logger       *logrus.Logger
}

// NewOrderListener создает новый экземпляр OrderListener для NATS Streaming.
func NewOrderListener(natsURL, clusterID, clientID string, cacheService *cache.Cache, logger *logrus.Logger) (*OrderListener, error) {
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Errorf("Ошибка подключения к NATS Streaming: %v", err)
		return nil, err
	}

	return &OrderListener{
		sc:           sc,
		cacheService: cacheService,
		logger:       logger,
	}, nil
}

// Start запускает слушатель сообщений.
func (ol *OrderListener) Start(ctx context.Context) {
	subscription, err := ol.sc.Subscribe(os.Getenv("NATS_CHANNEL_NAME"), func(msg *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			ol.logger.Errorf("Ошибка десериализации заказа: %v", err)
			return
		}

		// Кэширование заказа
		ol.cacheService.AddOrUpdateOrder(&order)
		ol.logger.Infof("Заказ с ID %s успешно обработан и добавлен в кэш", order.OrderUID)

		// Подтверждение обработки сообщения
		msg.Ack()

	}, stan.DurableName("order-listener-durable"), stan.SetManualAckMode(), stan.AckWait(stan.DefaultAckWait))
	if err != nil {
		ol.logger.Fatalf("Не удалось подписаться на канал %s: %v", os.Getenv("NATS_CHANNEL_NAME"), err)
	}
	defer subscription.Unsubscribe()

	<-ctx.Done() // Ожидание сигнала на завершение
}

// Stop останавливает слушателя и закрывает соединение с NATS Streaming.
func (ol *OrderListener) Stop() {
	if err := ol.sc.Close(); err != nil {
		ol.logger.Errorf("Ошибка при закрытии соединения с NATS Streaming: %v", err)
	}
}
