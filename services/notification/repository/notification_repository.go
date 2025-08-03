package repository

import (
	"github.com/SabinGhost19/go-micro-payment/services/notification/model"
	"gorm.io/gorm"
	"time"
)

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	Save(n *model.Notification) error
	UpdateStatus(id, status string) error
}

// pgRepo implements NotificationRepository using GORM
type pgRepo struct {
	db *gorm.DB
}

// NewPostgresNotificationRepository creates a new notification repository
func NewPostgresNotificationRepository(db *gorm.DB) NotificationRepository {
	return &pgRepo{db: db}
}

// save persists a notification to the database
func (r *pgRepo) Save(n *model.Notification) error {
	return r.db.Create(n).Error
}

// updateStatus updates the status of a notification
func (r *pgRepo) UpdateStatus(id, status string) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}
