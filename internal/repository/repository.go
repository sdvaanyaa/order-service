package repository

import (
	"context"
	"errors"
	"github.com/sdvaanyaa/order-service/internal/models"
)

var (
	ErrOrderNotFound = errors.New("timestamp not found")
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) error
	GetOrderByUID(ctx context.Context, uid string) (*models.Order, error)
	LoadAllOrders(ctx context.Context) (map[string]*models.Order, error)
}
