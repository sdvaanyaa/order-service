package postgres

import (
	"github.com/sdvaanyaa/order-service/internal/repository"
	"github.com/sdvaanyaa/order-service/pkg/pgdb"
	"log/slog"
)

type OrderRepo struct {
	db  *pgdb.Client
	log *slog.Logger
}

func New(db *pgdb.Client, log *slog.Logger) repository.OrderRepository {
	return &OrderRepo{
		db:  db,
		log: log,
	}
}
