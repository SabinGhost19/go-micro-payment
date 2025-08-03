package handler

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/proto/order"
	"github.com/SabinGhost19/go-micro-payment/services/order/service"
)

type OrderHandler struct {
	svc *service.OrderService
	orderpb.UnimplementedOrderServiceServer
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
	return h.svc.CreateOrder(ctx, req)
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
	return h.svc.GetOrder(ctx, req)
}

func (h *OrderHandler) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	return h.svc.ListOrders(ctx, req)
}
