// 业务错误码（与 backend/internal/constants/error_codes.go 对应）
export const ErrorCode = {
  OK: 0,
  BAD_REQUEST: 40000,
  UNAUTHORIZED: 40100,
  FORBIDDEN: 40300,
  NOT_FOUND: 40400,
  CONFLICT: 40900,
  RATE_LIMITED: 42900,
  VALIDATION_ERROR: 42200,
  INTERNAL_ERROR: 50000
} as const

export const ERROR_MESSAGES: Record<number, string> = {
  [ErrorCode.UNAUTHORIZED]: '未登录或登录已过期',
  [ErrorCode.FORBIDDEN]: '无权限执行该操作',
  [ErrorCode.NOT_FOUND]: '资源不存在',
  [ErrorCode.CONFLICT]: '资源状态冲突',
  [ErrorCode.RATE_LIMITED]: '请求过于频繁，请稍后再试',
  [ErrorCode.VALIDATION_ERROR]: '请求参数不合法'
}
