package ws

// Message: 客户端与服务器之间的消息
type WSMessage struct {
	// 消息类型
	Type    string  `json:"type"`     // single / group / heartbeat
	FromID  uint64  `json:"from_id"`  // 发送方ID
	ToID    uint64  `json:"to_id"`    // 接收方ID，单聊时为用户ID，群聊时为群ID
	Payload []byte  `json:"payload"`  // 消息内容
}
