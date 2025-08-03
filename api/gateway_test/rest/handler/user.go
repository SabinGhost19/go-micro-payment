package handler

import (
	"context"
	grpcclient "github.com/SabinGhost19/go-micro-payment/api/gateway_test/rest/grpcClient"
	"github.com/SabinGhost19/go-micro-payment/api/gateway_test/rest/helper"
	userpb "github.com/SabinGhost19/go-micro-payment/proto/user"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func RegisterUser(ctx *gin.Context) {
	req := &userpb.RegisterUserRequest{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		helper.SendError(ctx, http.StatusBadRequest, "Invalid JSON body", err.Error())
		return
	}

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := grpcclient.UserClient.RegisterUser(c, req)
	if err != nil {
		helper.HandleGrpcError(ctx, err)
		return
	}

	helper.SendSuccess(ctx, http.StatusCreated, res)
}

func AuthenticateUser(c *gin.Context) {
	var req userpb.AuthenticateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, "Invalid JSON body", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := grpcclient.UserClient.AuthenticateUser(ctx, &req)
	if err != nil {
		helper.HandleGrpcError(c, err)
		return
	}

	helper.SendSuccess(c, http.StatusOK, res)
}

func GetUser(c *gin.Context) {
	var req userpb.GetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.SendError(c, http.StatusBadRequest, "Invalid JSON body", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := grpcclient.UserClient.GetUser(ctx, &req)
	if err != nil {
		helper.HandleGrpcError(c, err)
		return
	}

	helper.SendSuccess(c, http.StatusOK, res)
}
