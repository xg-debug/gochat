package ws

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gochat/internal/model"
	"gochat/internal/pkg/db"
)

// 分发逻辑

type chatPayload struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	TempID      string `json:"tempId"`
	Extra       map[string]interface{} `json:"extra"`
}

func RouteMessage(hub *Hub, msg *WSMessage) {
	switch msg.Type {
	case "single":
		msg.Payload = enrichPayloadWithSender(msg.FromID, msg.Payload)
		if client := hub.getClient(msg.ToID); client != nil {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
		saveMessageAndAck(hub, msg, 1)
	case "group":
		msg.Payload = enrichPayloadWithSender(msg.FromID, msg.Payload)
		if sendGroupMessage(hub, msg) {
			saveMessageAndAck(hub, msg, 2)
		}
	case "call":
		if client := hub.getClient(msg.ToID); client != nil {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
	case "heartbeat":
		if client := hub.getClient(msg.FromID); client != nil {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
	case "read":
		handleReadReceipt(hub, msg)
	case "revoke":
		handleRevoke(hub, msg)
	default:
		// logger.Info("unknown message type: %s", msg.Type)
	}
}

func enrichPayloadWithSender(fromID uint64, payload []byte) []byte {
	if fromID == 0 || len(payload) == 0 {
		return payload
	}
	var p chatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return payload
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return payload
	}
	var account model.UserAccount
	if err := dbConn.First(&account, int64(fromID)).Error; err != nil {
		return payload
	}
	var profile model.UserProfile
	_ = dbConn.Where("user_id = ?", int64(fromID)).First(&profile).Error
	name := strings.TrimSpace(profile.Nickname)
	if name == "" {
		name = account.Username
	}
	if p.Extra == nil {
		p.Extra = map[string]interface{}{}
	}
	p.Extra["fromName"] = name
	if profile.Avatar != "" {
		p.Extra["fromAvatar"] = profile.Avatar
	}
	updated, err := json.Marshal(p)
	if err != nil {
		return payload
	}
	return updated
}

func sendGroupMessage(hub *Hub, msg *WSMessage) bool {
	if msg == nil || msg.ToID == 0 {
		return false
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return false
	}
	var memberCount int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", int64(msg.ToID), int64(msg.FromID)).Count(&memberCount)
	if memberCount == 0 {
		return false
	}
	var members []model.GroupMember
	if err := dbConn.Where("group_id = ?", int64(msg.ToID)).Find(&members).Error; err != nil {
		return false
	}
	data, _ := json.Marshal(msg)
	for _, m := range members {
		if uint64(m.UserID) == msg.FromID {
			continue
		}
		if client := hub.getClient(uint64(m.UserID)); client != nil {
			client.Send <- data
		}
	}
	return true
}

func saveMessageAndAck(hub *Hub, msg *WSMessage, chatType int8) {
	if msg == nil || msg.FromID == 0 || msg.ToID == 0 {
		return
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return
	}
	content := strings.TrimSpace(string(msg.Payload))
	contentType := "text"
	tempID := ""
	var payload chatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err == nil {
		if strings.TrimSpace(payload.Content) != "" {
			content = payload.Content
		}
		if strings.TrimSpace(payload.ContentType) != "" {
			contentType = strings.TrimSpace(payload.ContentType)
		}
		tempID = strings.TrimSpace(payload.TempID)
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
	case "audio":
		msgType = 5
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
	if err := dbConn.Create(&message).Error; err != nil {
		return
	}
	ackPayload, _ := json.Marshal(map[string]interface{}{
		"tempId":    tempID,
		"messageId": message.ID,
		"chatType":  chatType,
	})
	ackMsg := WSMessage{
		Type:    "ack",
		FromID:  msg.FromID,
		ToID:    msg.ToID,
		Payload: ackPayload,
	}
	if raw, err := json.Marshal(ackMsg); err == nil {
		if client := hub.getClient(msg.FromID); client != nil {
			client.Send <- raw
		}
	}
}

func handleReadReceipt(hub *Hub, msg *WSMessage) {
	if msg == nil || msg.FromID == 0 || msg.ToID == 0 {
		return
	}
	dbConn := db.GetDB()
	if dbConn != nil {
		dbConn.Model(&model.Message{}).
			Where("from_id = ? AND to_id = ? AND status < 1", int64(msg.ToID), int64(msg.FromID)).
			Update("status", 1)
	}
	if client := hub.getClient(msg.ToID); client != nil {
		data, _ := json.Marshal(msg)
		client.Send <- data
	}
}

func handleRevoke(hub *Hub, msg *WSMessage) {
	if msg == nil || msg.FromID == 0 {
		return
	}
	var payload chatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return
	}
	if payload.Extra == nil {
		return
	}
	rawID, ok := payload.Extra["messageId"]
	if !ok {
		return
	}
	var messageID int64
	switch v := rawID.(type) {
	case float64:
		messageID = int64(v)
	case string:
		fmt.Sscanf(v, "%d", &messageID)
	}
	if messageID <= 0 {
		return
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return
	}
	var message model.Message
	if err := dbConn.First(&message, messageID).Error; err != nil {
		return
	}
	if message.FromID != int64(msg.FromID) {
		return
	}
	dbConn.Model(&message).Updates(map[string]interface{}{
		"status":  2,
		"content": "",
	})
	if client := hub.getClient(msg.ToID); client != nil {
		data, _ := json.Marshal(msg)
		client.Send <- data
	}
}
