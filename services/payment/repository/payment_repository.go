package repository

import (
	"errors"
	"github.com/SabinGhost19/go-micro-payment/services/payment/model"
	"gorm.io/gorm"
	"time"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	Save(payment *model.Payment) error
	UpdateStatus(paymentID string, status model.PaymentStatus, message string) error
	FindByID(paymentID string) (*model.Payment, error)
}

// pgRepo implements PaymentRepository using GORM
type pgRepo struct {
	db *gorm.DB
}

// NewPostgresPaymentRepository creates a new payment repository
func NewPostgresPaymentRepository(db *gorm.DB) PaymentRepository {
	return &pgRepo{db: db}
}

// save persists a payment to the database
func (r *pgRepo) Save(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// updateStatus updates the status and message of a payment
func (r *pgRepo) UpdateStatus(paymentID string, status model.PaymentStatus, message string) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
		"message":    message,
	}).Error
}

// findByID retrieves a payment by its ID
func (r *pgRepo) FindByID(paymentID string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("id = ?", paymentID).First(&payment).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("not found")
	}
	return &payment, err
}
