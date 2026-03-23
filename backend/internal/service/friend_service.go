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

type friendService struct{}

var FriendService = &friendService{}

func (s *friendService) SearchUser(currentUserID int64, query request.SearchUserQuery) ([]response.UserSearchResult, error) {
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}

	var users []model.UserAccount
	if err := dbConn.Where("username LIKE ?", "%"+keyword+"%").Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}

	results := make([]response.UserSearchResult, 0, len(users))
	for _, u := range users {
		if u.ID == currentUserID {
			continue
		}

		var profile model.UserProfile
		_ = dbConn.Where("user_id = ?", u.ID).First(&profile).Error

		var count int64
		dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, u.ID).Count(&count)

		pending := false
		pendingFromMe := false
		if count == 0 {
			var pendingReq model.FriendRequest
			err := dbConn.Where("status = 0 AND ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))", currentUserID, u.ID, u.ID, currentUserID).
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

func (s *friendService) SendFriendRequest(currentUserID int64, req request.SendFriendRequest) error {
	if currentUserID == req.ToUserID {
		return errors.New("cannot add yourself")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}

	var count int64
	dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, req.ToUserID).Count(&count)
	if count > 0 {
		return errors.New("already friends")
	}

	var existing model.FriendRequest
	if err := dbConn.Where("from_user_id = ? AND to_user_id = ? AND status = 0", currentUserID, req.ToUserID).First(&existing).Error; err == nil {
		return nil
	}

	var reverse model.FriendRequest
	if err := dbConn.Where("from_user_id = ? AND to_user_id = ? AND status = 0", req.ToUserID, currentUserID).First(&reverse).Error; err == nil {
		return errors.New("user already requested you")
	}

	return dbConn.Create(&model.FriendRequest{FromUserID: currentUserID, ToUserID: req.ToUserID, Status: 0}).Error
}

func (s *friendService) ListFriendRequests(currentUserID int64) ([]response.FriendRequestItem, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	var requests []model.FriendRequest
	if err := dbConn.Where("to_user_id = ? AND status = 0", currentUserID).Find(&requests).Error; err != nil {
		return nil, err
	}

	result := make([]response.FriendRequestItem, 0, len(requests))
	for _, r := range requests {
		var account model.UserAccount
		var profile model.UserProfile
		_ = dbConn.First(&account, r.FromUserID).Error
		_ = dbConn.Where("user_id = ?", r.FromUserID).First(&profile).Error
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

func (s *friendService) HandleFriendRequest(currentUserID int64, req request.HandleFriendRequest) error {
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	var friendReq model.FriendRequest
	if err := dbConn.First(&friendReq, req.RequestID).Error; err != nil {
		return errors.New("request not found")
	}
	if friendReq.ToUserID != currentUserID {
		return errors.New("permission denied")
	}
	if friendReq.Status != 0 {
		return errors.New("request already handled")
	}

	tx := dbConn.Begin()
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

func (s *friendService) DeleteFriend(currentUserID int64, friendID int64) error {
	if friendID <= 0 {
		return errors.New("invalid friend id")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	tx := dbConn.Begin()
	if err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", currentUserID, friendID, friendID, currentUserID).Delete(&model.Friend{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *friendService) BlockFriend(currentUserID int64, friendID int64) error {
	if friendID <= 0 {
		return errors.New("invalid friend id")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	return dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, friendID).Update("status", 0).Error
}

func (s *friendService) UnblockFriend(currentUserID int64, friendID int64) error {
	if friendID <= 0 {
		return errors.New("invalid friend id")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return errors.New("db not ready")
	}
	return dbConn.Model(&model.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, friendID).Update("status", 1).Error
}

func (s *friendService) GetContacts(currentUserID int64) ([]response.ContactResponse, error) {
	// 统一复用 user service 的 contacts 逻辑
	return UserService.GetContacts(currentUserID)
}
