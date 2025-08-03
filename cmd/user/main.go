package main

import (
	"log"
	"net"
	"os"

	userpb "github.com/SabinGhost19/go-micro-payment/proto/user"
	"github.com/SabinGhost19/go-micro-payment/services/user/service"
	"github.com/joho/godotenv"

	"github.com/SabinGhost19/go-micro-payment/services/user/repository"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	//load env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, reading from system environment")
	}

	dsn := "host=localhost user=postgres password=155015 dbname=go_microservices_payment port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect DB:", err)
	}

	err = db.AutoMigrate(&repository.User{})
	if err != nil {
		return
	}

	repo := repository.NewUserRepository(db)
	//------------
	jwtSecret := os.Getenv("JWT_SECRET")
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET env variable not set")
	}
	if userServiceAddr == "" {
		log.Fatal("USER_SERVICE_ADDR env variable not set")
	}
	//----------------

	srv := service.NewUserService(repo, jwtSecret)
	grpcServer := grpc.NewServer()

	userpb.RegisterUserServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", userServiceAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("UserService running on " + userServiceAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}
}
