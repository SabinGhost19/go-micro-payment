package routes

import (
	"github.com/SabinGhost19/go-micro-payment/api/gateway_test/rest/handler"
	"github.com/gin-gonic/gin"
)

// router config
func NewRouter() *gin.Engine {
	r := gin.Default()

	// USER endpoints
	r.POST("/users", handler.RegisterUser)
	r.GET("/users/:id", handler.GetUser)
	r.POST("user/auth", handler.AuthenticateUser)
	//
	//// PRODUCT endpoints
	//r.GET("/products", handler.ListProducts)
	//r.GET("/products/:id", handler.GetProduct)
	//
	//// ORDER endpoints
	//r.POST("/orders", handler.CreateOrder)
	//r.GET("/orders/:id", handler.GetOrder)
	//
	//// PAYMENT endpoints
	//r.POST("/payments/initiate", handler.InitiatePayment)
	//r.GET("/payments/status/:id", handler.PaymentStatus)
	//
	//// INVENTORY endpoints
	//r.GET("/inventory/:product_id", handler.GetInventory)
	//
	//// NOTIFICATION endpoints
	//r.POST("/notifications/email", handler.SendEmail)
	//r.POST("/notifications/sms", handler.SendSMS)

	return r
}
