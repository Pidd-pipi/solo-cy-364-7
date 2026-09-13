package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/util"
)

// RequestLogger 请求日志：统一包含 request_id/method/path/status/latency_ms。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.NewString()
		c.Set("request_id", requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()
		util.GetLogger().Info(constants.LogRequestHandled,
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}
