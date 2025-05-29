-- +migrate Up
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