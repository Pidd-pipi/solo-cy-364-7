package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/config"
	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/util"
)

type bucket struct {
	count    int
	windowAt time.Time
}

// RateLimiter 基于内存滑动窗口的简单限流器，按客户端 IP 维度限流。
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	limit   int
	window  time.Duration
}

// NewRateLimiter 构造限流器。
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*bucket), limit: limit, window: window}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok || now.Sub(b.windowAt) > rl.window {
		rl.buckets[key] = &bucket{count: 1, windowAt: now}
		return true
	}
	b.count++
	return b.count <= rl.limit
}

// RateLimit 启用限流的中间件。
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	rl := NewRateLimiter(cfg.RateLimit, time.Duration(cfg.RateWindowSec)*time.Second)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			util.GetLogger().Warn(constants.LogRateLimited, "ip", c.ClientIP(), "path", c.FullPath())
			util.Fail(c, 429, constants.CodeRateLimited, constants.MsgRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitStrict 更严格限流（登录/导入等敏感接口使用）。
func RateLimitStrict(cfg *config.Config) gin.HandlerFunc {
	rl := NewRateLimiter(cfg.RateLimit/4+1, time.Duration(cfg.RateWindowSec)*time.Second)
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			util.GetLogger().Warn(constants.LogRateLimited, "ip", c.ClientIP(), "path", c.FullPath())
			util.Fail(c, 429, constants.CodeRateLimited, constants.MsgRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}
