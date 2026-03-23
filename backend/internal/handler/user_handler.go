package handler

import (
	"strings"
	"time"

	"gochat/internal/dto/request"
	"gochat/internal/pkg/auth"
	"gochat/internal/service"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	token, user, err := service.UserService.Login(req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": user})
}

func Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	token, user, err := service.UserService.Register(req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token, "user": user})
}

func Profile(c *gin.Context) {
	userID := c.GetInt64("userID")
	profile, err := service.UserService.Profile(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, profile)
}

func UpdateProfile(c *gin.Context) {
	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetInt64("userID")
	profile, err := service.UserService.UpdateProfile(userID, req)
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

func Conversations(c *gin.Context) {
	userID := c.GetInt64("userID")
	result, err := service.UserService.GetConversations(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func SearchConversations(c *gin.Context) {
	var req request.SearchConversationsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetInt64("userID")
	result, err := service.UserService.SearchConversations(userID, req.Keyword)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func Messages(c *gin.Context) {
	var req request.MessagesQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "conversationId required"})
		return
	}
	userID := c.GetInt64("userID")
	result, err := service.UserService.GetMessages(userID, req.ConversationID)
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

func Logout(c *gin.Context) {
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
