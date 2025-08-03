package model

import "time"

type Product struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"type:varchar(255)"`
	Stock     int32     `gorm:"type:integer;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
