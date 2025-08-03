package repository

import (
	"context"
	"errors"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/model"
	"gorm.io/gorm"
	"time"
)

type InventoryRepository interface {
	CheckStock(ctx context.Context, productID string) (int32, error)
	ReserveStock(ctx context.Context, productID string, quantity int32) error
	UpdateStock(ctx context.Context, productID string, quantity int32) (int32, error)
	SyncProduct(ctx context.Context, productID, name string, stock int32) error
}

type pgRepo struct {
	db *gorm.DB
}

func NewPostgresInventoryRepository(db *gorm.DB) InventoryRepository {
	return &pgRepo{db: db}
}

func (r *pgRepo) CheckStock(ctx context.Context, productID string) (int32, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("id = ?", productID).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.New("product not found")
	}
	if err != nil {
		return 0, err
	}
	return product.Stock, nil
}

func (r *pgRepo) ReserveStock(ctx context.Context, productID string, quantity int32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("product not found")
			}
			return err
		}
		if product.Stock < quantity {
			return errors.New("insufficient stock")
		}
		return tx.Model(&model.Product{}).Where("id = ?", productID).Update("stock", product.Stock-quantity).Error
	})
}

func (r *pgRepo) UpdateStock(ctx context.Context, productID string, delta int32) (int32, error) {
	var newStock int32
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("product not found")
			}
			return err
		}
		newStock = product.Stock + delta
		if newStock < 0 {
			return errors.New("stock cannot be negative")
		}
		return tx.Model(&model.Product{}).Where("id = ?", productID).Update("stock", newStock).Error
	})
	if err != nil {
		return 0, err
	}
	return newStock, nil
}

func (r *pgRepo) SyncProduct(ctx context.Context, productID, name string, stock int32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		err := tx.Where("id = ?", productID).First(&product).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create new product
			return tx.Create(&model.Product{
				ID:        productID,
				Name:      name,
				Stock:     stock,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}).Error
		}
		if err != nil {
			return err
		}
		// update existing product
		return tx.Model(&model.Product{}).Where("id = ?", productID).Updates(map[string]interface{}{
			"name":       name,
			"stock":      stock,
			"updated_at": time.Now(),
		}).Error
	})
}
