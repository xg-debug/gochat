package handler

import (
	"strings"

	"gochat/internal/dto/request"

	"github.com/gin-gonic/gin"
)

func (h *App) SearchUser(c *gin.Context) {
	var req request.SearchUserQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "keyword is required"})
		return
	}
	result, err := h.Friend.SearchUser(currentUserID(c), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *App) SendFriendRequest(c *gin.Context) {
	var req request.SendFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.Friend.SendFriendRequest(currentUserID(c), req); err != nil {
		switch err.Error() {
		case "cannot add yourself", "already friends", "user already requested you":
			c.JSON(400, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "request sent"})
}

func (h *App) ListFriendRequests(c *gin.Context) {
	result, err := h.Friend.ListFriendRequests(currentUserID(c))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *App) HandleFriendRequest(c *gin.Context) {
	var req request.HandleFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.Friend.HandleFriendRequest(currentUserID(c), req); err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			c.JSON(404, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "permission"), strings.Contains(err.Error(), "handled"):
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func (h *App) DeleteFriend(c *gin.Context) {
	var req request.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.Friend.DeleteFriend(currentUserID(c), req.FriendID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func (h *App) BlockFriend(c *gin.Context) {
	var req request.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.Friend.BlockFriend(currentUserID(c), req.FriendID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func (h *App) UnblockFriend(c *gin.Context) {
	var req request.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.Friend.UnblockFriend(currentUserID(c), req.FriendID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func (h *App) GetContacts(c *gin.Context) {
	result, err := h.Friend.GetContacts(currentUserID(c))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}
