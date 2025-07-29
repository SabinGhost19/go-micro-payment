package grpcclient

import (
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	notificationpb "github.com/SabinGhost19/go-micro-payment/proto/notification"
	orderpb "github.com/SabinGhost19/go-micro-payment/proto/order"
	paymentpb "github.com/SabinGhost19/go-micro-payment/proto/payment"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	userpb "github.com/SabinGhost19/go-micro-payment/proto/user"
	"google.golang.org/grpc"
	"log"
	"sync"
)

var (
	UserClient         userpb.UserServiceClient
	ProductClient      productpb.ProductServiceClient
	OrderClient        orderpb.OrderServiceClient
	PaymentClient      paymentpb.PaymentServiceClient
	InventoryClient    inventorypb.InventoryServiceClient
	NotificationClient notificationpb.NotificationServiceClient
)

// Inițializezi clienții gRPC (apeleză în main)
func InitGRPCClients(addrs map[string]string) {
	var wg sync.WaitGroup

	connect := func(addr string, setter func(conn *grpc.ClientConn)) {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Cannot connect to gRPC server at %s: %v", addr, err)
		}
		setter(conn)
	}
	wg.Add(6)
	go func() {
		connect(addrs["user"], func(conn *grpc.ClientConn) {
			UserClient = userpb.NewUserServiceClient(conn)
			wg.Done()
		})
		connect(addrs["product"], func(conn *grpc.ClientConn) {
			ProductClient = productpb.NewProductServiceClient(conn)
			wg.Done()
		})
		connect(addrs["order"], func(conn *grpc.ClientConn) {
			OrderClient = orderpb.NewOrderServiceClient(conn)
			wg.Done()
		})
		connect(addrs["payment"], func(conn *grpc.ClientConn) {
			PaymentClient = paymentpb.NewPaymentServiceClient(conn)
			wg.Done()
		})
		connect(addrs["inventory"], func(conn *grpc.ClientConn) {
			InventoryClient = inventorypb.NewInventoryServiceClient(conn)
			wg.Done()
		})
		connect(addrs["notification"], func(conn *grpc.ClientConn) {
			NotificationClient = notificationpb.NewNotificationServiceClient(conn)
			wg.Done()
		})
	}()
	wg.Wait()
}
