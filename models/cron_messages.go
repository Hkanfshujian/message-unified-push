package models

import (
	"errors"
	"fmt"
	"ops-message-unified-push/pkg/util"

	"gorm.io/gorm"
)

type CronMessages struct {
	UUIDModel

	Name       string `json:"name" gorm:"type:varchar(200) ;default:'';"`
	TemplateID string `json:"template_id" gorm:"type:varchar(36) ;column:task_id;default:'';"`
	Cron       string `json:"cron" gorm:"type:varchar(4096) ;default:'';"`
	Title      string `json:"title" gorm:"type:varchar(1000) ;default:'';"`
	Content    string `json:"content" gorm:"type:varchar(4096) ;default:'';"`
	//MarkDown string `json:"markdown" gorm:"type:varchar(4096) ;default:'';"`
	Url    string `json:"url" gorm:"type:varchar(4096) ;default:'';"`
	Enable int    `json:"enable" gorm:"type:int ;default:1;"`
}

func GenerateMsgUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("CM%s", newUUID)
}

func AddSendCronMsg(
	name string,
	templateID string,
	cron string,
	title string,
	url string,
	createdBy string,
) (string, error) {
	newUUID := GenerateMsgUniqueID()
	msg := CronMessages{
		UUIDModel: UUIDModel{
			ID:         newUUID,
			CreatedBy:  createdBy,
			ModifiedBy: createdBy,
		},
		Name:       name,
		TemplateID: templateID,
		Cron:       cron,
		Title:      title,
		Url:        url,
		Enable:     1,
	}
	if err := db.Create(&msg).Error; err != nil {
		return newUUID, err
	}
	return newUUID, nil
}

// GetCronMessages 获取所有任务
func GetCronMessages(pageNum int, pageSize int, name string, maps interface{}) ([]CronMessages, error) {
	var (
		msgs []CronMessages
		err  error
	)
	query := db.Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	err = query.Find(&msgs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return msgs, nil
}

// GetCronMessagesTotal 获取所有任务总数
func GetCronMessagesTotal(name string, maps interface{}) (int64, error) {
	var (
		err   error
		total int64
	)
	query := db.Model(&CronMessages{}).Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}

	err = query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func DeleteCronMsg(id string) error {
	if err := db.Where("id = ?", id).Delete(&CronMessages{}).Error; err != nil {
		return err
	}
	return nil
}

func EditCronMsg(id string, data interface{}) error {
	if err := db.Model(&CronMessages{}).Where("id = ? ", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func GetCronMsgByID(id string) (CronMessages, error) {
	var msg CronMessages
	err := db.Where("id = ? ", id).Take(&msg).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return msg, err
	}
	return msg, nil
}

// GetCronMsgByTemplateAndName 根据模板ID和名称获取最新的定时消息
func GetCronMsgByTemplateAndName(templateID, name string) (CronMessages, error) {
	var msg CronMessages
	err := db.Where("task_id = ? AND name = ?", templateID, name).
		Order("created_on DESC").
		Take(&msg).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return msg, err
	}
	return msg, nil
}

func CountCronMsgByTemplateID(templateID string) (int64, error) {
	if templateID == "" {
		return 0, nil
	}
	var count int64
	err := db.Model(&CronMessages{}).Where("task_id = ?", templateID).Count(&count).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return count, err
}

func GetCronMsgCountByTemplateIDs(templateIDs []string) (map[string]int64, error) {
	result := make(map[string]int64)
	if len(templateIDs) == 0 {
		return result, nil
	}

	type row struct {
		TemplateID string `gorm:"column:task_id"`
		Count      int64  `gorm:"column:cnt"`
	}

	var rows []row
	err := db.Table(GetSchema(CronMessages{})).
		Select("task_id, COUNT(*) as cnt").
		Where("task_id IN ?", templateIDs).
		Group("task_id").
		Scan(&rows).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return result, err
	}
	for _, r := range rows {
		result[r.TemplateID] = r.Count
	}
	return result, nil
}

func GetCronMsgsByTemplateID(templateID string) ([]CronMessages, error) {
	var msgs []CronMessages
	if templateID == "" {
		return msgs, nil
	}
	err := db.Where("task_id = ?", templateID).Order("created_on DESC").Find(&msgs).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return []CronMessages{}, nil
	}
	return msgs, err
}
