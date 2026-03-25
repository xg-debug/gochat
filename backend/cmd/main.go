package main

import (
	"fmt"

	"gochat/internal/config"
	"gochat/internal/handler"
	"gochat/internal/model"
	"gochat/internal/pkg/db"
	zlog "gochat/internal/pkg/zlog"
	"gochat/internal/router"
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

	hub := ws.NewHub(dbConn)
	go hub.Run()

	userSvc := service.NewUserService(dbConn, ws.IsOnline)
	friendSvc := service.NewFriendService(dbConn, ws.IsOnline)
	groupSvc := service.NewGroupService(dbConn)
	uploadSvc := service.NewUploadService(dbConn)
	h := handler.NewApp(userSvc, friendSvc, groupSvc, uploadSvc)

	r := gin.Default()
	router.RegisterRoutes(r, h, hub, zlog.GetLogger())

	addr := fmt.Sprintf("%s:%d", cfg.MainConfig.Host, cfg.MainConfig.Port)
	if err := r.Run(addr); err != nil {
		zlog.GetLogger().Fatal("server start failed", zap.Error(err))
	}
}
