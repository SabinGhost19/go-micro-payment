package service

import (
	"context"
	"time"

	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/product/repository"
)

type ProductService struct {
	repo repository.ProductRepository
	productpb.UnimplementedProductServiceServer
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Create product business logic
func (s *ProductService) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error) {
	p := &repository.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return &productpb.ProductResponse{
		ProductId:   p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// Exemplu GetProduct
func (s *ProductService) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error) {
	p, err := s.repo.GetByID(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &productpb.ProductResponse{
		ProductId:   p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}, nil
}
