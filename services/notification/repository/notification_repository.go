package repository

import (
	"github.com/SabinGhost19/go-micro-payment/services/notification/model"
	"gorm.io/gorm"
	"time"
)

type NotificationRepository interface {
	Save(n *model.Notification) error
	UpdateStatus(id, status string) error
}

type pgRepo struct {
	db *gorm.DB
}

func NewPostgresNotificationRepository(db *gorm.DB) NotificationRepository {
	return &pgRepo{db: db}
}

func (r *pgRepo) Save(n *model.Notification) error {
	return r.db.Create(n).Error
}

func (r *pgRepo) UpdateStatus(id, status string) error {
	return r.db.Model(&model.Notification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}).Error
}
