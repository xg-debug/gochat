package model

import "time"

// File 文件资源
type File struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"not null;index"`
	FileName  string    `gorm:"size:255"`
	FileURL   string    `gorm:"size:500"`
	FileSize  int64
	FileType  string    `gorm:"size:50"`
	CreatedAt time.Time
}

func (File) TableName() string {
	return "files"
}
