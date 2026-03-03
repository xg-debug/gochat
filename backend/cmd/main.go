package main

import (
	"fmt"

	"gochat/internal/config"
	"gochat/internal/handler"
	"gochat/internal/model"
	"gochat/internal/pkg/auth"
	"gochat/internal/pkg/db"
	zlog "gochat/internal/pkg/zlog"
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
	); err != nil {
		zlog.GetLogger().Fatal("db migrate failed", zap.Error(err))
	}

	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 静态文件服务：头像
	r.Static("/uploads/avatars", "./uploads/avatars")
	// 静态文件服务：聊天图片
	r.Static("/uploads/chat", "./uploads/chat")

	api := r.Group("/api")
	api.POST("/login", handler.Login)
	api.POST("/register", handler.Register)

	authorized := api.Group("/")
	authorized.Use(auth.AuthMiddleware())
	{
		authorized.POST("/logout", handler.Logout)
		authorized.GET("/profile", handler.Profile)
		authorized.PUT("/profile", handler.UpdateProfile)       // 更新个人信息
		authorized.POST("/upload/avatar", handler.UploadAvatar) // 头像上传
		authorized.POST("/upload/chat/image", handler.UploadChatImage)

		// 好友相关接口
		authorized.GET("/user/search", handler.SearchUser)
		authorized.POST("/friend/request", handler.SendFriendRequest)
		authorized.GET("/friend/requests", handler.ListFriendRequests)
		authorized.POST("/friend/handle", handler.HandleFriendRequest)
		authorized.GET("/contacts", handler.GetContacts) // 使用新的 GetContacts 实现

		authorized.GET("/conversations", handler.Conversations)
		authorized.GET("/messages", handler.Messages)
	}

	r.GET("/ws", auth.AuthMiddleware(), handler.WSHandler(hub, zlog.GetLogger()))

	addr := fmt.Sprintf("%s:%d", cfg.MainConfig.Host, cfg.MainConfig.Port)
	if err := r.Run(addr); err != nil {
		zlog.GetLogger().Fatal("server start failed", zap.Error(err))
	}
}
