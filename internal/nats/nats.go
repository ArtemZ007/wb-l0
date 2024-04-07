package nats

import (
	"encoding/json"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

// SubscriptionConfig содержит конфигурационные параметры для подписки в NATS Streaming.
type SubscriptionConfig struct {
	DurableName string        // DurableName - постоянное имя для подписки.
	AckWait     time.Duration // AckWait - время ожидания перед подтверждением сообщения.
}

// Subscribe подписывается на тему и обрабатывает сообщения через NATS Streaming.
// Возвращает подписку и ошибку. Функция принимает экземпляр cacheService для сохранения заказов в кэш.
func Subscribe(sc stan.Conn, subject string, subCfg SubscriptionConfig, cacheService *cache.Cache) (stan.Subscription, error) {
	options := []stan.SubscriptionOption{
		stan.DurableName(subCfg.DurableName), // Устанавливаем постоянное имя для подписки, чтобы не терять сообщения при перезапуске.
		stan.AckWait(subCfg.AckWait),         // Устанавливаем время ожидания подтверждения сообщения.
	}

	// Подписываемся на тему с заданными параметрами.
	sub, err := sc.Subscribe(subject, func(msg *stan.Msg) {
		var order model.Order
		// Десериализуем сообщение в структуру Order.
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			logrus.WithError(err).Error("Ошибка десериализации заказа")
			return
		}

		// Логируем получение нового заказа.
		logrus.WithFields(logrus.Fields{
			"orderUID":     order.OrderUID,
			"trackNumber":  order.TrackNumber,
			"customerID":   order.CustomerID,
			"dateCreated":  order.DateCreated,
			"deliveryCity": order.Delivery.City,
		}).Info("Новый заказ получен")

		// Проверяем, существует ли заказ в кэше.
		if _, found := cacheService.GetOrder(order.OrderUID); !found {
			// Если заказа нет в кэше, добавляем его.
			cacheService.AddOrder(&order) // Предполагается, что метод AddOrder обновлен и теперь не возвращает ошибку.
			logrus.WithFields(logrus.Fields{"orderUID": order.OrderUID}).Info("Заказ добавлен в кэш")
		} else {
			// Если заказ уже существует в кэше, логируем это.
			logrus.WithFields(logrus.Fields{"orderUID": order.OrderUID}).Info("Заказ уже существует в кэше")
		}

		// Подтверждаем обработку сообщения.
		if err := msg.Ack(); err != nil {
			logrus.WithError(err).Error("Ошибка подтверждения сообщения")
		}
	}, options...)

	if err != nil {
		return nil, err // Возвращаем ошибку без обертки, чтобы сохранить оригинальный тип ошибки.
	}

	logrus.WithFields(logrus.Fields{
		"subject":     subject,
		"durableName": subCfg.DurableName,
	}).Info("Успешная подписка на тему")

	return sub, nil
}
