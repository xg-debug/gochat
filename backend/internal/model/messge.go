package model

import "time"

type Message struct {
	ID          int64   `gorm:"primaryKey;autoIncrement"`
	FromID      int64   `gorm:"uniqueIndex:uk_msg_dedupe"`
	ToID        int64   `gorm:"uniqueIndex:uk_msg_dedupe"`
	ChatType    int8    `gorm:"uniqueIndex:uk_msg_dedupe"`
	ClientMsgID *string `gorm:"size:100;uniqueIndex:uk_msg_dedupe"`
	MsgType     int8
	Content     string `gorm:"type:text"`
	Status      int8   `gorm:"default:0"`
	CreatedAt   time.Time
}

func (Message) TableName() string {
	return "messages"
}
