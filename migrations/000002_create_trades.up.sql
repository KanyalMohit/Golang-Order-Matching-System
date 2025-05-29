-- +migrate Up
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