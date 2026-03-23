package handler

import (
	"strings"

	"gochat/internal/dto/request"
	"gochat/internal/service"

	"github.com/gin-gonic/gin"
)

func SearchUser(c *gin.Context) {
	var req request.SearchUserQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "keyword is required"})
		return
	}
	result, err := service.FriendService.SearchUser(c.GetInt64("userID"), req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func SendFriendRequest(c *gin.Context) {
	var req request.SendFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := service.FriendService.SendFriendRequest(c.GetInt64("userID"), req); err != nil {
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

func ListFriendRequests(c *gin.Context) {
	result, err := service.FriendService.ListFriendRequests(c.GetInt64("userID"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func HandleFriendRequest(c *gin.Context) {
	var req request.HandleFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := service.FriendService.HandleFriendRequest(c.GetInt64("userID"), req); err != nil {
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

func DeleteFriend(c *gin.Context) {
	var req request.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := service.FriendService.DeleteFriend(c.GetInt64("userID"), req.FriendID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func BlockFriend(c *gin.Context) {
	var req request.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := service.FriendService.BlockFriend(c.GetInt64("userID"), req.FriendID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func UnblockFriend(c *gin.Context) {
	var req request.FriendActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := service.FriendService.UnblockFriend(c.GetInt64("userID"), req.FriendID); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func GetContacts(c *gin.Context) {
	result, err := service.FriendService.GetContacts(c.GetInt64("userID"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}
