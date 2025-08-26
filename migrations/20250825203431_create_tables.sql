-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(255),
    locale VARCHAR(10),
    internal_signature TEXT,
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INTEGER,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR(10)
);

CREATE TABLE IF NOT EXISTS deliveries (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255),
    phone VARCHAR(20),
    zip VARCHAR(10),
    city VARCHAR(255),
    address TEXT,
    region VARCHAR(255),
    email VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS payments (
    transaction VARCHAR(255) PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    request_id VARCHAR(255),
    currency VARCHAR(10),
    provider VARCHAR(50),
    amount INTEGER,
    payment_dt BIGINT,
    bank VARCHAR(50),
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT,
    track_number VARCHAR(255),
    price INTEGER,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INTEGER,
    size VARCHAR(10),
    total_price INTEGER,
    nm_id BIGINT,
    brand VARCHAR(255),
    status INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
