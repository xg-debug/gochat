package service

import (
	"errors"
	"strings"
	"time"

	"gochat/internal/dto/request"
	"gochat/internal/dto/response"
	"gochat/internal/model"
	"gochat/internal/pkg/db"
)

type groupService struct{}

var GroupService = &groupService{}

func (s *groupService) CreateGroup(currentUserID int64, req request.CreateGroupRequest) (response.GroupResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return response.GroupResponse{}, errors.New("group name required")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return response.GroupResponse{}, errors.New("db not ready")
	}
	group := model.ChatGroup{Name: name, OwnerID: currentUserID, CreatedAt: time.Now()}
	if err := dbConn.Create(&group).Error; err != nil {
		return response.GroupResponse{}, err
	}
	member := model.GroupMember{GroupID: group.ID, UserID: currentUserID, Role: 2, CreatedAt: time.Now()}
	if err := dbConn.Create(&member).Error; err != nil {
		return response.GroupResponse{}, err
	}
	return response.GroupResponse{ID: group.ID, Name: group.Name, Avatar: group.Avatar, Notice: group.Notice, Role: 2}, nil
}

func (s *groupService) SearchGroup(keyword string) ([]response.GroupResponse, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("keyword required")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	var groups []model.ChatGroup
	if err := dbConn.Where("name LIKE ?", "%"+keyword+"%").Limit(20).Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make([]response.GroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, response.GroupResponse{ID: g.ID, Name: g.Name, Avatar: g.Avatar, Notice: g.Notice})
	}
	return result, nil
}

func (s *groupService) JoinGroup(currentUserID int64, req request.JoinGroupRequest) error {
	if req.GroupID <= 0 {
		return errors.New("invalid group id")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	var group model.ChatGroup
	if err := dbConn.First(&group, req.GroupID).Error; err != nil {
		return errors.New("group not found")
	}
	var count int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).Count(&count)
	if count > 0 {
		return nil
	}
	member := model.GroupMember{GroupID: req.GroupID, UserID: currentUserID, Role: 0, CreatedAt: time.Now()}
	return dbConn.Create(&member).Error
}

func (s *groupService) ListGroups(currentUserID int64) ([]response.GroupResponse, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	var members []model.GroupMember
	if err := dbConn.Where("user_id = ?", currentUserID).Find(&members).Error; err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []response.GroupResponse{}, nil
	}
	groupIDs := make([]int64, 0, len(members))
	for _, m := range members {
		groupIDs = append(groupIDs, m.GroupID)
	}
	var groups []model.ChatGroup
	if err := dbConn.Where("id in ?", groupIDs).Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make([]response.GroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, response.GroupResponse{ID: g.ID, Name: g.Name, Avatar: g.Avatar, Notice: g.Notice})
	}
	return result, nil
}

func (s *groupService) GetGroupProfile(currentUserID, groupID int64) (response.GroupResponse, error) {
	if groupID <= 0 {
		return response.GroupResponse{}, errors.New("invalid groupId")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return response.GroupResponse{}, errors.New("db not ready")
	}
	var group model.ChatGroup
	if err := dbConn.First(&group, groupID).Error; err != nil {
		return response.GroupResponse{}, errors.New("group not found")
	}
	role := int8(0)
	var member model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", groupID, currentUserID).First(&member).Error; err == nil {
		role = member.Role
	}
	return response.GroupResponse{ID: group.ID, Name: group.Name, Avatar: group.Avatar, Notice: group.Notice, Role: role}, nil
}

func (s *groupService) UpdateGroupProfile(currentUserID int64, req request.UpdateGroupProfileRequest) error {
	if req.GroupID <= 0 {
		return errors.New("invalid groupId")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	var member model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&member).Error; err != nil {
		return errors.New("not in group")
	}
	if member.Role == 0 {
		return errors.New("permission denied")
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
		return nil
	}
	return dbConn.Model(&model.ChatGroup{}).Where("id = ?", req.GroupID).Updates(updates).Error
}

