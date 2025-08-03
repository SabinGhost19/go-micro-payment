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
		"user":         "localhost:50051",
		"product":      "localhost:50052",
		"model":        "localhost:50053",
		"payment":      "localhost:50054",
		"inventory":    "localhost:50055",
		"notification": "localhost:50056",
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
