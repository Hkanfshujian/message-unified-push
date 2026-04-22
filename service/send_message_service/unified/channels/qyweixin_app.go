package channels

import (
	"encoding/json"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/message"
	"strings"
)

type QyWeiXinAppChannel struct {
	*BaseChannel
}

func NewQyWeiXinAppChannel() *QyWeiXinAppChannel {
	return &QyWeiXinAppChannel{
		BaseChannel: NewBaseChannel(MessageTypeQyWeiXinApp, []string{FormatTypeMarkdown, FormatTypeText}),
	}
}

type QyWeiXinAppAuth struct {
	CorpID     string `json:"corp_id"`
	CorpSecret string `json:"corp_secret"`
	AgentID    string `json:"agent_id"`
	ToUser     string `json:"to_user"`
}

func (c *QyWeiXinAppChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	// 1. 尝试将 msgObj 转换为特定的 auth 结构
	authData, ok := msgObj.(*QyWeiXinAppAuth)
	if !ok {
		// 如果转换失败，尝试从 JSON 重新解析（为了兼容不同的调用路径）
		authData = &QyWeiXinAppAuth{}
		if authJson, ok := msgObj.(string); ok {
			_ = json.Unmarshal([]byte(authJson), authData)
		} else if b, err := json.Marshal(msgObj); err == nil {
			_ = json.Unmarshal(b, authData)
		}
	}

	if authData.CorpID == "" || authData.CorpSecret == "" || authData.AgentID == "" {
		return "企业微信应用配置不完整 (CorpID/CorpSecret/AgentID)", ""
	}

	// 支持动态接收者：优先读取实例配置中的 to_user（由动态接收模式注入）
	if strings.TrimSpace(ins.Config) != "" {
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(ins.Config), &cfg); err == nil {
			if v, ok := cfg["to_user"]; ok {
				if toUser, ok := v.(string); ok && strings.TrimSpace(toUser) != "" {
					authData.ToUser = strings.TrimSpace(toUser)
				}
			}
		}
	}
	if strings.TrimSpace(authData.ToUser) == "" {
		return "企业微信应用接收者 to_user 为空，已阻止默认 @all 群发", ""
	}

	// 2. 选择合适的格式
	formatType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return fmt.Sprintf("消息格式化失败: %v", err), ""
	}
	if formattedContent == "" {
		return "消息内容不能为空", ""
	}

	// 3. 初始化发送器
	app := &message.QyWeiXinApp{
		CorpID:     authData.CorpID,
		CorpSecret: authData.CorpSecret,
		AgentID:    authData.AgentID,
		ToUser:     authData.ToUser,
	}

	// 4. 发送消息
	var sendErr error
	var msgId string

	if formatType == FormatTypeMarkdown {
		msgId, sendErr = app.SendMarkdown(formattedContent, content.GetAtMobiles()...)
	} else {
		msgId, sendErr = app.SendText(formattedContent, content.GetAtMobiles()...)
	}

	if sendErr != nil {
		return sendErr.Error(), ""
	}

	return msgId, ""
}
