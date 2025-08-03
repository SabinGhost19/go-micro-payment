package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/SabinGhost19/go-micro-payment/configs/utils"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	"github.com/SabinGhost19/go-micro-payment/services/notification/model"
	"github.com/SabinGhost19/go-micro-payment/services/notification/repository"
	orderModel "github.com/SabinGhost19/go-micro-payment/services/order/model"
	"log"
	"time"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type NotificationService struct {
	repo        repository.NotificationRepository
	emailSender EmailSender
	kafka       *kafka.Producer
}

// New creates a new NotificationService
func New(repo repository.NotificationRepository, emailSender EmailSender, kafka *kafka.Producer) *NotificationService {
	return &NotificationService{repo: repo, emailSender: emailSender, kafka: kafka}
}

// SendEmail sends an email and persists the notification
func (s *NotificationService) SendEmail(ctx context.Context, userID, to, subject, message, reference string) (*model.Notification, error) {
	status := "sent"
	err := s.emailSender.Send(to, subject, message)
	if err != nil {
		status = "failed"
	}

	n := &model.Notification{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		Type:      "email",
		Message:   message,
		Status:    status,
		Reference: reference,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// save notification to database
	if err := s.repo.Save(n); err != nil {
		return nil, err
	}

	// publish notification.sent event
	event := map[string]interface{}{
		"notification_id": n.ID,
		"type":            n.Type,
		"status":          n.Status,
		"reference":       n.Reference,
	}
	if err := s.kafka.SendMessage(ctx, "notification-events", n.ID, event); err != nil {
		log.Printf("failed to publish notification.sent event: %v", err)
	}

	return n, err
}

// ConsumeEvents listens for order and payment events to trigger notifications
func (s *NotificationService) ConsumeEvents(ctx context.Context) error {
	consumer, err := kafka.NewConsumer([]string{"kafka:9092"}, "notification-service-group")
	if err != nil {
		return err
	}
	defer consumer.Close()

	handler := &eventHandler{service: s}
	return consumer.Consume(ctx, []string{"order-events", "payment-status-updates"}, handler)
}

// eventHandler implements Sarama ConsumerGroupHandler for notification events
type eventHandler struct {
	service *NotificationService
}

// Setup is called when the consumer group session starts
func (h *eventHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called when the consumer group session ends
func (h *eventHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from the consumer group
func (h *eventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		switch msg.Topic {
		case "order-events":
			var event struct {
				OrderID string                 `json:"order_id"`
				UserID  string                 `json:"user_id"`
				Amount  float64                `json:"amount"`
				Items   []orderModel.OrderItem `json:"items"`
				Address string                 `json:"address"`
			}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("failed to unmarshal order event: %v", err)
				continue
			}
			// send order confirmation email
			subject := "Order Confirmation"
			body := fmt.Sprintf("Your order %s for $%.2f has been placed. Shipping to: %s", event.OrderID, event.Amount, event.Address)
			_, err := h.service.SendEmail(context.Background(), event.UserID, "user@example.com", subject, body, event.OrderID)
			if err != nil {
				log.Printf("failed to send order confirmation: %v", err)
			}

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
			// send payment status email
			subject := fmt.Sprintf("Payment %s", event.Status)
			body := fmt.Sprintf("Your payment for order %s is %s.", event.OrderID, event.Status)
			_, err := h.service.SendEmail(context.Background(), "", "user@example.com", subject, body, event.OrderID)
			if err != nil {
				log.Printf("failed to send payment status notification: %v", err)
			}
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
