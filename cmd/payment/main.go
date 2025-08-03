package main

import (
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	"github.com/SabinGhost19/go-micro-payment/proto/payment"
	"github.com/SabinGhost19/go-micro-payment/services/payment/handler"
	"github.com/SabinGhost19/go-micro-payment/services/payment/model"
	"github.com/SabinGhost19/go-micro-payment/services/payment/repository"
	"github.com/SabinGhost19/go-micro-payment/services/payment/service"
	"github.com/stripe/stripe-go/v74"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
)

// main initializes and runs the Payment Service
func main() {
	// load environment variables
	dbDSN := os.Getenv("DB_DSN")                         // e.g., "host=postgres user=admin password=secret dbname=payments port=5432 sslmode=disable"
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")} // e.g., ["kafka:9092"]
	grpcPort := os.Getenv("PAYMENT_SERVICE_GRPC_PORT")   // e.g., ":50052"
	stripeKey := os.Getenv("STRIPE_API_KEY")             // Stripe API key

	// set Stripe API key
	stripe.Key = stripeKey

	// initialize database
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// auto-migrate schema
	if err := db.AutoMigrate(&model.Payment{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// initialize repository, service, and handler
	repo := repository.NewPostgresPaymentRepository(db)
	svc := service.New(repo, kafkaProducer)
	h := handler.NewPaymentHandler(svc)

	// start gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, h)
	log.Printf("Payment Service gRPC server running on %s", grpcPort)

	// serve gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
