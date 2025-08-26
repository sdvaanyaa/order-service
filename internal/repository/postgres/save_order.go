package postgres

import (
	"context"
	"github.com/sdvaanyaa/order-service.git/internal/models"
)

func (r *OrderRepo) SaveOrder(ctx context.Context, order *models.Order) error {
	if err := r.insertOrder(ctx, order); err != nil {
		return err
	}

	if err := r.insertDelivery(ctx, order); err != nil {
		return err
	}

	if err := r.insertPayment(ctx, order); err != nil {
		return err
	}

	if err := r.insertItems(ctx, order); err != nil {
		return err
	}

	return nil
}

func (r *OrderRepo) insertOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature,
		                    customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)

	return err
}

func (r *OrderRepo) insertDelivery(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)

	return err
}

func (r *OrderRepo) insertPayment(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO payments (transaction, order_uid, request_id, currency, provider,
		                      amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		order.Payment.Transaction,
		order.OrderUID,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)

	return err
}

func (r *OrderRepo) insertItems(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO items (order_uid, chrt_id, track_number, price,
		                   rid, name, sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	for _, item := range order.Items {
		_, err := r.db.Exec(
			ctx,
			query,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
