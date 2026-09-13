package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应体 {code, message, data}。
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// OKMessage 成功响应并携带自定义文案。
func OKMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: message, Data: data})
}

// Fail 失败响应，data 为 nil。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message})
}
