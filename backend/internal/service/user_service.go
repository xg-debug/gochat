package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gochat/internal/dto/request"
	"gochat/internal/dto/response"
	"gochat/internal/model"
	"gochat/internal/pkg/auth"
	"gochat/internal/pkg/db"
	"gochat/internal/ws"

	"gorm.io/gorm"
)

type userService struct{}

var UserService = &userService{}

func (s *userService) Login(req request.LoginRequest) (string, response.UserResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return "", response.UserResponse{}, errors.New("username or password missing")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return "", response.UserResponse{}, errors.New("db not ready")
	}

	var account model.UserAccount
	if err := dbConn.Where("username = ?", req.Username).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", response.UserResponse{}, errors.New("invalid credentials")
		}
		return "", response.UserResponse{}, err
	}
	if err := auth.CheckPassword(account.PasswordHash, req.Password); err != nil {
		return "", response.UserResponse{}, errors.New("invalid credentials")
	}

	profile, err := s.loadOrEmptyProfile(int64(account.ID))
	if err != nil {
		return "", response.UserResponse{}, err
	}
	userResp := buildUserResponse(account, profile)

	token, err := auth.GenerateToken(int64(account.ID), account.Username)
	if err != nil {
		return "", response.UserResponse{}, err
	}
	return token, userResp, nil
}

func (s *userService) Register(req request.RegisterRequest) (string, response.UserResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Nickname = strings.TrimSpace(req.Nickname)
	if req.Username == "" || req.Password == "" {
		return "", response.UserResponse{}, errors.New("username or password missing")
	}
	dbConn := db.GetDB()
	if dbConn == nil {
		return "", response.UserResponse{}, errors.New("db not ready")
	}

	var existing model.UserAccount
	if err := dbConn.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return "", response.UserResponse{}, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", response.UserResponse{}, err
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return "", response.UserResponse{}, err
	}
	account := model.UserAccount{Username: req.Username, PasswordHash: hash, Status: 1}
	profile := model.UserProfile{Nickname: req.Nickname}
	if profile.Nickname == "" {
		profile.Nickname = req.Username
	}

	tx := dbConn.Begin()
	if err := tx.Create(&account).Error; err != nil {
		tx.Rollback()
		return "", response.UserResponse{}, err
	}
	profile.UserID = int64(account.ID)
	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		return "", response.UserResponse{}, err
	}
	if err := tx.Commit().Error; err != nil {
		return "", response.UserResponse{}, err
	}

	token, err := auth.GenerateToken(int64(account.ID), account.Username)
	if err != nil {
		return "", response.UserResponse{}, err
	}
	return token, buildUserResponse(account, profile), nil
}

func (s *userService) Profile(userID int64) (response.UserResponse, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return response.UserResponse{}, errors.New("db not ready")
	}
	var account model.UserAccount
	if err := dbConn.Where("id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.UserResponse{}, errors.New("user not found")
		}
		return response.UserResponse{}, err
	}
	profile, err := s.loadOrEmptyProfile(userID)
	if err != nil {
		return response.UserResponse{}, err
	}
	return buildUserResponse(account, profile), nil
}

