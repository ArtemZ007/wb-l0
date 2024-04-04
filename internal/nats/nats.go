package nats

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/pkg/errors"
)

// // NATSConfig структура для конфигурации подключения к NATS
// type NATSConfig struct {
// 	// ClusterID string
// 	// ClientID  string
// 	URL string
// }

// Connect устанавливает соединение с NATS Streaming сервером
type NatsConnection struct {
	*nats.Conn
}

func Connect() (*NatsConnection, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return &NatsConnection{}, errors.Wrap(err, "unable to connect nats")
	}

	ncStruct := &NatsConnection{nc}

	return ncStruct, nil
}

func (nc *NatsConnection) Close() error {
	nc.Conn.Close()
	return nil // Add this line to return nil error
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

// // Publish публикует сообщение в тему
// func Publish(conn stan.Conn, subject, message string) error {
// 	// Публикация сообщения
// 	err := conn.Publish(subject, []byte(message))
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("Published message to subject: %s", subject)
// 	return nil
// }
