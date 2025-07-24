package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	proto "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/SabinGhost19/go-micro-payment/services/product/handler"
	"github.com/SabinGhost19/go-micro-payment/services/product/repository"
	"github.com/SabinGhost19/go-micro-payment/services/product/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=productdb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Automigrate Product model
	err = db.AutoMigrate(&repository.Product{})
	if err != nil {
		log.Fatalf("failed to migrate product model: %v", err)
	}

	productRepo := repository.NewProductRepository(db)
	productSrv := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productSrv)

	grpcServer := grpc.NewServer()
	proto.RegisterProductServiceServer(grpcServer, productHandler)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Product Service started on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve grpc server: %v", err)
	}
}
