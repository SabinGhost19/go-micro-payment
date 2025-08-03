package service

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/SabinGhost19/go-micro-payment/configs/utils"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	"github.com/SabinGhost19/go-micro-payment/proto/order"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/order/model"
	"github.com/SabinGhost19/go-micro-payment/services/order/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

// InventoryGrpcClient defines the gRPC client interface for Inventory Service
type InventoryGrpcClient interface {
	CheckStock(ctx context.Context, productID string) (int32, error)
	ReserveStock(ctx context.Context, orderID string, items []inventorypb.StockItem) (bool, string, error)
}

// PaymentGrpcClient defines the gRPC client interface for Payment Service
type PaymentGrpcClient interface {
	InitiatePayment(ctx context.Context, orderID, userID string, amount float64, currency string) (paymentID, status string, err error)
}

// ProductGrpcClient defines the gRPC client interface for Product Service
type ProductGrpcClient interface {
	GetProduct(ctx context.Context, productID string) (*productpb.ProductResponse, error)
}

// OrderService handles order-related business logic
type OrderService struct {
	repo          repository.OrderRepository
	kafka         *kafka.Producer
	paymentGrpc   PaymentGrpcClient
	inventoryGrpc InventoryGrpcClient
	productGrpc   ProductGrpcClient
	orderpb.UnimplementedOrderServiceServer
}

// New creates a new OrderService
func New(repo repository.OrderRepository, kafka *kafka.Producer, paymentGrpc PaymentGrpcClient, inventoryGrpc InventoryGrpcClient, productGrpc ProductGrpcClient) *OrderService {
	return &OrderService{
		repo:          repo,
		kafka:         kafka,
		paymentGrpc:   paymentGrpc,
		inventoryGrpc: inventoryGrpc,
		productGrpc:   productGrpc,
	}
}

