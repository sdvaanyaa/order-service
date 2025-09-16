package repository

import (
	"context"
	"github.com/sdvaanyaa/order-service/internal/models"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) error
	GetOrderByUID(ctx context.Context, uid string) (*models.Order, error)
	LoadAllOrders(ctx context.Context) (map[string]*models.Order, error)
}
