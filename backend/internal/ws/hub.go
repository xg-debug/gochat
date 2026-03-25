package ws

import (
	"encoding/json"
	"sync"

	"gorm.io/gorm"
)

// Hub: 消息中心，统一管理所有在线连接
type Hub struct {
	// 所有注册的客户端
	Clients map[uint64]*Client
	mu      sync.RWMutex
	db      *gorm.DB

	// 广播消息到所有客户端
	Broadcast chan []byte

	// 注册新客户端
	Register chan *Client

	// 注销客户端
	Unregister chan *Client
}

// NewHub: 创建一个新的消息中心
func NewHub(db *gorm.DB) *Hub {
	hub := &Hub{
		Clients:    make(map[uint64]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		db:         db,
	}
	setDefaultHub(hub)
	return hub
}

// Run: 启动消息中心，处理注册、注销和广播消息
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.broadcastPresenceLocked(client.UserID, true)
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				h.broadcastPresenceLocked(client.UserID, false)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.UserID)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) broadcastPresenceLocked(userID uint64, online bool) {
	payload := map[string]interface{}{
		"content":     "",
		"contentType": "text",
		"extra": map[string]interface{}{
			"online": online,
		},
	}
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	msg := WSMessage{
		Type:    "presence",
		FromID:  userID,
		ToID:    0,
		Payload: payloadRaw,
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return
	}
	for _, client := range h.Clients {
		select {
		case client.Send <- raw:
		default:
		}
	}
}

func (h *Hub) isOnline(userID uint64) bool {
	h.mu.RLock()
	_, ok := h.Clients[userID]
	h.mu.RUnlock()
	return ok
}

func (h *Hub) getClient(userID uint64) *Client {
	h.mu.RLock()
	client := h.Clients[userID]
	h.mu.RUnlock()
	return client
}