// CreateOrder creates a new order and initiates payment
func (s *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
	// validate input
	if req.UserId == "" || len(req.Items) == 0 || req.Address == "" || req.Currency == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id, items, address, and currency are required")
	}

	// calculate total amount and validate stock
	var totalAmount float64
	stockItems := make([]inventorypb.StockItem, len(req.Items))
	for i, item := range req.Items {
		// fetch product details
		product, err := s.productGrpc.GetProduct(ctx, item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "failed to fetch product %s: %v", item.ProductId, err)
		}
		// check stock availability
		stock, err := s.inventoryGrpc.CheckStock(ctx, item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to check stock for product %s: %v", item.ProductId, err)
		}
		if stock < item.Quantity {
			return nil, status.Errorf(codes.FailedPrecondition, "insufficient stock for product %s", item.ProductId)
		}
		totalAmount += product.Price * float64(item.Quantity)
		stockItems[i] = inventorypb.StockItem{ProductId: item.ProductId, Quantity: item.Quantity}
	}

	// create order
	items := make([]model.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = model.OrderItem{
			ID:        utils.GenerateUUID(),
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			CreatedAt: time.Now(),
		}
	}
	order := &model.Order{
		ID:        utils.GenerateUUID(),
		UserID:    req.UserId,
		Items:     items,
		Address:   req.Address,
		Amount:    totalAmount,
		Status:    model.OrderPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// reserve stock
	success, message, err := s.inventoryGrpc.ReserveStock(ctx, order.ID, stockItems)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve stock: %s", message)
	}
	if !success {
		return nil, status.Errorf(codes.FailedPrecondition, "stock reservation failed: %s", message)
	}

	// save order to database
	if err := s.repo.Save(ctx, order); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save order: %v", err)
	}

	// initiate payment via gRPC
	paymentID, statusStr, err := s.paymentGrpc.InitiatePayment(ctx, order.ID, order.UserID, order.Amount, req.Currency)
	if err != nil {
		// update order status to FAILED if payment initiation fails
		_ = s.repo.UpdateStatus(ctx, order.ID, model.OrderFailed)
		return &orderpb.OrderResponse{
			OrderId:   order.ID,
			UserId:    order.UserID,
			Items:     req.Items,
			Address:   order.Address,
			Amount:    order.Amount,
			Status:    string(order.Status),
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
			UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
		}, status.Errorf(codes.Internal, "failed to initiate payment: %v", err)
	}

	// publish order.created event to Kafka
	event := map[string]interface{}{
		"order_id":   order.ID,
		"payment_id": paymentID,
		"user_id":    order.UserID,
		"amount":     order.Amount,
		"items":      order.Items,
		"address":    order.Address,
		"status":     statusStr,
	}
	if err := s.kafka.SendMessage(ctx, "order-events", order.ID, event); err != nil {
		log.Printf("failed to publish order.created event: %v", err)
	}

	return &orderpb.OrderResponse{
		OrderId:   order.ID,
		UserId:    order.UserID,
		Items:     req.Items,
		Address:   order.Address,
		Amount:    order.Amount,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
		UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateStatus updates the order status
func (s *OrderService) UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	return s.repo.UpdateStatus(ctx, orderID, status)
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
	order, err := s.repo.FindByID(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	items := make([]*orderpb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &orderpb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return &orderpb.OrderResponse{
		OrderId:   order.ID,
		UserId:    order.UserID,
		Items:     items,
		Address:   order.Address,
		Amount:    order.Amount,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
		UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// ListOrders retrieves orders for a user with pagination
func (s *OrderService) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	orders, err := s.repo.ListByUserID(ctx, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}

	resp := &orderpb.ListOrdersResponse{
		Orders: make([]*orderpb.OrderResponse, len(orders)),
	}
	for i, order := range orders {
		items := make([]*orderpb.OrderItem, len(order.Items))
		for j, item := range order.Items {
			items[j] = &orderpb.OrderItem{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
			}
		}
		resp.Orders[i] = &orderpb.OrderResponse{
			OrderId:   order.ID,
			UserId:    order.UserID,
			Items:     items,
			Address:   order.Address,
			Amount:    order.Amount,
			Status:    string(order.Status),
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
			UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
		}
	}

	return resp, nil
}

// ConsumePaymentUpdates listens for payment and stock status updates from Kafka
func (s *OrderService) ConsumePaymentUpdates(ctx context.Context) error {
	consumer, err := kafka.NewConsumer([]string{"kafka:9092"}, "order-service-group")
	if err != nil {
		return err
	}
	defer consumer.Close()

	handler := &paymentUpdateHandler{service: s}
	return consumer.Consume(ctx, []string{"payment-status-updates", "stock-events"}, handler)
}

// paymentUpdateHandler implements Sarama ConsumerGroupHandler for payment and stock events
type paymentUpdateHandler struct {
	service *OrderService
}

// Setup is called when the consumer group session starts
func (h *paymentUpdateHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called when the consumer group session ends
func (h *paymentUpdateHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from the consumer group
func (h *paymentUpdateHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		switch msg.Topic {
		case "payment-status-updates":
			var event struct {
				PaymentID string `json:"payment_id"`
				OrderID   string `json:"order_id"`
				Status    string `json:"status"`
			}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("failed to unmarshal payment event: %v", err)
				continue
			}
			// map payment status to order status
			var orderStatus model.OrderStatus
			switch event.Status {
			case "PAID":
				orderStatus = model.OrderPaid
			case "FAILED":
				orderStatus = model.OrderFailed
			default:
				continue
			}
			if err := h.service.UpdateStatus(context.Background(), event.OrderID, orderStatus); err != nil {
				log.Printf("failed to update order status: %v", err)
			}
		case "stock-events":
			var event struct {
				OrderID string `json:"order_id"`
				Status  string `json:"status"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("failed to unmarshal stock event: %v", err)
				continue
			}
			if event.Status == "failed" {
				if err := h.service.UpdateStatus(context.Background(), event.OrderID, model.OrderFailed); err != nil {
					log.Printf("failed to update order status: %v", err)
				}
			}
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
