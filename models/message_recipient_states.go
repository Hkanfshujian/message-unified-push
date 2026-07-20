package models

import "ops-message-unified-push/pkg/util"

type MessageRecipientState struct {
	ID              uint      `json:"id" gorm:"autoIncrement;type:integer;primaryKey"`
	UserID          int       `json:"user_id" gorm:"index:idx_user_message,unique;index"`
	MessageCategory string    `json:"message_category" gorm:"type:varchar(32);index:idx_user_message,unique;index"`
	MessageID       string    `json:"message_id" gorm:"type:varchar(64);index:idx_user_message,unique;index"`
	PushType        string    `json:"push_type" gorm:"type:varchar(64);default:'';index"`
	PushID          string    `json:"push_id" gorm:"type:varchar(64);default:'';index"`
	IsRead          bool      `json:"is_read" gorm:"default:false;index"`
	ReadAt          util.Time `json:"read_at"`
	IsDeleted       bool      `json:"is_deleted" gorm:"default:false;index"`
	DeletedAt       util.Time `json:"deleted_at"`
	SyncVersion     int64     `json:"sync_version" gorm:"default:1;index"`
	CreatedAt       util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime;index"`
	UpdatedAt       util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

func UpsertRecipientState(state *MessageRecipientState) error {
	var existing MessageRecipientState
	err := db.Where("user_id = ? AND message_category = ? AND message_id = ?", state.UserID, state.MessageCategory, state.MessageID).First(&existing).Error
	if err == nil {
		return nil
	}
	return db.Create(state).Error
}
