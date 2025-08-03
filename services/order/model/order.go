package model

import (
	"time"
)

// OrderStatus defines the possible statuses of an order
type OrderStatus string

const (
	OrderPending  OrderStatus = "PENDING"
	OrderPaid     OrderStatus = "PAID"
	OrderFailed   OrderStatus = "FAILED"
	OrderShipped  OrderStatus = "SHIPPED"
	OrderComplete OrderStatus = "COMPLETE"
)

// Order represents an order entity
type Order struct {
	ID         string      `gorm:"primaryKey"`
	UserID     string      `gorm:"index"`
	ProductIDs []string    `gorm:"type:jsonb"`
	Amount     float64     `gorm:"type:decimal(10,2)"`
	Status     OrderStatus `gorm:"type:varchar(20)"`
	CreatedAt  time.Time   `gorm:"autoCreateTime"`
	UpdatedAt  time.Time   `gorm:"autoUpdateTime"`
}
