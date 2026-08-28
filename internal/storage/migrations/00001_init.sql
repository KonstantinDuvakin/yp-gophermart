-- +goose Up
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    login           TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    current_balance DOUBLE PRECISION NOT NULL DEFAULT 0,
    withdrawn       DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE TABLE orders (
    number      TEXT PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    status      TEXT NOT NULL DEFAULT 'NEW',
    accrual     DOUBLE PRECISION,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user   ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);

CREATE TABLE withdrawals (
    id           BIGSERIAL PRIMARY KEY,
    order_number TEXT NOT NULL,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    sum          DOUBLE PRECISION NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_withdrawals_user ON withdrawals(user_id);

-- +goose Down
DROP TABLE withdrawals;
DROP TABLE orders;
DROP TABLE users;
