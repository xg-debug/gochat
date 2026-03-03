package ws

import (
	"encoding/json"
	"strings"
	"time"

	"gochat/internal/model"
	"gochat/internal/pkg/db"
)

// 分发逻辑

type chatPayload struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

func RouteMessage(hub *Hub, msg *WSMessage) {
	switch msg.Type {
	case "single":
		if client, ok := hub.Clients[msg.ToID]; ok {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
		saveMessage(msg, 1)
	case "group":
		if client, ok := hub.Clients[msg.ToID]; ok {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
	case "heartbeat":
		if client, ok := hub.Clients[msg.FromID]; ok {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
	default:
		// logger.Info("unknown message type: %s", msg.Type)
	}
}

func saveMessage(msg *WSMessage, chatType int8) {
	if msg == nil || msg.FromID == 0 || msg.ToID == 0 {
		return
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return
	}
	content := strings.TrimSpace(string(msg.Payload))
	contentType := "text"
	var payload chatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err == nil {
		if strings.TrimSpace(payload.Content) != "" {
			content = payload.Content
		}
		if strings.TrimSpace(payload.ContentType) != "" {
			contentType = strings.TrimSpace(payload.ContentType)
		}
	}
	if content == "" {
		return
	}
	msgType := int8(1)
	switch strings.ToLower(contentType) {
	case "image":
		msgType = 2
	case "file":
		msgType = 3
	case "video":
		msgType = 4
	default:
		msgType = 1
	}
	message := model.Message{
		FromID:    int64(msg.FromID),
		ToID:      int64(msg.ToID),
		ChatType:  chatType,
		MsgType:   msgType,
		Content:   content,
		Status:    0,
		CreatedAt: time.Now(),
	}
	dbConn.Create(&message)
}
