package repository

import (
	"errors"
	"github.com/SabinGhost19/go-micro-payment/services/order/model"
	"gorm.io/gorm"
	"time"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Save(order *model.Order) error
	UpdateStatus(orderID string, status model.OrderStatus) error
	FindByID(orderID string) (*model.Order, error)
}

// pgRepo implements OrderRepository using GORM
type pgRepo struct {
	db *gorm.DB
}

// NewPostgresOrderRepository creates a new order repository
func NewPostgresOrderRepository(db *gorm.DB) OrderRepository {
	return &pgRepo{db: db}
}

// save persists an order to the database
func (r *pgRepo) Save(order *model.Order) error {
	return r.db.Create(order).Error
}

// updateStatus updates the status of an order
func (r *pgRepo) UpdateStatus(orderID string, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}

// findByID retrieves an order by its ID
func (r *pgRepo) FindByID(orderID string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("id = ?", orderID).First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("not found")
	}
	return &order, err
}
