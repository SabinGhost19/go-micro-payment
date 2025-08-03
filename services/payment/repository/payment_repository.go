package repository

import (
	"errors"
	"github.com/SabinGhost19/go-micro-payment/services/payment/model"
	"gorm.io/gorm"
	"time"
)

type PaymentRepository interface {
	Save(payment *model.Payment) error
	UpdateStatus(paymentID string, status model.PaymentStatus, message string) error
	FindByID(paymentID string) (*model.Payment, error)
}

type pgRepo struct {
	db *gorm.DB
}

func NewPostgresPaymentRepository(db *gorm.DB) PaymentRepository {
	return &pgRepo{db: db}
}

func (r *pgRepo) Save(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *pgRepo) UpdateStatus(paymentID string, status model.PaymentStatus, message string) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
		"message":    message,
	}).Error
}

func (r *pgRepo) FindByID(paymentID string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("id = ?", paymentID).First(&payment).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New("not found")
	}
	return &payment, err
}
