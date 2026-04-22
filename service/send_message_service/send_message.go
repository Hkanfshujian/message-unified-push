package send_message_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/send_message_service/unified"
	"ops-message-unified-push/service/send_way_service"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	SendSuccess = 1
	SendFail    = 0
)

// 发送模式类型
const (
	SendModeTemplate = "template" // 模板模式
)

func errStrIsSuccess(errStr string) int {
	if errStr == "" {
		return SendSuccess
	}
	return SendFail
}

type SendMessageService struct {
	SendMode             string // 发送模式：task(任务模式) 或 template(模板模式)
	TaskID               string // 任务ID（任务模式）或模板ID（模板模式，用于日志记录）
	TemplateID           string // 模板ID（仅模板模式使用）
	PreferredContentType string // 模板模式下优先使用的内容格式(text/html/markdown)
	LogType              string // 日志类型：task/template/cron
	TaskType             string // 统计任务类型：task/template/cron
	Name                 string // 任务或模板名称（用于日志记录）
	Title                string
	Text                 string
	HTML                 string
	URL                  string
	MarkDown             string
	CallerIp             string

	// @提及相关字段
	AtMobiles []string
	AtUserIds []string
	AtAll     bool

	// 动态接收者（用于邮箱、微信公众号等支持群发的渠道）
	Recipients []string

	Status    int
	LogOutput []string
	MsgIDs    []string

	DefaultLogger *logrus.Entry
}

// LogsAndStatusMark 记录执行的日志和状态标记
func (sm *SendMessageService) LogsAndStatusMark(errStr string, status int) {
	sm.LogOutput = append(sm.LogOutput, errStr)
	if status == SendFail {
		sm.Status = SendFail
		sm.DefaultLogger.Errorf("%s, 状态：%d", strings.Trim(errStr, "\n"), status)
		return
	}
	// 成功链路细节默认使用 Debug，避免 INFO 级别刷屏
	sm.DefaultLogger.Debugf("%s, 状态：%d", strings.Trim(errStr, "\n"), status)
}

// AsyncSend 异步发送一个消息任务的所有实例
func (sm *SendMessageService) AsyncSend(task models.TaskIns) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("AsyncSend: Recovered from panic:", r)
		}
	}()

	// 限制并发异步发送数量
	constant.MaxSendTaskSemaphoreChan <- ""
	defer func() {
		<-constant.MaxSendTaskSemaphoreChan
	}()

	go func() {
		entry := logrus.WithFields(logrus.Fields{
			"prefix": "[Send Goroutine]",
		})
		_, err := sm.Send(task)
		if err != nil {
			entry.Errorf("任务[%s][%s]发送错误： %s", sm.TaskID, sm.Title, err)
		} else {
			entry.Infof("完成任务[%s][%s]发送", sm.TaskID, sm.Title)
		}
	}()
}

// SendPreCheck 发送前数据准备和预检查（仅模板模式）
func (sm *SendMessageService) SendPreCheck() (models.TaskIns, error) {
	errStr := ""
	entry := logrus.WithFields(logrus.Fields{
		"prefix": "[Message PreChecK]",
	})

	var task models.TaskIns

	mode := strings.TrimSpace(sm.SendMode)
	if mode == "" {
		mode = SendModeTemplate
	}
	if mode != SendModeTemplate {
		errStr = fmt.Sprintf("不支持的 SendMode: %s（已移除 task 模式）", mode)
		entry.Error(errStr)
		return task, errors.New(errStr)
	}

	if sm.TemplateID == "" {
		errStr = "模板模式下 TemplateID 不能为空"
		entry.Error(errStr)
		return task, errors.New(errStr)
	}

	insList, err := models.GetTemplateInsList(sm.TemplateID)
	if err != nil {
		errStr = fmt.Sprintf("模板[%s]实例查询失败：%s", sm.TemplateID, err)
		entry.Error(errStr)
		return task, errors.New(errStr)
	}
	if len(insList) == 0 {
		errStr = fmt.Sprintf("模板[%s]没有关联任何实例！", sm.TemplateID)
		entry.Error(errStr)
		return task, errors.New(errStr)
	}
	task.ID = sm.TaskID
	task.InsData = make([]models.SendTasksInsRes, 0, len(insList))
	for _, ins := range insList {
		task.InsData = append(task.InsData, ins)
	}
	entry.Infof("模板[%s]加载了 %d 个实例", sm.TemplateID, len(insList))
	return task, nil
}

