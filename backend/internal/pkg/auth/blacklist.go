package auth

import (
	"sync"
	"time"
)

type tokenBlacklist struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

var blacklist = &tokenBlacklist{
	tokens: make(map[string]time.Time),
}

func RevokeToken(token string, expiresAt time.Time) {
	if token == "" {
		return
	}
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(72 * time.Hour)
	}
	blacklist.mu.Lock()
	blacklist.tokens[token] = expiresAt
	blacklist.mu.Unlock()
}

func IsTokenRevoked(token string) bool {
	if token == "" {
		return false
	}
	now := time.Now()
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	if expiresAt, ok := blacklist.tokens[token]; ok {
		if expiresAt.Before(now) {
			delete(blacklist.tokens, token)
			return false
		}
		return true
	}
	return false
}
