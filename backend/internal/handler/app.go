package handler

import (
	"gochat/internal/service"

	"github.com/gin-gonic/gin"
)

type App struct {
	User   *service.UserService
	Friend *service.FriendService
	Group  *service.GroupService
	Upload *service.UploadService
}

func NewApp(user *service.UserService, friend *service.FriendService, group *service.GroupService, upload *service.UploadService) *App {
	return &App{User: user, Friend: friend, Group: group, Upload: upload}
}

func currentUserID(c *gin.Context) int64 {
	if userID := c.GetInt64("userID"); userID > 0 {
		return userID
	}
	if value, ok := c.Get("user_id"); ok {
		if parsed, ok := value.(uint64); ok {
			return int64(parsed)
		}
	}
	if value, ok := c.Get("userId"); ok {
		if parsed, ok := value.(uint64); ok {
			return int64(parsed)
		}
	}
	return 0
}
