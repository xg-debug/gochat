package model

import "time"

// Call 音视频通话记录
type Call struct {
	ID        int64      `gorm:"primaryKey;autoIncrement"`
	CallerID  int64      `gorm:"not null;index"`
	CalleeID  int64      `gorm:"not null;index"`
	CallType  int8       `gorm:"comment:1语音 2视频"`
	Status    int8       `gorm:"comment:0呼叫中 1接通 2拒绝 3挂断"`
	StartTime *time.Time
	EndTime   *time.Time
}

func (Call) TableName() string {
	return "calls"
}
