package nats

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ArtemZ007/wb-l0/internal/domain/service"
	"github.com/ArtemZ007/wb-l0/internal/model"
	"github.com/ArtemZ007/wb-l0/pkg/logger"
	"github.com/nats-io/stan.go"
)

// ListenerConfig конфигурация для слушателя NATS Streaming.
type ListenerConfig struct {
	ClusterID string
	ClientID  string
	NatsURL   string
	Subject   string
	QGroup    string
	Durable   string
}

// OrderListener слушатель сообщений о заказах через NATS Streaming.
type OrderListener struct {
	sc             stan.Conn
	service        *service.Service
	listenerConfig ListenerConfig
}

// NewOrderListener создает новый экземпляр OrderListener для NATS Streaming.
func NewOrderListener(svc *service.Service, config ListenerConfig) (*OrderListener, error) {
	sc, err := stan.Connect(config.ClusterID, config.ClientID, stan.NatsURL(config.NatsURL))
	if err != nil {
		logger.Error("Ошибка подключения к NATS Streaming:", err)
		return nil, err
	}

	return &OrderListener{
		sc:             sc,
		service:        svc,
		listenerConfig: config,
	}, nil
}

// Start запускает слушатель сообщений в отдельной горутине.
func (ol *OrderListener) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	subscription, err := ol.sc.QueueSubscribe(ol.listenerConfig.Subject, ol.listenerConfig.QGroup, func(msg *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			logger.Error("Ошибка десериализации заказа:", err)
			return
		}

		// Передача заказа в сервис для дальнейшей обработки
		if err := ol.service.ProcessOrder(ctx, &order); err != nil {
			logger.Error("Ошибка при обработке заказа:", err)
			return
		}

		// Подтверждение обработки сообщения
		msg.Ack()

	}, stan.DurableName(ol.listenerConfig.Durable), stan.SetManualAckMode(), stan.AckWait(stan.DefaultAckWait))
	if err != nil {
		logger.Fatalf("Не удалось подписаться на тему %s: %v", ol.listenerConfig.Subject, err)
	}
	defer subscription.Unsubscribe()

	<-ctx.Done() // Ожидание сигнала на завершение
}

// Stop останавливает слушателя и закрывает соединение с NATS Streaming.
func (ol *OrderListener) Stop() {
	if err := ol.sc.Close(); err != nil {
		logger.Error("Ошибка при закрытии соединения с NATS Streaming:", err)
	}
}
