package service

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/SabinGhost19/go-micro-payment/configs/utils"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// ProductGrpcClient defines the gRPC client interface for Product Service
type ProductGrpcClient interface {
	GetProduct(ctx context.Context, productID string) (*productpb.ProductResponse, error)
}

// InventoryService handles inventory-related business logic
type InventoryService struct {
	repo        repository.InventoryRepository
	kafka       *kafka.Producer
	productGrpc ProductGrpcClient
	inventorypb.UnimplementedInventoryServiceServer
}

// NewInventoryService creates a new InventoryService
func NewInventoryService(repo repository.InventoryRepository, kafka *kafka.Producer, productGrpc ProductGrpcClient) *InventoryService {
	return &InventoryService{repo: repo, kafka: kafka, productGrpc: productGrpc}
}

// CheckStock checks the available stock for a product
func (s *InventoryService) CheckStock(ctx context.Context, req *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	// validate product existence
	if _, err := s.productGrpc.GetProduct(ctx, req.ProductId); err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	stock, err := s.repo.CheckStock(ctx, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check stock: %v", err)
	}
	return &inventorypb.CheckStockResponse{
		ProductId: req.ProductId,
		Available: stock,
	}, nil
}

// ReserveStock reserves stock for an order
func (s *InventoryService) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	for _, item := range req.Items {
		// validate product existence
		if _, err := s.productGrpc.GetProduct(ctx, item.ProductId); err != nil {
			return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
		}
		if err := s.repo.ReserveStock(ctx, item.ProductId, item.Quantity); err != nil {
			// publish stock reservation failure event
			event := map[string]interface{}{
				"order_id":   req.OrderId,
				"product_id": item.ProductId,
				"quantity":   item.Quantity,
				"status":     "failed",
				"message":    err.Error(),
			}
			if err := s.kafka.SendMessage(ctx, "stock-events", utils.GenerateUUID(), event); err != nil {
				log.Printf("failed to publish stock.reserved event: %v", err)
			}
			return &inventorypb.ReserveStockResponse{
				OrderId: req.OrderId,
				Success: false,
				Message: err.Error(),
			}, err
		}
	}

	// publish stock reservation success event
	event := map[string]interface{}{
		"order_id": req.OrderId,
		"items":    req.Items,
		"status":   "reserved",
	}
	if err := s.kafka.SendMessage(ctx, "stock-events", req.OrderId, event); err != nil {
		log.Printf("failed to publish stock.reserved event: %v", err)
	}

	return &inventorypb.ReserveStockResponse{
		OrderId: req.OrderId,
		Success: true,
		Message: "Stock reserved successfully",
	}, nil
}

// UpdateStock updates the stock level for a product
func (s *InventoryService) UpdateStock(ctx context.Context, req *inventorypb.UpdateStockRequest) (*inventorypb.UpdateStockResponse, error) {
	// validate product existence
	if _, err := s.productGrpc.GetProduct(ctx, req.ProductId); err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	newStock, err := s.repo.UpdateStock(ctx, req.ProductId, req.StockDelta)
	if err != nil {
		// publish stock update failure event
		event := map[string]interface{}{
			"product_id":  req.ProductId,
			"stock_delta": req.StockDelta,
			"status":      "failed",
			"message":     err.Error(),
		}
		if err := s.kafka.SendMessage(ctx, "stock-events", utils.GenerateUUID(), event); err != nil {
			log.Printf("failed to publish stock.updated event: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to update stock: %v", err)
	}

	// publish stock update success event
	event := map[string]interface{}{
		"product_id": req.ProductId,
		"new_stock":  newStock,
		"status":     "updated",
	}
	if err := s.kafka.SendMessage(ctx, "stock-events", req.ProductId, event); err != nil {
		log.Printf("failed to publish stock.updated event: %v", err)
	}

	return &inventorypb.UpdateStockResponse{
		ProductId: req.ProductId,
		NewStock:  newStock,
	}, nil
}

// ConsumeProductEvents listens for product events to sync inventory
func (s *InventoryService) ConsumeProductEvents(ctx context.Context) error {
	consumer, err := kafka.NewConsumer([]string{"kafka:9092"}, "inventory-service-group")
	if err != nil {
		return err
	}
	defer consumer.Close()

	handler := &productEventHandler{service: s}
	return consumer.Consume(ctx, []string{"product-events"}, handler)
}

// productEventHandler implements Sarama ConsumerGroupHandler for product events
type productEventHandler struct {
	service *InventoryService
}

// Setup is called when the consumer group session starts
func (h *productEventHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called when the consumer group session ends
func (h *productEventHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes product event messages
func (h *productEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event struct {
			ProductID string `json:"product_id"`
			Name      string `json:"name"`
			Stock     int32  `json:"stock"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("failed to unmarshal product event: %v", err)
			continue
		}

		switch event.Status {
		case "created", "updated":
			if err := h.service.repo.SyncProduct(context.Background(), event.ProductID, event.Name, event.Stock); err != nil {
				log.Printf("failed to sync product %s: %v", event.ProductID, err)
			}
		case "deleted":
			//if err := h.service.repo.Delete(context.Background(), event.ProductID); err != nil {
			//	log.Printf("failed to delete product %s: %v", event.ProductID, err)
			//}
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
