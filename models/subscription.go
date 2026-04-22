package models

import (
	"errors"
	"fmt"
	"ops-message-unified-push/pkg/util"

	"gorm.io/gorm"
)

// Subscription 订阅规则
type Subscription struct {
	UUIDModel
	Name      string `json:"name" gorm:"type:varchar(200);not null"`
	SourceID  string `json:"source_id" gorm:"type:varchar(12);not null"`
	Topic     string `json:"topic" gorm:"type:varchar(200);not null"`
	Tag       string `json:"tag" gorm:"type:varchar(200)"`
	GroupName string `json:"group_name" gorm:"type:varchar(200)"`

	// 正则配置
	ValidateRegex string `json:"validate_regex" gorm:"type:text"`  // 验证正则
	ExtractRegex  string `json:"extract_regex" gorm:"type:text"`   // 提取正则
	ExtractField  string `json:"extract_field" gorm:"type:varchar(100)"`

	// 模板
	TemplateID string `json:"template_id" gorm:"type:varchar(12);not null"`

	// 订阅发送配置
	// 历史字段 consume_mode 原本用于 push/pull；当前复用为模板内容格式 text/html/markdown
	ConsumeMode string `json:"consume_mode" gorm:"type:varchar(20);default:'text'"`

	// 状态
	Enabled     int    `json:"enabled" gorm:"default:1"`
	Status      string `json:"status" gorm:"type:varchar(20);default:'stopped'"`    // running/stopped/error

	// 统计
	TotalConsumed   int       `json:"total_consumed" gorm:"default:0"`
	TotalSent       int       `json:"total_sent" gorm:"default:0"`
	TotalFailed     int       `json:"total_failed" gorm:"default:0"`
	LastConsumeTime util.Time `json:"last_consume_time"`
}

// GenerateSubscriptionUniqueID 生成唯一ID: SUB + 9位随机（总长12，匹配数据库字段）
func GenerateSubscriptionUniqueID() string {
	newUUID := util.GenerateRandomString(9)
	return fmt.Sprintf("SUB%s", newUUID)
}

// AddSubscription 添加订阅规则
func AddSubscription(name, sourceID, topic, tag, groupName, validateRegex, extractRegex, extractField, templateID, consumeMode, createdBy string) (string, error) {
	newUUID := GenerateSubscriptionUniqueID()
	subscription := Subscription{
		UUIDModel: UUIDModel{
			ID:         newUUID,
			CreatedBy:  createdBy,
			ModifiedBy: createdBy,
		},
		Name:          name,
		SourceID:      sourceID,
		Topic:         topic,
		Tag:           tag,
		GroupName:     groupName,
		ValidateRegex: validateRegex,
		ExtractRegex:  extractRegex,
		ExtractField:  extractField,
		TemplateID:    templateID,
		Enabled:       1,
		ConsumeMode:   consumeMode,
		Status:        "stopped",
	}

	if err := db.Create(&subscription).Error; err != nil {
		return newUUID, err
	}
	return newUUID, nil
}

// GetSubscriptions 获取订阅列表
func GetSubscriptions(pageNum, pageSize int, name, sourceID, status, templateID string) ([]Subscription, error) {
	var subscriptions []Subscription
	query := db.Model(&Subscription{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if templateID != "" {
		query = query.Where("template_id = ?", templateID)
	}

	query = query.Order("created_on DESC")
	if pageSize > 0 && pageNum > 0 {
		query = query.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	}

	err := query.Find(&subscriptions).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return subscriptions, nil
}

// GetSubscriptionsTotal 获取订阅总数
func GetSubscriptionsTotal(name, sourceID, status, templateID string) (int64, error) {
	var total int64
	query := db.Model(&Subscription{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if templateID != "" {
		query = query.Where("template_id = ?", templateID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetSubscriptionByID 根据ID获取订阅
func GetSubscriptionByID(id string) (Subscription, error) {
	var subscription Subscription
	err := db.Where("id = ?", id).First(&subscription).Error
	if err != nil {
		return subscription, err
	}
	return subscription, nil
}

// UpdateSubscription 更新订阅
func UpdateSubscription(id string, data map[string]interface{}) error {
	return db.Model(&Subscription{}).Where("id = ?", id).Updates(data).Error
}

// DeleteSubscription 删除订阅
func DeleteSubscription(id string) error {
	return db.Where("id = ?", id).Delete(&Subscription{}).Error
}

// UpdateSubscriptionStatus 更新订阅状态
func UpdateSubscriptionStatus(id, status string, errorMsg *string) error {
	data := map[string]interface{}{
		"status": status,
	}
	if errorMsg != nil {
		if *errorMsg != "" {
			data["status"] = "error"
		}
	}
	return db.Model(&Subscription{}).Where("id = ?", id).Updates(data).Error
}

// IncrSubscriptionConsumed 增加消费计数
func IncrSubscriptionConsumed(id string) error {
	return db.Model(&Subscription{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_consumed":  gorm.Expr("total_consumed + 1"),
			"last_consume_time": util.TimeNow(),
		}).Error
}

// IncrSubscriptionSent 增加发送计数
func IncrSubscriptionSent(id string) error {
	return db.Model(&Subscription{}).Where("id = ?", id).
		Update("total_sent", gorm.Expr("total_sent + 1")).Error
}

// IncrSubscriptionFailed 增加失败计数
func IncrSubscriptionFailed(id string) error {
	return db.Model(&Subscription{}).Where("id = ?", id).
		Update("total_failed", gorm.Expr("total_failed + 1")).Error
}

// GetEnabledSubscriptions 获取所有启用的订阅（用于启动时加载）
func GetEnabledSubscriptions() ([]Subscription, error) {
	var subscriptions []Subscription
	err := db.Where("enabled = ?", 1).Find(&subscriptions).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return subscriptions, nil
}

// GetSubscriptionsBySourceID 根据数据源ID获取订阅列表
func GetSubscriptionsBySourceID(sourceID string) ([]Subscription, error) {
	var subscriptions []Subscription
	err := db.Where("source_id = ?", sourceID).Find(&subscriptions).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return subscriptions, nil
}

// GetByUUID 根据 UUID 获取订阅
func (s *Subscription) GetByUUID(id string) error {
	return db.Where("id = ?", id).First(s).Error
}

// UpdateStatus 更新订阅状态
func (s *Subscription) UpdateStatus(status string) error {
	s.Status = status
	return db.Model(s).Update("status", status).Error
}

// UpdateSubscriptionStats 更新订阅统计
func UpdateSubscriptionStats(id string, matched, sendStatus int) error {
	updates := map[string]interface{}{
		"last_consume_time": util.TimeNow(),
	}

	if matched == 1 {
		updates["total_consumed"] = gorm.Expr("total_consumed + 1")
	}

	if sendStatus == 1 {
		updates["total_sent"] = gorm.Expr("total_sent + 1")
	} else if sendStatus == 2 {
		updates["total_failed"] = gorm.Expr("total_failed + 1")
	}

	return db.Model(&Subscription{}).Where("id = ?", id).Updates(updates).Error
}
