package interfaces

import (
	"context"
	"database/sql"

	"github.com/ArtemZ007/wb-l0/internal/domain/model"
)

type ICacheInterface interface {
	LoadOrdersFromDB(ctx context.Context, db *sql.DB) error
	GetOrder(id string) (*model.Order, bool)
	GetAllOrderIDs() []string
	DeleteOrder(id string) bool
	AddOrUpdateOrder(order *model.Order) error
}
