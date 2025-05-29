package repository

import (
	"database/sql"
	"orderSystem/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

// Repository defines database operations for the order matching system
type Repository interface {
	SaveOrder(order *models.Order) error
	UpdateOrder(order *models.Order) error
	GetOrder(orderID uint64) (*models.Order, error)
	SaveTrade(trade *models.Trade) error
	GetOrderBook(symbol string) ([]*models.Order, error)
	GetTrades(symbol string) ([]*models.Trade, error)
	BeginTx() (*sql.Tx, error)
	SaveOrderTx(tx *sql.Tx, order *models.Order) error
	UpdateOrderTx(tx *sql.Tx, order *models.Order) error
	SaveTradeTx(tx *sql.Tx, trade *models.Trade) error
}

// MySQLRepository implements Repository using MySQL
type MySQLRepository struct {
	db *sql.DB
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

// BeginTx starts a new transaction
func (r *MySQLRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// SaveOrder persists a new order to the database
func (r *MySQLRepository) SaveOrder(order *models.Order) error {
	query := `
		INSERT INTO orders (order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, order.OrderID, order.Symbol, order.Side, order.Type, order.Price,
		order.InitialQuantity, order.RemainingQuantity, order.Status, order.CreatedAt)
	return err
}

// SaveOrderTx persists a new order to the database within a transaction
func (r *MySQLRepository) SaveOrderTx(tx *sql.Tx, order *models.Order) error {
	query := `
		INSERT INTO orders (order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, order.OrderID, order.Symbol, order.Side, order.Type, order.Price,
		order.InitialQuantity, order.RemainingQuantity, order.Status, order.CreatedAt)
	return err
}

// UpdateOrder updates an existing order in the database
func (r *MySQLRepository) UpdateOrder(order *models.Order) error {
	query := `
		UPDATE orders
		SET remaining_quantity = ?, status = ?
		WHERE order_id = ?`
	_, err := r.db.Exec(query, order.RemainingQuantity, order.Status, order.OrderID)
	return err
}

// UpdateOrderTx updates an existing order in the database within a transaction
func (r *MySQLRepository) UpdateOrderTx(tx *sql.Tx, order *models.Order) error {
	query := `
		UPDATE orders
		SET remaining_quantity = ?, status = ?
		WHERE order_id = ?`
	_, err := tx.Exec(query, order.RemainingQuantity, order.Status, order.OrderID)
	return err
}

// GetOrder retrieves an order by its ID
func (r *MySQLRepository) GetOrder(orderID uint64) (*models.Order, error) {
	query := `
		SELECT order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at
		FROM orders
		WHERE order_id = ?`
	row := r.db.QueryRow(query, orderID)
	order := &models.Order{}
	var price sql.NullFloat64
	err := row.Scan(&order.OrderID, &order.Symbol, &order.Side, &order.Type, &price,
		&order.InitialQuantity, &order.RemainingQuantity, &order.Status, &order.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	order.Price = price
	return order, nil
}

// SaveTrade persists a trade to the database
func (r *MySQLRepository) SaveTrade(trade *models.Trade) error {
	query := `
		INSERT INTO trades (symbol, buy_order_id, sell_order_id, price, quantity, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, trade.Symbol, trade.BuyOrderID, trade.SellOrderID, trade.Price,
		trade.Quantity, trade.CreatedAt)
	return err
}

// SaveTradeTx persists a trade to the database within a transaction
func (r *MySQLRepository) SaveTradeTx(tx *sql.Tx, trade *models.Trade) error {
	query := `
		INSERT INTO trades (symbol, buy_order_id, sell_order_id, price, quantity, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, trade.Symbol, trade.BuyOrderID, trade.SellOrderID, trade.Price,
		trade.Quantity, trade.CreatedAt)
	return err
}

// GetOrderBook retrieves all open orders for a given symbol
func (r *MySQLRepository) GetOrderBook(symbol string) ([]*models.Order, error) {
	query := `
		SELECT order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at
		FROM orders
		WHERE symbol = ? AND status = 'open'`
	rows, err := r.db.Query(query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		order := &models.Order{}
		var price sql.NullFloat64
		if err := rows.Scan(&order.OrderID, &order.Symbol, &order.Side, &order.Type, &price,
			&order.InitialQuantity, &order.RemainingQuantity, &order.Status, &order.CreatedAt); err != nil {
			return nil, err
		}
		order.Price = price
		orders = append(orders, order)
	}
	return orders, nil
}

// GetTrades retrieves all trades for a given symbol
func (r *MySQLRepository) GetTrades(symbol string) ([]*models.Trade, error) {
	query := `
		SELECT trade_id, symbol, buy_order_id, sell_order_id, price, quantity, created_at
		FROM trades
		WHERE symbol = ?`
	rows, err := r.db.Query(query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []*models.Trade
	for rows.Next() {
		trade := &models.Trade{}
		if err := rows.Scan(&trade.TradeID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID,
			&trade.Price, &trade.Quantity, &trade.CreatedAt); err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}
	return trades, nil
}
