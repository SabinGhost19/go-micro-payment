package main

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	orderpb "github.com/SabinGhost19/go-micro-payment/proto/order"
	paymentpb "github.com/SabinGhost19/go-micro-payment/proto/payment"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/order/handler"
	"github.com/SabinGhost19/go-micro-payment/services/order/model"
	"github.com/SabinGhost19/go-micro-payment/services/order/repository"
	"github.com/SabinGhost19/go-micro-payment/services/order/service"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
)

// paymentGrpcClient implements the PaymentGrpcClient interface
type paymentGrpcClient struct {
	client paymentpb.PaymentServiceClient
}

// InitiatePayment calls the Payment Service's gRPC endpoint
func (c *paymentGrpcClient) InitiatePayment(ctx context.Context, orderID, userID string, amount float64, currency string) (string, string, error) {
	resp, err := c.client.InitiatePayment(ctx, &paymentpb.InitiatePaymentRequest{
		OrderId:  orderID,
		UserId:   userID,
		Amount:   amount,
		Currency: currency,
	})
	if err != nil {
		return "", "", err
	}
	return resp.PaymentId, resp.Status, nil
}

// inventoryGrpcClient implements the InventoryGrpcClient interface
type inventoryGrpcClient struct {
	client inventorypb.InventoryServiceClient
}

// CheckStock calls the Inventory Service's gRPC endpoint
func (c *inventoryGrpcClient) CheckStock(ctx context.Context, productID string) (int32, error) {
	resp, err := c.client.CheckStock(ctx, &inventorypb.CheckStockRequest{ProductId: productID})
	if err != nil {
		return 0, err
	}
	return resp.Available, nil
}

// ReserveStock calls the Inventory Service's gRPC endpoint
func (c *inventoryGrpcClient) ReserveStock(ctx context.Context, orderID string, items []inventorypb.StockItem) (bool, string, error) {

	// convert []inventorypb.StockItem -> []*inventorypb.StockItem
	pbItems := make([]*inventorypb.StockItem, len(items))
	for i := range items {
		pbItems[i] = &items[i]
	}

	resp, err := c.client.ReserveStock(ctx, &inventorypb.ReserveStockRequest{
		OrderId: orderID,
		Items:   pbItems,
	})
	if err != nil {
		return false, "", err
	}
	return resp.Success, resp.Message, nil
}

// productGrpcClient implements the ProductGrpcClient interface
type productGrpcClient struct {
	client productpb.ProductServiceClient
}

// GetProduct calls the Product Service's gRPC endpoint
func (c *productGrpcClient) GetProduct(ctx context.Context, productID string) (*productpb.ProductResponse, error) {
	return c.client.GetProduct(ctx, &productpb.GetProductRequest{ProductId: productID})
}

// main initializes and runs the Order Service
func main() {
	// load environment variables
	dbDSN := os.Getenv("DB_DSN")                                // e.g., "host=postgres user=admin password=secret dbname=orders port=5432 sslmode=disable"
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}        // e.g., ["kafka:9092"]
	grpcPort := os.Getenv("ORDER_SERVICE_GRPC_PORT")            // e.g., ":50051"
	paymentServiceAddr := os.Getenv("PAYMENT_SERVICE_ADDR")     // e.g., "payment-service:50052"
	inventoryServiceAddr := os.Getenv("INVENTORY_SERVICE_ADDR") // e.g., "inventory-service:50054"
	productServiceAddr := os.Getenv("PRODUCT_SERVICE_ADDR")     // e.g., "product-service:50055"

	// initialize database
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// auto-migrate schema
	if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// initialize gRPC client for Payment Service
	conn, err := grpc.Dial(paymentServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to Payment Service: %v", err)
	}
	defer conn.Close()
	paymentClient := &paymentGrpcClient{client: paymentpb.NewPaymentServiceClient(conn)}

	// initialize gRPC client for Inventory Service
	conn, err = grpc.Dial(inventoryServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to Inventory Service: %v", err)
	}
	defer conn.Close()
	inventoryClient := &inventoryGrpcClient{client: inventorypb.NewInventoryServiceClient(conn)}

	// initialize gRPC client for Product Service
	conn, err = grpc.Dial(productServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to Product Service: %v", err)
	}
	defer conn.Close()
	productClient := &productGrpcClient{client: productpb.NewProductServiceClient(conn)}

	// initialize repository, service, and handler
	repo := repository.NewPostgresOrderRepository(db)
	svc := service.New(repo, kafkaProducer, paymentClient, inventoryClient, productClient)
	h := handler.NewOrderHandler(svc)

	// start gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, h)
	log.Printf("Order Service gRPC server running on %s", grpcPort)

	// start Kafka consumer for payment and stock updates
	go func() {
		if err := svc.ConsumePaymentUpdates(context.Background()); err != nil {
			log.Fatalf("failed to start Kafka consumer: %v", err)
		}
	}()

	// serve gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
