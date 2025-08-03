package repository

import (
	"context"
	"errors"
	"github.com/SabinGhost19/go-micro-payment/services/product/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	GetByID(ctx context.Context, id string) (*model.Product, error)
	List(ctx context.Context, limit, offset int) ([]*model.Product, error)
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id string) error
}

type productRepo struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *productRepo) GetByID(ctx context.Context, id string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("product not found")
	}
	return &product, err
}

func (r *productRepo) List(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	var products []*model.Product
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

func (r *productRepo) Update(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.Product
		if err := tx.Where("id = ?", p.ID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("product not found")
			}
			return err
		}
		return tx.Save(p).Error
	})
}

func (r *productRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Product{}).Error
}
