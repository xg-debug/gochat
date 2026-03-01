package model

import "time"

type UserAccount struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:50;unique;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Email        string    `gorm:"size:100"`
	Phone        string    `gorm:"size:20"`
	Status       int8      `gorm:"default:1"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserAccount) TableName() string {
	return "user_accounts"
}