func (s *userService) UpdateProfile(userID int64, req request.UpdateProfileRequest) (response.UserResponse, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return response.UserResponse{}, errors.New("db not ready")
	}

	var profile model.UserProfile
	if err := dbConn.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			profile = model.UserProfile{UserID: userID}
			if err := dbConn.Create(&profile).Error; err != nil {
				return response.UserResponse{}, err
			}
		} else {
			return response.UserResponse{}, err
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
	if req.Phone != nil {
		updates["phone"] = strings.TrimSpace(*req.Phone)
	}
	if req.Location != nil {
		updates["location"] = strings.TrimSpace(*req.Location)
	}
	if req.Birthday != nil {
		birthdayText := strings.TrimSpace(*req.Birthday)
		if birthdayText == "" {
			updates["birthday"] = nil
		} else {
			birthday, err := time.Parse("2006-01-02", birthdayText)
			if err != nil {
				return response.UserResponse{}, errors.New("invalid birthday format, expected YYYY-MM-DD")
			}
			updates["birthday"] = birthday
		}
	}

	if len(updates) > 0 {
		if err := dbConn.Model(&model.UserProfile{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			return response.UserResponse{}, err
		}
	}

	return s.Profile(userID)
}

func (s *userService) GetContacts(userID int64) ([]response.ContactResponse, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	var accounts []model.UserAccount
	if err := dbConn.Where("id <> ?", userID).Order("id asc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, int64(account.ID))
	}
	profileMap := map[int64]model.UserProfile{}
	if len(ids) > 0 {
		var profiles []model.UserProfile
		if err := dbConn.Where("user_id in ?", ids).Find(&profiles).Error; err == nil {
			for _, p := range profiles {
				profileMap[p.UserID] = p
			}
		}
	}

	result := make([]response.ContactResponse, 0, len(accounts))
	for _, account := range accounts {
		profile := profileMap[int64(account.ID)]
		name := strings.TrimSpace(profile.Nickname)
		if name == "" {
			name = account.Username
		}
		result = append(result, response.ContactResponse{
			ID:     fmt.Sprintf("u_%d", account.ID),
			Name:   name,
			Avatar: profile.Avatar,
			Online: ws.IsOnline(uint64(account.ID)),
		})
	}
	return result, nil
}

func (s *userService) GetConversations(userID int64) ([]response.ConversationResponse, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}

	var friends []model.Friend
	if err := dbConn.Where("user_id = ? AND status = 1", userID).Find(&friends).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(friends))
	for _, f := range friends {
		ids = append(ids, f.FriendID)
	}

	var accounts []model.UserAccount
	if len(ids) > 0 {
		if err := dbConn.Where("id in ?", ids).Order("id asc").Find(&accounts).Error; err != nil {
			return nil, err
		}
	}

	profileMap := map[int64]model.UserProfile{}
	if len(ids) > 0 {
		var profiles []model.UserProfile
		if err := dbConn.Where("user_id in ?", ids).Find(&profiles).Error; err == nil {
			for _, p := range profiles {
				profileMap[p.UserID] = p
			}
		}
	}

	result := make([]response.ConversationResponse, 0, len(accounts)+4)
	for _, account := range accounts {
		profile := profileMap[int64(account.ID)]
		name := strings.TrimSpace(profile.Nickname)
		if name == "" {
			name = account.Username
		}
		lastMessage := s.latestSingleMessageText(userID, int64(account.ID))
		result = append(result, response.ConversationResponse{
			ID:          fmt.Sprintf("u_%d", account.ID),
			Name:        name,
			Avatar:      profile.Avatar,
			LastMessage: lastMessage,
			Unread:      s.countUnreadSingle(userID, int64(account.ID)),
			Online:      ws.IsOnline(uint64(account.ID)),
		})
	}

	var members []model.GroupMember
	if err := dbConn.Where("user_id = ?", userID).Find(&members).Error; err == nil && len(members) > 0 {
		groupIDs := make([]int64, 0, len(members))
		for _, m := range members {
			groupIDs = append(groupIDs, m.GroupID)
		}
		var groups []model.ChatGroup
		if err := dbConn.Where("id in ?", groupIDs).Find(&groups).Error; err == nil {
			for _, g := range groups {
				result = append(result, response.ConversationResponse{
					ID:          fmt.Sprintf("g_%d", g.ID),
					Name:        g.Name,
					Avatar:      g.Avatar,
					LastMessage: s.latestGroupMessageText(g.ID),
					Unread:      s.countUnreadGroup(userID, g.ID),
					Online:      false,
				})
			}
		}
	}
	return result, nil
}

func (s *userService) SearchConversations(userID int64, keyword string) ([]response.ConversationResponse, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return s.GetConversations(userID)
	}
	all, err := s.GetConversations(userID)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(keyword)
	result := make([]response.ConversationResponse, 0, len(all))
	for _, item := range all {
		if strings.Contains(strings.ToLower(item.Name), lower) || strings.Contains(strings.ToLower(item.LastMessage), lower) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *userService) GetMessages(userID int64, conversationID string) ([]response.MessageResponse, error) {
	dbConn := db.GetDB()
	if dbConn == nil {
		return nil, errors.New("db not ready")
	}
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return nil, errors.New("conversationId required")
	}

	var messages []model.Message
	if strings.HasPrefix(conversationID, "g_") {
		groupID, err := parseID(strings.TrimPrefix(conversationID, "g_"))
		if err != nil || groupID <= 0 {
			return nil, errors.New("invalid conversationId")
		}
		var count int64
		dbConn.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
		if count == 0 {
			return nil, errors.New("not in group")
		}
		if err := dbConn.Where("chat_type = ? AND to_id = ?", 2, groupID).Order("created_at asc").Find(&messages).Error; err != nil {
			return nil, err
		}
	} else {
		peerID, err := parseID(strings.TrimPrefix(conversationID, "u_"))
		if err != nil || peerID <= 0 {
			return nil, errors.New("invalid conversationId")
		}
		if err := dbConn.Where("(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)", userID, peerID, peerID, userID).
			Order("created_at asc").Find(&messages).Error; err != nil {
			return nil, err
		}
	}

	avatarMap := s.buildAvatarMap(messages)
	result := make([]response.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		content := msg.Content
		status := "delivered"
		if msg.Status == 1 {
			status = "read"
		} else if msg.Status == 2 {
			status = "revoked"
			content = "[已撤回]"
		}
		result = append(result, response.MessageResponse{
			ID:          fmt.Sprintf("m_%d", msg.ID),
			FromID:      fmt.Sprintf("u_%d", msg.FromID),
			FromAvatar:  avatarMap[msg.FromID],
			Content:     content,
			ContentType: msgTypeToContentType(msg.MsgType),
			Time:        msg.CreatedAt.UnixMilli(),
			Status:      status,
		})
	}
	return result, nil
}

