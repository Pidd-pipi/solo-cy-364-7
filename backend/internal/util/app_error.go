package util

import (
	"errors"
	"fmt"

	"github.com/ld/storeinventory/internal/constants"
)

// 哨兵错误：仓储层与业务层统一使用 errors.Is 判断。
var (
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrValidation     = errors.New("validation failed")
	ErrRateLimited    = errors.New("rate limited")
	ErrStockNotEnough = errors.New("stock not enough")
)

// AppError 业务错误：携带业务码与 HTTP 状态码，由 error_handler 统一转 JSON。
type AppError struct {
	Code       int
	HTTPStatus int
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误。
func NewAppError(httpStatus, code int, message string, err error) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: message, Err: err}
}

func BadRequest(message string, err error) *AppError {
	return NewAppError(400, constants.CodeBadRequest, message, err)
}

func Unauthorized(message string, err error) *AppError {
	return NewAppError(401, constants.CodeUnauthorized, message, err)
}

func Forbidden(message string, err error) *AppError {
	return NewAppError(403, constants.CodeForbidden, message, err)
}

func NotFound(message string, err error) *AppError {
	return NewAppError(404, constants.CodeNotFound, message, err)
}

func Conflict(message string, err error) *AppError {
	return NewAppError(409, constants.CodeConflict, message, err)
}

func Validation(message string, err error) *AppError {
	return NewAppError(422, constants.CodeValidationError, message, err)
}

func Internal(message string, err error) *AppError {
	return NewAppError(500, constants.CodeInternalError, message, err)
}

// AsAppError 将任意 error 归一为 AppError。
func AsAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	switch {
	case errors.Is(err, ErrNotFound):
		return NotFound(constants.MsgNotFound, err)
	case errors.Is(err, ErrConflict):
		return Conflict(constants.MsgConflict, err)
	case errors.Is(err, ErrUnauthorized):
		return Unauthorized(constants.MsgUnauthorized, err)
	case errors.Is(err, ErrForbidden):
		return Forbidden(constants.MsgForbidden, err)
	case errors.Is(err, ErrValidation):
		return Validation(constants.MsgInvalidRequest, err)
	case errors.Is(err, ErrRateLimited):
		return NewAppError(429, constants.CodeRateLimited, constants.MsgRateLimited, err)
	case errors.Is(err, ErrStockNotEnough):
		return BadRequest(constants.MsgStockNotEnough, err)
	}
	return Internal(constants.MsgInternalError, err)
}
