package handler

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/proto/payment"
	"github.com/SabinGhost19/go-micro-payment/services/payment/service"
	"time"
)

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) InitiatePayment(ctx context.Context, req *paymentpb.InitiatePaymentRequest) (*paymentpb.PaymentResponse, error) {
	p, err := h.svc.InitiatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.Currency)
	if err != nil {
		return &paymentpb.PaymentResponse{
			PaymentId: p.ID,
			OrderId:   p.OrderID,
			Status:    string(p.Status),
			Provider:  p.Provider,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
			UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
			Message:   p.Message,
		}, err
	}
	return &paymentpb.PaymentResponse{
		PaymentId: p.ID,
		OrderId:   p.OrderID,
		Status:    string(p.Status),
		Provider:  p.Provider,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		Message:   p.Message,
	}, nil
}

func (h *PaymentHandler) CheckPaymentStatus(ctx context.Context, req *paymentpb.CheckPaymentStatusRequest) (*paymentpb.PaymentResponse, error) {
	p, err := h.svc.Repo.FindByID(req.PaymentId)
	if err != nil {
		return nil, err
	}
	return &paymentpb.PaymentResponse{
		PaymentId: p.ID,
		OrderId:   p.OrderID,
		Status:    string(p.Status),
		Provider:  p.Provider,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		Message:   p.Message,
	}, nil
}
