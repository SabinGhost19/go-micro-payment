package handler

import (
	"context"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/service"
)

type InventoryHandler struct {
	inventorypb.UnimplementedInventoryServiceServer
	svc *service.InventoryService
}

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) CheckStock(ctx context.Context, req *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	return h.svc.CheckStock(ctx, req)
}

func (h *InventoryHandler) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	return h.svc.ReserveStock(ctx, req)
}

func (h *InventoryHandler) UpdateStock(ctx context.Context, req *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
	return h.svc.UpdateStock(ctx, req)
}
