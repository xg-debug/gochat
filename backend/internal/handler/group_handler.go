package handler

import (
	"fmt"
	"strings"
	"time"

	"gochat/internal/model"
	"gochat/internal/pkg/db"

	"github.com/gin-gonic/gin"
)

type createGroupRequest struct {
	Name string `json:"name"`
}

type joinGroupRequest struct {
	GroupID int64 `json:"groupId"`
}

type groupResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Avatar string `json:"avatar"`
	Notice string `json:"notice"`
	Role   int8   `json:"role"`
}

func CreateGroup(c *gin.Context) {
	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(400, gin.H{"error": "group name required"})
		return
	}

	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	group := model.ChatGroup{
		Name:      name,
		OwnerID:   currentUserID,
		CreatedAt: time.Now(),
	}
	if err := dbConn.Create(&group).Error; err != nil {
		c.JSON(500, gin.H{"error": "create group failed"})
		return
	}

	member := model.GroupMember{
		GroupID:   group.ID,
		UserID:    currentUserID,
		Role:      2,
		CreatedAt: time.Now(),
	}
	dbConn.Create(&member)

	c.JSON(200, groupResponse{ID: group.ID, Name: group.Name, Avatar: group.Avatar, Notice: group.Notice, Role: 2})
}

func SearchGroup(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		c.JSON(400, gin.H{"error": "keyword required"})
		return
	}
	dbConn := db.GetDB()

	var groups []model.ChatGroup
	if err := dbConn.Where("name LIKE ?", "%"+keyword+"%").Limit(20).Find(&groups).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	result := make([]groupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, groupResponse{ID: g.ID, Name: g.Name, Avatar: g.Avatar, Notice: g.Notice})
	}
	c.JSON(200, result)
}

func JoinGroup(c *gin.Context) {
	var req joinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if req.GroupID <= 0 {
		c.JSON(400, gin.H{"error": "invalid group id"})
		return
	}
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var group model.ChatGroup
	if err := dbConn.First(&group, req.GroupID).Error; err != nil {
		c.JSON(404, gin.H{"error": "group not found"})
		return
	}

	var count int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).Count(&count)
	if count > 0 {
		c.JSON(200, gin.H{"message": "already joined"})
		return
	}

	member := model.GroupMember{
		GroupID:   req.GroupID,
		UserID:    currentUserID,
		Role:      0,
		CreatedAt: time.Now(),
	}
	if err := dbConn.Create(&member).Error; err != nil {
		c.JSON(500, gin.H{"error": "join failed"})
		return
	}

	c.JSON(200, gin.H{"message": "joined"})
}

func ListGroups(c *gin.Context) {
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()

	var members []model.GroupMember
	if err := dbConn.Where("user_id = ?", currentUserID).Find(&members).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	groupIDs := make([]int64, 0, len(members))
	for _, m := range members {
		groupIDs = append(groupIDs, m.GroupID)
	}
	if len(groupIDs) == 0 {
		c.JSON(200, []groupResponse{})
		return
	}

	var groups []model.ChatGroup
	if err := dbConn.Where("id in ?", groupIDs).Find(&groups).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	result := make([]groupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, groupResponse{ID: g.ID, Name: g.Name, Avatar: g.Avatar, Notice: g.Notice})
	}
	c.JSON(200, result)
}

