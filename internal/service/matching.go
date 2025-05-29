package service

import (
	"database/sql"
	"orderSystem/internal/models"
	"orderSystem/internal/repository"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OrderBook manages the in-memory order book
type OrderBook struct {
	Bids  map[string][]*models.OrderBookEntry
	Asks  map[string][]*models.OrderBookEntry
	mutex sync.RWMutex
}

// NewOrderBook initializes a new order book
func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids: make(map[string][]*models.OrderBookEntry),
		Asks: make(map[string][]*models.OrderBookEntry),
	}
}

// MatchingService handles order matching logic
type MatchingService struct {
	orderBook *OrderBook
	repo      repository.Repository
	logger    *zap.Logger
}

// NewMatchingService creates a new matching service
func NewMatchingService(repo repository.Repository, logger *zap.Logger) *MatchingService {
	service := &MatchingService{
		orderBook: NewOrderBook(),
		repo:      repo,
		logger:    logger,
	}

	// Load open orders from database
	orders, err := repo.GetOrderBook("BTC-USD") // TODO: Load for all symbols
	if err != nil {
		logger.Error("Failed to load order book", zap.Error(err))
	} else {
		for _, order := range orders {
			service.addToOrderBook(order)
		}
	}

	return service
}

// PlaceOrder processes a new order and attempts to match it
func (s *MatchingService) PlaceOrder(order *models.Order) ([]*models.Trade, error) {
	s.orderBook.mutex.Lock()
	defer s.orderBook.mutex.Unlock()

	// Assign order ID and initialize fields
	order.OrderID = uint64(uuid.New().ID())
	order.Status = models.StatusOpen
	order.CreatedAt = time.Now()

	// Validate order parameters
	if order.Symbol == "" || order.InitialQuantity <= 0 {
		s.logger.Error("Invalid order parameters", zap.Any("order", order))
		return nil, models.ErrInvalidOrder
	}
	if order.Type == models.TypeLimit && (!order.Price.Valid || order.Price.Float64 <= 0) {
		s.logger.Error("Invalid price for limit order", zap.Any("order", order))
		return nil, models.ErrInvalidOrder
	}
	if order.Type == models.TypeMarket {
		order.Price = sql.NullFloat64{Valid: false} // Market orders have no price
	}

	// Begin database transaction
	tx, err := s.repo.BeginTx()
	if err != nil {
		s.logger.Error("Failed to start transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback()

	// Save order to database
	if err := s.repo.SaveOrderTx(tx, order); err != nil {
		s.logger.Error("Failed to save order", zap.Error(err))
		return nil, err
	}

	// Match order
	var trades []*models.Trade
	remainingQty := order.RemainingQuantity
	if order.Type == models.TypeMarket {
		trades, remainingQty, err = s.matchMarketOrder(tx, order)
	} else {
		trades, remainingQty, err = s.matchLimitOrder(tx, order)
	}
	if err != nil {
		s.logger.Error("Matching failed", zap.Error(err))
		return nil, err
	}

	// Update order status and quantity
	order.RemainingQuantity = remainingQty
	if order.RemainingQuantity == 0 {
		order.Status = models.StatusFilled
	} else if order.Type == models.TypeMarket {
		order.Status = models.StatusCanceled
	}
	if err := s.repo.UpdateOrderTx(tx, order); err != nil {
		s.logger.Error("Failed to update order", zap.Error(err))
		return nil, err
	}

	// Save trades
	for _, trade := range trades {
		if err := s.repo.SaveTradeTx(tx, trade); err != nil {
			s.logger.Error("Failed to save trade", zap.Error(err))
			return nil, err
		}
	}

	// Add to order book if limit order and still open
	if order.Type == models.TypeLimit && order.Status == models.StatusOpen {
		s.addToOrderBook(order)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, err
	}

	return trades, nil
}

// matchLimitOrder matches a limit order against the order book
func (s *MatchingService) matchLimitOrder(tx *sql.Tx, order *models.Order) ([]*models.Trade, float64, error) {
	var trades []*models.Trade
	remainingQty := order.RemainingQuantity
	oppositeSide := s.orderBook.Asks[order.Symbol]
	if order.Side == models.SideSell {
		oppositeSide = s.orderBook.Bids[order.Symbol]
	}

	// Sort opposite side by price (bids: descending, asks: ascending)
	sort.Slice(oppositeSide, func(i, j int) bool {
		if order.Side == models.SideSell {
			return oppositeSide[i].Price > oppositeSide[j].Price
		}
		return oppositeSide[i].Price < oppositeSide[j].Price
	})

	for _, entry := range oppositeSide {
		if remainingQty == 0 {
			break
		}
		if (order.Side == models.SideBuy && entry.Price > order.Price.Float64) ||
			(order.Side == models.SideSell && entry.Price < order.Price.Float64) {
			continue
		}

		for _, restingOrder := range entry.Orders {
			if remainingQty == 0 {
				break
			}
			matchQty := min(remainingQty, restingOrder.RemainingQuantity)
			tradePrice := restingOrder.Price.Float64
			trade := &models.Trade{
				TradeID:     uint64(uuid.New().ID()),
				Symbol:      order.Symbol,
				BuyOrderID:  order.OrderID,
				SellOrderID: restingOrder.OrderID,
				Price:       tradePrice,
				Quantity:    matchQty,
				CreatedAt:   time.Now(),
			}
			if order.Side == models.SideSell {
				trade.BuyOrderID, trade.SellOrderID = restingOrder.OrderID, order.OrderID
			}

			trades = append(trades, trade)
			remainingQty -= matchQty
			restingOrder.RemainingQuantity -= matchQty

			if restingOrder.RemainingQuantity == 0 {
				restingOrder.Status = models.StatusFilled
			}
			if err := s.repo.UpdateOrderTx(tx, restingOrder); err != nil {
				s.logger.Error("Failed to update resting order", zap.Error(err))
				return nil, 0, err
			}
			s.removeFromOrderBook(restingOrder)
		}
	}

	return trades, remainingQty, nil
}

// matchMarketOrder matches a market order against the order book
func (s *MatchingService) matchMarketOrder(tx *sql.Tx, order *models.Order) ([]*models.Trade, float64, error) {
	var trades []*models.Trade
	remainingQty := order.RemainingQuantity
	oppositeSide := s.orderBook.Asks[order.Symbol]
	if order.Side == models.SideSell {
		oppositeSide = s.orderBook.Bids[order.Symbol]
	}

	// Sort opposite side by price (bids: descending, asks: ascending)
	sort.Slice(oppositeSide, func(i, j int) bool {
		if order.Side == models.SideSell {
			return oppositeSide[i].Price > oppositeSide[j].Price
		}
		return oppositeSide[i].Price < oppositeSide[j].Price
	})

	for _, entry := range oppositeSide {
		if remainingQty == 0 {
			break
		}

		for _, restingOrder := range entry.Orders {
			if remainingQty == 0 {
				break
			}
			matchQty := min(remainingQty, restingOrder.RemainingQuantity)
			tradePrice := restingOrder.Price.Float64
			trade := &models.Trade{
				TradeID:     uint64(uuid.New().ID()),
				Symbol:      order.Symbol,
				BuyOrderID:  order.OrderID,
				SellOrderID: restingOrder.OrderID,
				Price:       tradePrice,
				Quantity:    matchQty,
				CreatedAt:   time.Now(),
			}
			if order.Side == models.SideSell {
				trade.BuyOrderID, trade.SellOrderID = restingOrder.OrderID, order.OrderID
			}

			trades = append(trades, trade)
			remainingQty -= matchQty
			restingOrder.RemainingQuantity -= matchQty

			if restingOrder.RemainingQuantity == 0 {
				restingOrder.Status = models.StatusFilled
			}
			if err := s.repo.UpdateOrderTx(tx, restingOrder); err != nil {
				s.logger.Error("Failed to update resting order", zap.Error(err))
				return nil, 0, err
			}
			s.removeFromOrderBook(restingOrder)
		}
	}

	return trades, remainingQty, nil
}

// addToOrderBook adds a limit order to the order book
func (s *MatchingService) addToOrderBook(order *models.Order) {
	side := s.orderBook.Bids
	if order.Side == models.SideSell {
		side = s.orderBook.Asks
	}

	entries, exists := side[order.Symbol]
	if !exists {
		entries = []*models.OrderBookEntry{}
	}

	for _, entry := range entries {
		if entry.Price == order.Price.Float64 {
			entry.Orders = append(entry.Orders, order)
			side[order.Symbol] = entries
			return
		}
	}

	entries = append(entries, &models.OrderBookEntry{
		Price:  order.Price.Float64,
		Orders: []*models.Order{order},
	})
	side[order.Symbol] = entries
}

// removeFromOrderBook removes an order from the order book
func (s *MatchingService) removeFromOrderBook(order *models.Order) {
	side := s.orderBook.Bids
	if order.Side == models.SideSell {
		side = s.orderBook.Asks
	}

	entries, exists := side[order.Symbol]
	if !exists {
		return
	}

	for i, entry := range entries {
		if entry.Price == order.Price.Float64 {
			for j, o := range entry.Orders {
				if o.OrderID == order.OrderID {
					entry.Orders = append(entry.Orders[:j], entry.Orders[j+1:]...)
					if len(entry.Orders) == 0 {
						entries = append(entries[:i], entries[i+1:]...)
					}
					break
				}
			}
			break
		}
	}
	side[order.Symbol] = entries
}

// CancelOrder cancels an existing order
func (s *MatchingService) CancelOrder(orderID uint64) error {
	s.orderBook.mutex.Lock()
	defer s.orderBook.mutex.Unlock()

	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		s.logger.Error("Failed to get order", zap.Error(err))
		return err
	}
	if order.Status != models.StatusOpen {
		s.logger.Warn("Attempt to cancel non-open order", zap.Uint64("order_id", orderID))
		return models.ErrOrderNotOpen
	}

	order.Status = models.StatusCanceled
	if err := s.repo.UpdateOrder(order); err != nil {
		s.logger.Error("Failed to update order status", zap.Error(err))
		return err
	}

	s.removeFromOrderBook(order)
	s.logger.Info("Order canceled", zap.Uint64("order_id", orderID))
	return nil
}

// GetOrderBook retrieves the current order book for a symbol
func (s *MatchingService) GetOrderBook(symbol string) ([]*models.Order, error) {
	orders, err := s.repo.GetOrderBook(symbol)
	if err != nil {
		s.logger.Error("Failed to get order book", zap.Error(err))
		return nil, err
	}
	return orders, nil
}

// GetTrades retrieves all trades for a symbol
func (s *MatchingService) GetTrades(symbol string) ([]*models.Trade, error) {
	trades, err := s.repo.GetTrades(symbol)
	if err != nil {
		s.logger.Error("Failed to get trades", zap.Error(err))
		return nil, err
	}
	return trades, nil
}

// GetOrder retrieves an order by ID
func (s *MatchingService) GetOrder(orderID uint64) (*models.Order, error) {
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		s.logger.Error("Failed to get order", zap.Error(err))
		return nil, err
	}
	return order, nil
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
