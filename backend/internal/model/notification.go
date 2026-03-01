package model

import "time"

// Notification 系统通知
type Notification struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;index"`
	Content   string    `gorm:"size:500"`
	IsRead    int8      `gorm:"default:0;comment:0未读 1已读"`
	CreatedAt time.Time
}

func (Notification) TableName() string {
	return "notifications"
}
