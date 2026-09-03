-- +goose Up
CREATE TABLE users (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    current_balance NUMERIC(12, 2) NOT NULL DEFAULT 0,
    withdrawn       NUMERIC(12, 2) NOT NULL DEFAULT 0
);

CREATE TABLE orders (
    number      VARCHAR(255) PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    status      VARCHAR(20) NOT NULL DEFAULT 'NEW',
    accrual     NUMERIC(12, 2),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_user   ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);

CREATE TABLE withdrawals (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_number VARCHAR(255) NOT NULL,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    sum          NUMERIC(12, 2) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_withdrawals_user ON withdrawals(user_id);

-- +goose Down
DROP TABLE withdrawals;
DROP TABLE orders;
DROP TABLE users;
