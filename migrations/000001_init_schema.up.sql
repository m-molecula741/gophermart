/*
 * SQL dialect: PostgreSQL
 */

-- +goose Up
-- +goose StatementBegin
-- language=PostgreSQL

-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы заказов
CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(255) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL,
    accrual DECIMAL(10, 2) NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_status CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED'))
);

-- Создание таблицы балансов
CREATE TABLE IF NOT EXISTS balances (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    current DECIMAL(10, 2) NOT NULL DEFAULT 0,
    withdrawn DECIMAL(10, 2) NOT NULL DEFAULT 0,
    CHECK (current >= 0)
);

-- Создание таблицы списаний
CREATE TABLE IF NOT EXISTS withdrawals (
    order_number VARCHAR(255) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    sum DECIMAL(10, 2) NOT NULL CHECK (sum > 0),
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_users_login ON users(login); 