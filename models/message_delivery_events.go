package models

import "ops-message-unified-push/pkg/util"

type MessageDeliveryEvent struct {
	ID              uint      `json:"id" gorm:"autoIncrement;type:integer;primaryKey"`
	UserID          int       `json:"user_id" gorm:"index"`
	MessageCategory string    `json:"category" gorm:"type:varchar(32);index"`
	MessageID       string    `json:"message_id" gorm:"type:varchar(64);index"`
	EventType       string    `json:"event_type" gorm:"type:varchar(32);index"`
	SyncVersion     int64     `json:"sync_version" gorm:"index"`
	OccurredAt      util.Time `json:"occurred_at" gorm:"autoCreateTime;index"`
}

func AddMessageDeliveryEvent(event *MessageDeliveryEvent) error {
	return db.Create(event).Error
}
