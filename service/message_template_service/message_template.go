package message_template_service

import (
	"encoding/json"
	"errors"
	"ops-message-unified-push/models"
	"strings"
)

type TemplateService struct {
	ID               string
	Name             string
	Description      string
	TextTemplate     string
	HTMLTemplate     string
	MarkdownTemplate string
	Placeholders     string
	AtMobiles        string
	AtUserIds        string
	IsAtAll          bool
	Status           string
	Text             string
	
	PageNum  int
	PageSize int
}

// Placeholder 占位符定义
type Placeholder struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Default string `json:"default"`
}

// validateTemplateID 校验模板ID格式：TP + 10位字母数字
func validateTemplateID(id string) error {
	if id == "" {
		return nil
	}
	if len(id) != 12 {
		return errors.New("模板ID必须为12位，格式为TP开头+10位字母或数字")
	}
	if !strings.HasPrefix(id, "TP") {
		return errors.New("模板ID必须以TP开头")
	}
	for _, ch := range id[2:] {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			continue
		}
		return errors.New("模板ID只能包含字母和数字")
	}
	return nil
}

// Add 添加消息模板
func (s *TemplateService) Add() error {
	if err := s.validatePlaceholders(); err != nil {
		return err
	}

	if err := validateTemplateID(s.ID); err != nil {
		return err
	}

	if s.ID != "" {
		exists, err := models.ExistTemplateByID(s.ID)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("模板ID已存在，请更换后再试")
		}
	}

	newUUID := s.ID
	if newUUID == "" {
		newUUID = models.GenerateTemplateUniqueID()
	}

	model := models.Template{
		UUIDModel: models.UUIDModel{
			ID: newUUID,
		},
		Name:             s.Name,
		Description:      s.Description,
		TextTemplate:     s.TextTemplate,
		HTMLTemplate:     s.HTMLTemplate,
		MarkdownTemplate: s.MarkdownTemplate,
		Placeholders:     s.Placeholders,
		AtMobiles:        s.AtMobiles,
		AtUserIds:        s.AtUserIds,
		IsAtAll:          s.IsAtAll,
		Status:           s.Status,
	}
	
	return model.Add()
}

// Update 更新消息模板
func (s *TemplateService) Update() error {
	if err := s.validatePlaceholders(); err != nil {
		return err
	}
	
	model := models.Template{
		UUIDModel: models.UUIDModel{
			ID: s.ID,
		},
		Name:             s.Name,
		Description:      s.Description,
		TextTemplate:     s.TextTemplate,
		HTMLTemplate:     s.HTMLTemplate,
		MarkdownTemplate: s.MarkdownTemplate,
		Placeholders:     s.Placeholders,
		AtMobiles:        s.AtMobiles,
		AtUserIds:        s.AtUserIds,
		IsAtAll:          s.IsAtAll,
		Status:           s.Status,
	}
	
	return model.Update()
}

// Delete 删除消息模板
func (s *TemplateService) Delete() error {
	if s.ID == "" {
		return errors.New("模板ID不能为空")
	}

	// 检查是否存在关联的定时消息
	count, err := models.CountCronMsgByTemplateID(s.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("当前模板仍关联定时消息，请先在【定时消息】中删除相关任务后再删除模板")
	}

	model := models.Template{
		UUIDModel: models.UUIDModel{
			ID: s.ID,
		},
	}
	return model.Delete()
}

// Get 获取单个消息模板
func (s *TemplateService) Get() (*models.TemplateResult, error) {
	return models.GetTemplateByID(s.ID)
}

// GetAll 获取消息模板列表
func (s *TemplateService) GetAll() ([]models.TemplateResult, error) {
	templates, err := models.GetTemplates(s.PageNum, s.PageSize, s.Text, s.getMaps())
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(templates))
	for _, t := range templates {
		if t.ID != "" {
			ids = append(ids, t.ID)
		}
	}

	countMap, err := models.GetCronMsgCountByTemplateIDs(ids)
	if err != nil {
		return nil, err
	}

	for i := range templates {
		templates[i].CronMsgCount = countMap[templates[i].ID]
	}

	return templates, nil
}

// Count 获取消息模板总数
func (s *TemplateService) Count() (int64, error) {
	return models.GetTemplatesTotal(s.Text, s.getMaps())
}

// ExistByID 检查模板是否存在
func (s *TemplateService) ExistByID() (bool, error) {
	return models.ExistTemplateByID(s.ID)
}

// RenderTemplate 渲染模板（替换占位符）
func (s *TemplateService) RenderTemplate(templateContent string, params map[string]string) string {
	result := templateContent
	
	for key, value := range params {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	return result
}

// PreviewTemplate 预览模板效果
func (s *TemplateService) PreviewTemplate(params map[string]string) (map[string]string, error) {
	template, err := s.Get()
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	
	if template.TextTemplate != "" {
		result["text"] = s.RenderTemplate(template.TextTemplate, params)
	}
	
	if template.HTMLTemplate != "" {
		result["html"] = s.RenderTemplate(template.HTMLTemplate, params)
	}
	
	if template.MarkdownTemplate != "" {
		result["markdown"] = s.RenderTemplate(template.MarkdownTemplate, params)
	}
	
	return result, nil
}

// validatePlaceholders 验证占位符格式
func (s *TemplateService) validatePlaceholders() error {
	if s.Placeholders == "" {
		return nil
	}
	
	var placeholders []Placeholder
	if err := json.Unmarshal([]byte(s.Placeholders), &placeholders); err != nil {
		return errors.New("占位符格式错误，必须是有效的JSON数组")
	}
	
	for _, p := range placeholders {
		if p.Key == "" {
			return errors.New("占位符的key不能为空")
		}
	}
	
	return nil
}

// getMaps 获取查询条件
func (s *TemplateService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	
	if s.Status != "" {
		maps["status"] = s.Status
	}
	
	return maps
}
