package ws

// Hub: 消息中心，统一管理所有在线连接
type Hub struct {
	// 所有注册的客户端
	Clients map[uint64]*Client

	// 广播消息到所有客户端
	Broadcast chan []byte

	// 注册新客户端
	Register chan *Client

	// 注销客户端
	Unregister chan *Client
}

// NewHub: 创建一个新的消息中心
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint64]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}
		

// Run: 启动消息中心，处理注册、注销和广播消息
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.UserID] = client
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.UserID)
				}
			}
		}
	}
}