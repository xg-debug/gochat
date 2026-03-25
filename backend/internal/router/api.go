package router

import (
	"time"

	"gochat/internal/handler"
	"gochat/internal/pkg/auth"
	"gochat/internal/pkg/middleware"
	"gochat/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterRoutes(r *gin.Engine, h *handler.App, hub *ws.Hub, logger *zap.Logger) {
	authLimiter := middleware.NewMemoryRateLimiter(30, time.Minute, middleware.KeyByIP)
	userLimiter := middleware.NewMemoryRateLimiter(120, time.Minute, middleware.KeyByUserOrIP)
	wsLimiter := middleware.NewMemoryRateLimiter(40, time.Minute, middleware.KeyByUserOrIP)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 静态文件服务：头像
	r.Static("/uploads/avatars", "./uploads/avatars")
	// 静态文件服务：聊天图片
	r.Static("/uploads/chat", "./uploads/chat")
	// 静态文件服务：聊天文件/语音
	r.Static("/uploads/files", "./uploads/files")
	r.Static("/uploads/audio", "./uploads/audio")
	r.Static("/uploads/groups", "./uploads/groups")

	api := r.Group("/api")
	api.POST("/login", authLimiter.Middleware(), h.Login)
	api.POST("/register", authLimiter.Middleware(), h.Register)

	authorized := api.Group("/")
	authorized.Use(auth.AuthMiddleware())
	{
		authorized.POST("/logout", h.Logout)
		authorized.GET("/profile", h.Profile)
		authorized.PUT("/profile", h.UpdateProfile)
		authorized.POST("/upload/avatar", userLimiter.Middleware(), h.UploadAvatar)
		authorized.POST("/upload/chat/image", userLimiter.Middleware(), h.UploadChatImage)
		authorized.POST("/upload/chat/file", userLimiter.Middleware(), h.UploadChatFile)
		authorized.POST("/upload/chat/audio", userLimiter.Middleware(), h.UploadChatAudio)
		authorized.POST("/upload/group/avatar", userLimiter.Middleware(), h.UploadGroupAvatar)

		// 好友相关接口
		authorized.GET("/user/search", h.SearchUser)
		authorized.POST("/friend/request", userLimiter.Middleware(), h.SendFriendRequest)
		authorized.GET("/friend/requests", h.ListFriendRequests)
		authorized.POST("/friend/handle", h.HandleFriendRequest)
		authorized.POST("/friend/delete", h.DeleteFriend)
		authorized.POST("/friend/block", h.BlockFriend)
		authorized.POST("/friend/unblock", h.UnblockFriend)
		authorized.GET("/contacts", h.GetContacts)

		authorized.GET("/conversations", h.Conversations)
		authorized.GET("/conversations/search", h.SearchConversations)
		authorized.GET("/messages", h.Messages)

		// 群聊相关
		authorized.POST("/group/create", h.CreateGroup)
		authorized.GET("/group/search", h.SearchGroup)
		authorized.POST("/group/join", h.JoinGroup)
		authorized.GET("/groups", h.ListGroups)
		authorized.GET("/group/profile", h.GetGroupProfile)
		authorized.PUT("/group/profile", h.UpdateGroupProfile)
		authorized.GET("/group/members", h.ListGroupMembers)
		authorized.POST("/group/kick", h.KickGroupMember)
		authorized.POST("/group/admin", h.SetGroupAdmin)
		authorized.GET("/group/inviteable", h.ListInviteableFriends)
		authorized.POST("/group/invite", h.InviteGroupMember)
	}

	r.GET("/ws", auth.AuthMiddleware(), wsLimiter.Middleware(), handler.WSHandler(hub, logger))
}
