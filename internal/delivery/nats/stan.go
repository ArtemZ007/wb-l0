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

// OrderListener прослушивает сообщения на тему NATS Streaming и обрабатывает их.
type OrderListener struct {
	sc           stan.Conn      // Соединение с NATS Streaming
	cacheService *cache.Cache   // Сервис кэширования для хранения заказов
	logger       *logrus.Logger // Логгер для записи сообщений
}

// NewOrderListener создает новый экземпляр OrderListener для NATS Streaming.
// NewOrderListener создает новый экземпляр OrderListener для NATS Streaming.
func NewOrderListener(natsURL, clusterID, clientID string, cacheService *cache.Cache, logger *logrus.Logger) (*OrderListener, error) {
	logger.Infof("Проверка параметров подключения к NATS Streaming...")
	sc, err := connectToNATSStreaming(clusterID, clientID, natsURL, logger)
	if err != nil {
		return nil, err
	}

	return &OrderListener{
		sc:           sc,
		cacheService: cacheService,
		logger:       logger,
	}, nil
}

// Подключение к NATS Streaming
func connectToNATSStreaming(clusterID, clientID, natsURL string, logger *logrus.Logger) (stan.Conn, error) {
	logger.Infof("Подключение к NATS Streaming.")
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Errorf("Ошибка подключения к NATS Streaming: %v", err)
		return nil, err
	}
	return sc, nil
}

// Start запускает слушатель сообщений.
func (ol *OrderListener) Start(ctx context.Context) {
	natsSubject := os.Getenv("NATS_SUBJECT")
	if natsSubject == "" {
		ol.logger.Errorf("Переменная окружения NATS_SUBJECT не задана")
		return
	}

	ol.logger.Infof("Подписка на канал NATS: %s", natsSubject)
	subscription, err := ol.sc.Subscribe(natsSubject, func(msg *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			ol.logger.Errorf("Ошибка десериализации заказа: %v", err)
			return
		}

		ol.cacheService.AddOrUpdateOrder(&order)
		ol.logger.Infof("Заказ с ID %s успешно обработан и добавлен в кэш", order.OrderUID)

		msg.Ack()

	}, stan.DurableName("order-listener-durable"), stan.SetManualAckMode(), stan.AckWait(stan.DefaultAckWait))
	if err != nil {
		ol.logger.Fatalf("Не удалось подписаться на канал %s: %v", natsSubject, err)
	}
	defer subscription.Unsubscribe()

	<-ctx.Done()
}

// Stop останавливает слушателя и закрывает соединение с NATS Streaming.
func (ol *OrderListener) Stop() {
	if err := ol.sc.Close(); err != nil {
		ol.logger.Errorf("Ошибка при закрытии соединения с NATS Streaming: %v", err)
	}
}
