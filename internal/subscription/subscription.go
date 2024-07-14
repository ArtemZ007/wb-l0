package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/cache"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/nats-io/stan.go"
)

// Listener обрабатывает сообщения NATS Streaming для заказов.
type Listener struct {
	conn         stan.Conn
	cacheService *cache.Service
	orderService database.IOrderService
	log          logger.Logger
	subscription stan.Subscription
}

// NewListener создает новый экземпляр Listener.
func NewListener(natsURL, clusterID, clientID string, cacheService *cache.Service, orderService database.IOrderService, log logger.Logger) (*Listener, error) {
	log.Info("Подключение к NATS Streaming", natsURL, clusterID, clientID)
	conn, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Error("Не удалось подключиться к NATS Streaming", err)
		return nil, err
	}
	return &Listener{
		conn:         conn,
		cacheService: cacheService,
		orderService: orderService,
		log:          log,
	}, nil
}

// Start начинает прослушивание сообщений на указанной теме.
func (l *Listener) Start(ctx context.Context) error {
	subject := "orders"
	var err error
	l.subscription, err = l.conn.Subscribe(subject, l.handleMessage, stan.DurableName("order-listener-durable"), stan.SetManualAckMode(), stan.AckWait(30*time.Second))
	if err != nil {
		l.log.Error("Ошибка подписки на тему", "subject", subject, "error", err)
		return err
	}

	l.log.Info("Успешно подписан на канал", subject)

	// Ожидание завершения контекста для остановки слушателя
	go func() {
		<-ctx.Done()
		if err := l.Stop(); err != nil {
			l.log.Error("Ошибка при остановке слушателя", err)
		}
	}()

	return nil
}

// handleMessage обрабатывает полученные сообщения.
func (l *Listener) handleMessage(msg *stan.Msg) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		l.log.Error("Ошибка десериализации заказа", err)
		return
	}

	// Сохранение заказа в базе данных
	if err := l.orderService.SaveOrder(context.Background(), &order); err != nil {
		l.log.Error("Ошибка сохранения заказа в базе данных", err)
		return
	}
	l.log.Info("Заказ сохранен в базе данных", order.OrderUID)

	// Сохранение заказа в кэше
	if err := l.cacheService.AddOrUpdateOrder(&order); err != nil {
		l.log.Error("Ошибка сохранения заказа в кэше", err)
		return
	}
	l.log.Info("Заказ сохранен в кэше", order.OrderUID)

	// Подтверждение получения сообщения
	if err := msg.Ack(); err != nil {
		l.log.Error("Ошибка подтверждения сообщения: ", err)
	}
}

// Stop останавливает слушателя и закрывает соединение с NATS Streaming.
func (l *Listener) Stop() error {
	if l.subscription != nil {
		if err := l.subscription.Unsubscribe(); err != nil {
			l.log.Error("Ошибка отписки от темы: ", err)
			return err
		}
	}
	if err := l.conn.Close(); err != nil {
		l.log.Error("Ошибка закрытия соединения с NATS Streaming: ", err)
		return err
	}
	return nil
}
