package middleware

import (
	"sync"
	"time"
	"github.com/gin-gonic/gin"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Limit      int           // 时间窗口内允许的最大请求数
	WindowSize time.Duration // 时间窗口大小
	KeyFunc    func(*gin.Context) string // 生成速率限制键的函数
}

// RateLimiter 速率限制器
type RateLimiter struct {
	mu      sync.Mutex
	windows map[string][]time.Time
	config  *RateLimitConfig
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		windows: make(map[string][]time.Time),
		config:  config,
	}
}

// RateLimit 速率限制中间件
func RateLimit(config *RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)
	
	// 启动清理过期数据的协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 记录日志但不中断程序，防止cleanup() 发生 panic 会导致整个协程崩溃
				log.Errorf("RateLimiter cleanup panicked: %v", r)
			}
		}()
		for {
			time.Sleep(time.Minute)
			limiter.cleanup()
		}
	}()
	
	return func(c *gin.Context) {
		key := config.KeyFunc(c)
		if key == "" {
			key = c.ClientIP() // 默认使用客户端IP作为键
		}
		
		if !limiter.Allow(key) {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrTooManyRequests)
			c.JSON(statusCode, errorResponse)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	
	// 清理过期的时间戳
	window := rl.windows[key]
	valid := make([]time.Time, 0)
	for _, t := range window {
		if now.Sub(t) < rl.config.WindowSize {
			valid = append(valid, t)
		}
	}
	
	// 检查是否超过限制
	if len(valid) >= rl.config.Limit {
		rl.windows[key] = valid
		return false
	}
	
	// 添加当前时间戳
	valid = append(valid, now)
	rl.windows[key] = valid
	return true
}

// cleanup 清理过期的窗口数据
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	for key, window := range rl.windows {
		valid := make([]time.Time, 0)
		for _, t := range window {
			if now.Sub(t) < rl.config.WindowSize {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.windows, key)
		} else {
			rl.windows[key] = valid
		}
	}
}
