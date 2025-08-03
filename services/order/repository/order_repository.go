package repository

import (
	"context"
	"errors"
	"github.com/SabinGhost19/go-micro-payment/services/order/model"
	"gorm.io/gorm"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Save(ctx context.Context, order *model.Order) error
	UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error
	FindByID(ctx context.Context, orderID string) (*model.Order, error)
	ListByUserID(ctx context.Context, userID string, page, pageSize int32) ([]*model.Order, error)
}

// pgRepo implements OrderRepository using GORM
type pgRepo struct {
	db *gorm.DB
}

// NewPostgresOrderRepository creates a new order repository
func NewPostgresOrderRepository(db *gorm.DB) OrderRepository {
	return &pgRepo{db: db}
}

// Save persists an order and its items to the database
func (r *pgRepo) Save(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, item := range order.Items {
			item.OrderID = order.ID
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateStatus updates the status of an order
func (r *pgRepo) UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// FindByID retrieves an order by its ID
func (r *pgRepo) FindByID(ctx context.Context, orderID string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", orderID).First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("order not found")
	}
	return &order, err
}

// ListByUserID retrieves orders for a user with pagination
func (r *pgRepo) ListByUserID(ctx context.Context, userID string, page, pageSize int32) ([]*model.Order, error) {
	var orders []*model.Order
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).Limit(int(pageSize)).Offset(int(offset)).Find(&orders).Error
	return orders, err
}
