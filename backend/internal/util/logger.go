package util

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

// NewLogger 创建结构化日志器。
func NewLogger(env string) *slog.Logger {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// SetLogger 注入全局日志器，供无构造器注入能力的工具函数使用。
func SetLogger(l *slog.Logger) {
	defaultLogger = l
}

// GetLogger 获取全局日志器；未初始化时回退到标准输出 logger。
func GetLogger() *slog.Logger {
	if defaultLogger == nil {
		defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return defaultLogger
}
