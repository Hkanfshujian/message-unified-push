package models

import "ops-message-unified-push/pkg/util"

type BusinessPushMessage struct {
	ID             string    `json:"id" gorm:"type:varchar(32);primaryKey"`
	Type           string    `json:"type" gorm:"type:varchar(64);default:'business';index"`
	Title          string    `json:"title" gorm:"type:varchar(200);not null;index"`
	Summary        string    `json:"summary" gorm:"type:varchar(500);default:''"`
	Content        string    `json:"content" gorm:"type:text"`
	ThumbnailURL   string    `json:"thumbnail_url" gorm:"type:varchar(500);default:''"`
	ThumbnailRatio string    `json:"thumbnail_ratio" gorm:"type:varchar(20);default:'16:9'"`
	SourceSubject  string    `json:"source_subject" gorm:"type:varchar(200);default:''"`
	SourceType     string    `json:"source_type" gorm:"type:varchar(64);default:'';index"`
	SourceID       string    `json:"source_id" gorm:"type:varchar(64);default:'';index"`
	PushTime       util.Time `json:"push_time" gorm:"index"`
	TargetURL      string    `json:"target_url" gorm:"type:varchar(500);default:''"`
	CreatedAt      util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime;index"`
}

func AddBusinessPushMessage(message *BusinessPushMessage) error {
	return db.Create(message).Error
}

func GetBusinessPushMessageByID(id string) (*BusinessPushMessage, error) {
	var message BusinessPushMessage
	if err := db.Where("id = ?", id).First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}
