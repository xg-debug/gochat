package model

import "time"

type Message struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	FromID    int64
	ToID      int64
	ChatType  int8
	MsgType   int8
	Content   string    `gorm:"type:text"`
	Status    int8      `gorm:"default:0"`
	CreatedAt time.Time
}

func (Message) TableName() string {
	return "messages"
}
