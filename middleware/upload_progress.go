package middleware

import (
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leleo886/lopic/internal/websocket"
)

// UploadProgressMiddleware 上传进度跟踪中间件
type UploadProgressMiddleware struct {
	Hub *websocket.Hub
}

// NewUploadProgressMiddleware 创建上传进度中间件实例
func NewUploadProgressMiddleware(hub *websocket.Hub) *UploadProgressMiddleware {
	return &UploadProgressMiddleware{
		Hub: hub,
	}
}

// FileProgress 跟踪单个文件的上传进度
type FileProgress struct {
	FileName string
	FileSize int64
	Readed   int64
	UploadID string
	LastSent time.Time
}

// ProgressReader 包装了io.Reader，用于跟踪文件读取进度
type ProgressReader struct {
	r         io.ReadCloser // 改为io.ReadCloser
	Size      int64
	Readed    int64
	UserID    uint
	Hub       *websocket.Hub
	LastSent  time.Time
	TotalSent bool
}

// Read 实现io.Reader接口，读取数据并发送进度通知
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	pr.Readed += int64(n)

	// 计算总进度百分比
	progress := int(float64(pr.Readed) / float64(pr.Size) * 100)

	// 发送总进度通知（基于时间的节流：每500ms发送一次）
	if pr.Hub != nil && !pr.TotalSent {
		now := time.Now()
		timeThreshold := now.Sub(pr.LastSent) > 500*time.Millisecond
		completionThreshold := pr.Readed >= pr.Size || err == io.EOF

		if timeThreshold || completionThreshold {
			// 发送总进度
			pr.Hub.BroadcastToUser(pr.UserID, "upload_progress", map[string]interface{}{
				"upload_id": "total",
				"file_name": "Total",
				"progress":  progress,
				"read":      pr.Readed,
				"total":     pr.Size,
			})
			pr.LastSent = now

			// 标记完成
			if completionThreshold {
				pr.TotalSent = true
			}
		}
	}

	return n, err
}

// Close 实现io.Closer接口，关闭底层的reader
func (pr *ProgressReader) Close() error {
	if pr.r != nil {
		return pr.r.Close()
	}
	return nil
}

// Handle 处理上传进度跟踪
func (m *UploadProgressMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只处理multipart/form-data请求
		contentType := c.GetHeader("Content-Type")
		if contentType != "" && contentType[:19] == "multipart/form-data" {
			// 获取用户ID
			userID, exists := c.Get("user_id")
			if !exists {
				c.Next()
				return
			}

			// 获取Content-Length
			contentLength := c.GetHeader("Content-Length")
			if contentLength == "" {
				c.Next()
				return
			}

			// 解析Content-Length
			size, err := strconv.ParseInt(contentLength, 10, 64)
			if err != nil {
				c.Next()
				return
			}

			// 包装请求体，用于实时跟踪进度
			pr := &ProgressReader{
				r:        c.Request.Body,
				Size:     size,
				Readed:   0,
				UserID:   userID.(uint),
				Hub:      m.Hub,
				LastSent: time.Now(),
			}

			// 替换请求体
			c.Request.Body = pr

			// 继续处理请求
			c.Next()
		} else {
			c.Next()
		}
	}
}
