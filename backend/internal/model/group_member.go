package model

import "time"

// GroupMember 群成员
type GroupMember struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	GroupID   int64     `gorm:"not null;index"`
	UserID    int64     `gorm:"not null;index"`
	Role      int8      `gorm:"default:0;comment:0成员 1管理员 2群主"`
	CreatedAt time.Time
}

func (GroupMember) TableName() string {
	return "group_members"
}