func (s *groupService) ListGroupMembers(currentUserID, groupID int64) ([]response.GroupMemberResponse, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid groupId")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	var count int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, currentUserID).Count(&count)
	if count == 0 {
		return nil, errors.New("not in group")
	}

	var members []model.GroupMember
	if err := dbConn.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	userIDs := make([]int64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	var accounts []model.UserAccount
	_ = dbConn.Where("id in ?", userIDs).Find(&accounts).Error
	accountMap := map[int64]model.UserAccount{}
	for _, a := range accounts {
		accountMap[a.ID] = a
	}
	var profiles []model.UserProfile
	_ = dbConn.Where("user_id in ?", userIDs).Find(&profiles).Error
	profileMap := map[int64]model.UserProfile{}
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}

	result := make([]response.GroupMemberResponse, 0, len(members))
	for _, m := range members {
		account := accountMap[m.UserID]
		profile := profileMap[m.UserID]
		nickname := profile.Nickname
		if nickname == "" {
			nickname = account.Username
		}
		result = append(result, response.GroupMemberResponse{UserID: m.UserID, Nickname: nickname, Username: account.Username, Avatar: profile.Avatar, Role: m.Role})
	}
	return result, nil
}

func (s *groupService) KickGroupMember(currentUserID int64, req request.KickGroupMemberRequest) error {
	if req.GroupID <= 0 || req.UserID <= 0 {
		return errors.New("invalid params")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	var current model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&current).Error; err != nil {
		return errors.New("not in group")
	}
	if current.Role == 0 {
		return errors.New("permission denied")
	}
	if current.UserID == req.UserID {
		return errors.New("cannot kick yourself")
	}
	return dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Delete(&model.GroupMember{}).Error
}

func (s *groupService) SetGroupAdmin(currentUserID int64, req request.SetGroupAdminRequest) error {
	if req.GroupID <= 0 || req.UserID <= 0 {
		return errors.New("invalid params")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	var current model.GroupMember
	if err := dbConn.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&current).Error; err != nil {
		return errors.New("not in group")
	}
	if current.Role != 2 {
		return errors.New("only owner")
	}
	role := int8(0)
	if req.Action == "set" {
		role = 1
	}
	return dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Update("role", role).Error
}

func (s *groupService) ListInviteableFriends(currentUserID, groupID int64) ([]response.GroupMemberResponse, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid groupId")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	var count int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, currentUserID).Count(&count)
	if count == 0 {
		return nil, errors.New("not in group")
	}

	var friends []model.Friend
	if err := dbConn.Where("user_id = ? AND status = 1", currentUserID).Find(&friends).Error; err != nil {
		return nil, err
	}
	friendIDs := make([]int64, 0, len(friends))
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendID)
	}
	if len(friendIDs) == 0 {
		return []response.GroupMemberResponse{}, nil
	}

	var members []model.GroupMember
	_ = dbConn.Where("group_id = ?", groupID).Find(&members).Error
	memberMap := map[int64]bool{}
	for _, m := range members {
		memberMap[m.UserID] = true
	}

	var accounts []model.UserAccount
	_ = dbConn.Where("id in ?", friendIDs).Find(&accounts).Error
	accountMap := map[int64]model.UserAccount{}
	for _, a := range accounts {
		accountMap[a.ID] = a
	}
	var profiles []model.UserProfile
	_ = dbConn.Where("user_id in ?", friendIDs).Find(&profiles).Error
	profileMap := map[int64]model.UserProfile{}
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}

	result := make([]response.GroupMemberResponse, 0, len(friendIDs))
	for _, id := range friendIDs {
		if memberMap[id] {
			continue
		}
		account, ok := accountMap[id]
		if !ok {
			continue
		}
		profile := profileMap[id]
		nickname := profile.Nickname
		if nickname == "" {
			nickname = account.Username
		}
		result = append(result, response.GroupMemberResponse{UserID: id, Nickname: nickname, Username: account.Username, Avatar: profile.Avatar, Role: 0})
	}
	return result, nil
}

func (s *groupService) InviteGroupMember(currentUserID int64, req request.InviteGroupMemberRequest) error {
	if req.GroupID <= 0 || req.UserID <= 0 {
		return errors.New("invalid params")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	var count int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).Count(&count)
	if count == 0 {
		return errors.New("not in group")
	}
	var friendCount int64
	dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ? AND status = 1", currentUserID, req.UserID).Count(&friendCount)
	if friendCount == 0 {
		return errors.New("can only invite friends")
	}
	var memberCount int64
	dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Count(&memberCount)
	if memberCount > 0 {
		return nil
	}
	member := model.GroupMember{GroupID: req.GroupID, UserID: req.UserID, Role: 0, CreatedAt: time.Now()}
	return dbConn.Create(&member).Error
}
