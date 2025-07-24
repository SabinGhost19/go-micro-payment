package repository

import "context"

type InventoryReposory interface {
	CheckStock(ctx context.Context, productId string) (int32, error)
	ReserveStock(ctx context.Context, productId string, quantity int32) error
	UpdateStock(ctx context.Context, productId string, quantity int32) error
}
