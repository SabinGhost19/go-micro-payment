package model

import "time"

// Notification represents a notification entity
type Notification struct {
	ID        string    `gorm:"primaryKey"`
	UserID    string    `gorm:"index"`
	Type      string    `gorm:"type:varchar(20)"`
	Message   string    `gorm:"type:text"`
	Status    string    `gorm:"type:varchar(20)"`
	Reference string    `gorm:"type:varchar(36)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
