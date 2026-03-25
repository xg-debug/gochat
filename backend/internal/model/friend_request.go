package model

import "time"

// FriendRequest 好友申请
type FriendRequest struct {
	ID         int64 `gorm:"primaryKey;autoIncrement"`
	FromUserID int64 `gorm:"not null;uniqueIndex:uk_friend_req_status"`
	ToUserID   int64 `gorm:"not null;uniqueIndex:uk_friend_req_status"`
	Status     int8  `gorm:"default:0;comment:0待处理 1同意 2拒绝;uniqueIndex:uk_friend_req_status"`
	CreatedAt  time.Time
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