// Send 发送一个消息任务的所有实例
func (sm *SendMessageService) Send(task models.TaskIns) (string, error) {
	sm.Status = SendSuccess

	idForLog := ""
	switch sm.LogType {
	case "cron_message":
		idForLog = sm.TaskID
		if idForLog == "" && sm.TemplateID != "" && sm.Name != "" {
			if msg, err := models.GetCronMsgByTemplateAndName(sm.TemplateID, sm.Name); err == nil && msg.ID != "" {
				idForLog = msg.ID
			}
		}
	case "template":
		if sm.TemplateID != "" {
			idForLog = sm.TemplateID
		} else {
			idForLog = sm.TaskID
		}
	default:
		idForLog = sm.TaskID
	}
	label := "任务ID"
	if sm.LogType == "template" {
		label = "模板ID"
	}
	sm.LogsAndStatusMark(fmt.Sprintf("%s: %s", label, idForLog), sm.Status)

	for idx, ins := range task.InsData {
		way, err := models.GetWayByID(ins.WayID)
		if err != nil {
			errStr := fmt.Sprintf("渠道[%s]信息不存在！跳过这个实例的发送", ins.WayID)
			sm.LogsAndStatusMark(errStr, SendFail)
			continue
		}

		// 暂停了实例的发送
		if ins.Enable != 1 {
			//sm.LogsAndStatusMark("该实例发送已经被暂停，跳过发送！\n", sm.Status)
			continue
		}

		wayService := send_way_service.SendWay{
			ID:   fmt.Sprintf("%s", way.ID),
			Name: way.Name,
			Auth: way.Auth,
			Type: way.Type,
		}

		sm.LogsAndStatusMark(fmt.Sprintf(">> 实例 %d", idx+1), sm.Status)
		sm.LogsAndStatusMark(fmt.Sprintf("开始发送，实例: %s", ins.WayID), sm.Status)
		sm.LogsAndStatusMark(fmt.Sprintf("实例渠道名: %s", way.Name), sm.Status)
		preferredType := strings.TrimSpace(sm.PreferredContentType)
		if preferredType != "" {
			sm.LogsAndStatusMark(
				fmt.Sprintf("实例类型: %s (订阅模板格式: %s)", ins.WayType, preferredType),
				sm.Status,
			)
		} else {
			sm.LogsAndStatusMark(
				fmt.Sprintf("实例类型: %s + %s", ins.WayType, ins.ContentType),
				sm.Status,
			)
		}
		// sm.LogsAndStatusMark(fmt.Sprintf("实例配置: %s", ins.Config), sm.Status)

		// 发送渠道的校验
		errStr, msgObj := wayService.ValidateDiffWay()
		if errStr != "" {
			sm.LogsAndStatusMark(fmt.Sprintf("实例渠道认证校验失败: %s", errStr), SendFail)
			continue
		}

		// 使用新的Channel架构发送消息
		channelRegistry := unified.GetGlobalChannelRegistry()
		channel, ok := channelRegistry.GetChannel(way.Type)
		if !ok {
			sm.LogsAndStatusMark(fmt.Sprintf("发送失败：未知渠道类型 %s 的发信实例: %s\n", way.Type, ins.ID), SendFail)
			continue
		}

		// 根据发送模式构建消息内容
		var unifiedContent *unified.UnifiedMessageContent
		if sm.SendMode == SendModeTemplate {
			// 模板模式：根据实例的 ContentType 精确发送对应类型的内容
			unifiedContent = sm.BuildTemplateContent(ins.SendTasksIns)
			if unifiedContent == nil {
				sm.LogsAndStatusMark(fmt.Sprintf("模板内容为空，实例类型: %s", ins.ContentType), SendFail)
				continue
			}
			sm.LogsAndStatusMark(
				fmt.Sprintf("模板实际发送格式: %s", sm.resolveTemplateContentType(ins.SendTasksIns)),
				sm.Status,
			)
		} else {
			// 任务模式：使用现有逻辑（支持内容类型回退）
			typeC, content := sm.GetSendMsg(ins.SendTasksIns)
			if content == "" {
				sm.LogsAndStatusMark(fmt.Sprintf("发送内容为空，设置的类型: %s，实际检测的类型: %s", ins.SendTasksIns.ContentType, typeC), SendFail)
				continue
			}
			// 构建统一消息内容（支持@功能）
			unifiedContent = &unified.UnifiedMessageContent{
				Title:     sm.Title,
				Text:      sm.Text,
				HTML:      sm.HTML,
				Markdown:  sm.MarkDown,
				URL:       sm.URL,
				AtMobiles: sm.AtMobiles,
				AtUserIds: sm.AtUserIds,
				AtAll:     sm.AtAll,
			}
		}

		// 处理动态接收者（邮箱、微信公众号等支持群发的渠道）
		isDynamicMode := sm.isDynamicRecipientMode(ins.SendTasksIns)

		// 检查：如果提供了recipients但实例未启用动态模式，给出提示
		if !isDynamicMode && sm.supportsDynamicRecipient(way.Type) && len(sm.Recipients) > 0 {
			sm.LogsAndStatusMark(fmt.Sprintf("警告：提供了 %d 个接收者，但实例未启用动态接收模式（allowMultiRecip=true），将使用实例配置的固定接收者", len(sm.Recipients)), sm.Status)
		}

		if isDynamicMode && sm.supportsDynamicRecipient(way.Type) && len(sm.Recipients) > 0 {
			// 动态接收模式：使用API传入的Recipients列表（群发）
			sm.LogsAndStatusMark(fmt.Sprintf("动态接收模式[共 %d 个接收者]", len(sm.Recipients)), sm.Status)
			for recipientIdx, recipient := range sm.Recipients {
				recipient = strings.TrimSpace(recipient)
				if recipient == "" {
					sm.LogsAndStatusMark("发送失败：接收者为空，已拦截", SendFail)
					continue
				}
				sm.LogsAndStatusMark(fmt.Sprintf(">>> 接收者 %d/%d: %s", recipientIdx+1, len(sm.Recipients), recipient), sm.Status)

				// 临时修改实例配置中的接收者
				modifiedIns := sm.modifyInsRecipient(ins.SendTasksIns, recipient, way.Type)

				// 使用 SendUnified 方法发送
				res, errMsg := channel.SendUnified(msgObj, modifiedIns, unifiedContent)
				if res != "" {
					sm.recordMsgID(fmt.Sprintf("实例[%s]接收者[%s]", ins.WayID, recipient), res)
					sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s", res), sm.Status)
				} else {
					sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
				}
			}
		} else {
			// 固定接收模式：使用实例配置的to_account
			sm.LogsAndStatusMark("固定接收模式", sm.Status)
			res, errMsg := channel.SendUnified(msgObj, ins.SendTasksIns, unifiedContent)
			if res != "" {
				sm.recordMsgID(fmt.Sprintf("实例[%s]", ins.WayID), res)
				sm.LogsAndStatusMark(fmt.Sprintf("返回内容：%s\n", res), sm.Status)
			} else {
				sm.LogsAndStatusMark(sm.TransError(errMsg), errStrIsSuccess(errMsg))
			}
		}

	}

	// 追加记录发送内容
	sm.AppendSendContent()
	// 日志写到数据库
	sm.RecordSendLog()
	// 更新统计数据（任务级别：一次任务算一次）
	sm.UpdateSendStats()

	totalOutputLog := strings.Join(sm.LogOutput, "\n")
	if sm.Status == SendSuccess {
		return totalOutputLog, nil
	} else {
		return totalOutputLog, errors.New("发送过程中有失败，请检查详细日志")
	}
}

