package service

import (
	"context"
	_ "google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/status"
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

func (s *ProductService) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductResponse, error) {

	product := &repository.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	p, err := s.repo.Update(ctx, product)

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

func (s *ProductService) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {

}
