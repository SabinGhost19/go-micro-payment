package helper

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

func HandleGrpcError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		SendError(c, http.StatusInternalServerError, "Unexpected server error", err.Error())
		return
	}

	switch {
	case strings.Contains(strings.ToLower(st.Message()), "not found"):
		SendError(c, http.StatusNotFound, "Resource not found", st.Message())
	case strings.Contains(strings.ToLower(st.Message()), "invalid credentials"):
		SendError(c, http.StatusUnauthorized, "Invalid credentials", st.Message())
	case strings.Contains(strings.ToLower(st.Message()), "missing required field"):
		SendError(c, http.StatusBadRequest, "Missing required field", st.Message())
	default:
		SendError(c, http.StatusInternalServerError, "Internal server error", st.Message())
	}
}
