package models

import (
	"database/sql"
	"errors"
	"time"
)

// OrderSide represents the side of an order (buy or sell)
type OrderSide string

// OrderType represents the type of an order (limit or market)
type OrderType string

// OrderStatus represents the status of an order
type OrderStatus string

// Constants for order attributes
const (
	SideBuy        OrderSide   = "buy"
	SideSell       OrderSide   = "sell"
	TypeLimit      OrderType   = "limit"
	TypeMarket     OrderType   = "market"
	StatusOpen     OrderStatus = "open"
	StatusFilled   OrderStatus = "filled"
	StatusCanceled OrderStatus = "canceled"
)

// Custom errors for order operations
var (
	ErrInvalidOrder   = errors.New("invalid order parameters")
	ErrOrderNotFound  = errors.New("order not found")
	ErrOrderNotOpen   = errors.New("order is not open")
)

// Order represents a trading order
type Order struct {
	OrderID           uint64
	Symbol            string
	Side              OrderSide
	Type              OrderType
	Price             sql.NullFloat64 // Changed to sql.NullFloat64
	InitialQuantity   float64
	RemainingQuantity float64
	Status            OrderStatus
	CreatedAt         time.Time
}

// Trade represents an executed trade
type Trade struct {
	TradeID     uint64
	Symbol      string
	BuyOrderID  uint64
	SellOrderID uint64
	Price       float64
	Quantity    float64
	CreatedAt   time.Time
}

// OrderBookEntry represents orders at a specific price level
type OrderBookEntry struct {
	Price  float64
	Orders []*Order
}