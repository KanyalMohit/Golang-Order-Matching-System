# Order Matching System

A high-performance order matching engine built with Go, implementing a limit order book with support for limit and market orders.

## Features

- Limit and market order support
- Price-time priority matching
- Real-time order book management
- RESTful API interface
- MySQL database for persistence
- Transaction support for atomic operations
- Concurrent order processing with mutex locks

## Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- Docker (optional, for containerized deployment)

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/order-matching-system.git
cd order-matching-system
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
```bash
# Create the database
mysql -u root -p -e "CREATE DATABASE order_matching_system;"

# Run migrations
go run cmd/migrate/main.go
```

4. Configure the environment:
```bash
cp .env.example .env
# Edit .env with your database credentials and other settings
```

5. Run the server:
```bash
go run cmd/server/main.go
```

## API Endpoints

### Orders

#### Place Order
```http
POST /api/v1/orders
Content-Type: application/json

{
    "symbol": "BTC-USD",
    "side": "buy",
    "type": "limit",
    "price": 50000,
    "quantity": 1
}
```

#### Get Order
```http
GET /api/v1/orders/{order_id}
```

#### Cancel Order
```http
DELETE /api/v1/orders/{order_id}
```

### Order Book

#### Get Order Book
```http
GET /api/v1/orderbook/{symbol}
```

### Trades

#### Get Trades
```http
GET /api/v1/trades/{symbol}
```

## Order Types

### Limit Orders
- Specify both price and quantity
- Match against existing orders at the specified price or better
- Remain in the order book if not fully matched

### Market Orders
- Specify only quantity
- Match against existing limit orders at the best available price
- Execute immediately at the best available price
- Cancel if not fully matched

## Matching Rules

1. Price-Time Priority
   - Orders are matched first by price (best price first)
   - Within the same price level, orders are matched by time (first in, first out)

2. Matching Process
   - Buy orders match against the lowest ask price
   - Sell orders match against the highest bid price
   - Market orders match against the best available price
   - Partial fills are supported

## Database Schema

### Orders Table
```sql
CREATE TABLE orders (
    order_id BIGINT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    side ENUM('buy', 'sell') NOT NULL,
    type ENUM('limit', 'market') NOT NULL,
    price DECIMAL(20,8),
    initial_quantity DECIMAL(20,8) NOT NULL,
    remaining_quantity DECIMAL(20,8) NOT NULL,
    status ENUM('open', 'filled', 'canceled') NOT NULL,
    created_at TIMESTAMP NOT NULL,
    INDEX idx_symbol_status (symbol, status)
);
```

### Trades Table
```sql
CREATE TABLE trades (
    trade_id BIGINT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id BIGINT NOT NULL,
    sell_order_id BIGINT NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (buy_order_id) REFERENCES orders(order_id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(order_id),
    CHECK (price > 0),
    CHECK (quantity > 0)
);
```

## Example Usage

### Place a Limit Sell Order
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC-USD",
    "side": "sell",
    "type": "limit",
    "price": 50000,
    "quantity": 1
  }'
```

### Place a Market Buy Order
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC-USD",
    "side": "buy",
    "type": "market",
    "quantity": 1
  }'
```

### Get Order Book
```bash
curl http://localhost:8080/api/v1/orderbook/BTC-USD
```

## Error Handling

The system handles various error conditions:
- Invalid order parameters
- Insufficient liquidity
- Database errors
- Concurrent order processing conflicts

## Performance Considerations

- In-memory order book for fast matching
- Database transactions for data consistency
- Mutex locks for concurrent access
- Efficient price-time priority sorting

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

