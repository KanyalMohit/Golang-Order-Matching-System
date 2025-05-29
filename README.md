Order Matching System
A simplified stock exchange order matching system implemented in Go with a MySQL backend.
Features

Supports limit and market orders for buy and sell sides
Price-time priority matching algorithm
RESTful API with JSON endpoints
MySQL persistence with raw SQL
Structured logging with Zap
Environment-based configuration

Prerequisites

Go 1.22 or higher
MySQL 8.0 or higher
Git

Setup

Clone the repository:
git clone <repository-url>
cd order-matching-system


Install dependencies:
go mod download


Set up environment:Create a .env file in the root directory:
DB_DSN=user:password@tcp(localhost:3306)/order_matching?parseTime=true
SERVER_ADDR=:8080


Set up MySQL:

Create a database named order_matching.
Run the schema script:mysql -u user -p order_matching < sql/schema.sql




Build and run:
go build -o order-matching-system ./cmd/server
./order-matching-system

The server runs on http://localhost:8080 (or as specified in .env).


API Endpoints
All endpoints consume/produce JSON.
Place Order

POST /orders
Request Body:{
    "symbol": "BTCUSD",
    "side": "buy",
    "type": "limit",
    "price": 50000.00,
    "quantity": 1.5
}


Response (200 OK):{
    "order_id": 123456,
    "status": "open",
    "trades": []
}


cURL:curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d '{"symbol":"BTCUSD","side":"buy","type":"limit","price":50000.00,"quantity":1.5}'



Cancel Order

DELETE /orders/{orderId}
Response (200 OK):{"message": "Order canceled"}


cURL:curl -X DELETE http://localhost:8080/orders/123456



Query Order Book

GET /orderbook?symbol={symbol}
Response (200 OK):[
    {
        "order_id": 123456,
        "symbol": "BTCUSD",
        "side": "buy",
        "type": "limit",
        "price": 50000.00,
        "initial_quantity": 1.5,
        "remaining_quantity": 1.5,
        "status": "open",
        "created_at": "2025-05-29T11:07:00Z"
    }
]


cURL:curl http://localhost:8080/orderbook?symbol=BTCUSD



List Trades

GET /trades?symbol={symbol}
Response (200 OK):[
    {
        "trade_id": 1,
        "symbol": "BTCUSD",
        "buy_order_id": 123456,
        "sell_order_id": 123457,
        "price": 50000.00,
        "quantity": 1.0,
        "created_at": "2025-05-29T11:07:00Z"
    }
]


cURL:curl http://localhost:8080/trades?symbol=BTCUSD



Get Order Status

GET /orders/{orderId}
Response (200 OK):{
    "order_id": 123456,
    "symbol": "BTCUSD",
    "side": "buy",
    "type": "limit",
    "price": 50000.00,
    "initial_quantity": 1.5,
    "remaining_quantity": 1.5,
    "status": "open",
    "created_at": "2025-05-29T11:07:00Z"
}


cURL:curl http://localhost:8080/orders/123456



Design Decisions

Package Structure: Organized into cmd, internal (with api, models, repository, service, config) for encapsulation and modularity.
Repository Pattern: Separates database logic from business logic.
Logging: Uses Zap for structured logging.
Configuration: Environment variables via .env file for flexibility.
Error Handling: Custom errors (ErrInvalidOrder, etc.) for clarity.
Validation: Stricter input validation using Gin's binding (e.g., alphanum for symbol, required_if for price).

Assumptions

Sequential order processing (single-threaded for simplicity).
Symbol names are alphanumeric and up to 10 characters.
Prices and quantities use DECIMAL(10,2) for precision.
MySQL runs locally on port 3306 unless specified in .env.

Future Improvements

Add unit tests for service and repository layers.
Implement connection pooling and retry mechanisms for database.
Optimize order book for high-performance matching.
Add authentication for API endpoints.

