package ws

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gochat/internal/model"
)

// 分发逻辑

type chatPayload struct {
	Content     string                 `json:"content"`
	ContentType string                 `json:"contentType"`
	TempID      string                 `json:"tempId"`
	Extra       map[string]interface{} `json:"extra"`
}

type groupCallSession struct {
	GroupID      uint64
	HostID       uint64
	CallType     string
	Participants map[uint64]time.Time
	UpdatedAt    time.Time
}

var (
	groupCallMu       sync.RWMutex
	groupCallSessions = map[uint64]*groupCallSession{}
	groupCallTTL      = 30 * time.Minute
)

func RouteMessage(hub *Hub, msg *WSMessage) {
	switch msg.Type {
	case "single":
		if !canSendSingle(hub, msg.FromID, msg.ToID) {
			sendAckError(hub, msg, "forbidden")
			return
		}
		msg.Payload = enrichPayloadWithSender(hub, msg.FromID, msg.Payload)
		if client := hub.getClient(msg.ToID); client != nil {
			data, _ := json.Marshal(msg)
			client.Send <- data
		}
		saveMessageAndAck(hub, msg, 1)
	case "group":
		msg.Payload = enrichPayloadWithSender(hub, msg.FromID, msg.Payload)
		if sendGroupMessage(hub, msg) {
			saveMessageAndAck(hub, msg, 2)
		}
	case "call":
		if isGroupCallSignal(msg) {
			sendGroupCallSignal(hub, msg)
			return
		}
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

func canSendSingle(hub *Hub, fromID, toID uint64) bool {
	if hub == nil || hub.db == nil || fromID == 0 || toID == 0 || fromID == toID {
		return false
	}
	var outbound model.Friend
	if err := hub.db.Where("user_id = ? AND friend_id = ?", int64(fromID), int64(toID)).First(&outbound).Error; err != nil {
		return false
	}
	var inbound model.Friend
	if err := hub.db.Where("user_id = ? AND friend_id = ?", int64(toID), int64(fromID)).First(&inbound).Error; err != nil {
		return false
	}
	return outbound.Status == 1 && inbound.Status == 1
}

func sendAckError(hub *Hub, msg *WSMessage, reason string) {
	if hub == nil || msg == nil || msg.FromID == 0 {
		return
	}
	tempID := ""
	var payload chatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err == nil {
		tempID = strings.TrimSpace(payload.TempID)
	}
	ackPayload, _ := json.Marshal(map[string]interface{}{
		"tempId": tempID,
		"error":  reason,
	})
	ackMsg := WSMessage{Type: "ack", FromID: msg.FromID, ToID: msg.ToID, Payload: ackPayload}
	if raw, err := json.Marshal(ackMsg); err == nil {
		if client := hub.getClient(msg.FromID); client != nil {
			client.Send <- raw
		}
	}
}

func isGroupCallSignal(msg *WSMessage) bool {
	if msg == nil || msg.ToID == 0 {
		return false
	}
	var payload chatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return false
	}
	if payload.Extra == nil {
		return false
	}
	scope, _ := payload.Extra["scope"].(string)
	return strings.EqualFold(scope, "group")
}

func parseGroupCallAction(payload chatPayload) (string, string) {
	if payload.Extra == nil {
		return "", "video"
	}
	action, _ := payload.Extra["action"].(string)
	callType, _ := payload.Extra["callType"].(string)
	callType = strings.ToLower(strings.TrimSpace(callType))
	if callType != "audio" && callType != "video" {
		callType = "video"
	}
	return strings.ToLower(strings.TrimSpace(action)), callType
}

func getGroupCallSession(groupID uint64) *groupCallSession {
	groupCallMu.Lock()
	defer groupCallMu.Unlock()
	session := groupCallSessions[groupID]
	if session == nil {
		return nil
	}
	if time.Since(session.UpdatedAt) > groupCallTTL {
		delete(groupCallSessions, groupID)
		return nil
	}
	copied := &groupCallSession{
		GroupID:      session.GroupID,
		HostID:       session.HostID,
		CallType:     session.CallType,
		Participants: make(map[uint64]time.Time, len(session.Participants)),
		UpdatedAt:    session.UpdatedAt,
	}
	for id, ts := range session.Participants {
		copied.Participants[id] = ts
	}
	return copied
}

func upsertGroupCallSession(groupID, hostID, userID uint64, callType string) {
	if groupID == 0 || userID == 0 {
		return
	}
	groupCallMu.Lock()
	defer groupCallMu.Unlock()
	now := time.Now()
	session := groupCallSessions[groupID]
	if session == nil || time.Since(session.UpdatedAt) > groupCallTTL {
		session = &groupCallSession{
			GroupID:      groupID,
			HostID:       hostID,
			CallType:     callType,
			Participants: map[uint64]time.Time{},
		}
		groupCallSessions[groupID] = session
	}
	if hostID > 0 {
		session.HostID = hostID
	}
	if callType == "audio" || callType == "video" {
		session.CallType = callType
	}
	session.Participants[userID] = now
	session.UpdatedAt = now
}

func removeGroupCallParticipant(groupID, userID uint64) bool {
	groupCallMu.Lock()
	defer groupCallMu.Unlock()
	session := groupCallSessions[groupID]
	if session == nil {
		return false
	}
	delete(session.Participants, userID)
	session.UpdatedAt = time.Now()
	if len(session.Participants) == 0 {
		delete(groupCallSessions, groupID)
		return true
	}
	return false
}

func sendGroupCallStateToUser(hub *Hub, userID, groupID uint64) {
	if hub == nil || userID == 0 || groupID == 0 {
		return
	}
	extra := map[string]interface{}{
		"scope":   "group",
		"groupId": groupID,
		"action":  "group-state",
		"active":  false,
	}
	fromID := uint64(0)
	session := getGroupCallSession(groupID)
	if session != nil {
		extra["active"] = true
		extra["hostId"] = session.HostID
		extra["callType"] = session.CallType
		participants := make([]uint64, 0, len(session.Participants))
		for participantID := range session.Participants {
			participants = append(participants, participantID)
		}
		extra["participants"] = participants
		fromID = session.HostID
	}
	payloadRaw, err := json.Marshal(chatPayload{Content: "", ContentType: "text", Extra: extra})
	if err != nil {
		return
	}
	resp := WSMessage{Type: "call", FromID: fromID, ToID: groupID, Payload: payloadRaw}
	if raw, err := json.Marshal(resp); err == nil {
		if client := hub.getClient(userID); client != nil {
			client.Send <- raw
		}
	}
}

func sendGroupCallSignal(hub *Hub, msg *WSMessage) bool {
	if msg == nil || msg.ToID == 0 || msg.FromID == 0 || hub == nil || hub.db == nil {
		return false
	}
	var payload chatPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return false
	}
	action, callType := parseGroupCallAction(payload)
	var memberCount int64
	hub.db.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", int64(msg.ToID), int64(msg.FromID)).
		Count(&memberCount)
	if memberCount == 0 {
		return false
	}
	if action == "group-state-query" {
		sendGroupCallStateToUser(hub, msg.FromID, msg.ToID)
		return true
	}
	var members []model.GroupMember
	if err := hub.db.Where("group_id = ?", int64(msg.ToID)).Find(&members).Error; err != nil {
		return false
	}
	switch action {
	case "group-initiate":
		upsertGroupCallSession(msg.ToID, msg.FromID, msg.FromID, callType)
	case "group-join":
		upsertGroupCallSession(msg.ToID, msg.FromID, msg.FromID, callType)
	case "group-offer", "group-answer", "group-candidate":
		upsertGroupCallSession(msg.ToID, 0, msg.FromID, callType)
	case "group-leave", "group-reject":
		removeGroupCallParticipant(msg.ToID, msg.FromID)
	}
	data, _ := json.Marshal(msg)
	for _, member := range members {
		if uint64(member.UserID) == msg.FromID {
			continue
		}
		if client := hub.getClient(uint64(member.UserID)); client != nil {
			client.Send <- data
		}
	}
	return true
}

