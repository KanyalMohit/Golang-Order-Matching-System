package api

import "orderSystem/internal/models"

// PlaceOrderRequest defines the request body for placing an order
type PlaceOrderRequest struct {
	Symbol   string           `json:"symbol" binding:"required,alphanum,max=10"`
	Side     models.OrderSide `json:"side" binding:"required,oneof=buy sell"`
	Type     models.OrderType `json:"type" binding:"required,oneof=limit market"`
	Price    float64          `json:"price" binding:"required_if=Type limit"`
	Quantity float64          `json:"quantity" binding:"required,gt=0"`
}

// PlaceOrderResponse defines the response for placing an order
type PlaceOrderResponse struct {
	OrderID uint64             `json:"order_id"`
	Status  models.OrderStatus `json:"status"`
	Trades  []*models.Trade    `json:"trades"`
}

// ErrorResponse defines an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
