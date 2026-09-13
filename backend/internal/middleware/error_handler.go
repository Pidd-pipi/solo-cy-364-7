package middleware

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/util"
)

// ErrorHandler 统一错误响应格式：将 handler 抛出的错误转为 {code,message} JSON。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		appErr := util.AsAppError(err)
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			appErr = util.Validation(constants.MsgInvalidRequest, verr)
		}
		slog.Error("request error", "path", c.FullPath(), "method", c.Request.Method, "error", err)
		util.Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
	}
}
