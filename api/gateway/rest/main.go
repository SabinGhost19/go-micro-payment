package main

import (
	"api/gateway/rest/grpcclient"
	"api/gateway/rest/handler"
	"api/gateway/rest/router"
	"log"
	"os"
)

func main() {
	// Inițializezi clienții gRPC pentru fiecare microserviciu (setați porturile după serviciile voastre)
	grpcclient.InitGRPCClients(map[string]string{
		"user":         "localhost:50051",
		"product":      "localhost:50052",
		"order":        "localhost:50053",
		"payment":      "localhost:50054",
		"inventory":    "localhost:50055",
		"notification": "localhost:50056",
	})

	// Obții routerul Gin complet routat
	r := router.NewRouter()
	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway running on port %s\n", port)
	r.Run(":" + port)
}
