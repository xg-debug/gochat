
package model

import "time"

// Friend 好友关系（单向）
type Friend struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;index"`
	FriendID  int64     `gorm:"not null;index"`
	Status    int8      `gorm:"default:1;comment:1正常 0拉黑"`
	CreatedAt time.Time
}

func (Friend) TableName() string {
	return "friends"
}
