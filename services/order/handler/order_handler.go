package handler

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/proto/order"
	"github.com/SabinGhost19/go-micro-payment/services/order/service"
	"time"
)

// OrderHandler implements the gRPC OrderService server
type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	svc *service.OrderService
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// CreateOrder handles the gRPC CreateOrder request
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
	// extract product IDs from order items
	var productIDs []string
	for _, item := range req.Items {
		for i := 0; i < int(item.Quantity); i++ {
			productIDs = append(productIDs, item.ProductId)
		}
	}

	// calculate total amount (assuming item prices are fetched elsewhere or passed)
	amount := 0.0 // placeholder: in a real system, fetch product prices
	for _, item := range req.Items {
		// assume price is fetched or passed; for simplicity, use a dummy value
		amount += float64(item.Quantity) * 10.0 // replace with actual price lookup
	}

	// create order
	order, err := h.svc.CreateOrder(ctx, req.UserId, productIDs, amount, "USD")
	if err != nil {
		return nil, err
	}

	// convert to gRPC response
	return &orderpb.OrderResponse{
		OrderId:     order.ID,
		UserId:      order.UserID,
		Items:       req.Items, // simplified; map to OrderItem if needed
		Status:      string(order.Status),
		TotalAmount: order.Amount,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetOrder retrieves an order by ID
func (h *OrderHandler) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
	order, err := h.svc.Repo.FindByID(req.OrderId)
	if err != nil {
		return nil, err
	}

	// convert to gRPC response
	items := make([]*orderpb.OrderItem, 0, len(order.ProductIDs))
	for _, pid := range order.ProductIDs {
		items = append(items, &orderpb.OrderItem{ProductId: pid, Quantity: 1}) // simplified
	}

	return &orderpb.OrderResponse{
		OrderId:     order.ID,
		UserId:      order.UserID,
		Items:       items,
		Status:      string(order.Status),
		TotalAmount: order.Amount,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ListOrders lists orders for a user (simplified)
func (h *OrderHandler) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	// implement pagination logic with GORM if needed
	return &orderpb.ListOrdersResponse{}, nil // placeholder
}
