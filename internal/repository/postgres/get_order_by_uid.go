package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/sdvaanyaa/order-service/internal/models"
	"github.com/sdvaanyaa/order-service/internal/repository"
)

func (r *OrderRepo) GetOrderByUID(ctx context.Context, uid string) (*models.Order, error) {
	order := &models.Order{OrderUID: uid}

	if err := r.getOrder(ctx, uid, order); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrOrderNotFound
		}
		return nil, err
	}

	if err := r.getDelivery(ctx, uid, order); err != nil {
		return nil, err
	}

	if err := r.getPayment(ctx, uid, order); err != nil {
		return nil, err
	}

	if err := r.getItems(ctx, uid, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (r *OrderRepo) getOrder(ctx context.Context, uid string, order *models.Order) error {
	query := `
		SELECT track_number, entry, locale, internal_signature, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`

	return r.db.QueryRow(ctx, query, uid).Scan(
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
}

func (r *OrderRepo) getDelivery(ctx context.Context, uid string, order *models.Order) error {
	query := `
		SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid = $1
	`

	return r.db.QueryRow(ctx, query, uid).Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
}

func (r *OrderRepo) getPayment(ctx context.Context, uid string, order *models.Order) error {
	query := `
		SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments WHERE order_uid = $1
	`

	return r.db.QueryRow(ctx, query, uid).Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDt,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
}

func (r *OrderRepo) getItems(ctx context.Context, uid string, order *models.Order) error {
	query := `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`

	rows, err := r.db.Query(ctx, query, uid)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item

		err = rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return err
		}

		order.Items = append(order.Items, item)
	}

	return nil
}
