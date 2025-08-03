package service

import (
	"context"
	"fmt"
	"github.com/SabinGhost19/go-micro-payment/configs/utils"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	"github.com/SabinGhost19/go-micro-payment/services/payment/model"
	"github.com/SabinGhost19/go-micro-payment/services/payment/repository"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"log"
	"time"
)

// PaymentService handles payment-related business logic
type PaymentService struct {
	Repo  repository.PaymentRepository
	kafka *kafka.Producer
}

// New creates a new PaymentService
func New(repo repository.PaymentRepository, kafka *kafka.Producer) *PaymentService {
	return &PaymentService{Repo: repo, kafka: kafka}
}

// initiatePayment initiates a payment via Stripe
func (s *PaymentService) InitiatePayment(ctx context.Context, orderID, userID string, amount float64, currency string) (*model.Payment, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:    stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{Name: stripe.String("Order " + orderID)},
					UnitAmount:  stripe.Int64(int64(amount * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("https://your-app.com/payment/success"),
		CancelURL:  stripe.String("https://your-app.com/payment/cancel"),
	}

	sessionObj, err := session.New(params)
	status := model.PaymentPending
	message := "Stripe session initiated"
	if err != nil {
		status = model.PaymentFailed
		message = err.Error()
	}

	payment := &model.Payment{
		ID:              utils.GenerateUUID(),
		OrderID:         orderID,
		UserID:          userID,
		Amount:          amount,
		Currency:        currency,
		StripeSessionID: sessionObj.ID,
		Status:          status,
		Provider:        "stripe",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Message:         message,
	}

	// save payment to database
	if err := s.Repo.Save(payment); err != nil {
		return nil, fmt.Errorf("db failed: %w", err)
	}

	// publish payment.created event
	event := map[string]interface{}{
		"payment_id": payment.ID,
		"order_id":   payment.OrderID,
		"status":     payment.Status,
	}
	if err := s.kafka.SendMessage(ctx, "payment-events", payment.ID, event); err != nil {
		log.Printf("failed to publish payment.created event: %v", err)
	}

	return payment, nil
}

// updateStatus updates the payment status and publishes an event
func (s *PaymentService) UpdateStatus(ctx context.Context, paymentID string, status model.PaymentStatus, message string) error {
	if err := s.Repo.UpdateStatus(paymentID, status, message); err != nil {
		return err
	}

	// publish payment.status-updated event
	event := map[string]interface{}{
		"payment_id": paymentID,
		"status":     status,
	}
	if err := s.kafka.SendMessage(ctx, "payment-status-updates", paymentID, event); err != nil {
		log.Printf("failed to publish payment.status-updated event: %v", err)
	}
	return nil
}
