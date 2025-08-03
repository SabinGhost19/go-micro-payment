package main

import (
	grpcclient "github.com/SabinGhost19/go-micro-payment/api/gateway_test/rest/grpcClient"
	"github.com/SabinGhost19/go-micro-payment/api/gateway_test/rest/routes"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {

	//load env
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Println("No .env file found, reading from system environment")
	}

	//init go clients
	err := grpcclient.InitGRPCClients(map[string]string{
		"user": "localhost:50056",
	})
	if err != nil {
		log.Fatal("Error initializing grpc clients :", err)
		return
	}

	//get gin router
	r := routes.NewRouter()
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway running on port %s\n", port)
	r.Run(port)
}
