package handler

import (
	"errors"
	"fmt"
	"strings"

	"gochat/internal/model"
	"gochat/internal/pkg/auth"
	"gochat/internal/pkg/db"
	zlog "gochat/internal/pkg/zlog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type contactResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type conversationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	LastMessage string `json:"lastMessage"`
	Unread      int    `json:"unread"`
}

type messageResponse struct {
	ID          string `json:"id"`
	FromID      string `json:"fromId"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	Time        int64  `json:"time"`
	Status      string `json:"status"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		c.JSON(400, gin.H{"error": "username or password missing"})
		return
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(500, gin.H{"error": "db not ready"})
		return
	}
	var account model.UserAccount
	if err := dbConn.Where("username = ?", req.Username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	if err := auth.CheckPassword(account.PasswordHash, req.Password); err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	var profile model.UserProfile
	if err := dbConn.Where("user_id = ?", account.ID).First(&profile).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
	}
	nickname := profile.Nickname
	if strings.TrimSpace(nickname) == "" {
		nickname = account.Username
	}
	token, err := auth.GenerateToken(account.ID, account.Username)
	if err != nil {
		zlog.Error("generate token failed")
		c.JSON(500, gin.H{"error": "token error"})
		return
	}
	c.JSON(200, gin.H{
		"token": token,
		"user": userResponse{
			ID:       account.ID,
			Username: account.Username,
			Nickname: nickname,
			Avatar:   profile.Avatar,
		},
	})
}

func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Username == "" || req.Password == "" {
		c.JSON(400, gin.H{"error": "username or password missing"})
		return
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(500, gin.H{"error": "db not ready"})
		return
	}
	var existing model.UserAccount
	if err := dbConn.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		c.JSON(409, gin.H{"error": "username already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "password error"})
		return
	}
	account := model.UserAccount{
		Username:     req.Username,
		PasswordHash: hash,
		Status:       1,
	}
	profile := model.UserProfile{
		Nickname: req.Nickname,
	}
	if profile.Nickname == "" {
		profile.Nickname = req.Username
	}
	tx := dbConn.Begin()
	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	profile.UserID = account.ID
	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	token, err := auth.GenerateToken(account.ID, account.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "token error"})
		return
	}
	c.JSON(200, gin.H{
		"token": token,
		"user": userResponse{
			ID:       account.ID,
			Username: account.Username,
			Nickname: profile.Nickname,
			Avatar:   profile.Avatar,
		},
	})
}

func Profile(c *gin.Context) {
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(500, gin.H{"error": "db not ready"})
		return
	}
	userID := int64(c.GetUint64("user_id"))
	if userID <= 0 {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	var account model.UserAccount
	if err := dbConn.Where("id = ?", userID).First(&account).Error; err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	var profile model.UserProfile
	if err := dbConn.Where("user_id = ?", userID).First(&profile).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	nickname := profile.Nickname
	if strings.TrimSpace(nickname) == "" {
		nickname = account.Username
	}
	c.JSON(200, userResponse{
		ID:       account.ID,
		Username: account.Username,
		Nickname: nickname,
		Avatar:   profile.Avatar,
	})
}

func Contacts(c *gin.Context) {
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(500, gin.H{"error": "db not ready"})
		return
	}
	userID := int64(c.GetUint64("user_id"))
	if userID <= 0 {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	var accounts []model.UserAccount
	if err := dbConn.Where("id <> ?", userID).Order("id asc").Find(&accounts).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	profileMap := map[int64]model.UserProfile{}
	if len(ids) > 0 {
		var profiles []model.UserProfile
		if err := dbConn.Where("user_id in ?", ids).Find(&profiles).Error; err == nil {
			for _, profile := range profiles {
				profileMap[profile.UserID] = profile
			}
		}
	}
	result := make([]contactResponse, 0, len(accounts))
	for _, account := range accounts {
		profile := profileMap[account.ID]
		name := strings.TrimSpace(profile.Nickname)
		if name == "" {
			name = account.Username
		}
		result = append(result, contactResponse{
			ID:     fmt.Sprintf("u_%d", account.ID),
			Name:   name,
			Avatar: profile.Avatar,
		})
	}
	c.JSON(200, result)
}

func Conversations(c *gin.Context) {
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(500, gin.H{"error": "db not ready"})
		return
	}
	userID := int64(c.GetUint64("user_id"))
	if userID <= 0 {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	var accounts []model.UserAccount
	if err := dbConn.Where("id <> ?", userID).Order("id asc").Find(&accounts).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	profileMap := map[int64]model.UserProfile{}
	if len(ids) > 0 {
		var profiles []model.UserProfile
		if err := dbConn.Where("user_id in ?", ids).Find(&profiles).Error; err == nil {
			for _, profile := range profiles {
				profileMap[profile.UserID] = profile
			}
		}
	}
	result := make([]conversationResponse, 0, len(accounts))
	for _, account := range accounts {
		profile := profileMap[account.ID]
		name := strings.TrimSpace(profile.Nickname)
		if name == "" {
			name = account.Username
		}
		lastMessage := ""
		var msg model.Message
		err := dbConn.
			Where("(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)", userID, account.ID, account.ID, userID).
			Order("created_at desc").
			Limit(1).
			Find(&msg).Error
		if err == nil && msg.ID > 0 {
			lastMessage = msg.Content
		}
		result = append(result, conversationResponse{
			ID:          fmt.Sprintf("u_%d", account.ID),
			Name:        name,
			Avatar:      profile.Avatar,
			LastMessage: lastMessage,
			Unread:      0,
		})
	}
	c.JSON(200, result)
}

func Messages(c *gin.Context) {
	dbConn := db.GetDB()
	if dbConn == nil {
		c.JSON(500, gin.H{"error": "db not ready"})
		return
	}
	userID := int64(c.GetUint64("user_id"))
	if userID <= 0 {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	conversationID := strings.TrimSpace(c.Query("conversationId"))
	if conversationID == "" {
		c.JSON(400, gin.H{"error": "conversationId required"})
		return
	}
	peerIDStr := strings.TrimPrefix(conversationID, "u_")
	peerID, err := parseInt64(peerIDStr)
	if err != nil || peerID <= 0 {
		c.JSON(400, gin.H{"error": "invalid conversationId"})
		return
	}
	var messages []model.Message
	if err := dbConn.
		Where("(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)", userID, peerID, peerID, userID).
		Order("created_at asc").
		Find(&messages).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	result := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		result = append(result, messageResponse{
			ID:          fmt.Sprintf("m_%d", msg.ID),
			FromID:      fmt.Sprintf("u_%d", msg.FromID),
			Content:     msg.Content,
			ContentType: "text",
			Time:        msg.CreatedAt.UnixMilli(),
			Status:      "delivered",
		})
	}
	c.JSON(200, result)
}

func parseInt64(value string) (int64, error) {
	var parsed int64
	_, err := fmt.Sscanf(value, "%d", &parsed)
	return parsed, err
}
