package handler

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/proto/notification"
	"github.com/SabinGhost19/go-micro-payment/services/notification/service"
)

// NotificationHandler implements the gRPC NotificationService server
type NotificationHandler struct {
	notificationpb.UnimplementedNotificationServiceServer
	svc *service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// SendEmail handles the gRPC SendEmail request
func (h *NotificationHandler) SendEmail(ctx context.Context, req *notificationpb.SendEmailRequest) (*notificationpb.NotifyResponse, error) {
	n, err := h.svc.SendEmail(ctx, req.UserId, req.To, req.Subject, req.Body, "")
	if err != nil {
		return &notificationpb.NotifyResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}
	return &notificationpb.NotifyResponse{
		Success: true,
		Message: n.Status,
	}, nil
}

// SendSMS handles the gRPC SendSMS request (placeholder)
func (h *NotificationHandler) SendSMS(ctx context.Context, req *notificationpb.SendSMSRequest) (*notificationpb.NotifyResponse, error) {
	return &notificationpb.NotifyResponse{
		Success: false,
		Message: "SMS not implemented",
	}, nil
}
