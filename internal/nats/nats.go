package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NatsConnection структура для управления соединением с NATS.
type NatsConnection struct {
	*nats.Conn
}

// NATSConfig представляет конфигурацию для подключения к NATS Streaming.
type NATSConfig struct {
	ClusterID string // Идентификатор кластера NATS Streaming
	ClientID  string // Уникальный идентификатор клиента в рамках кластера
	URL       string // URL для подключения к NATS
}

// SubscriptionConfig представляет конфигурацию для подписки.
type SubscriptionConfig struct {
	DurableName string
	AckWait     time.Duration
}

// Connect устанавливает соединение с NATS сервером и возвращает экземпляр NatsConnection или ошибку.
func Connect(cfg NATSConfig) (*NatsConnection, error) {
	nc, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to NATS")
	}

	logrus.WithFields(logrus.Fields{
		"url": cfg.URL,
	}).Info("Successfully connected to NATS")

	return &NatsConnection{nc}, nil
}

// Close закрывает соединение с NATS сервером. Безопасно вызывается даже если соединение уже закрыто.
func (nc *NatsConnection) Close() {
	if nc.Conn != nil {
		nc.Conn.Close()
		logrus.Info("NATS connection closed")
	}
}

// ConnectToNATSStreaming создает подключение к NATS Streaming и возвращает соединение или ошибку.
func ConnectToNATSStreaming(cfg NATSConfig) (stan.Conn, error) {
	sc, err := stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(cfg.URL))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to NATS Streaming")
	}

	logrus.WithFields(logrus.Fields{
		"clusterID": cfg.ClusterID,
		"clientID":  cfg.ClientID,
		"url":       cfg.URL,
	}).Info("Successfully connected to NATS Streaming")

	return sc, nil
}

// Subscribe подписывается на тему и обрабатывает сообщения через NATS Streaming. Возвращает подписку и ошибку.
func Subscribe(sc stan.Conn, subject string, cb stan.MsgHandler, subCfg SubscriptionConfig) (stan.Subscription, error) {
	options := []stan.SubscriptionOption{
		stan.DurableName(subCfg.DurableName),
		stan.AckWait(subCfg.AckWait),
	}

	sub, err := sc.Subscribe(subject, cb, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to subscribe to subject")
	}
	logrus.WithFields(logrus.Fields{
		"subject":     subject,
		"durableName": subCfg.DurableName,
	}).Info("Subscribed to subject with durable name")

	return sub, nil
}