type updateGroupRequest struct {
	GroupID int64  `json:"groupId"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Notice  string `json:"notice"`
}

func GetGroupProfile(c *gin.Context) {
	groupIDStr := strings.TrimSpace(c.Query("groupId"))
	if groupIDStr == "" {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	var groupID int64
	if _, err := fmt.Sscanf(groupIDStr, "%d", &groupID); err != nil || groupID <= 0 {
		c.JSON(400, gin.H{"error": "invalid groupId"})
		return
	}
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()
	var group model.ChatGroup
	if err := dbConn.First(&group, groupID).Error; err != nil {
		c.JSON(404, gin.H{"error": "group not found"})
		return
	}
	var member model.GroupMember
	role := int8(0)
	if err := dbConn.Where("group_id = ? AND user_id = ?", groupID, currentUserID).First(&member).Error; err == nil {
		role = member.Role
	}
	c.JSON(200, groupResponse{ID: group.ID, Name: group.Name, Avatar: group.Avatar, Notice: group.Notice, Role: role})
}

func UpdateGroupProfile(c *gin.Context) {
	var req updateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if req.GroupID <= 0 {
		c.JSON(400, gin.H{"error": "invalid groupId"})
		return
	}
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()
	var member model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&member).Error; err != nil {
		c.JSON(403, gin.H{"error": "not in group"})
		return
	}
	if member.Role == 0 {
		c.JSON(403, gin.H{"error": "permission denied"})
		return
	}
	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Avatar) != "" {
		updates["avatar"] = strings.TrimSpace(req.Avatar)
	}
	if req.Notice != "" {
		updates["notice"] = req.Notice
	}
	if len(updates) == 0 {
		c.JSON(200, gin.H{"message": "no changes"})
		return
	}
	if err := dbConn.Model(&model.ChatGroup{}).Where("id = ?", req.GroupID).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

type groupMemberResponse struct {
	UserID   int64  `json:"userId"`
	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Role     int8   `json:"role"`
}

func ListGroupMembers(c *gin.Context) {
	groupIDStr := strings.TrimSpace(c.Query("groupId"))
	if groupIDStr == "" {
		c.JSON(400, gin.H{"error": "groupId required"})
		return
	}
	var groupID int64
	if _, err := fmt.Sscanf(groupIDStr, "%d", &groupID); err != nil || groupID <= 0 {
		c.JSON(400, gin.H{"error": "invalid groupId"})
		return
	}
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()
	var count int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, currentUserID).Count(&count)
	if count == 0 {
		c.JSON(403, gin.H{"error": "not in group"})
		return
	}
	var members []model.GroupMember
	if err := dbConn.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	userIDs := make([]int64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	var accounts []model.UserAccount
	dbConn.Where("id in ?", userIDs).Find(&accounts)
	accountMap := map[int64]model.UserAccount{}
	for _, a := range accounts {
		accountMap[a.ID] = a
	}
	var profiles []model.UserProfile
	dbConn.Where("user_id in ?", userIDs).Find(&profiles)
	profileMap := map[int64]model.UserProfile{}
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}
	result := make([]groupMemberResponse, 0, len(members))
	for _, m := range members {
		account := accountMap[m.UserID]
		profile := profileMap[m.UserID]
		nickname := profile.Nickname
		if nickname == "" {
			nickname = account.Username
		}
		result = append(result, groupMemberResponse{
			UserID:   m.UserID,
			Nickname: nickname,
			Username: account.Username,
			Avatar:   profile.Avatar,
			Role:     m.Role,
		})
	}
	c.JSON(200, result)
}

type kickMemberRequest struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
}

func KickGroupMember(c *gin.Context) {
	var req kickMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if req.GroupID <= 0 || req.UserID <= 0 {
		c.JSON(400, gin.H{"error": "invalid params"})
		return
	}
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()
	var current model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&current).Error; err != nil {
		c.JSON(403, gin.H{"error": "not in group"})
		return
	}
	if current.Role == 0 {
		c.JSON(403, gin.H{"error": "permission denied"})
		return
	}
	if current.UserID == req.UserID {
		c.JSON(400, gin.H{"error": "cannot kick yourself"})
		return
	}
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Delete(&model.GroupMember{}).Error; err != nil {
		c.JSON(500, gin.H{"error": "kick failed"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

type adminRequest struct {
	GroupID int64 `json:"groupId"`
	UserID  int64 `json:"userId"`
	Action  string `json:"action"`
}

func SetGroupAdmin(c *gin.Context) {
	var req adminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if req.GroupID <= 0 || req.UserID <= 0 {
		c.JSON(400, gin.H{"error": "invalid params"})
		return
	}
	currentUserID := c.GetInt64("userID")
	dbConn := db.GetDB()
	var current model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&current).Error; err != nil {
		c.JSON(403, gin.H{"error": "not in group"})
		return
	}
	if current.Role != 2 {
		c.JSON(403, gin.H{"error": "only owner"})
		return
	}
	role := int8(0)
	if req.Action == "set" {
		role = 1
	} else if req.Action == "unset" {
		role = 0
	} else {
		c.JSON(400, gin.H{"error": "invalid action"})
		return
	}
	if err := dbConn.Model(&model.GroupMember{}).
		Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).
		Update("role", role).Error; err != nil {
		c.JSON(500, gin.H{"error": "update failed"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
