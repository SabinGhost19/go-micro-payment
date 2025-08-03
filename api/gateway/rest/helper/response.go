package helper

import (
	"github.com/gin-gonic/gin"
)

type StandardErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}

type StandardSuccessResponse struct {
	Data interface{} `json:"data"`
	Code int         `json:"code"`
}

func SendError(c *gin.Context, status int, errMsg string, details string) {
	c.JSON(status, StandardErrorResponse{
		Error:   errMsg,
		Details: details,
		Code:    status,
	})
}

func SendSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, StandardSuccessResponse{
		Data: data,
		Code: status,
	})
}
