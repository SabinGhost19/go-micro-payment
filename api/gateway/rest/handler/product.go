package handler

import (
	"context"
	grpcclient "github.com/SabinGhost19/go-micro-payment/api/gateway/rest/grpcClient"
	productpb "github.com/SabinGhost19/go-micro-payment/proto/product"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ListProducts(c *gin.Context) {
	req := &productpb.ListProductsRequest{}
	resp, err := grpcclient.ProductClient.ListProducts(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Products)
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")
	req := &productpb.GetProductRequest{ProductId: id}
	resp, err := grpcclient.ProductClient.GetProduct(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Product)
}
