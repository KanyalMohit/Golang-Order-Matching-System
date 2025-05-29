CREATE DATABASE IF NOT EXISTS order_matching;

USE order_matching;

CREATE TABLE orders (
    order_id BIGINT UNSIGNED PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    side ENUM('buy', 'sell') NOT NULL,
    type ENUM('limit', 'market') NOT NULL,
    price DECIMAL(10,2) DEFAULT NULL,
    initial_quantity DECIMAL(10,2) NOT NULL,
    remaining_quantity DECIMAL(10,2) NOT NULL,
    status ENUM('open', 'filled', 'canceled') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_symbol_status (symbol, status),
    CHECK (initial_quantity >= 0),
    CHECK (remaining_quantity >= 0),
    CHECK (price > 0 OR price IS NULL),
    CHECK (remaining_quantity <= initial_quantity)
);

CREATE TABLE trades (
    trade_id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    symbol VARCHAR(10) NOT NULL,
    buy_order_id BIGINT UNSIGNED NOT NULL,
    sell_order_id BIGINT UNSIGNED NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (buy_order_id) REFERENCES orders(order_id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(order_id),
    CHECK (price > 0),
    CHECK (quantity > 0)
);