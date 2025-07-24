package service

import (
	"context"

	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/repository"
)

type InventoryService struct {
	inventorypb.UnimplementedInventoryServiceServer
	inventoryRepository repository.InventoryReposory
}

func NewInvetoryService() *InventoryService {
	return &InventoryService{}
}

func (c *InventoryService) CheckStock(ctx context.Context, in *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error, string) {
	stock, err := c.inventoryRepository.CheckStock(ctx, in.ProductId)
	if err != nil {
		return nil, err, "Failed to check stock"
	}

	if stock == 0 {
		return &inventorypb.CheckStockResponse{ProductId: in.ProductId, Available: 0}, nil, "Product is out of stock"
	}

	return &inventorypb.CheckStockResponse{
		ProductId: in.ProductId,
		Available: stock,
	}, nil, "Product is available"
}

func (c *InventoryService) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
}

func (c *InventoryService) UpdateStock(ctx context.Context, req *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
}
