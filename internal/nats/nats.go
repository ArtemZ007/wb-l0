package nats

import (
	"log"

	"github.com/nats-io/stan.go"
)

// NATSConfig структура для конфигурации подключения к NATS
type NATSConfig struct {
	ClusterID string
	ClientID  string
	URL       string
}

// Connect устанавливает соединение с NATS Streaming сервером
func Connect(cfg *NATSConfig) (stan.Conn, error) {
	conn, err := stan.Connect(cfg.ClusterID, cfg.ClientID, stan.NatsURL(cfg.URL))
	if err != nil {
		return nil, err
	}
	log.Println("Connected to NATS Streaming server")
	return conn, nil
}

// Subscribe подписывается на тему и обрабатывает сообщения
func Subscribe(conn stan.Conn, subject string, cb stan.MsgHandler) (stan.Subscription, error) {
	// Подписка на тему с использованием обработчика сообщений
	sub, err := conn.Subscribe(subject, cb, stan.DurableName("my-durable"))
	if err != nil {
		return nil, err
	}
	log.Printf("Subscribed to subject: %s", subject)
	return sub, nil
}

// Publish публикует сообщение в тему
func Publish(conn stan.Conn, subject, message string) error {
	// Публикация сообщения
	err := conn.Publish(subject, []byte(message))
	if err != nil {
		return err
	}
	log.Printf("Published message to subject: %s", subject)
	return nil
}
