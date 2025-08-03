package main

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	inventorypb "github.com/SabinGhost19/go-micro-payment/proto/inventory"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/product/handler"
	"github.com/SabinGhost19/go-micro-payment/services/product/model"
	"github.com/SabinGhost19/go-micro-payment/services/product/repository"
	"github.com/SabinGhost19/go-micro-payment/services/product/service"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
)

// inventoryGrpcClient implements the InventoryGrpcClient interface
type inventoryGrpcClient struct {
	client inventorypb.InventoryServiceClient
}

// UpdateStock calls the Inventory Service's gRPC endpoint
func (c *inventoryGrpcClient) UpdateStock(ctx context.Context, productID string, stockDelta int32) (int32, error) {
	resp, err := c.client.UpdateStock(ctx, &inventorypb.UpdateStockRequest{
		ProductId:  productID,
		StockDelta: stockDelta,
	})
	if err != nil {
		return 0, err
	}
	return resp.NewStock, nil
}

// main initializes and runs the Product Service
func main() {
	// load environment variables
	dbDSN := os.Getenv("DB_DSN")                                // e.g., "host=postgres user=admin password=secret dbname=products port=5432 sslmode=disable"
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}        // e.g., ["kafka:9092"]
	grpcPort := os.Getenv("PRODUCT_SERVICE_GRPC_PORT")          // e.g., ":50055"
	inventoryServiceAddr := os.Getenv("INVENTORY_SERVICE_ADDR") // e.g., "inventory-service:50054"

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

	// initialize gRPC client for Inventory Service
	conn, err := grpc.Dial(inventoryServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to Inventory Service: %v", err)
	}
	defer conn.Close()
	inventoryClient := &inventoryGrpcClient{client: inventorypb.NewInventoryServiceClient(conn)}

	// initialize repository, service, and handler
	repo := repository.NewProductRepository(db)
	svc := service.NewProductService(repo, kafkaProducer, inventoryClient)
	h := handler.NewProductHandler(svc)

	// start gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	productpb.RegisterProductServiceServer(grpcServer, h)
	log.Printf("Product Service gRPC server running on %s", grpcPort)

	// serve gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