// AppendSendContent 添加发送内容
func (sm *SendMessageService) AppendSendContent() {
	sm.LogOutput = append(sm.LogOutput, fmt.Sprintf(">> 发送的内容:"))
	if sm.Text != "" {
		sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("Text: %s \n", sm.Text))
	}
	if sm.HTML != "" {
		sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("HTML: %s \n", sm.HTML))
	}
	if sm.MarkDown != "" {
		sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("MarkDown: %s \n", sm.MarkDown))
	}
	if len(sm.MsgIDs) > 0 {
		sm.LogOutput = append(sm.LogOutput, ">> 消息ID列表:")
		for idx, id := range sm.MsgIDs {
			sm.LogOutput = append(sm.LogOutput, fmt.Sprintf("%d. %s", idx+1, id))
		}
	}
}

// RecordSendLog 记录发送日志
func (sm *SendMessageService) RecordSendLog() {
	// 确定日志类型
	logType := sm.LogType
	if logType == "" {
		logType = "task"
		if sm.SendMode == SendModeTemplate {
			logType = "template"
		}
	}

	log := models.SendTasksLogs{
		Log:      strings.Join(sm.LogOutput, "\n"),
		TaskID:   sm.TaskID,
		Type:     logType,
		Name:     sm.Name,
		Status:   &sm.Status,
		CallerIp: sm.CallerIp,
	}
	err := log.Add()
	if err != nil {
		sm.DefaultLogger.Errorf("添加日志失败！原因是：%s", err)
	}
}

