package model

import "time"

// ChatGroup 群聊
type ChatGroup struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"size:100;not null"`
	OwnerID   int64     `gorm:"not null;index"`
	CreatedAt time.Time
}

func (ChatGroup) TableName() string {
	return "chat_groups"
}
