package main

import (
	grpcclient "github.com/SabinGhost19/go-micro-payment/api/gateway/rest/grpcClient"
	"github.com/SabinGhost19/go-micro-payment/api/gateway/rest/routes"
	"log"
	"os"
)

func main() {
	//init go clients
	grpcclient.InitGRPCClients(map[string]string{
		"user":         "localhost:50056",
		"product":      "localhost:50055",
		"order":        "localhost:50051",
		"payment":      "localhost:50052",
		"inventory":    "localhost:50054",
		"notification": "localhost:50053",
	})
	
	//get gin router
	r := routes.NewRouter()
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway running on port %s\n", port)
	r.Run(":" + port)
}
