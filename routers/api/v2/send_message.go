package v2

import (
	"encoding/json"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/e"
	utilpkg "ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/send_message_service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SendMessageByTemplateReq struct {
	Token        string                 `json:"token" validate:"required" label:"模板token"`
	Title        string                 `json:"title" validate:"required" label:"消息标题"`
	Placeholders map[string]interface{} `json:"placeholders" label:"占位符"`
	Recipients   []string               `json:"recipients" label:"接收者列表"`
	WaitResult   bool                   `json:"wait_result" label:"是否同步等待发送结果"`
}

type TemplatePlaceholder struct {
	Key     string `json:"key"`
	Default string `json:"default"`
}

// DoSendMessageByTemplate 使用模板发送消息
func DoSendMessageByTemplate(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  SendMessageByTemplateReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	// 清洗 recipients：去空白、去重
	normalizedRecipients := normalizeRecipients(req.Recipients)
	if len(req.Recipients) > 0 && len(normalizedRecipients) == 0 {
		appG.CResponse(http.StatusBadRequest, "recipients 参数不能为空字符串，请传入有效接收者", nil)
		return
	}
	req.Recipients = normalizedRecipients

	// 解析 token 为模板 ID
	templateID, err := utilpkg.DecryptTokenHex(req.Token, 71) // 71 为简单对称密钥
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("token解析失败：%v", err), nil)
		return
	}

	// 获取模板
	template, err := models.GetTemplateByID(templateID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("模板不存在：%s", err), nil)
		return
	}

	// 检查模板状态
	if template.Status != "enabled" {
		appG.CResponse(http.StatusBadRequest, "模板已禁用", nil)
		return
	}

	// 填充默认占位符
	if template.Placeholders != "" {
		var placeholderDefs []TemplatePlaceholder
		if err := json.Unmarshal([]byte(template.Placeholders), &placeholderDefs); err == nil {
			if req.Placeholders == nil {
				req.Placeholders = make(map[string]interface{})
			}
			for _, def := range placeholderDefs {
				// 如果请求中没有该占位符，且配置了默认值，则使用默认值
				if _, exists := req.Placeholders[def.Key]; !exists && def.Default != "" {
					req.Placeholders[def.Key] = def.Default
				}
			}
		} else {
			logrus.Errorf("解析模板占位符配置失败: %v", err)
		}
	}

	// 替换占位符
	textContent := replacePlaceholders(template.TextTemplate, req.Placeholders)
	htmlContent := replacePlaceholders(template.HTMLTemplate, req.Placeholders)
	markdownContent := replacePlaceholders(template.MarkdownTemplate, req.Placeholders)

	// 解析@提醒配置
	var atMobiles []string
	var atUserIds []string
	if template.AtMobiles != "" {
		atMobiles = strings.Split(template.AtMobiles, ",")
		// 去除空格
		for i := range atMobiles {
			atMobiles[i] = strings.TrimSpace(atMobiles[i])
		}
	}
	if template.AtUserIds != "" {
		atUserIds = strings.Split(template.AtUserIds, ",")
		// 去除空格
		for i := range atUserIds {
			atUserIds[i] = strings.TrimSpace(atUserIds[i])
		}
	}

	// 获取模板关联的实例列表
	insList, err := models.GetTemplateInsList(templateID)
	if err != nil || len(insList) == 0 {
		appG.CResponse(http.StatusBadRequest, "模板没有配置发送实例", nil)
		return
	}

	// 过滤启用的实例
	var enabledCount int
	for _, ins := range insList {
		if ins.Enable == 1 {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		appG.CResponse(http.StatusBadRequest, "模板没有启用的发送实例", nil)
		return
	}

	// 使用发送服务进行发送
	// 将模板ID作为TaskID传入，用于日志记录
	msgService := send_message_service.SendMessageService{
		SendMode:   send_message_service.SendModeTemplate, // 明确标记为模板模式
		TaskID:     templateID,                            // 使用模板ID作为TaskID（用于日志记录）
		TemplateID: templateID,                            // 模板ID
		Name:       template.Name,                         // 模板名称
		Title:      req.Title,
		Text:       textContent,
		HTML:       htmlContent,
		MarkDown:   markdownContent,
		CallerIp:   c.ClientIP(),
		AtMobiles:  atMobiles,
		AtUserIds:  atUserIds,
		AtAll:      template.IsAtAll,
		Recipients: req.Recipients, // 动态接收者列表
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": "[Template Send]",
		}),
	}

	// 发送前检查
	task, err := msgService.SendPreCheck()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("发送检查不通过：%s", err), nil)
		return
	}

	if req.WaitResult {
		// 同步发送：直接返回最终结果，便于 curl 联调定位失败原因
		result, sendErr := msgService.Send(task)
		returnContent := extractReturnContent(result)
		if sendErr != nil {
			appG.CResponse(http.StatusBadRequest, sendErr.Error(), map[string]interface{}{
				"token":          req.Token,
				"count":          enabledCount,
				"return_content": returnContent,
				"status":         "failed",
			})
			return
		}
		appG.CResponse(http.StatusOK, "success", map[string]interface{}{
			"token":          req.Token,
			"count":          enabledCount,
			"return_content": returnContent,
			"status":         "success",
		})
		return
	}

	// 异步发送（默认）
	msgService.AsyncSend(task)
	appG.CResponse(http.StatusOK, "accepted", map[string]interface{}{
		"token": req.Token,
		"count": enabledCount,
		"mode":  "async",
	})
}

func extractReturnContent(detail string) string {
	lines := strings.Split(detail, "\n")
	// 优先返回最关键的“返回内容”
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "返回内容：") {
			return strings.TrimSpace(strings.TrimPrefix(line, "返回内容："))
		}
	}
	// 失败场景兜底提取
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "发送失败：") {
			return strings.TrimSpace(strings.TrimPrefix(line, "发送失败："))
		}
	}
	// 最后兜底：取最后一行非空内容
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}
	return ""
}

func normalizeRecipients(recipients []string) []string {
	if len(recipients) == 0 {
		return recipients
	}
	result := make([]string, 0, len(recipients))
	seen := map[string]bool{}
	for _, r := range recipients {
		v := strings.TrimSpace(r)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		result = append(result, v)
	}
	return result
}

// replacePlaceholders 替换模板中的占位符
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
