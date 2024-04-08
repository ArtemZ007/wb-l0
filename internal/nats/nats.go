package nats

import (
	"encoding/json"
	"time"

	"github.com/ArtemZ007/wb-l0/internal/cache"
	"github.com/ArtemZ007/wb-l0/internal/model"
	"github.com/ArtemZ007/wb-l0/internal/validator"
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
		stan.DurableName(subCfg.DurableName),
		stan.AckWait(subCfg.AckWait),
	}

	messageHandler := func(msg *stan.Msg) {
		handleMessage(msg, cacheService)
	}

	sub, err := sc.Subscribe(subject, messageHandler, options...)
	if err != nil {
		logrus.WithError(err).Error("Ошибка при подписке на тему")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"subject":     subject,
		"durableName": subCfg.DurableName,
	}).Info("Успешная подписка на тему")

	return sub, nil
}

// handleMessage обрабатывает полученное сообщение из NATS Streaming.
func handleMessage(msg *stan.Msg, cacheService *cache.Cache) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		logrus.WithError(err).Error("Ошибка десериализации заказа")
		return
	}

	// Валидация полученного заказа
	if validationErrors := validator.ValidateOrder(order); len(validationErrors) > 0 {
		logrus.WithFields(logrus.Fields{
			"orderUID": order.OrderUID,
			"errors":   validationErrors,
		}).Error("Ошибка валидации заказа")
		return
	}

	// Получение объекта sql.DB
	db := database.GetDB() // Это пример. Вам нужно будет заменить его на вашу реализацию.

	if err := cacheService.AddOrder(db, order.OrderUID, &order); err != nil {
		logrus.WithFields(logrus.Fields{"orderUID": order.OrderUID, "error": err}).Error("Ошибка при добавлении заказа в кэш")
	} else {
		logrus.WithFields(logrus.Fields{"orderUID": order.OrderUID}).Info("Заказ успешно добавлен в кэш")
	}

	if err := msg.Ack(); err != nil {
		logrus.WithError(err).Error("Ошибка подтверждения сообщения")
	}
}
