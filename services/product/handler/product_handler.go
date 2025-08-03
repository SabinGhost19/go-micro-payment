package handler

import (
	"context"

	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/product/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	productService *service.ProductService
	productpb.UnimplementedProductServiceServer
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error) {
	// Putem adăuga validări minimale aici, ex:
	if req.Name == "" || req.Price <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Name and Price are required")
	}
	return h.productService.CreateProduct(ctx, req)
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error) {
	if req.ProductId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ProductId is required")
	}
	return h.productService.GetProduct(ctx, req)
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductResponse, error) {
	if req.ProductId == "" || req.Price <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "ProductId and Price are required")
	}

	return h.productService.UpdateProduct(ctx, req)
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.DeleteProductResponse, error) {
	if req.ProductId == "" {
		return &productpb.DeleteProductResponse{Success: false}, status.Errorf(codes.InvalidArgument, "ProductId is required")
	}
	return h.productService.DeleteProduct(ctx, req)
}
