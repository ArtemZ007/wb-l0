package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/interfaces"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/nats-io/stan.go"
)

// OrderListener реализует слушатель сообщений NATS Streaming для заказов.
type OrderListener struct {
	sc           stan.Conn
	cacheService cache.Cache
	orderService interfaces.IOrderService // Изменено на IOrderService
	logger       logger.ILogger
	subscription stan.Subscription
}

// NewOrderListener создает новый экземпляр OrderListener.
func NewOrderListener(natsURL, clusterID, clientID string, cacheService cache.Cache, orderService interfaces.IOrderService, logger *logger.Logger) (*OrderListener, error) {
	// Логирование процесса подключения к NATS Streaming
	logger.Info("Подключение к NATS Streaming ", natsURL, " ", clusterID, " ", clientID, " ")
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Error("Не удалось подключиться к NATS Streaming", err)
		return nil, err
	}
	return &OrderListener{
		sc:           sc,
		cacheService: cacheService,
		orderService: orderService, // Сохраняем переданный orderService
		logger:       logger,
	}, nil
}

// Start Ваши методы Start и Stop останутся без изменений, за исключением использования orderService для работы с заказами.
// Start Измененный конструктор для включения orderService
// Измененный метод Start для включения логики сохранения в базу данных
// Start начинает прослушивание сообщений на заданную тему.
func (ol *OrderListener) Start(ctx context.Context) error {
	// Your existing code...
	natsSubject := "orders"
	var err error
	ol.subscription, err = ol.sc.Subscribe(natsSubject, func(msg *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			ol.logger.Error("Ошибка при десериализации заказа", err)
			return
		}

		// Попытка сохранить заказ с использованием сервиса заказов
		if err := ol.orderService.SaveOrder(ctx, &order); err != nil {
			ol.logger.Error("Ошибка при сохранении заказа в базу данных", err)
			return
		}
		// Логирование успешного сохранения заказа в базу данных
		ol.logger.Info("Заказ сохранен в базу данных ", order.OrderUID)

		// Попытка сохранить заказ в кэш
		if err := ol.cacheService.AddOrUpdateOrder(&order); err != nil {
			ol.logger.Error("Ошибка при сохранении заказа в кэш", err)
			return
		}

		// Логирование успешного сохранения заказа в кэш
		ol.logger.Info("Заказ сохранен в кэш ", order.OrderUID)

		if err := msg.Ack(); err != nil {
			ol.logger.Error("Ошибка подтверждения сообщения", err)
		}
	}, stan.DurableName("order-listener-durable"), stan.SetManualAckMode(), stan.AckWait(30*time.Second))

	if err != nil {
		ol.logger.Error("Ошибка при подписке на тему", "subject", natsSubject, "error", err)
		return err
	}

	// Логирование успешной подписки на канал
	ol.logger.Info("Успешно подписан на канал ", natsSubject)

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
