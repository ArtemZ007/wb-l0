package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/nats-io/stan.go"
)

// OrderListener реализует слушатель сообщений NATS Streaming для заказов.
type OrderListener struct {
	sc           stan.Conn             // Соединение с NATS Streaming
	cacheService cache.ICacheInterface // Сервис кэширования
	logger       logger.ILogger        // Интерфейс логирования
	subscription stan.Subscription     // Подписка на сообщения
}

// NewOrderListener создает новый экземпляр OrderListener.
func NewOrderListener(natsURL, clusterID, clientID string, cacheService cache.ICacheInterface, logger logger.ILogger) (*OrderListener, error) {
	// Логирование процесса подключения к NATS Streaming
	logger.Info("Подключение к NATS Streaming", "URL", natsURL, "ClusterID", clusterID, "ClientID", clientID)
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Error("Не удалось подключиться к NATS Streaming", err)
		return nil, err
	}
	return &OrderListener{
		sc:           sc,
		cacheService: cacheService,
		logger:       logger,
	}, nil
}

// Start начинает прослушивание сообщений на заданную тему.
func (ol *OrderListener) Start(ctx context.Context) error {
	natsSubject := "order.created"

	// Логирование начала подписки на тему
	ol.logger.Info("Подписка на тему", "subject", natsSubject)
	var err error
	// Сохранение подписки в структуре для последующего управления
	ol.subscription, err = ol.sc.Subscribe(natsSubject, func(msg *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			ol.logger.Error("Ошибка при десериализации заказа", err)
			return
		}

		if err := ol.cacheService.AddOrUpdateOrder(&order); err != nil {
			ol.logger.Error("Ошибка при добавлении заказа в кэш", err)
			return
		}

		ol.logger.Info("Заказ добавлен в кэш", "orderUID", order.OrderUID)

		// Подтверждение обработки сообщения
		if err := msg.Ack(); err != nil {
			ol.logger.Error("Ошибка подтверждения сообщения", err)
		}
	}, stan.DurableName("order-listener-durable"), stan.SetManualAckMode(), stan.AckWait(30*time.Second))
	if err != nil {
		ol.logger.Error("Ошибка при подписке на тему", "subject", natsSubject, "error", err)
		return err
	}

	// Ожидание сигнала от контекста для остановки слушателя
	go func() {
		<-ctx.Done()
		if err := ol.Stop(); err != nil {
			ol.logger.Error("Ошибка при остановке слушателя", err)
		}
	}()

	return nil
}

// Stop останавливает слушателя и закрывает соединение с NATS Streaming.
func (ol *OrderListener) Stop() error {
	// Отписка от темы, если подписка была инициализирована
	if ol.subscription != nil {
		if err := ol.subscription.Unsubscribe(); err != nil {
			ol.logger.Error("Ошибка при отписке от темы", err)
			return err
		}
	}
	// Закрытие соединения с NATS Streaming
	if err := ol.sc.Close(); err != nil {
		ol.logger.Error("Ошибка при закрытии соединения с NATS Streaming", err)
		return err
	}
	return nil
}
