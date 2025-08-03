package main

import (
	"context"
	"github.com/SabinGhost19/go-micro-payment/internal/kafka"
	"github.com/SabinGhost19/go-micro-payment/proto/notification"
	"github.com/SabinGhost19/go-micro-payment/services/notification/handler"
	"github.com/SabinGhost19/go-micro-payment/services/notification/model"
	"github.com/SabinGhost19/go-micro-payment/services/notification/repository"
	"github.com/SabinGhost19/go-micro-payment/services/notification/service"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
)

// mockEmailSender is a placeholder for an email sending implementation
type mockEmailSender struct{}

func (s *mockEmailSender) Send(to, subject, body string) error {
	log.Printf("Mock sending email to %s: %s - %s", to, subject, body)
	return nil
}

// main initializes and runs the Notification Service
func main() {
	// load environment variables
	dbDSN := os.Getenv("DB_DSN")                            // e.g., "host=postgres user=admin password=secret dbname=notifications port=5432 sslmode=disable"
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}    // e.g., ["kafka:9092"]
	grpcPort := os.Getenv("NOTIFICATION_SERVICE_GRPC_PORT") // e.g., ":50053"

	// initialize database
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// auto-migrate schema
	if err := db.AutoMigrate(&model.Notification{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(kafkaBrokers)
	if err != nil {
		log.Fatalf("failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// initialize repository, service, and handler
	repo := repository.NewPostgresNotificationRepository(db)
	emailSender := &mockEmailSender{} // replace with real implementation
	svc := service.New(repo, emailSender, kafkaProducer)
	h := handler.NewNotificationHandler(svc)

	// start gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(grpcServer, h)
	log.Printf("Notification Service gRPC server running on %s", grpcPort)

	// start Kafka consumer for order and payment events
	go func() {
		if err := svc.ConsumeEvents(context.Background()); err != nil {
			log.Fatalf("failed to start Kafka consumer: %v", err)
		}
	}()

	// serve gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
