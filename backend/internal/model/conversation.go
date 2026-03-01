package model

import "time"

type Conversation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	UserID      int64
	PeerID      int64
	ChatType    int8
	UnreadCount int
	UpdatedAt   time.Time
}

func (Conversation) TableName() string {
	return "conversations"
}
