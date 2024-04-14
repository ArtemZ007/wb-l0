// Package subscription Пакет subscription предоставляет функциональность для прослушивания сообщений NATS Streaming
// и их соответствующей обработки. Включает реализацию слушателя,
// который может подписываться на определенные темы и обрабатывать входящие сообщения,
// десериализуя их в объекты заказов и сохраняя их в базу данных.
package subscription

import (
	"context"
	"encoding/json"
	"github.com/ArtemZ007/wb-l0/internal/domain/model"
	"github.com/ArtemZ007/wb-l0/internal/repository/database"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

// Listener описывает слушателя сообщений из NATS Streaming.
type Listener struct {
	orderRepo *database.Service // Использование конкретного типа для репозитория
	logger    *logrus.Logger    // Использование *logrus.Logger для логгирования
}

// NewListener создает новый экземпляр Listener.
// Принимает репозиторий для работы с заказами и логгер.
func NewListener(orderRepo *database.Service, logger *logrus.Logger) *Listener {
	return &Listener{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

// Start запускает слушателя для прослушивания сообщений из NATS Streaming.
// Параметры: ctx для контроля завершения, url, clusterID, clientID, subject для подключения к NATS.
func (l *Listener) Start(ctx context.Context, url, clusterID, clientID, subject string) error {
	l.logger.Infof("Запуск слушателя NATS Streaming. URL: %s, ClusterID: %s, ClientID: %s, Subject: %s", url, clusterID, clientID, subject)

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url))
	if err != nil {
		l.logger.WithError(err).Error("Ошибка подключения к NATS Streaming.")
		return err
	}
	defer func() {
		if err := sc.Close(); err != nil {
			l.logger.WithError(err).Error("Ошибка при закрытии соединения с NATS Streaming.")
		}
	}()

	sub, err := sc.Subscribe(subject, func(msg *stan.Msg) {
		l.handleMessage(ctx, msg) // Передаем контекст в обработчик сообщений
	}, stan.DurableName("my-durable"))
	if err != nil {
		l.logger.WithError(err).Error("Ошибка подписки на NATS Streaming.")
		return err
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			l.logger.WithError(err).Error("Ошибка при отписке от NATS Streaming.")
		}
	}()

	<-ctx.Done()
	l.logger.Info("Остановка слушателя NATS Streaming.")
	return nil
}

// handleMessage обрабатывает полученные сообщения, десериализуя их в объекты заказов и сохраняя их в базу данных.
func (l *Listener) handleMessage(ctx context.Context, msg *stan.Msg) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		l.logger.WithError(err).Error("Ошибка десериализации сообщения.")
		return
	}

	if err := l.orderRepo.SaveOrder(ctx, &order); err != nil {
		l.logger.WithError(err).WithField("orderUID", order.OrderUID).Error("Ошибка сохранения заказа в базу данных.")
		return
	}

	l.logger.WithField("orderUID", order.OrderUID).Info("Заказ успешно сохранен в базу данных.")
}

// Stop останавливает слушателя и освобождает ресурсы.
func (l *Listener) Stop() {
	// Здесь должна быть реализация остановки слушателя, если это необходимо.
	l.logger.Info("Слушатель NATS Streaming остановлен.")
}
