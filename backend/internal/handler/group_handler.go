package handler

import (
	"strings"

	"gochat/internal/dto/request"
	"gochat/internal/service"

	"github.com/gin-gonic/gin"
)

func CreateGroup(c *gin.Context) {
	var req request.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	result, err := service.GroupService.CreateGroup(c.GetInt64("userID"), req)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, result)
}

func SearchGroup(c *gin.Context) {
	keyword := c.Query("keyword")
	result, err := service.GroupService.SearchGroup(keyword)
	if err != nil {
		if strings.Contains(err.Error(), "required") {
			c.JSON(400, gin.H{"error": err.Error()})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, result)
}

func JoinGroup(c *gin.Context) {
	var req request.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := service.GroupService.JoinGroup(c.GetInt64("userID"), req); err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "not found"):
			c.JSON(404, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "joined"})
}

func ListGroups(c *gin.Context) {
	result, err := service.GroupService.ListGroups(c.GetInt64("userID"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func GetGroupProfile(c *gin.Context) {
	var query request.GroupIDQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	result, err := service.GroupService.GetGroupProfile(c.GetInt64("userID"), query.GroupID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "not found"):
			c.JSON(404, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, result)
}

func UpdateGroupProfile(c *gin.Context) {
	var req request.UpdateGroupProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := service.GroupService.UpdateGroupProfile(c.GetInt64("userID"), req); err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "permission"), strings.Contains(err.Error(), "not in group"):
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func ListGroupMembers(c *gin.Context) {
	var query request.GroupIDQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	result, err := service.GroupService.ListGroupMembers(c.GetInt64("userID"), query.GroupID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
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

func KickGroupMember(c *gin.Context) {
	var req request.KickGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := service.GroupService.KickGroupMember(c.GetInt64("userID"), req); err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "yourself"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "permission"), strings.Contains(err.Error(), "not in group"):
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func SetGroupAdmin(c *gin.Context) {
	var req request.SetGroupAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := service.GroupService.SetGroupAdmin(c.GetInt64("userID"), req); err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "only owner") || strings.Contains(err.Error(), "not in group"):
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

func ListInviteableFriends(c *gin.Context) {
	var query request.GroupIDQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	result, err := service.GroupService.ListInviteableFriends(c.GetInt64("userID"), query.GroupID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
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

func InviteGroupMember(c *gin.Context) {
	var req request.InviteGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := service.GroupService.InviteGroupMember(c.GetInt64("userID"), req); err != nil {
		switch {
		case strings.Contains(err.Error(), "invalid"):
			c.JSON(400, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "not in group"), strings.Contains(err.Error(), "friends"):
			c.JSON(403, gin.H{"error": err.Error()})
		default:
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
