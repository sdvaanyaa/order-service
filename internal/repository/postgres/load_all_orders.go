package postgres

import (
	"context"
	"github.com/sdvaanyaa/order-service.git/internal/models"
)

func (r *OrderRepo) LoadAllOrders(ctx context.Context) (map[string]*models.Order, error) {
	cache := make(map[string]*models.Order)

	rows, err := r.db.Query(ctx, `SELECT order_uid FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid string

		err = rows.Scan(&uid)
		if err != nil {
			return nil, err
		}

		order, err := r.GetOrderByUID(ctx, uid)
		if err != nil {
			return nil, err
		}

		cache[uid] = order
	}

	return cache, nil
}
