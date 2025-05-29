package api

import (
	"database/sql"
	"net/http"
	"orderSystem/internal/models"
	"orderSystem/internal/service"

	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler manages API endpoints
type Handler struct {
	service *service.MatchingService
	logger  *zap.Logger
}

// NewHandler creates a new API handler
func NewHandler(s *service.MatchingService, logger *zap.Logger) *Handler {
	return &Handler{service: s, logger: logger}
}

// SetupRoutes configures API routes
func SetupRoutes(router *gin.Engine, h *Handler) {
	router.POST("/orders", h.placeOrder)
	router.DELETE("/orders/:orderId", h.cancelOrder)
	router.GET("/orderbook", h.getOrderBook)
	router.GET("/trades", h.getTrades)
	router.GET("/orders/:orderId", h.getOrder)
}

// placeOrder handles POST /orders
func (h *Handler) placeOrder(c *gin.Context) {
	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	price := sql.NullFloat64{Valid: false}
	if req.Type == models.TypeLimit {
		price = sql.NullFloat64{Float64: req.Price, Valid: true}
	}

	order := &models.Order{
		Symbol:            req.Symbol,
		Side:              req.Side,
		Type:              req.Type,
		Price:             price,
		InitialQuantity:   req.Quantity,
		RemainingQuantity: req.Quantity,
	}

	trades, err := h.service.PlaceOrder(order)
	if err != nil {
		h.logger.Error("Failed to place order", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, PlaceOrderResponse{
		OrderID: order.OrderID,
		Status:  order.Status,
		Trades:  trades,
	})
}

// cancelOrder handles DELETE /orders/:orderId
func (h *Handler) cancelOrder(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid order ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	if err := h.service.CancelOrder(orderID); err != nil {
		h.logger.Error("Failed to cancel order", zap.Error(err))
		if err == models.ErrOrderNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		} else if err == models.ErrOrderNotOpen {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Order is not open"})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order canceled"})
}

// getOrderBook handles GET /orderbook?symbol={symbol}
func (h *Handler) getOrderBook(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		h.logger.Warn("Missing symbol parameter")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Symbol is required"})
		return
	}

	orders, err := h.service.GetOrderBook(symbol)
	if err != nil {
		h.logger.Error("Failed to get order book", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// getTrades handles GET /trades?symbol={symbol}
func (h *Handler) getTrades(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		h.logger.Warn("Missing symbol parameter")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Symbol is required"})
		return
	}

	trades, err := h.service.GetTrades(symbol)
	if err != nil {
		h.logger.Error("Failed to get trades", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// getOrder handles GET /orders/:orderId
func (h *Handler) getOrder(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid order ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order ID"})
		return
	}

	order, err := h.service.GetOrder(orderID)
	if err == models.ErrOrderNotFound {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Order not found"})
		return
	} else if err != nil {
		h.logger.Error("Failed to get order", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}