func (s *userService) loadOrEmptyProfile(userID int64) (model.UserProfile, error) {
	dbConn := db.GetDB()
	var profile model.UserProfile
	if err := dbConn.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.UserProfile{}, nil
		}
		return model.UserProfile{}, err
	}
	return profile, nil
}

func buildUserResponse(account model.UserAccount, profile model.UserProfile) response.UserResponse {
	nickname := strings.TrimSpace(profile.Nickname)
	if nickname == "" {
		nickname = account.Username
	}
	return response.UserResponse{
		ID:        account.ID,
		Username:  account.Username,
		Nickname:  nickname,
		Avatar:    profile.Avatar,
		Signature: profile.Signature,
		Gender:    profile.Gender,
		Phone:     profile.Phone,
		Location:  profile.Location,
		Birthday:  formatBirthday(profile.Birthday),
	}
}

func formatBirthday(birthday *time.Time) string {
	if birthday == nil {
		return ""
	}
	return birthday.Format("2006-01-02")
}

func parseID(value string) (int64, error) {
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

func mapMsgPreview(msg model.Message) string {
	if msg.Status == 2 {
		return "[已撤回]"
	}
	switch msg.MsgType {
	case 2:
		return "[图片]"
	case 3:
		return "[文件]"
	case 4:
		return "[视频]"
	case 5:
		return "[语音]"
	default:
		return msg.Content
	}
}

func (s *userService) latestSingleMessageText(userID, peerID int64) string {
	dbConn := db.GetDB()
	var msg model.Message
	err := dbConn.Where("(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)", userID, peerID, peerID, userID).
		Order("created_at desc").Limit(1).Find(&msg).Error
	if err != nil || msg.ID == 0 {
		return ""
	}
	return mapMsgPreview(msg)
}

func (s *userService) latestGroupMessageText(groupID int64) string {
	dbConn := db.GetDB()
	var msg model.Message
	err := dbConn.Where("chat_type = ? AND to_id = ?", 2, groupID).Order("created_at desc").Limit(1).Find(&msg).Error
	if err != nil || msg.ID == 0 {
		return ""
	}
	return mapMsgPreview(msg)
}

func (s *userService) countUnreadSingle(userID, peerID int64) int {
	dbConn := db.GetDB()
	var count int64
	dbConn.Model(&model.Message{}).
		Where("from_id = ? AND to_id = ? AND status = 0", peerID, userID).
		Count(&count)
	return int(count)
}

func (s *userService) countUnreadGroup(userID, groupID int64) int {
	// 当前模型没有每用户群消息游标，先与现有逻辑保持一致返回0。
	return 0
}

func (s *userService) buildAvatarMap(messages []model.Message) map[int64]string {
	dbConn := db.GetDB()
	avatarMap := map[int64]string{}
	if len(messages) == 0 {
		return avatarMap
	}
	fromIDs := make([]int64, 0, len(messages))
	seen := map[int64]struct{}{}
	for _, message := range messages {
		if _, ok := seen[message.FromID]; ok {
			continue
		}
		seen[message.FromID] = struct{}{}
		fromIDs = append(fromIDs, message.FromID)
	}
	var profiles []model.UserProfile
	if err := dbConn.Where("user_id in ?", fromIDs).Find(&profiles).Error; err == nil {
		for _, p := range profiles {
			avatarMap[p.UserID] = p.Avatar
		}
	}
	return avatarMap
}
