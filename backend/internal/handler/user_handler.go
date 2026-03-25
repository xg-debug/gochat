package handler

import (
	"strings"
	"time"

	"gochat/internal/dto/request"
	"gochat/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

func (h *App) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	token, user, err := h.User.Login(req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": user})
}

func (h *App) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	token, user, err := h.User.Register(req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": user})
}

func (h *App) Profile(c *gin.Context) {
	profile, err := h.User.Profile(currentUserID(c))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, profile)
}

func (h *App) UpdateProfile(c *gin.Context) {
	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	profile, err := h.User.UpdateProfile(currentUserID(c), req)
	if err != nil {
		if strings.Contains(err.Error(), "birthday") {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, profile)
}

func (h *App) Conversations(c *gin.Context) {
	result, err := h.User.GetConversations(currentUserID(c))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *App) SearchConversations(c *gin.Context) {
	var req request.SearchConversationsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	result, err := h.User.SearchConversations(currentUserID(c), req.Keyword)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *App) Messages(c *gin.Context) {
	var req request.MessagesQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "conversationId required"})
		return
	}
	result, err := h.User.GetMessages(currentUserID(c), req.ConversationID, req.Cursor, req.Limit)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "required"), strings.Contains(err.Error(), "invalid"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "not in group"):
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, result)
}

func (h *App) Logout(c *gin.Context) {
	token := auth.ExtractToken(c)
	if token == "" {
		c.JSON(400, gin.H{"error": "missing token"})
		return
	}
	claims, err := auth.ParseToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	if claims.ExpiresAt != nil {
		auth.RevokeToken(token, claims.ExpiresAt.Time)
	} else {
		auth.RevokeToken(token, time.Now().Add(72*time.Hour))
	}
	c.JSON(200, gin.H{"message": "ok"})
}
