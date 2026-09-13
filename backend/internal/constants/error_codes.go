package constants

// 业务错误码，前后端共享定义（frontend/src/constants/errorCodes.ts 对应实现）。
const (
	CodeOK              = 0
	CodeBadRequest      = 40000
	CodeUnauthorized    = 40100
	CodeForbidden       = 40300
	CodeNotFound        = 40400
	CodeConflict        = 40900
	CodeRateLimited     = 42900
	CodeValidationError = 42200
	CodeInternalError   = 50000
)
