package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
)

const (
	// 写入消息的等待时间
	writeWait = 10 * time.Second

	// 读取消息的pong等待时间
	pongWait = 60 * time.Second

	// 发送ping消息的间隔时间，必须小于pongWait
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 表示一个WebSocket连接客户端
type Client struct {
	// Hub实例
	hub *Hub

	// WebSocket连接
	conn *websocket.Conn

	// 发送消息的通道
	send chan *Message

	// 用户ID
	userID uint
}

// readPump 从WebSocket连接读取消息并处理
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("websocket read error: %v", err)
			}
			break
		}
		// 这里可以处理客户端发送的消息，目前我们只需要服务器发送消息
	}
}

// writePump 向WebSocket连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// 序列化消息为JSON
			jsonData, err := json.Marshal(message)
			if err != nil {
				log.Errorf("failed to marshal websocket message: %v", err)
				return
			}

			w.Write(jsonData)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleWebSocket 处理WebSocket连接请求
func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			statusCode, errorResponse := cerrors.NewErrorResponse(cerrors.ErrUnauthorized)
			c.JSON(statusCode, errorResponse)
			return
		}

		// 升级HTTP连接为WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Errorf("failed to upgrade connection: %v", err)
			return
		}

		// 创建新的客户端
		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan *Message, 256),
			userID: userID.(uint),
		}

		// 注册客户端
		client.hub.register <- client

		// 启动goroutine处理读写操作
		go client.writePump()
		go client.readPump()
	}
}
