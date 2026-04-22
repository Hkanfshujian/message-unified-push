package models

import (
	"errors"
	"fmt"
	"ops-message-unified-push/pkg/util"

	"gorm.io/gorm"
)

// MQSource 消息队列数据源
type MQSource struct {
	UUIDModel
	Name           string    `json:"name" gorm:"type:varchar(200);not null"`
	Type           string    `json:"type" gorm:"type:varchar(50);not null"` // rocketmq/kafka/rabbitmq
	Enabled        int       `json:"enabled" gorm:"default:1"`              // 0-禁用 1-启用

	// RocketMQ 配置
	NamesrvAddr string `json:"namesrv_addr" gorm:"type:varchar(500)"` // NameServer 地址
	AccessKey   string `json:"access_key" gorm:"type:varchar(200)"`   // 访问密钥（可选）
	SecretKey   string `json:"secret_key" gorm:"type:varchar(200)"`   // 密钥（可选）

	// 连接测试
	LastTestStatus string    `json:"last_test_status" gorm:"type:varchar(20)"` // success/failed
	LastTestTime   util.Time `json:"last_test_time"`
	TestError      string    `json:"test_error" gorm:"type:text"`

	// 非数据库字段
	BindingCount int `json:"binding_count" gorm:"-"` // 绑定的订阅数量
}

// GenerateMQSourceUniqueID 生成唯一ID: MS + 10位随机
func GenerateMQSourceUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("MS%s", newUUID)
}

// AddMQSource 添加消息队列数据源
func AddMQSource(name, mqType, namesrvAddr, accessKey, secretKey, createdBy string) (string, error) {
	newUUID := GenerateMQSourceUniqueID()
	source := MQSource{
		UUIDModel: UUIDModel{
			ID:         newUUID,
			CreatedBy:  createdBy,
			ModifiedBy: createdBy,
		},
		Name:        name,
		Type:        mqType,
		NamesrvAddr: namesrvAddr,
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Enabled:     1,
	}

	if err := db.Create(&source).Error; err != nil {
		return newUUID, err
	}
	return newUUID, nil
}

// GetMQSources 获取数据源列表
func GetMQSources(pageNum, pageSize int, name, mqType, status string) ([]MQSource, error) {
	var sources []MQSource
	query := db.Model(&MQSource{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if mqType != "" {
		query = query.Where("type = ?", mqType)
	}
	if status != "" {
		if status == "untested" {
			query = query.Where("last_test_status = '' OR last_test_status IS NULL")
		} else {
			query = query.Where("last_test_status = ?", status)
		}
	}

	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	}

	err := query.Find(&sources).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return sources, nil
}

// GetMQSourcesTotal 获取数据源总数
func GetMQSourcesTotal(name, mqType, status string) (int64, error) {
	var total int64
	query := db.Model(&MQSource{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if mqType != "" {
		query = query.Where("type = ?", mqType)
	}
	if status != "" {
		if status == "untested" {
			query = query.Where("last_test_status = '' OR last_test_status IS NULL")
		} else {
			query = query.Where("last_test_status = ?", status)
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetMQSourceByID 根据ID获取数据源
func GetMQSourceByID(id string) (MQSource, error) {
	var source MQSource
	err := db.Where("id = ?", id).First(&source).Error
	if err != nil {
		return source, err
	}
	return source, nil
}

// UpdateMQSource 更新数据源
func UpdateMQSource(id string, data map[string]interface{}) error {
	return db.Model(&MQSource{}).Where("id = ?", id).Updates(data).Error
}

// DeleteMQSource 删除数据源
func DeleteMQSource(id string) error {
	return db.Where("id = ?", id).Delete(&MQSource{}).Error
}

// UpdateMQSourceTestStatus 更新测试状态
func UpdateMQSourceTestStatus(id, status, errorMsg string) error {
	data := map[string]interface{}{
		"last_test_status": status,
		"last_test_time":   util.TimeNow(),
	}
	if errorMsg != "" {
		data["test_error"] = errorMsg
	} else {
		data["test_error"] = ""
	}
	return db.Model(&MQSource{}).Where("id = ?", id).Updates(data).Error
}

// GetMQSourceBindingCount 获取数据源绑定的订阅数量
func GetMQSourceBindingCount(sourceID string) (int64, error) {
	var count int64
	err := db.Model(&Subscription{}).Where("source_id = ?", sourceID).Count(&count).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	return count, nil
}

// GetMQSourceBindingCounts 批量获取数据源的绑定数量
func GetMQSourceBindingCounts(sourceIDs []string) (map[string]int64, error) {
	result := make(map[string]int64)
	if len(sourceIDs) == 0 {
		return result, nil
	}

	type row struct {
		SourceID string `gorm:"column:source_id"`
		Count    int64  `gorm:"column:cnt"`
	}

	var rows []row
	err := db.Table(GetSchema(Subscription{})).
		Select("source_id, COUNT(*) as cnt").
		Where("source_id IN ?", sourceIDs).
		Group("source_id").
		Scan(&rows).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return result, err
	}
	for _, r := range rows {
		result[r.SourceID] = r.Count
	}
	return result, nil
}

// TestConnection 测试连接（占位，Phase 2 实现）
func (s *MQSource) TestConnection() error {
	// TODO: Phase 2 实现真实连接测试
	return nil
}