func enrichPayloadWithSender(hub *Hub, fromID uint64, payload []byte) []byte {
	if fromID == 0 || len(payload) == 0 || hub == nil || hub.db == nil {
		return payload
	}
	var p chatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return payload
	}
	var account model.UserAccount
	if err := hub.db.First(&account, int64(fromID)).Error; err != nil {
		return payload
	}
	var profile model.UserProfile
	_ = hub.db.Where("user_id = ?", int64(fromID)).First(&profile).Error
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
	if msg == nil || msg.ToID == 0 || hub == nil || hub.db == nil {
		return false
	}
	var memberCount int64
	hub.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", int64(msg.ToID), int64(msg.FromID)).Count(&memberCount)
	if memberCount == 0 {
		return false
	}
	var members []model.GroupMember
	if err := hub.db.Where("group_id = ?", int64(msg.ToID)).Find(&members).Error; err != nil {
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
	if msg == nil || msg.FromID == 0 || msg.ToID == 0 || hub == nil || hub.db == nil {
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
	message := model.Message{FromID: int64(msg.FromID), ToID: int64(msg.ToID), ChatType: chatType, MsgType: msgType, Content: content, Status: 0, CreatedAt: time.Now()}
	if err := hub.db.Create(&message).Error; err != nil {
		return
	}
	ackPayload, _ := json.Marshal(map[string]interface{}{"tempId": tempID, "messageId": message.ID, "chatType": chatType})
	ackMsg := WSMessage{Type: "ack", FromID: msg.FromID, ToID: msg.ToID, Payload: ackPayload}
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
	if hub != nil && hub.db != nil {
		hub.db.Model(&model.Message{}).
			Where("from_id = ? AND to_id = ? AND status < 1", int64(msg.ToID), int64(msg.FromID)).
			Update("status", 1)
	}
	if client := hub.getClient(msg.ToID); client != nil {
		data, _ := json.Marshal(msg)
		client.Send <- data
	}
}

func handleRevoke(hub *Hub, msg *WSMessage) {
	if msg == nil || msg.FromID == 0 || hub == nil || hub.db == nil {
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
	var message model.Message
	if err := hub.db.First(&message, messageID).Error; err != nil {
		return
	}
	if message.FromID != int64(msg.FromID) {
		return
	}
	hub.db.Model(&message).Updates(map[string]interface{}{"status": 2, "content": ""})

	data, _ := json.Marshal(msg)
	if message.ChatType == 2 {
		var members []model.GroupMember
		if err := hub.db.Where("group_id = ?", message.ToID).Find(&members).Error; err == nil {
			for _, member := range members {
				if uint64(member.UserID) == msg.FromID {
					continue
				}
				if client := hub.getClient(uint64(member.UserID)); client != nil {
					client.Send <- data
				}
			}
		}
		return
	}
	if client := hub.getClient(msg.ToID); client != nil {
		client.Send <- data
	}
}
