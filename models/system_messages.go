package models

import "ops-message-unified-push/pkg/util"

type SystemMessage struct {
	ID                 string    `json:"id" gorm:"type:varchar(32);primaryKey"`
	Type               string    `json:"type" gorm:"type:varchar(64);default:'announcement';index"`
	Title              string    `json:"title" gorm:"type:varchar(200);not null;index"`
	Summary            string    `json:"summary" gorm:"type:varchar(500);default:''"`
	Content            string    `json:"content" gorm:"type:text"`
	PublishTime        util.Time `json:"publish_time" gorm:"index"`
	EffectiveStartTime util.Time `json:"effective_start_time" gorm:"index"`
	EffectiveEndTime   util.Time `json:"effective_end_time" gorm:"index"`
	IsPinned           bool      `json:"is_pinned" gorm:"default:false;index"`
	Status             string    `json:"status" gorm:"type:varchar(32);default:'published';index"`
	CreatedBy          string    `json:"created_by" gorm:"type:varchar(100);default:''"`
	ModifiedBy         string    `json:"modified_by" gorm:"type:varchar(100);default:''"`
	CreatedAt          util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime;index"`
	UpdatedAt          util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

func AddSystemMessage(message *SystemMessage) error {
	return db.Create(message).Error
}

func UpdateSystemMessage(id string, data map[string]interface{}) error {
	return db.Model(&SystemMessage{}).Where("id = ?", id).Updates(data).Error
}

func GetSystemMessageByID(id string) (*SystemMessage, error) {
	var message SystemMessage
	if err := db.Where("id = ?", id).First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}
