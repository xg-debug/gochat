package model

import "time"

type UserProfile struct {
	UserID    int64     `gorm:"primaryKey"`
	Nickname  string    `gorm:"size:50"`
	Avatar    string    `gorm:"size:255"`
	Gender    int8      `gorm:"default:0"`
	Signature string    `gorm:"size:255"`
	Birthday  *time.Time
	Location  string    `gorm:"size:100"`
	UpdatedAt time.Time
}

func (UserProfile) TableName() string {
	return "user_profiles"
}
