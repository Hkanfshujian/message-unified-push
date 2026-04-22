package models

import (
	"fmt"
	"strings"
)

// GetTemplateInstancesByTemplateIDAndEnable 根据模板ID和启用状态获取实例
func GetTemplateInstancesByTemplateIDAndEnable(templateID string, enable int, result interface{}) error {
	return db.Where("template_id = ? AND enable = ?", templateID, enable).Find(result).Error
}

type SendTasksIns struct {
	UUIDModel

	TaskID      string `json:"task_id"  gorm:"type:varchar(12) ;default:'';index"`
	TemplateID  string `json:"template_id"  gorm:"type:varchar(12) ;default:'';index"` // 模板ID
	WayID       string `json:"way_id" gorm:"type:varchar(12) ;default:'';index"`
	WayType     string `json:"way_type" gorm:"type:varchar(100) ;default:'';index"`
	ContentType string `json:"content_type" gorm:"type:varchar(100) ;default:'';index"`
	Config      string `json:"config" gorm:"type:text ;"`
	Extra       string `json:"extra" gorm:"type:text ;"`
	Enable      int    `json:"enable" gorm:"type:int ;default:1;"`
}

type SendTasksInsRes struct {
	SendTasksIns
	WayName string `json:"way_name"`
}

type TaskIns struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	InsData []SendTasksInsRes `json:"ins_data"`
}

// InsEmailConfig 实例里面的邮箱config
type InsEmailConfig struct {
	ToAccount string `json:"to_account" validate:"required,email" label:"收件邮箱"`
}

// InsWeChatAccountConfig 实例里面的邮箱config
type InsWeChatAccountConfig struct {
	ToAccount string `json:"to_account" validate:"required" label:"收件微信Openid"`
}

// InsEmailConfig 实例里面的邮箱config
type InsDtalkConfig struct {
}

// InsQyWeiXinConfig 实例里面的企业微信config
type InsQyWeiXinConfig struct {
}

// InsFeishuConfig 实例里面的飞书config
type InsFeishuConfig struct {
}

// InsCustomConfig 实例里面的自定义config
type InsCustomConfig struct {
}

// InsAliyunSMSConfig 实例里面的阿里云短信config
type InsAliyunSMSConfig struct {
	PhoneNumber  string `json:"phone_number" validate:"required" label:"手机号码"`
	TemplateCode string `json:"template_code" validate:"required" label:"短信模板CODE"`
}

// InsTelegramConfig 实例里面的Telegram config
type InsTelegramConfig struct {
}

// InsBarkConfig 实例里面的Bark config
type InsBarkConfig struct {
}

// InsPushMeConfig 实例里面的PushMe config
type InsPushMeConfig struct {
}

// InsNtfyConfig 实例里面的Ntfy config
type InsNtfyConfig struct {
}

// InsGotifyConfig 实例里面的Gotify config
type InsGotifyConfig struct {
}

// InsQyWeiXinAppConfig 实例里面的企业微信应用 config
type InsQyWeiXinAppConfig struct {
}

// ManyAddTaskIns 批量添加实例
func ManyAddTaskIns(taskIns []SendTasksIns) error {
	tx := db.Begin()
	for _, ins := range taskIns {
		// 存在就跳过这条ins记录
		err := db.Where("id = ?", ins.ID).Take(&SendTasksIns{}).Error
		if err == nil {
			continue
		}
		if err := tx.Create(&ins).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// AddTaskInsOne 添加一条实例
func AddTaskInsOne(ins SendTasksIns) error {
	if err := db.Create(&ins).Error; err != nil {
		return err
	}
	return nil
}

// DeleteMsgTaskIns 删除一条实例
func DeleteMsgTaskIns(id string) error {
	if err := db.Where("id = ?", id).Delete(&SendTasksIns{}).Error; err != nil {
		return err
	}
	return nil
}

// UpdateMsgTaskIns 更新实例
func UpdateMsgTaskIns(id string, data map[string]interface{}) error {
	if err := db.Model(&SendTasksIns{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

// GetTemplateInsList 获取模板关联的实例列表（包含渠道名称）
func GetTemplateInsList(templateID string) ([]SendTasksInsRes, error) {
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var insList []SendTasksInsRes

	err := db.
		Table(insTable).
		Select(fmt.Sprintf("%s.*, %s.name as way_name", insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.template_id = ?", insTable), templateID).
		Order(fmt.Sprintf("%s.created_on DESC", insTable)).
		Scan(&insList).Error

	if err != nil {
		return nil, err
	}
	return insList, nil
}

func GetTemplateInsByIDs(ids []string) ([]SendTasksInsRes, error) {
	if len(ids) == 0 {
		return []SendTasksInsRes{}, nil
	}
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var insList []SendTasksInsRes

	err := db.
		Table(insTable).
		Select(fmt.Sprintf("%s.*, %s.name as way_name", insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.id IN ?", insTable), ids).
		Scan(&insList).Error

	if err != nil {
		return nil, err
	}
	return insList, nil
}

func GetTemplateInsByTemplateIDs(templateIDs []string) ([]SendTasksInsRes, error) {
	if len(templateIDs) == 0 {
		return []SendTasksInsRes{}, nil
	}
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var insList []SendTasksInsRes

	err := db.
		Table(insTable).
		Select(fmt.Sprintf("%s.*, %s.name as way_name", insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.template_id IN ?", insTable), templateIDs).
		Scan(&insList).Error

	if err != nil {
		return nil, err
	}
	return insList, nil
}

func FindWayUsageNames(wayID string) ([]string, error) {
	if strings.TrimSpace(wayID) == "" {
		return []string{}, nil
	}
	insTable := GetSchema(SendTasksIns{})
	templateTable := GetSchema(Template{})
	var rows []struct {
		TemplateID   string `gorm:"column:template_id"`
		TemplateName string `gorm:"column:template_name"`
	}
	err := db.
		Table(insTable).
		Select(fmt.Sprintf("DISTINCT %s.template_id AS template_id, %s.name AS template_name", insTable, templateTable)).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.template_id = %s.id", templateTable, insTable, templateTable)).
		Where(fmt.Sprintf("%s.way_id = ?", insTable), wayID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.TemplateName)
		if name == "" {
			templateID := strings.TrimSpace(row.TemplateID)
			if templateID == "" {
				name = "未知模板"
			} else {
				name = "模板ID:" + templateID
			}
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names, nil
}
