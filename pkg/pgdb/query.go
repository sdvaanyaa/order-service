package pgdb

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
	"strings"
	"time"
)

func (c *Client) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()

	var rows pgx.Rows
	var err error

	if tx := extractTx(ctx); tx != nil {
		rows, err = tx.Query(ctx, sql, args...)
	} else {
		rows, err = c.conn.Query(ctx, sql, args...)
	}

	c.logQuery(sql, time.Since(start), err)
	return rows, err
}

func (c *Client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx := extractTx(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}

	return c.conn.QueryRow(ctx, sql, args...)
}

func (c *Client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()

	var tag pgconn.CommandTag
	var err error

	if tx := extractTx(ctx); tx != nil {
		tag, err = tx.Exec(ctx, sql, args...)
	} else {
		tag, err = c.conn.Exec(ctx, sql, args...)
	}

	c.logQuery(sql, time.Since(start), err)
	return tag, err
}

func (c *Client) logQuery(sql string, duration time.Duration, err error) {
	fields := strings.Fields(sql)
	operation := "UNKNOWN"
	if len(fields) > 0 {
		operation = strings.ToUpper(fields[0])
	}

	cleanSQL := strings.Join(fields, " ")
	attrs := []slog.Attr{
		slog.String("operation", operation),
		slog.String("sql", cleanSQL),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.LogAttrs(context.Background(), slog.LevelError, "query failed", attrs...)
		return
	}

	c.log.LogAttrs(context.Background(), slog.LevelDebug, "query succeeded", attrs...)
}
