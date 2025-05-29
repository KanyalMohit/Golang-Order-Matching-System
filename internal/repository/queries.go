package repository

const (
	SaveOrder = `INSERT INTO orders (order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at) VALUES (?,?,?,?,?,?,?,?,?)`

	UpdateOrder = `UPDATE orders SET remaining_quantity = ?, status = ? WHERE order_id = ? `

	GetOrder = `SELECT order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status,
	created_at FROM orders WHERE order_id=?`

	SaveTrade = `INSERT INTO trades (symbol, buy_order_id, sell_order_id, price, quantity, created_at) VALUES (?,?,?,?,?,?)`

	GetOrderBook = `SELECT order_id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at FROM orders WHERE symbol = ? AND status = 'open'`

	GetTrades = `SELECT trade_id, symbol, buy_order_id, sell_order_id, price, quantity, created_at FROM trades WHERE symbol = ? `
)
