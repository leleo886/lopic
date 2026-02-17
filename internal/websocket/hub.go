package websocket

import (
	"sync"
)

// Hub 管理所有活跃的WebSocket连接
type Hub struct {
	// 注册的连接，按用户ID组织
	clients map[uint]map[*Client]bool

	// 广播消息到特定用户的通道
	broadcast chan *Message

	// 注册请求通道
	register chan *Client

	// 注销请求通道
	unregister chan *Client

	// 互斥锁，保护clients映射
	mutex sync.RWMutex
}

// Message WebSocket消息结构
type Message struct {
	UserID  uint        `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewHub 创建一个新的Hub实例
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[uint]map[*Client]bool),
	}
}

// Run 启动Hub的消息处理循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if _, ok := h.clients[client.userID]; !ok {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.userID]; ok {
				if _, ok := h.clients[client.userID][client]; ok {
					delete(h.clients[client.userID], client)
					close(client.send)
					if len(h.clients[client.userID]) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.Lock()
			if clients, ok := h.clients[message.UserID]; ok {
				for client := range clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.clients, message.UserID)
						}
					}
				}
			}
			h.mutex.Unlock()
		}
	}
}

// BroadcastToUser 向特定用户广播消息
func (h *Hub) BroadcastToUser(userID uint, msgType string, payload interface{}) {
	h.broadcast <- &Message{
		UserID:  userID,
		Type:    msgType,
		Payload: payload,
	}
}
