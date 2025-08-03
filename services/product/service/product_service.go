package service

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/configs/utils"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/product/model"
	"github.com/SabinGhost19/go-micro-payment/services/product/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

// InventoryGrpcClient defines the gRPC client interface for Inventory Service
type InventoryGrpcClient interface {
	UpdateStock(ctx context.Context, productID string, stockDelta int32) (int32, error)
}

// ProductService handles product-related business logic
type ProductService struct {
	repo          repository.ProductRepository
	kafka         *kafka.Producer
	inventoryGrpc InventoryGrpcClient
	productpb.UnimplementedProductServiceServer
}

// NewProductService creates a new ProductService
func NewProductService(repo repository.ProductRepository, kafka *kafka.Producer, inventoryGrpc InventoryGrpcClient) *ProductService {
	return &ProductService{repo: repo, kafka: kafka, inventoryGrpc: inventoryGrpc}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error) {
	if req.Name == "" || req.Price <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "name and price are required")
	}

	p := &model.Product{
		ID:          utils.GenerateUUID(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	// publish product.created event
	event := map[string]interface{}{
		"product_id": p.ID,
		"name":       p.Name,
		"price":      p.Price,
		"stock":      p.Stock,
		"status":     "created",
	}
	if err := s.kafka.SendMessage(ctx, "product-events", p.ID, event); err != nil {
		log.Printf("failed to publish product.created event: %v", err)
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

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error) {
	p, err := s.repo.GetByID(ctx, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
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

// ListProducts retrieves a paginated list of products
func (s *ProductService) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10 // default page size
	}
	offset := int(req.Page-1) * pageSize
	if offset < 0 {
		offset = 0
	}

	products, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	resp := &productpb.ListProductsResponse{
		Products: make([]*productpb.ProductResponse, len(products)),
	}
	for i, p := range products {
		resp.Products[i] = &productpb.ProductResponse{
			ProductId:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		}
	}

	return resp, nil
}

// UpdateProduct updates a product's details
func (s *ProductService) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductResponse, error) {
	p, err := s.repo.GetByID(ctx, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	// update fields if provided
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Price > 0 {
		p.Price = req.Price
	}
	if req.Stock != p.Stock {
		// call Inventory Service to update stock
		newStock, err := s.inventoryGrpc.UpdateStock(ctx, req.ProductId, req.Stock-p.Stock)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update stock: %v", err)
		}
		p.Stock = newStock
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	// publish product.updated event
	event := map[string]interface{}{
		"product_id": p.ID,
		"name":       p.Name,
		"price":      p.Price,
		"stock":      p.Stock,
		"status":     "updated",
	}
	if err := s.kafka.SendMessage(ctx, "product-events", p.ID, event); err != nil {
		log.Printf("failed to publish product.updated event: %v", err)
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

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	if err := s.repo.Delete(ctx, req.ProductId); err != nil {
		return &productpb.DeleteProductResponse{Success: false}, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	// publish product.deleted event
	event := map[string]interface{}{
		"product_id": req.ProductId,
		"status":     "deleted",
	}
	if err := s.kafka.SendMessage(ctx, "product-events", req.ProductId, event); err != nil {
		log.Printf("failed to publish product.deleted event: %v", err)
	}

	return &productpb.DeleteProductResponse{Success: true}, nil
}
