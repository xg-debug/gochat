package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gochat/internal/model"
	"gochat/internal/pkg/auth"
	"gochat/internal/pkg/db"
	zlog "gochat/internal/pkg/zlog"
	"gochat/internal/ws"

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
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Gender    int8   `json:"gender"`
}

type contactResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Online bool   `json:"online"`
}

type conversationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	LastMessage string `json:"lastMessage"`
	Unread      int    `json:"unread"`
	Online      bool   `json:"online"`
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
			ID:        account.ID,
			Username:  account.Username,
			Nickname:  nickname,
			Avatar:    profile.Avatar,
			Signature: profile.Signature,
			Gender:    profile.Gender,
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
			ID:        account.ID,
			Username:  account.Username,
			Nickname:  profile.Nickname,
			Avatar:    profile.Avatar,
			Signature: profile.Signature,
			Gender:    profile.Gender,
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
		ID:        account.ID,
		Username:  account.Username,
		Nickname:  nickname,
		Avatar:    profile.Avatar,
		Signature: profile.Signature,
		Gender:    profile.Gender,
	})
}

// UpdateProfile 更新个人信息
func UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname  *string `json:"nickname"`
		Avatar    *string `json:"avatar"`
		Signature *string `json:"signature"`
		Gender    *int8   `json:"gender"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var profile model.UserProfile
	if err := dbConn.Where("user_id = ?", currentUserID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create if not exists
			profile = model.UserProfile{
				UserID: currentUserID,
			}
			dbConn.Create(&profile)
		} else {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
	}

	updates := map[string]interface{}{}
	if req.Nickname != nil {
		updates["nickname"] = strings.TrimSpace(*req.Nickname)
	}
	if req.Avatar != nil {
		updates["avatar"] = strings.TrimSpace(*req.Avatar)
	}
	if req.Signature != nil {
		updates["signature"] = strings.TrimSpace(*req.Signature)
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}

	if len(updates) > 0 {
		if err := dbConn.Model(&profile).Updates(updates).Error; err != nil {
			c.JSON(500, gin.H{"error": "update failed"})
			return
		}
	}

	// Refresh profile data
	dbConn.Where("user_id = ?", currentUserID).First(&profile)

	var account model.UserAccount
	dbConn.First(&account, currentUserID)

	nickname := strings.TrimSpace(profile.Nickname)
	if nickname == "" {
		nickname = account.Username
	}
	c.JSON(200, userResponse{
		ID:        account.ID,
		Username:  account.Username,
		Nickname:  nickname,
		Avatar:    profile.Avatar,
		Signature: profile.Signature,
		Gender:    profile.Gender,
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
			Online: ws.IsOnline(uint64(account.ID)),
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
	var friends []model.Friend
	if err := dbConn.Where("user_id = ? AND status = 1", userID).Find(&friends).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	ids := make([]int64, 0, len(friends))
	for _, f := range friends {
		ids = append(ids, f.FriendID)
	}
	var accounts []model.UserAccount
	if len(ids) > 0 {
		if err := dbConn.Where("id in ?", ids).Order("id asc").Find(&accounts).Error; err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
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
			if msg.Status == 2 {
				lastMessage = "[已撤回]"
			} else if msg.MsgType == 2 {
				lastMessage = "[图片]"
			} else if msg.MsgType == 3 {
				lastMessage = "[文件]"
			} else if msg.MsgType == 4 {
				lastMessage = "[视频]"
			} else if msg.MsgType == 5 {
				lastMessage = "[语音]"
			} else {
				lastMessage = msg.Content
			}
		}
		result = append(result, conversationResponse{
			ID:          fmt.Sprintf("u_%d", account.ID),
			Name:        name,
			Avatar:      profile.Avatar,
			LastMessage: lastMessage,
			Unread:      0,
			Online:      ws.IsOnline(uint64(account.ID)),
		})
	}
	// groups
	var members []model.GroupMember
	if err := dbConn.Where("user_id = ?", userID).Find(&members).Error; err == nil && len(members) > 0 {
		groupIDs := make([]int64, 0, len(members))
		for _, m := range members {
			groupIDs = append(groupIDs, m.GroupID)
		}
		var groups []model.ChatGroup
		if err := dbConn.Where("id in ?", groupIDs).Find(&groups).Error; err == nil {
			for _, g := range groups {
				lastMessage := ""
				var msg model.Message
				err := dbConn.
					Where("chat_type = ? AND to_id = ?", 2, g.ID).
					Order("created_at desc").
					Limit(1).
					Find(&msg).Error
				if err == nil && msg.ID > 0 {
					if msg.Status == 2 {
						lastMessage = "[已撤回]"
					} else if msg.MsgType == 2 {
						lastMessage = "[图片]"
					} else if msg.MsgType == 3 {
						lastMessage = "[文件]"
					} else if msg.MsgType == 4 {
						lastMessage = "[视频]"
					} else if msg.MsgType == 5 {
						lastMessage = "[语音]"
					} else {
						lastMessage = msg.Content
					}
				}
				result = append(result, conversationResponse{
					ID:          fmt.Sprintf("g_%d", g.ID),
					Name:        g.Name,
					Avatar:      g.Avatar,
					LastMessage: lastMessage,
					Unread:      0,
					Online:      false,
				})
			}
		}
	}
	c.JSON(200, result)
}

func SearchConversations(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		Conversations(c)
		return
	}
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
	kwLike := "%" + keyword + "%"

	var friends []model.Friend
	if err := dbConn.Where("user_id = ? AND status = 1", userID).Find(&friends).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	friendIDs := make([]int64, 0, len(friends))
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendID)
	}
	var accounts []model.UserAccount
	if len(friendIDs) > 0 {
		if err := dbConn.Where("id in ? AND username LIKE ?", friendIDs, kwLike).Find(&accounts).Error; err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
		var matchedProfiles []model.UserProfile
		_ = dbConn.Where("user_id in ? AND nickname LIKE ?", friendIDs, kwLike).Find(&matchedProfiles).Error
		profileMatchedIDs := map[int64]bool{}
		for _, profile := range matchedProfiles {
			profileMatchedIDs[profile.UserID] = true
		}
		if len(profileMatchedIDs) > 0 {
			exists := map[int64]bool{}
			for _, a := range accounts {
				exists[a.ID] = true
			}
			for id := range profileMatchedIDs {
				if exists[id] {
					continue
				}
				var acc model.UserAccount
				if err := dbConn.Where("id = ?", id).First(&acc).Error; err == nil {
					accounts = append(accounts, acc)
				}
			}
		}
	}
	result := make([]conversationResponse, 0, len(accounts))
	if len(accounts) > 0 {
		ids := make([]int64, 0, len(accounts))
		for _, a := range accounts {
			ids = append(ids, a.ID)
		}
		profileMap := map[int64]model.UserProfile{}
		var profiles []model.UserProfile
		_ = dbConn.Where("user_id in ?", ids).Find(&profiles).Error
		for _, p := range profiles {
			profileMap[p.UserID] = p
		}
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
				Online:      ws.IsOnline(uint64(account.ID)),
			})
		}
	}

	var members []model.GroupMember
	if err := dbConn.Where("user_id = ?", userID).Find(&members).Error; err == nil && len(members) > 0 {
		groupIDs := make([]int64, 0, len(members))
		for _, m := range members {
			groupIDs = append(groupIDs, m.GroupID)
		}
		var groups []model.ChatGroup
		if err := dbConn.Where("id in ? AND name LIKE ?", groupIDs, kwLike).Find(&groups).Error; err == nil {
			for _, g := range groups {
				result = append(result, conversationResponse{
					ID:          fmt.Sprintf("g_%d", g.ID),
					Name:        g.Name,
					Avatar:      g.Avatar,
					LastMessage: "",
					Unread:      0,
					Online:      false,
				})
			}
		}
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
	var messages []model.Message
	if strings.HasPrefix(conversationID, "g_") {
		groupIDStr := strings.TrimPrefix(conversationID, "g_")
		groupID, err := parseInt64(groupIDStr)
		if err != nil || groupID <= 0 {
			c.JSON(400, gin.H{"error": "invalid conversationId"})
			return
		}
		var count int64
		dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
		if count == 0 {
			c.JSON(403, gin.H{"error": "not in group"})
			return
		}
		if err := dbConn.
			Where("chat_type = ? AND to_id = ?", 2, groupID).
			Order("created_at asc").
			Find(&messages).Error; err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
	} else {
		peerIDStr := strings.TrimPrefix(conversationID, "u_")
		peerID, err := parseInt64(peerIDStr)
		if err != nil || peerID <= 0 {
			c.JSON(400, gin.H{"error": "invalid conversationId"})
			return
		}
		if err := dbConn.
			Where("(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)", userID, peerID, peerID, userID).
			Order("created_at asc").
			Find(&messages).Error; err != nil {
			c.JSON(500, gin.H{"error": "db error"})
			return
		}
	}
	result := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		content := msg.Content
		status := "delivered"
		if msg.Status == 1 {
			status = "read"
		} else if msg.Status == 2 {
			status = "revoked"
			content = "[已撤回]"
		}
		result = append(result, messageResponse{
			ID:          fmt.Sprintf("m_%d", msg.ID),
			FromID:      fmt.Sprintf("u_%d", msg.FromID),
			Content:     content,
			ContentType: msgTypeToContentType(msg.MsgType),
			Time:        msg.CreatedAt.UnixMilli(),
			Status:      status,
		})
	}
	c.JSON(200, result)
}

func parseInt64(value string) (int64, error) {
	var parsed int64
	_, err := fmt.Sscanf(value, "%d", &parsed)
	return parsed, err
}

func msgTypeToContentType(msgType int8) string {
	switch msgType {
	case 2:
		return "image"
	case 3:
		return "file"
	case 4:
		return "video"
	case 5:
		return "audio"
	default:
		return "text"
	}
}

func Logout(c *gin.Context) {
	token := auth.ExtractToken(c)
	if token == "" {
		c.JSON(400, gin.H{"error": "missing token"})
		return
	}
	claims, err := auth.ParseToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}
	if claims.ExpiresAt != nil {
		auth.RevokeToken(token, claims.ExpiresAt.Time)
	} else {
		auth.RevokeToken(token, time.Now().Add(72*time.Hour))
	}
	c.JSON(200, gin.H{"message": "ok"})
}
