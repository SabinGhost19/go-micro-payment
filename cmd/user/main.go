package main

import (
	"log"
	"net"
	"os"

	userpb "github.com/SabinGhost19/go-micro-payment/proto/user"

	"github.com/SabinGhost19/go-micro-payment/services/user/service"

	"github.com/SabinGhost19/go-micro-payment/services/user/repository"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=155015 dbname=go_microservices_payment port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect DB:", err)
	}

	db.AutoMigrate(&repository.User{})
	repo := repository.NewUserRepository(db)
	jwtSecret := os.Getenv("JWT_SECRET")

	srv := service.NewUserService(repo, jwtSecret)
	grpcServer := grpc.NewServer()

	userpb.RegisterUserServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("UserService running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}
}