// UpdateSendStats 更新发送统计数据（任务级别）
func (sm *SendMessageService) UpdateSendStats() {
	// 获取当前日期
	currentDay := sm.getCurrentDay()

	// 确定任务类型
	taskType := sm.TaskType
	if taskType == "" {
		taskType = "task"
		if sm.SendMode == SendModeTemplate {
			taskType = "template"
		}
	}

	// 根据任务的最终状态更新统计（一次任务只记录一次）
	var status string
	if sm.Status == SendSuccess {
		status = "success"
	} else {
		status = "failed"
	}

	// 更新统计：每次任务执行记录为1次
	err := models.IncrementSendStats(sm.TaskID, taskType, currentDay, status, 1)
	if err != nil {
		logrus.Errorf("更新发送统计失败：%s", err)
	}
}

// getCurrentDay 获取当前日期（YYYY-MM-DD格式）
func (sm *SendMessageService) getCurrentDay() string {
	return util.GetNowTimeStr()[:10]
}

// TransError 转化错误
func (sm *SendMessageService) TransError(err string) string {
	if err == "" {
		return "发送成功！\n"
	} else {
		return fmt.Sprintf("发送失败：%s！\n", err)
	}
}

// BuildTemplateContent 构建模板模式的消息内容
// 模板模式：根据实例的 ContentType 精确匹配对应类型的内容，只传递该类型的内容
func (sm *SendMessageService) BuildTemplateContent(ins models.SendTasksIns) *unified.UnifiedMessageContent {
	contentType := sm.resolveTemplateContentType(ins)

	// 内容类型映射表
	contentMap := map[string]string{
		unified.FormatTypeText:     sm.Text,
		unified.FormatTypeHTML:     sm.HTML,
		unified.FormatTypeMarkdown: sm.MarkDown,
	}

	// 检查内容是否存在
	contentValue, exists := contentMap[contentType]
	if !exists {
		logrus.Warnf("模板模式：未知的内容类型 %s", contentType)
		return nil
	}
	if contentValue == "" {
		logrus.Warnf("模板模式：实例要求的 %s 类型内容为空", contentType)
		return nil
	}

	// 构建消息内容，只填充实例要求的类型
	content := &unified.UnifiedMessageContent{
		Title:     sm.Title,
		URL:       sm.URL,
		AtMobiles: sm.AtMobiles,
		AtUserIds: sm.AtUserIds,
		AtAll:     sm.AtAll,
	}

	// 根据类型填充对应字段
	switch contentType {
	case unified.FormatTypeText:
		content.Text = contentValue
	case unified.FormatTypeHTML:
		content.HTML = contentValue
	case unified.FormatTypeMarkdown:
		content.Markdown = contentValue
	}

	return content
}

