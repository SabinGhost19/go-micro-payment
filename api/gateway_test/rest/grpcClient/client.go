package grpcclient

import (
	"fmt"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	notificationpb "github.com/SabinGhost19/go-micro-payment/proto/notification"
	orderpb "github.com/SabinGhost19/go-micro-payment/proto/order"
	paymentpb "github.com/SabinGhost19/go-micro-payment/proto/payment"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	userpb "github.com/SabinGhost19/go-micro-payment/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func InitGRPCClients(addrs map[string]string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 6)

	connect := func(addr string, setter func(conn *grpc.ClientConn)) {
		conn, err := grpc.Dial(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			errCh <- fmt.Errorf("connect %s: %w", addr, err)
			return
		}
		setter(conn)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		connect(addrs["user"], func(conn *grpc.ClientConn) { UserClient = userpb.NewUserServiceClient(conn) })
	}()
	//go func() {
	//	defer wg.Done()
	//	connect(addrs["product"], func(conn *grpc.ClientConn) { ProductClient = productpb.NewProductServiceClient(conn) })
	//}()
	//go func() {
	//	defer wg.Done()
	//	connect(addrs["order"], func(conn *grpc.ClientConn) { OrderClient = orderpb.NewOrderServiceClient(conn) })
	//}()
	//go func() {
	//	defer wg.Done()
	//	connect(addrs["payment"], func(conn *grpc.ClientConn) { PaymentClient = paymentpb.NewPaymentServiceClient(conn) })
	//}()
	//go func() {
	//	defer wg.Done()
	//	connect(addrs["inventory"], func(conn *grpc.ClientConn) { InventoryClient = inventorypb.NewInventoryServiceClient(conn) })
	//}()
	//go func() {
	//	defer wg.Done()
	//	connect(addrs["notification"], func(conn *grpc.ClientConn) { NotificationClient = notificationpb.NewNotificationServiceClient(conn) })
	//}()

	wg.Wait()
	close(errCh)

	if len(errCh) > 0 {
		var first error
		for e := range errCh {
			if first == nil {
				first = e
			}
			log.Println("grpc dial error:", e)
		}
		return first
	}
	return nil
}
