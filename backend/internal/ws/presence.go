package ws

import "sync"

var (
	defaultHub     *Hub
	defaultHubLock sync.RWMutex
)

func setDefaultHub(hub *Hub) {
	defaultHubLock.Lock()
	defaultHub = hub
	defaultHubLock.Unlock()
}

func IsOnline(userID uint64) bool {
	defaultHubLock.RLock()
	hub := defaultHub
	defaultHubLock.RUnlock()
	if hub == nil {
		return false
	}
	return hub.isOnline(userID)
}

func SendToUser(userID uint64, payload []byte) bool {
	defaultHubLock.RLock()
	hub := defaultHub
	defaultHubLock.RUnlock()
	if hub == nil {
		return false
	}
	client := hub.getClient(userID)
	if client == nil {
		return false
	}
	client.Send <- payload
	return true
}