func (sm *SendMessageService) resolveTemplateContentType(ins models.SendTasksIns) string {
	contentType := strings.ToLower(strings.TrimSpace(sm.PreferredContentType))
	if contentType == "" {
		contentType = strings.ToLower(ins.ContentType)
	}
	return contentType
}

func (sm *SendMessageService) recordMsgID(scene, msgID string) {
	id := strings.TrimSpace(msgID)
	if id == "" {
		return
	}
	for _, existing := range sm.MsgIDs {
		if existing == id {
			return
		}
	}
	sm.MsgIDs = append(sm.MsgIDs, id)
	sm.DefaultLogger.Infof("发送成功[%s] msgId=%s", scene, id)
}

// supportsDynamicRecipient 判断渠道是否支持动态接收者
func (sm *SendMessageService) supportsDynamicRecipient(wayType string) bool {
	// 支持动态接收者的渠道类型
	supportedTypes := map[string]bool{
		constant.MessageTypeEmail:           true,
		constant.MessageTypeWeChatOFAccount: true,
		constant.MessageTypeAliyunSMS:       true,
		constant.MessageTypeQyWeiXinApp:     true,
		// 可以继续添加其他支持动态接收者的渠道
	}
	return supportedTypes[wayType]
}

// isDynamicRecipientMode 判断实例是否配置为动态接收模式
func (sm *SendMessageService) isDynamicRecipientMode(ins models.SendTasksIns) bool {
	// 解析实例配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ins.Config), &config); err != nil {
		return false
	}

	// 检查allowMultiRecip字段
	// allowMultiRecip=true: 动态模式
	// allowMultiRecip=false或不存在: 固定模式
	if allowMultiRecip, ok := config["allowMultiRecip"]; ok {
		if allow, ok := allowMultiRecip.(bool); ok {
			return allow
		}
	}

	// 默认为固定模式（兼容历史数据）
	return false
}

// modifyInsRecipient 临时修改实例配置中的接收者
func (sm *SendMessageService) modifyInsRecipient(ins models.SendTasksIns, recipient string, wayType string) models.SendTasksIns {
	// 创建副本
	modifiedIns := ins

	// 根据渠道类型修改配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ins.Config), &config); err != nil {
		sm.LogsAndStatusMark(fmt.Sprintf("解析实例配置失败: %s", err.Error()), SendFail)
		return ins
	}

	// 根据渠道类型写入对应的接收者字段
	switch wayType {
	case constant.MessageTypeQyWeiXinApp:
		config["to_user"] = recipient
	case constant.MessageTypeAliyunSMS:
		config["phone_number"] = recipient
	default:
		config["to_account"] = recipient
	}

	// 序列化回JSON
	modifiedConfigBytes, err := json.Marshal(config)
	if err != nil {
		sm.LogsAndStatusMark(fmt.Sprintf("序列化实例配置失败: %s", err.Error()), SendFail)
		return ins
	}

	modifiedIns.Config = string(modifiedConfigBytes)
	return modifiedIns
}

// GetSendMsg 获取对应消息内容（任务模式使用）
// 先根据实例设置的类型取，取不到或者取到的是空，则使用text发送
func (sm *SendMessageService) GetSendMsg(ins models.SendTasksIns) (string, string) {
	data := map[string]string{}
	data[unified.FormatTypeText] = sm.Text
	data[unified.FormatTypeHTML] = sm.HTML
	data[unified.FormatTypeMarkdown] = sm.MarkDown
	content, ok := data[strings.ToLower(ins.ContentType)]
	if !ok || len(content) == 0 {
		content, ok := data[unified.FormatTypeText]
		if !ok {
			logrus.Error("text节点数据为空！")
			return unified.FormatTypeText, ""
		} else {
			logrus.Error(fmt.Sprintf("没有找到%s对应的消息，使用text消息替代！", ins.ContentType))
			return unified.FormatTypeText, content
		}
	} else {
		return strings.ToLower(ins.ContentType), content
	}
}
