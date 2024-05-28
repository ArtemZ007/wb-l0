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

// OrderListener implements a NATS Streaming message listener for orders.
type OrderListener struct {
	sc           stan.Conn
	cacheService cache.Cache
	orderService interfaces.IOrderService
	logger       logger.ILogger
	subscription stan.Subscription
}

// NewOrderListener creates a new instance of OrderListener.
func NewOrderListener(natsURL, clusterID, clientID string, cacheService cache.Cache, orderService interfaces.IOrderService, logger logger.ILogger) (*OrderListener, error) {
	logger.Info("Connecting to NATS Streaming ", natsURL, " ", clusterID, " ", clientID)
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		logger.Error("Failed to connect to NATS Streaming", err)
		return nil, err
	}
	return &OrderListener{
		sc:           sc,
		cacheService: cacheService,
		orderService: orderService,
		logger:       logger,
	}, nil
}

// Start begins listening for messages on the specified subject.
func (ol *OrderListener) Start(ctx context.Context) error {
	natsSubject := "orders"
	var err error
	ol.subscription, err = ol.sc.Subscribe(natsSubject, ol.handleMessage, stan.DurableName("order-listener-durable"), stan.SetManualAckMode(), stan.AckWait(30*time.Second))
	if err != nil {
		ol.logger.Error("Error subscribing to subject", "subject", natsSubject, "error", err)
		return err
	}

	ol.logger.Info("Successfully subscribed to channel ", natsSubject)

	go func() {
		<-ctx.Done()
		if err := ol.Stop(); err != nil {
			ol.logger.Error("Error stopping listener", err)
		}
	}()

	return nil
}

// handleMessage processes received messages.
func (ol *OrderListener) handleMessage(msg *stan.Msg) {
	var order model.Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		ol.logger.Error("Error deserializing order", err)
		return
	}

	if err := ol.orderService.SaveOrder(context.Background(), &order); err != nil {
		ol.logger.Error("Error saving order to database", err)
		return
	}
	ol.logger.Info("Order saved to database ", order.OrderUID)

	if err := ol.cacheService.AddOrUpdateOrder(&order); err != nil {
		ol.logger.Error("Error saving order to cache", err)
		return
	}
	ol.logger.Info("Order saved to cache ", order.OrderUID)

	if err := msg.Ack(); err != nil {
		ol.logger.Error("Error acknowledging message", err)
	}
}

// Stop stops the listener and closes the connection to NATS Streaming.
func (ol *OrderListener) Stop() error {
	if ol.subscription != nil {
		if err := ol.subscription.Unsubscribe(); err != nil {
			ol.logger.Error("Error unsubscribing from subject", err)
			return err
		}
	}
	if err := ol.sc.Close(); err != nil {
		ol.logger.Error("Error closing connection to NATS Streaming", err)
		return err
	}
	return nil
}
