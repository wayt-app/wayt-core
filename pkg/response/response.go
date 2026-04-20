package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

func BadRequest(c *gin.Context, message string, err error) {
	resp := ErrorResponse{Success: false, Message: message}
	if err != nil {
		resp.Error = err.Error()
	}
	c.JSON(http.StatusBadRequest, resp)
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{Success: false, Message: message})
}

func InternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false, Message: "internal server error", Error: err.Error(),
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{Success: false, Message: "unauthorized"})
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, ErrorResponse{Success: false, Message: "forbidden"})
}
