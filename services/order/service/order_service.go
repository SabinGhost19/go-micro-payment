package service

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/SabinGhost19/go-micro-payment/configs/utils"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	"github.com/SabinGhost19/go-micro-payment/services/order/model"
	"github.com/SabinGhost19/go-micro-payment/services/order/repository"
	"log"
	"time"
)

// PaymentGrpcClient defines the gRPC client interface for Payment Service
type PaymentGrpcClient interface {
	InitiatePayment(ctx context.Context, orderID, userID string, amount float64, currency string) (paymentID, status string, err error)
}

// OrderService handles order-related business logic
type OrderService struct {
	Repo        repository.OrderRepository
	kafka       *kafka.Producer
	paymentGrpc PaymentGrpcClient
}

// New creates a new OrderService
func New(repo repository.OrderRepository, kafka *kafka.Producer, paymentGrpc PaymentGrpcClient) *OrderService {
	return &OrderService{Repo: repo, kafka: kafka, paymentGrpc: paymentGrpc}
}

// createOrder creates a new order and initiates payment
func (s *OrderService) CreateOrder(ctx context.Context, userID string, productIDs []string, amount float64, currency string) (*model.Order, error) {
	order := &model.Order{
		ID:         utils.GenerateUUID(),
		UserID:     userID,
		ProductIDs: productIDs,
		Amount:     amount,
		Status:     model.OrderPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// save order to database
	if err := s.Repo.Save(order); err != nil {
		return nil, err
	}

	// initiate payment via gRPC
	paymentID, status, err := s.paymentGrpc.InitiatePayment(ctx, order.ID, order.UserID, order.Amount, currency)
	if err != nil {
		// update order status to FAILED if payment initiation fails
		_ = s.Repo.UpdateStatus(order.ID, model.OrderFailed)
		return order, err
	}

	// publish order.created event to Kafka
	event := map[string]interface{}{
		"order_id":   order.ID,
		"payment_id": paymentID,
		"user_id":    order.UserID,
		"amount":     order.Amount,
		"products":   order.ProductIDs,
		"status":     status,
	}
	if err := s.kafka.SendMessage(ctx, "order-events", order.ID, event); err != nil {
		log.Printf("failed to publish order.created event: %v", err)
	}

	return order, nil
}

// updateStatus updates the order status
func (s *OrderService) UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	return s.Repo.UpdateStatus(orderID, status)
}

// consumePaymentUpdates listens for payment status updates from Kafka
func (s *OrderService) ConsumePaymentUpdates(ctx context.Context) error {
	consumer, err := kafka.NewConsumer([]string{"kafka:9092"}, "order-service-group")
	if err != nil {
		return err
	}
	defer consumer.Close()

	handler := &paymentUpdateHandler{service: s}
	return consumer.Consume(ctx, []string{"payment-status-updates"}, handler)
}

// paymentUpdateHandler implements Sarama ConsumerGroupHandler for payment updates
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

		// update order status
		if err := h.service.UpdateStatus(context.Background(), event.OrderID, orderStatus); err != nil {
			log.Printf("failed to update order status: %v", err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
