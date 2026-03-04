package handler

import (
	"errors"
	"fmt"

	"gochat/internal/model"
	"gochat/internal/pkg/db"
	"gochat/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type searchUserRequest struct {
	Keyword string `form:"keyword" binding:"required"`
}

type friendRequestAction struct {
	ToUserID int64 `json:"toUserId" binding:"required"`
}

type friendAction struct {
	FriendID int64 `json:"friendId" binding:"required"`
}

type handleFriendRequest struct {
	RequestID int64  `json:"requestId" binding:"required"`
	Action    string `json:"action" binding:"required,oneof=accept reject"`
}

type userSearchResult struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	IsFriend bool   `json:"isFriend"`
	Pending  bool   `json:"pending"`
	PendingFromMe bool `json:"pendingFromMe"`
}

// SearchUser 搜索用户
func SearchUser(c *gin.Context) {
	var req searchUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "keyword is required"})
		return
	}

	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var users []model.UserAccount
	keyword := "%" + req.Keyword + "%"

	if err := dbConn.Where("username LIKE ?", keyword).Limit(20).Find(&users).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	var results []userSearchResult
	for _, u := range users {
		if u.ID == currentUserID {
			continue
		}

		var profile model.UserProfile
		dbConn.Where("user_id = ?", u.ID).First(&profile)

		// Check if friend
		var count int64
		dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, u.ID).Count(&count)
		pending := false
		pendingFromMe := false
		if count == 0 {
			var pendingReq model.FriendRequest
			err := dbConn.
				Where("status = 0 AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))",
					currentUserID, u.ID, u.ID, currentUserID).
				First(&pendingReq).Error
			if err == nil {
				pending = true
				if pendingReq.FromUserID == currentUserID {
					pendingFromMe = true
				}
			}
		}

		nickname := profile.Nickname
		if nickname == "" {
			nickname = u.Username
		}

		results = append(results, userSearchResult{
			ID:       u.ID,
			Username: u.Username,
			Nickname: nickname,
			Avatar:   profile.Avatar,
			IsFriend: count > 0,
			Pending:  pending,
			PendingFromMe: pendingFromMe,
		})
	}

	c.JSON(200, results)
}

// SendFriendRequest 发送好友请求
func SendFriendRequest(c *gin.Context) {
	var req friendRequestAction
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetInt64("userID")
	if currentUserID == req.ToUserID {
		c.JSON(400, gin.H{"error": "cannot add yourself"})
		return
	}

	dbConn := db.GetDB()

	// Check if already friends
	var count int64
	dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, req.ToUserID).Count(&count)
	if count > 0 {
		c.JSON(400, gin.H{"error": "already friends"})
		return
	}

	// Check existing pending request
	var existingReq model.FriendRequest
	err := dbConn.Where("from_user_id = ? AND to_user_id = ? AND status = 0", currentUserID, req.ToUserID).First(&existingReq).Error
	if err == nil {
		c.JSON(200, gin.H{"message": "request already sent"})
		return
	}
	// Check if target already requested current user
	var reverseReq model.FriendRequest
	if err := dbConn.Where("from_user_id = ? AND to_user_id = ? AND status = 0", req.ToUserID, currentUserID).First(&reverseReq).Error; err == nil {
		c.JSON(400, gin.H{"error": "user already requested you"})
		return
	}

	newReq := model.FriendRequest{
		FromUserID: currentUserID,
		ToUserID:   req.ToUserID,
		Status:     0, // Pending
	}

	if err := dbConn.Create(&newReq).Error; err != nil {
		c.JSON(500, gin.H{"error": "failed to send request"})
		return
	}

	c.JSON(200, gin.H{"message": "request sent"})
}

// ListFriendRequests 获取好友请求列表
func ListFriendRequests(c *gin.Context) {
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var requests []model.FriendRequest
	if err := dbConn.Where("to_user_id = ? AND status = 0", currentUserID).Find(&requests).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	type requestWithProfile struct {
		ID         int64  `json:"id"`
		FromUserID int64  `json:"fromUserId"`
		Username   string `json:"username"`
		Nickname   string `json:"nickname"`
		Avatar     string `json:"avatar"`
		Time       int64  `json:"time"`
	}

	var result []requestWithProfile
	for _, r := range requests {
		var account model.UserAccount
		var profile model.UserProfile
		dbConn.First(&account, r.FromUserID)
		dbConn.Where("user_id = ?", r.FromUserID).First(&profile)

		nickname := profile.Nickname
		if nickname == "" {
			nickname = account.Username
		}

		result = append(result, requestWithProfile{
			ID:         r.ID,
			FromUserID: r.FromUserID,
			Username:   account.Username,
			Nickname:   nickname,
			Avatar:     profile.Avatar,
			Time:       r.CreatedAt.Unix(),
		})
	}

	c.JSON(200, result)
}

