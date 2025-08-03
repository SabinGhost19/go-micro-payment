package model

import "time"

type PaymentStatus string

const (
	PaymentPending PaymentStatus = "PENDING"
	PaymentPaid    PaymentStatus = "PAID"
	PaymentFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID              string        `gorm:"primaryKey"`
	OrderID         string        `gorm:"index"`
	UserID          string        `gorm:"index"`
	Amount          float64       `gorm:"type:decimal(10,2)"`
	Currency        string        `gorm:"type:varchar(3)"`
	StripeSessionID string        `gorm:"type:varchar(255)"`
	Status          PaymentStatus `gorm:"type:varchar(20)"`
	Provider        string        `gorm:"type:varchar(50)"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"`
	Message         string        `gorm:"type:text"`
}
