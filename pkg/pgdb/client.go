package pgdb

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sdvaanyaa/order-service/internal/config"
	"log/slog"
)

type Client struct {
	conn *pgxpool.Pool
	log  *slog.Logger
}

func New(cfg config.PostgresConfig, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}
	dsn := cfg.DSN()
	log.Info("connecting to database",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.String("database", cfg.Database),
	)
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Error("error connecting to database", slog.Any("error", err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err = conn.Ping(context.Background()); err != nil {
		log.Error("error pinging database", slog.Any("error", err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info("database connection established")
	return &Client{
		conn: conn,
		log:  log,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
	c.log.Info("database connection closed")
}