// HandleFriendRequest 处理好友请求
func HandleFriendRequest(c *gin.Context) {
	var req handleFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var friendReq model.FriendRequest
	if err := dbConn.First(&friendReq, req.RequestID).Error; err != nil {
		c.JSON(404, gin.H{"error": "request not found"})
		return
	}

	if friendReq.ToUserID != currentUserID {
		c.JSON(403, gin.H{"error": "permission denied"})
		return
	}

	if friendReq.Status != 0 {
		c.JSON(400, gin.H{"error": "request already handled"})
		return
	}

	tx := dbConn.Begin()

	status := int8(2) // Reject
	if req.Action == "accept" {
		status = 1 // Accept
	}

	if err := tx.Model(&friendReq).Update("status", status).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}

	if req.Action == "accept" {
		// Create bidirectional friendship
		f1 := model.Friend{UserID: friendReq.FromUserID, FriendID: friendReq.ToUserID, Status: 1}
		f2 := model.Friend{UserID: friendReq.ToUserID, FriendID: friendReq.FromUserID, Status: 1}

		if err := tx.Create(&f1).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "create friend failed"})
			return
		}
		if err := tx.Create(&f2).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "create friend failed"})
			return
		}
	}

	tx.Commit()
	c.JSON(200, gin.H{"message": "success"})
}

// DeleteFriend 删除好友（双向）
func DeleteFriend(c *gin.Context) {
	var req friendAction
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	currentUserID := c.GetInt64("userID")
	if req.FriendID <= 0 {
		c.JSON(400, gin.H{"error": "invalid friendId"})
		return
	}
	dbConn := db.GetDB()
	if err := dbConn.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		currentUserID, req.FriendID, req.FriendID, currentUserID).
		Delete(&model.Friend{}).Error; err != nil {
		c.JSON(500, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

// BlockFriend 拉黑好友（单向）
func BlockFriend(c *gin.Context) {
	var req friendAction
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	currentUserID := c.GetInt64("userID")
	if req.FriendID <= 0 {
		c.JSON(400, gin.H{"error": "invalid friendId"})
		return
	}
	dbConn := db.GetDB()
	if err := dbConn.Model(&model.Friend{}).
		Where("user_id = ? AND friend_id = ?", currentUserID, req.FriendID).
		Update("status", 0).Error; err != nil {
		c.JSON(500, gin.H{"error": "block failed"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

// UnblockFriend 解除拉黑
func UnblockFriend(c *gin.Context) {
	var req friendAction
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	currentUserID := c.GetInt64("userID")
	if req.FriendID <= 0 {
		c.JSON(400, gin.H{"error": "invalid friendId"})
		return
	}
	dbConn := db.GetDB()
	if err := dbConn.Model(&model.Friend{}).
		Where("user_id = ? AND friend_id = ?", currentUserID, req.FriendID).
		Update("status", 1).Error; err != nil {
		c.JSON(500, gin.H{"error": "unblock failed"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

// UpdateProfile 更新个人信息
func UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Signature string `json:"signature"`
		Gender    int8   `json:"gender"`
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
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Signature != "" {
		updates["signature"] = req.Signature
	}
	updates["gender"] = req.Gender

	if err := dbConn.Model(&profile).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}

	// Refresh profile data
	dbConn.Where("user_id = ?", currentUserID).First(&profile)

	var account model.UserAccount
	dbConn.First(&account, currentUserID)

	c.JSON(200, userResponse{
		ID:       account.ID,
		Username: account.Username,
		Nickname: profile.Nickname,
		Avatar:   profile.Avatar,
	})
}

// GetContacts 获取联系人列表（重写原有的mock实现）
func GetContacts(c *gin.Context) {
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var friends []model.Friend
	if err := dbConn.Where("user_id = ? AND status = 1", currentUserID).Find(&friends).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	var contacts []contactResponse
	for _, f := range friends {
		var profile model.UserProfile
		var account model.UserAccount

		dbConn.First(&account, f.FriendID)
		dbConn.Where("user_id = ?", f.FriendID).First(&profile)

		nickname := profile.Nickname
		if nickname == "" {
			nickname = account.Username
		}

		contacts = append(contacts, contactResponse{
			ID:     fmt.Sprintf("u_%d", account.ID),
			Name:   nickname,
			Avatar: profile.Avatar,
			Online: ws.IsOnline(uint64(account.ID)),
		})
	}

	c.JSON(200, contacts)
}
