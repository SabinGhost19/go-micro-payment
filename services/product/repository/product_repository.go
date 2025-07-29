package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Description string
	Price       float64
	Stock       int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, limit, offset int) ([]*Product, error)
	Update(ctx context.Context, p *Product) (*Product, error)
	Delete(ctx context.Context, id string) error
	ReturnProductById(ctx context.Context, id string) (*Product, error)
}

type productRepo struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, p *Product) error {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(p).Error
}

// helper
func (r *productRepo) ReturnProductById(ctx context.Context, id string) (*Product, error) {
	product := &Product{}
	r.db.WithContext(ctx).Where("id = ?", id).First(&product)
	if product.ID == "" {
		return nil, errors.New("product not found")
	}
	return product, nil
}
func (r *productRepo) GetByID(ctx context.Context, id string) (*Product, error) {
	prod, err := r.ReturnProductById(ctx, id)
	return prod, err
}

func (r *productRepo) List(ctx context.Context, limit, offset int) ([]*Product, error) {
	var products []*Product
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepo) Update(ctx context.Context, p *Product) (*Product, error) {
	p.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Save(p).Error
	return p, err
}

func (r *productRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Product{}, "id = ?", id).Error
}
