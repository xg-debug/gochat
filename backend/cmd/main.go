package main

import (
	"fmt"

	"gochat/internal/config"
	"gochat/internal/handler"
	"gochat/internal/model"
	"gochat/internal/pkg/auth"
	"gochat/internal/pkg/db"
	zlog "gochat/internal/pkg/zlog"
	"gochat/internal/service"
	"gochat/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.GetConfig()
	dbConn, err := db.Init(cfg)
	if err != nil {
		zlog.GetLogger().Fatal("db init failed", zap.Error(err))
	}
	if err := dbConn.AutoMigrate(
		&model.UserAccount{},
		&model.UserProfile{},
		&model.Friend{},
		&model.FriendRequest{},
		&model.Message{},
		&model.File{},
		&model.ChatGroup{},
		&model.GroupMember{},
	); err != nil {
		zlog.GetLogger().Fatal("db migrate failed", zap.Error(err))
	}

	hub := ws.NewHub()
	go hub.Run()

	userSvc := service.NewUserService(dbConn, ws.IsOnline)
	friendSvc := service.NewFriendService(dbConn, ws.IsOnline)
	groupSvc := service.NewGroupService(dbConn)
	uploadSvc := service.NewUploadService(dbConn)
	h := handler.NewApp(userSvc, friendSvc, groupSvc, uploadSvc)

	r := gin.Default()
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
	api.POST("/login", h.Login)
	api.POST("/register", h.Register)

	authorized := api.Group("/")
	authorized.Use(auth.AuthMiddleware())
	{
		authorized.POST("/logout", h.Logout)
		authorized.GET("/profile", h.Profile)
		authorized.PUT("/profile", h.UpdateProfile) // 更新个人信息
		authorized.POST("/upload/avatar", h.UploadAvatar)
		authorized.POST("/upload/chat/image", h.UploadChatImage)
		authorized.POST("/upload/chat/file", h.UploadChatFile)
		authorized.POST("/upload/chat/audio", h.UploadChatAudio)
		authorized.POST("/upload/group/avatar", h.UploadGroupAvatar)

		// 好友相关接口
		authorized.GET("/user/search", h.SearchUser)
		authorized.POST("/friend/request", h.SendFriendRequest)
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

	r.GET("/ws", auth.AuthMiddleware(), handler.WSHandler(hub, zlog.GetLogger()))

	addr := fmt.Sprintf("%s:%d", cfg.MainConfig.Host, cfg.MainConfig.Port)
	if err := r.Run(addr); err != nil {
		zlog.GetLogger().Fatal("server start failed", zap.Error(err))
	}
}
