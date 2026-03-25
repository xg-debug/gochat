package handler

import (
	"strings"

	"gochat/internal/dto/request"

	"github.com/gin-gonic/gin"
)

func (h *App) CreateGroup(c *gin.Context) {
	var req request.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	result, err := h.Group.CreateGroup(currentUserID(c), req)
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

func (h *App) SearchGroup(c *gin.Context) {
	keyword := c.Query("keyword")
	result, err := h.Group.SearchGroup(keyword)
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

func (h *App) JoinGroup(c *gin.Context) {
	var req request.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Group.JoinGroup(currentUserID(c), req); err != nil {
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

func (h *App) ListGroups(c *gin.Context) {
	result, err := h.Group.ListGroups(currentUserID(c))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func (h *App) GetGroupProfile(c *gin.Context) {
	var query request.GroupIDQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	result, err := h.Group.GetGroupProfile(currentUserID(c), query.GroupID)
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

func (h *App) UpdateGroupProfile(c *gin.Context) {
	var req request.UpdateGroupProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Group.UpdateGroupProfile(currentUserID(c), req); err != nil {
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

func (h *App) ListGroupMembers(c *gin.Context) {
	var query request.GroupIDQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	result, err := h.Group.ListGroupMembers(currentUserID(c), query.GroupID)
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

func (h *App) KickGroupMember(c *gin.Context) {
	var req request.KickGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Group.KickGroupMember(currentUserID(c), req); err != nil {
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

func (h *App) SetGroupAdmin(c *gin.Context) {
	var req request.SetGroupAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Group.SetGroupAdmin(currentUserID(c), req); err != nil {
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

func (h *App) ListInviteableFriends(c *gin.Context) {
	var query request.GroupIDQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	result, err := h.Group.ListInviteableFriends(currentUserID(c), query.GroupID)
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

func (h *App) InviteGroupMember(c *gin.Context) {
	var req request.InviteGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if err := h.Group.InviteGroupMember(currentUserID(c), req); err != nil {
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
