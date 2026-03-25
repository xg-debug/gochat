package service

import (
	"errors"
	"strings"
	"time"

	"gochat/internal/dto/request"
	"gochat/internal/dto/response"
	"gochat/internal/model"

	"gorm.io/gorm"
)

type GroupService struct {
	db *gorm.DB
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{db: db}
}

func (s *GroupService) CreateGroup(currentUserID int64, req request.CreateGroupRequest) (response.GroupResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return response.GroupResponse{}, errors.New("group name required")
	}
	group := model.ChatGroup{Name: name, OwnerID: currentUserID, CreatedAt: time.Now()}
	if err := s.db.Create(&group).Error; err != nil {
		return response.GroupResponse{}, err
	}
	member := model.GroupMember{GroupID: group.ID, UserID: currentUserID, Role: 2, CreatedAt: time.Now()}
	if err := s.db.Create(&member).Error; err != nil {
		return response.GroupResponse{}, err
	}
	return response.GroupResponse{ID: group.ID, Name: group.Name, Avatar: group.Avatar, Notice: group.Notice, Role: 2}, nil
}

func (s *GroupService) SearchGroup(keyword string) ([]response.GroupResponse, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("keyword required")
	}
	var groups []model.ChatGroup
	if err := s.db.Where("name LIKE ?", "%"+keyword+"%").Limit(20).Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make([]response.GroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, response.GroupResponse{ID: g.ID, Name: g.Name, Avatar: g.Avatar, Notice: g.Notice})
	}
	return result, nil
}

func (s *GroupService) JoinGroup(currentUserID int64, req request.JoinGroupRequest) error {
	if req.GroupID <= 0 {
		return errors.New("invalid group id")
	}
	var group model.ChatGroup
	if err := s.db.First(&group, req.GroupID).Error; err != nil {
		return errors.New("group not found")
	}
	var count int64
	s.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).Count(&count)
	if count > 0 {
		return nil
	}
	member := model.GroupMember{GroupID: req.GroupID, UserID: currentUserID, Role: 0, CreatedAt: time.Now()}
	return s.db.Create(&member).Error
}

func (s *GroupService) ListGroups(currentUserID int64) ([]response.GroupResponse, error) {
	var members []model.GroupMember
	if err := s.db.Where("user_id = ?", currentUserID).Find(&members).Error; err != nil {
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
	if err := s.db.Where("id in ?", groupIDs).Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make([]response.GroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, response.GroupResponse{ID: g.ID, Name: g.Name, Avatar: g.Avatar, Notice: g.Notice})
	}
	return result, nil
}

func (s *GroupService) GetGroupProfile(currentUserID, groupID int64) (response.GroupResponse, error) {
	if groupID <= 0 {
		return response.GroupResponse{}, errors.New("invalid groupId")
	}
	var group model.ChatGroup
	if err := s.db.First(&group, groupID).Error; err != nil {
		return response.GroupResponse{}, errors.New("group not found")
	}
	role := int8(0)
	var member model.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, currentUserID).First(&member).Error; err == nil {
		role = member.Role
	}
	return response.GroupResponse{ID: group.ID, Name: group.Name, Avatar: group.Avatar, Notice: group.Notice, Role: role}, nil
}

func (s *GroupService) UpdateGroupProfile(currentUserID int64, req request.UpdateGroupProfileRequest) error {
	if req.GroupID <= 0 {
		return errors.New("invalid groupId")
	}
	var member model.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&member).Error; err != nil {
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
	return s.db.Model(&model.ChatGroup{}).Where("id = ?", req.GroupID).Updates(updates).Error
}

func (s *GroupService) ListGroupMembers(currentUserID, groupID int64) ([]response.GroupMemberResponse, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid groupId")
	}
	var count int64
	s.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, currentUserID).Count(&count)
	if count == 0 {
		return nil, errors.New("not in group")
	}

	var members []model.GroupMember
	if err := s.db.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	userIDs := make([]int64, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	var accounts []model.UserAccount
	_ = s.db.Where("id in ?", userIDs).Find(&accounts).Error
	accountMap := map[int64]model.UserAccount{}
	for _, a := range accounts {
		accountMap[a.ID] = a
	}
	var profiles []model.UserProfile
	_ = s.db.Where("user_id in ?", userIDs).Find(&profiles).Error
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

func (s *GroupService) KickGroupMember(currentUserID int64, req request.KickGroupMemberRequest) error {
	if req.GroupID <= 0 || req.UserID <= 0 {
		return errors.New("invalid params")
	}
	var current model.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&current).Error; err != nil {
		return errors.New("not in group")
	}
	if current.Role == 0 {
		return errors.New("permission denied")
	}
	if current.UserID == req.UserID {
		return errors.New("cannot kick yourself")
	}
	return s.db.Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Delete(&model.GroupMember{}).Error
}

func (s *GroupService) SetGroupAdmin(currentUserID int64, req request.SetGroupAdminRequest) error {
	if req.GroupID <= 0 || req.UserID <= 0 {
		return errors.New("invalid params")
	}
	var current model.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).First(&current).Error; err != nil {
		return errors.New("not in group")
	}
	if current.Role != 2 {
		return errors.New("only owner")
	}
	role := int8(0)
	if req.Action == "set" {
		role = 1
	}
	return s.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Update("role", role).Error
}

func (s *GroupService) ListInviteableFriends(currentUserID, groupID int64) ([]response.GroupMemberResponse, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid groupId")
	}
	var count int64
	s.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, currentUserID).Count(&count)
	if count == 0 {
		return nil, errors.New("not in group")
	}

	var friends []model.Friend
	if err := s.db.Where("user_id = ? AND status = 1", currentUserID).Find(&friends).Error; err != nil {
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
	_ = s.db.Where("group_id = ?", groupID).Find(&members).Error
	memberMap := map[int64]bool{}
	for _, m := range members {
		memberMap[m.UserID] = true
	}

	var accounts []model.UserAccount
	_ = s.db.Where("id in ?", friendIDs).Find(&accounts).Error
	accountMap := map[int64]model.UserAccount{}
	for _, a := range accounts {
		accountMap[a.ID] = a
	}
	var profiles []model.UserProfile
	_ = s.db.Where("user_id in ?", friendIDs).Find(&profiles).Error
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

func (s *GroupService) InviteGroupMember(currentUserID int64, req request.InviteGroupMemberRequest) error {
	if req.GroupID <= 0 || req.UserID <= 0 {
		return errors.New("invalid params")
	}
	var count int64
	s.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, currentUserID).Count(&count)
	if count == 0 {
		return errors.New("not in group")
	}
	var friendCount int64
	s.db.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ? AND status = 1", currentUserID, req.UserID).Count(&friendCount)
	if friendCount == 0 {
		return errors.New("can only invite friends")
	}
	var memberCount int64
	s.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", req.GroupID, req.UserID).Count(&memberCount)
	if memberCount > 0 {
		return nil
	}
	member := model.GroupMember{GroupID: req.GroupID, UserID: req.UserID, Role: 0, CreatedAt: time.Now()}
	return s.db.Create(&member).Error
}
