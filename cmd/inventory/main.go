package main

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/handler"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/model"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/repository"
	"github.com/SabinGhost19/go-micro-payment/services/inventory/service"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
)

// productGrpcClient implements the ProductGrpcClient interface
type productGrpcClient struct {
	client productpb.ProductServiceClient
}

// GetProduct calls the Product Service's gRPC endpoint
func (c *productGrpcClient) GetProduct(ctx context.Context, productID string) (*productpb.ProductResponse, error) {
	return c.client.GetProduct(ctx, &productpb.GetProductRequest{ProductId: productID})
}

// main initializes and runs the Inventory Service
func main() {
	// load environment variables
	dbDSN := os.Getenv("DB_DSN")                            // e.g., "host=postgres user=admin password=secret dbname=inventory port=5432 sslmode=disable"
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}    // e.g., ["kafka:9092"]
	grpcPort := os.Getenv("INVENTORY_SERVICE_GRPC_PORT")    // e.g., ":50054"
	productServiceAddr := os.Getenv("PRODUCT_SERVICE_ADDR") // e.g., "product-service:50055"

	// initialize database
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// auto-migrate schema
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// initialize gRPC client for Product Service
	conn, err := grpc.Dial(productServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to Product Service: %v", err)
	}
	defer conn.Close()
	productClient := &productGrpcClient{client: productpb.NewProductServiceClient(conn)}

	// initialize repository, service, and handler
	repo := repository.NewPostgresInventoryRepository(db)
	svc := service.NewInventoryService(repo, kafkaProducer, productClient)
	h := handler.NewInventoryHandler(svc)

	// start gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, h)
	log.Printf("Inventory Service gRPC server running on %s", grpcPort)

	// start Kafka consumer for product events
	go func() {
		if err := svc.ConsumeProductEvents(context.Background()); err != nil {
			log.Fatalf("failed to start Kafka consumer: %v", err)
		}
	}()

	// serve gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
