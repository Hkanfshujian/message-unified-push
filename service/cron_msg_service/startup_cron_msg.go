package cron_msg_service

import (
	"encoding/json"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/service/cron_service"
	"ops-message-unified-push/service/send_message_service"
	"strings"

	"github.com/sirupsen/logrus"
)

type MsgCronTask struct {
}

func (s MsgCronTask) Register() {
	// 获取所有的定时消息任务
	limit := 10000
	filter := make(map[string]interface{})
	filter["enable"] = 1
	data, err := models.GetCronMessages(0, limit, "", filter)
	if err != nil {
		logrus.Errorf("获取定时消息任务失败！原因：%s", err.Error())
		return
	}
	if len(data) == 0 {
		logrus.Infof("没有定时消息任务需要注册")
		return
	}
	//注册定时任务
	for _, msg := range data {
		AddCronMsgToCronServer(msg)
	}
	length := len(data)
	if length > 0 {
		logrus.Infof("完成用户自定义的定时消息注册，注册个数：%d", length)
	}
}

// AddCronMsgToCronServer 注册定时任务到定时服务
func AddCronMsgToCronServer(msg models.CronMessages) {
	if msg.Enable != 1 {
		return
	}
	taskId := cron_service.AddTask(cron_service.ScheduledTask{
		Schedule: msg.Cron,
		Job: func() {
			CronMsgSendF(msg)
		},
	})
	constant.CronMsgIdMapMemoryCache[msg.ID] = taskId
	logrus.Infof("新增定时消息成功，消息id: %s，消息名: %s，当前任务总数：%d", msg.ID, msg.Name, len(constant.CronMsgIdMapMemoryCache))
}

// 执行任务的构造函数
func CronMsgSendF(msg models.CronMessages) {
	logrus.Infof("开始只能执行定时消息发送任务: %s，消息名: %s", msg.ID, msg.Name)
	template, err := models.GetTemplateByID(msg.TemplateID)
	if err != nil {
		logrus.Infof("消息模板不存在: %s ", msg.TemplateID)
		return
	}

	if template.Status != "enabled" {
		logrus.Infof("消息模板已禁用: %s ", msg.TemplateID)
		return
	}

	textContent, htmlContent, markdownContent := buildTemplateContent(template)
	sender := send_message_service.SendMessageService{
		SendMode:   send_message_service.SendModeTemplate,
		TaskID:     msg.ID,
		TemplateID: template.ID,
		LogType:    "cron_message",
		TaskType:   "cron_message",
		Name:       msg.Name,
		Title:      msg.Name,
		Text:       textContent,
		HTML:       htmlContent,
		MarkDown:   markdownContent,
		URL:        "",
		CallerIp:   fmt.Sprintf("[CronTemplate] [%s] ID: %s", template.Name, template.ID),
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": "[Cron Message]",
		}),
	}
	taskData, _ := sender.SendPreCheck()
	_, err = sender.Send(taskData)
	if err != nil {
		logrus.Errorf("执行定时消息失败：%s", err.Error())
		return
	}
}

// UpdateCronMsgToCronServer 更新定时服务的任务
func UpdateCronMsgToCronServer(msg models.CronMessages) {
	if entryId, ok := constant.CronMsgIdMapMemoryCache[msg.ID]; ok {
		// 先删除之前的定时任务
		delete(constant.CronMsgIdMapMemoryCache, msg.ID)
		cron_service.RemoveTask(entryId)
		// 再注册新的定时任务
		AddCronMsgToCronServer(msg)
	} else {
		// 注册新的定时任务
		AddCronMsgToCronServer(msg)
	}
	logrus.Infof("完成定时消息的定时更新，消息id: %s，当前任务总数：%d", msg.ID, len(constant.CronMsgIdMapMemoryCache))
}

// RemoveCronMsgToCronServer 删除定时任务中心的任务
func RemoveCronMsgToCronServer(msg models.CronMessages) {
	if entryId, ok := constant.CronMsgIdMapMemoryCache[msg.ID]; ok {
		// 先删除之前的定时任务
		delete(constant.CronMsgIdMapMemoryCache, msg.ID)
		cron_service.RemoveTask(entryId)
	}
	logrus.Infof("删除定时消息完成，消息id: %s，剩余任务总数：%d", msg.ID, len(constant.CronMsgIdMapMemoryCache))
}

// StartUpMsgCronTask 启动注册定时任务
func StartUpUserSetupMsgCronTask() {
	MsgCronTask{}.Register()
}

// SendCronMessage 发送定时消息（用于立即发送）
func SendCronMessage(msg models.CronMessages, callerIP string) error {
	template, err := models.GetTemplateByID(msg.TemplateID)
	if err != nil {
		return fmt.Errorf("消息模板不存在: %s", msg.TemplateID)
	}

	if template.Status != "enabled" {
		return fmt.Errorf("消息模板已禁用: %s", msg.TemplateID)
	}

	textContent, htmlContent, markdownContent := buildTemplateContent(template)
	sender := send_message_service.SendMessageService{
		SendMode:   send_message_service.SendModeTemplate,
		TaskID:     msg.ID,
		TemplateID: template.ID,
		LogType:    "cron_message",
		TaskType:   "cron_message",
		Name:       msg.Name,
		Title:      msg.Name,
		Text:       textContent,
		HTML:       htmlContent,
		MarkDown:   markdownContent,
		URL:        "",
		CallerIp:   callerIP,
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": "[Manual Send Cron Message]",
		}),
	}

	// 预检查
	taskData, err := sender.SendPreCheck()
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "没有关联任何实例") {
			return fmt.Errorf("该消息模板尚未配置发送实例，请先在【消息模板】页面为模板 [%s] 添加至少一个发送实例", template.Name)
		}
		return fmt.Errorf("发送预检查失败: %s", errMsg)
	}

	// 发送消息
	_, err = sender.Send(taskData)
	if err != nil {
		return fmt.Errorf("发送失败: %s", err.Error())
	}

	logrus.Infof("立即发送定时消息成功，消息id: %s，消息名: %s", msg.ID, msg.Name)
	return nil
}

type TemplatePlaceholder struct {
	Key     string `json:"key"`
	Default string `json:"default"`
}

func buildTemplateContent(template *models.TemplateResult) (string, string, string) {
	placeholders := buildTemplateDefaults(template.Placeholders)
	textContent := replacePlaceholders(template.TextTemplate, placeholders)
	htmlContent := replacePlaceholders(template.HTMLTemplate, placeholders)
	markdownContent := replacePlaceholders(template.MarkdownTemplate, placeholders)
	return textContent, htmlContent, markdownContent
}

func buildTemplateDefaults(placeholders string) map[string]interface{} {
	result := map[string]interface{}{}
	if placeholders == "" {
		return result
	}
	var defs []TemplatePlaceholder
	if err := json.Unmarshal([]byte(placeholders), &defs); err != nil {
		return result
	}
	for _, def := range defs {
		if def.Key == "" {
			continue
		}
		if def.Default != "" {
			result[def.Key] = def.Default
		}
	}
	return result
}

func replacePlaceholders(template string, placeholders map[string]interface{}) string {
	if template == "" || placeholders == nil {
		return template
	}
	result := template
	for key, value := range placeholders {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}
