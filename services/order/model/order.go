package model

import "time"

// OrderStatus defines the possible statuses of an order
type OrderStatus string

const (
	OrderPending OrderStatus = "PENDING"
	OrderPaid    OrderStatus = "PAID"
	OrderFailed  OrderStatus = "FAILED"
)

// Order represents an order entity
type Order struct {
	ID        string      `gorm:"primaryKey;type:uuid"`
	UserID    string      `gorm:"index;type:varchar(36)"`
	Items     []OrderItem `gorm:"foreignKey:OrderID"`
	Address   string      `gorm:"type:varchar(255)"`
	Amount    float64     `gorm:"type:decimal(10,2)"`
	Status    OrderStatus `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time   `gorm:"autoCreateTime"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	OrderID   string    `gorm:"index;type:varchar(36)"`
	ProductID string    `gorm:"type:varchar(36)"`
	Quantity  int32     `gorm:"type:integer;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
