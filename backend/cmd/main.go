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
	if err := dbConn.AutoMigrate(&model.UserAccount{}, &model.UserProfile{}); err != nil {
		zlog.GetLogger().Fatal("db migrate failed", zap.Error(err))
	}

	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api := r.Group("/api")
	api.POST("/login", handler.Login)
	api.POST("/register", handler.Register)
	api.GET("/profile", auth.AuthMiddleware(), handler.Profile)
	api.GET("/contacts", auth.AuthMiddleware(), handler.Contacts)
	api.GET("/conversations", auth.AuthMiddleware(), handler.Conversations)
	api.GET("/messages", auth.AuthMiddleware(), handler.Messages)

	r.GET("/ws", auth.AuthMiddleware(), handler.WSHandler(hub, zlog.GetLogger()))

	addr := fmt.Sprintf("%s:%d", cfg.MainConfig.Host, cfg.MainConfig.Port)
	if err := r.Run(addr); err != nil {
		zlog.GetLogger().Fatal("server start failed", zap.Error(err))
	}
}
