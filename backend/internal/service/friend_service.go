package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gochat/internal/dto/request"
	"gochat/internal/dto/response"
	"gochat/internal/model"

	"gorm.io/gorm"
)

type FriendService struct {
	db       *gorm.DB
	isOnline func(userID uint64) bool
}

func NewFriendService(db *gorm.DB, isOnline func(userID uint64) bool) *FriendService {
	if isOnline == nil {
		isOnline = func(uint64) bool { return false }
	}
	return &FriendService{db: db, isOnline: isOnline}
}

func (s *FriendService) SearchUser(currentUserID int64, query request.SearchUserQuery) ([]response.UserSearchResult, error) {
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}

	var users []model.UserAccount
	if err := s.db.Where("username LIKE ?", "%"+keyword+"%").Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}

	results := make([]response.UserSearchResult, 0, len(users))
	for _, u := range users {
		if u.ID == currentUserID {
			continue
		}

		var profile model.UserProfile
		_ = s.db.Where("user_id = ?", u.ID).First(&profile).Error

		var count int64
		s.db.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, u.ID).Count(&count)

		pending := false
		pendingFromMe := false
		if count == 0 {
			var pendingReq model.FriendRequest
			err := s.db.Where("status = 0 AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))", currentUserID, u.ID, u.ID, currentUserID).
				First(&pendingReq).Error
			if err == nil {
				pending = true
				pendingFromMe = pendingReq.FromUserID == currentUserID
			}
		}

		nickname := profile.Nickname
		if nickname == "" {
			nickname = u.Username
		}
		results = append(results, response.UserSearchResult{
			ID:            u.ID,
			Username:      u.Username,
			Nickname:      nickname,
			Avatar:        profile.Avatar,
			IsFriend:      count > 0,
			Pending:       pending,
			PendingFromMe: pendingFromMe,
		})
	}
	return results, nil
}

func (s *FriendService) SendFriendRequest(currentUserID int64, req request.SendFriendRequest) error {
	if currentUserID == req.ToUserID {
		return errors.New("cannot add yourself")
	}

	var count int64
	s.db.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, req.ToUserID).Count(&count)
	if count > 0 {
		return errors.New("already friends")
	}

	var existing model.FriendRequest
	if err := s.db.Where("from_user_id = ? AND to_user_id = ? AND status = 0", currentUserID, req.ToUserID).First(&existing).Error; err == nil {
		return nil
	}

	var reverse model.FriendRequest
	if err := s.db.Where("from_user_id = ? AND to_user_id = ? AND status = 0", req.ToUserID, currentUserID).First(&reverse).Error; err == nil {
		return errors.New("user already requested you")
	}

	if err := s.db.Create(&model.FriendRequest{FromUserID: currentUserID, ToUserID: req.ToUserID, Status: 0}).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil
		}
		return err
	}
	return nil
}

func (s *FriendService) ListFriendRequests(currentUserID int64) ([]response.FriendRequestItem, error) {
	var requests []model.FriendRequest
	if err := s.db.Where("to_user_id = ? AND status = 0", currentUserID).Find(&requests).Error; err != nil {
		return nil, err
	}

	result := make([]response.FriendRequestItem, 0, len(requests))
	for _, r := range requests {
		var account model.UserAccount
		var profile model.UserProfile
		_ = s.db.First(&account, r.FromUserID).Error
		_ = s.db.Where("user_id = ?", r.FromUserID).First(&profile).Error
		nickname := profile.Nickname
		if nickname == "" {
			nickname = account.Username
		}
		result = append(result, response.FriendRequestItem{
			ID:         r.ID,
			FromUserID: r.FromUserID,
			Username:   account.Username,
			Nickname:   nickname,
			Avatar:     profile.Avatar,
			Time:       r.CreatedAt.Unix(),
		})
	}
	return result, nil
}

func (s *FriendService) HandleFriendRequest(currentUserID int64, req request.HandleFriendRequest) error {
	var friendReq model.FriendRequest
	if err := s.db.First(&friendReq, req.RequestID).Error; err != nil {
		return errors.New("request not found")
	}
	if friendReq.ToUserID != currentUserID {
		return errors.New("permission denied")
	}
	if friendReq.Status != 0 {
		return errors.New("request already handled")
	}

	tx := s.db.Begin()
	status := int8(2)
	if req.Action == "accept" {
		status = 1
	}
	if err := tx.Model(&friendReq).Update("status", status).Error; err != nil {
		tx.Rollback()
		return err
	}
	if req.Action == "accept" {
		f1 := model.Friend{UserID: friendReq.FromUserID, FriendID: friendReq.ToUserID, Status: 1, CreatedAt: time.Now()}
		f2 := model.Friend{UserID: friendReq.ToUserID, FriendID: friendReq.FromUserID, Status: 1, CreatedAt: time.Now()}
		if err := tx.Create(&f1).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&f2).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (s *FriendService) DeleteFriend(currentUserID int64, friendID int64) error {
	if friendID <= 0 {
		return errors.New("invalid friend id")
	}
	tx := s.db.Begin()
	if err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", currentUserID, friendID, friendID, currentUserID).Delete(&model.Friend{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *FriendService) BlockFriend(currentUserID int64, friendID int64) error {
	if friendID <= 0 {
		return errors.New("invalid friend id")
	}
	return s.db.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, friendID).Update("status", 0).Error
}

func (s *FriendService) UnblockFriend(currentUserID int64, friendID int64) error {
	if friendID <= 0 {
		return errors.New("invalid friend id")
	}
	return s.db.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, friendID).Update("status", 1).Error
}

func (s *FriendService) GetContacts(currentUserID int64) ([]response.ContactResponse, error) {
	var friends []model.Friend
	if err := s.db.Where("user_id = ? AND status = 1", currentUserID).Find(&friends).Error; err != nil {
		return nil, err
	}
	if len(friends) == 0 {
		return []response.ContactResponse{}, nil
	}

	friendIDs := make([]int64, 0, len(friends))
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendID)
	}

	var accounts []model.UserAccount
	if err := s.db.Where("id in ?", friendIDs).Find(&accounts).Error; err != nil {
		return nil, err
	}

	var profiles []model.UserProfile
	_ = s.db.Where("user_id in ?", friendIDs).Find(&profiles).Error
	profileMap := make(map[int64]model.UserProfile, len(profiles))
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}

	result := make([]response.ContactResponse, 0, len(accounts))
	for _, account := range accounts {
		profile := profileMap[account.ID]
		name := strings.TrimSpace(profile.Nickname)
		if name == "" {
			name = account.Username
		}
		result = append(result, response.ContactResponse{
			ID:     fmt.Sprintf("u_%d", account.ID),
			Name:   name,
			Avatar: profile.Avatar,
			Online: s.isOnline(uint64(account.ID)),
		})
	}
	return result, nil
}
