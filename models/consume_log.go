package models

import (
	"ops-message-unified-push/pkg/util"
)

// ConsumeLog 消费日志
type ConsumeLog struct {
	ID              uint         `gorm:"autoIncrement;type:bigint;primaryKey" json:"id"`
	SubscriptionID  string       `json:"subscription_id" gorm:"type:varchar(12);not null;index"`
	MsgID           string       `json:"msg_id" gorm:"type:varchar(100);index"`
	Topic           string       `json:"topic" gorm:"type:varchar(200)"`
	Tag             string       `json:"tag" gorm:"type:varchar(200)"`
	RawMessage      string       `json:"raw_message" gorm:"type:text"`
	Matched         int          `json:"matched" gorm:"default:0;index"` // 0-未匹配 1-匹配
	ExtractedValues util.JSONMap `json:"extracted_values" gorm:"type:json"`
	SendStatus      int          `json:"send_status" gorm:"default:0;index"` // 0-未发送 1-成功 2-失败
	SendError       string       `json:"send_error" gorm:"type:text"`
	ConsumeTime     util.Time    `json:"consume_time" gorm:"autoCreateTime;index"`
}

// CreateConsumeLog 创建消费日志
func CreateConsumeLog(log *ConsumeLog) error {
	return db.Create(log).Error
}

// Create 创建消费日志（方法版本）
func (l *ConsumeLog) Create() error {
	return db.Create(l).Error
}

// GetConsumeLogByID 根据ID获取消费日志
func GetConsumeLogByID(id uint) (*ConsumeLog, error) {
	var log ConsumeLog
	if err := db.Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetConsumeLogs 获取消费日志列表
func GetConsumeLogs(pageNum, pageSize int, subscriptionID, matched, sendStatus, startTime, endTime string) ([]ConsumeLog, error) {
	var logs []ConsumeLog
	query := db.Model(&ConsumeLog{})

	if subscriptionID != "" {
		query = query.Where("subscription_id = ?", subscriptionID)
	}
	if matched != "" {
		query = query.Where("matched = ?", matched)
	}
	if sendStatus != "" {
		query = query.Where("send_status = ?", sendStatus)
	}
	if startTime != "" {
		query = query.Where("consume_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("consume_time <= ?", endTime)
	}

	query = query.Order("consume_time DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	}

	err := query.Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// GetConsumeLogsTotal 获取消费日志总数
func GetConsumeLogsTotal(subscriptionID, matched, sendStatus, startTime, endTime string) (int64, error) {
	var total int64
	query := db.Model(&ConsumeLog{})

	if subscriptionID != "" {
		query = query.Where("subscription_id = ?", subscriptionID)
	}
	if matched != "" {
		query = query.Where("matched = ?", matched)
	}
	if sendStatus != "" {
		query = query.Where("send_status = ?", sendStatus)
	}
	if startTime != "" {
		query = query.Where("consume_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("consume_time <= ?", endTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// CleanOldConsumeLogs 清理旧日志（保留最近N天）
func CleanOldConsumeLogs(days int) error {
	if days <= 0 {
		return nil
	}
	cutoffTime := util.TimeNow().AddDate(0, 0, -days)
	return db.Where("consume_time < ?", cutoffTime).Delete(&ConsumeLog{}).Error
}

// CleanConsumeLogsByCount 按数量清理日志（保留最近N条）
func CleanConsumeLogsByCount(keepCount int) error {
	if keepCount <= 0 {
		return nil
	}

	// 获取需要保留的最小ID
	var minLog ConsumeLog
	err := db.Order("consume_time DESC").Offset(keepCount - 1).First(&minLog).Error
	if err != nil {
		return nil // 日志数量不足，无需清理
	}

	return db.Where("id < ?", minLog.ID).Delete(&ConsumeLog{}).Error
}

// DeleteOutDateConsumeLogs 按数量清理消费日志（保留最近N条）
func DeleteOutDateConsumeLogs(keepNum int) (int, error) {
	if keepNum <= 0 {
		return 0, nil
	}

	var threshold ConsumeLog
	result := db.Model(&ConsumeLog{}).
		Select("id").
		Order("consume_time DESC").
		Offset(keepNum - 1).
		Limit(1).
		First(&threshold)

	if result.Error != nil {
		return 0, nil
	}

	deleteResult := db.Where("id < ?", threshold.ID).Delete(&ConsumeLog{})
	if deleteResult.Error != nil {
		return 0, deleteResult.Error
	}

	return int(deleteResult.RowsAffected), nil
}
