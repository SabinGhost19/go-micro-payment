package handler

import (
	"context"

	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/service"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
	inventorypb.UnimplementedInventoryServiceServer
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (c *InventoryHandler) CheckStock(ctx context.Context, in *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	return c.inventoryService.CheckStock(ctx, in)
}

func (*InventoryHandler) ReserveStock(context.Context, *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
}

func (*InventoryHandler) UpdateStock(context.Context, *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
}
