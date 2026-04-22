package mq_consumer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/send_message_service"
	"ops-message-unified-push/service/subscription_rule"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ConsumerManager 消费者管理器
type ConsumerManager struct {
	mu        sync.RWMutex
	consumers map[string]rocketmq.PushConsumer // subscriptionID -> consumer
}

// NewConsumerManager 创建消费者管理器
func NewConsumerManager() *ConsumerManager {
	return &ConsumerManager{
		consumers: make(map[string]rocketmq.PushConsumer),
	}
}

// StartSubscription 启动订阅
func (cm *ConsumerManager) StartSubscription(subscriptionID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查是否已经在运行
	if _, exists := cm.consumers[subscriptionID]; exists {
		return fmt.Errorf("订阅 %s 已在运行中", subscriptionID)
	}

	// 获取订阅信息
	sub := models.Subscription{}
	if err := sub.GetByUUID(subscriptionID); err != nil {
		return fmt.Errorf("获取订阅信息失败: %w", err)
	}

	if sub.Status != "running" {
		return fmt.Errorf("订阅状态不是 running")
	}

	// 获取数据源信息
	source, err := models.GetMQSourceByID(sub.SourceID)
	if err != nil {
		return fmt.Errorf("获取数据源信息失败: %w", err)
	}

	if source.Enabled != 1 {
		return fmt.Errorf("数据源 %s 已禁用", source.Name)
	}

	// 创建 Push Consumer
	consumerOpts := []consumer.Option{
		consumer.WithGroupName(sub.GroupName),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(parseNamesrvAddrs(source.NamesrvAddr))),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithMaxReconsumeTimes(3),
	}
	if source.AccessKey != "" && source.SecretKey != "" {
		consumerOpts = append(consumerOpts, consumer.WithCredentials(primitive.Credentials{
			AccessKey: source.AccessKey,
			SecretKey: source.SecretKey,
		}))
	}
	c, err := rocketmq.NewPushConsumer(consumerOpts...)
	if err != nil {
		return fmt.Errorf("创建消费者失败: %w", err)
	}

	// 订阅 Topic
	tag := FormatTag(sub.Tag)
	err = c.Subscribe(sub.Topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tag,
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		return cm.handleMessages(subscriptionID, msgs)
	})
	if err != nil {
		return fmt.Errorf("订阅 Topic 失败: %w", err)
	}

	// 启动消费者
	err = c.Start()
	if err != nil {
		return fmt.Errorf("启动消费者失败: %w", err)
	}

	// 保存消费者实例
	cm.consumers[subscriptionID] = c

	logrus.Infof("订阅 %s 启动成功, Topic: %s, Tag: %s", subscriptionID, sub.Topic, tag)
	return nil
}

// StopSubscription 停止订阅
func (cm *ConsumerManager) StopSubscription(subscriptionID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	c, exists := cm.consumers[subscriptionID]
	if !exists {
		return fmt.Errorf("订阅 %s 未在运行", subscriptionID)
	}

	err := c.Shutdown()
	if err != nil {
		return fmt.Errorf("停止消费者失败: %w", err)
	}

	delete(cm.consumers, subscriptionID)
	logrus.Infof("订阅 %s 已停止", subscriptionID)
	return nil
}

// StopAll 停止所有订阅
func (cm *ConsumerManager) StopAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for id, c := range cm.consumers {
		err := c.Shutdown()
		if err != nil {
			logrus.Errorf("停止订阅 %s 失败: %v", id, err)
		} else {
			logrus.Infof("订阅 %s 已停止", id)
		}
	}
	cm.consumers = make(map[string]rocketmq.PushConsumer)
}

// StartAllRunning 启动所有运行中的订阅
func (cm *ConsumerManager) StartAllRunning() {
	// 获取所有 running 状态的订阅
	subscriptions, err := models.GetSubscriptions(0, 0, "", "", "running", "")
	if err != nil {
		logrus.Errorf("获取运行中的订阅失败: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		logrus.Info("没有需要启动的订阅")
		return
	}

	logrus.Infof("开始启动 %d 个运行中的订阅", len(subscriptions))
	successCount := 0
	failCount := 0

	for _, sub := range subscriptions {
		// 使用新的 goroutine 启动，避免阻塞
		go func(subID string) {
			// 等待 1 秒再启动，避免并发问题
			time.Sleep(1 * time.Second)
			if err := cm.StartSubscription(subID); err != nil {
				logrus.Errorf("启动订阅 %s 失败: %v", subID, err)
			} else {
				logrus.Infof("订阅 %s 启动成功", subID)
			}
		}(sub.ID)
	}

	logrus.Infof("订阅启动任务已提交，成功: %d, 失败: %d", successCount, failCount)
}

// handleMessages 处理消息
func (cm *ConsumerManager) handleMessages(subscriptionID string, msgs []*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for _, msg := range msgs {
		err := cm.processSingleMessage(subscriptionID, msg)
		if err != nil {
			logrus.Errorf("处理消息失败 [Sub: %s, MsgID: %s]: %v", subscriptionID, msg.MsgId, err)
			// 返回重试，RocketMQ 会自动重试
			return consumer.ConsumeRetryLater, err
		}
	}
	return consumer.ConsumeSuccess, nil
}

// processSingleMessage 处理单条消息
func (cm *ConsumerManager) processSingleMessage(subscriptionID string, msg *primitive.MessageExt) error {
	// 获取订阅信息
	sub := models.Subscription{}
	if err := sub.GetByUUID(subscriptionID); err != nil {
		return fmt.Errorf("获取订阅信息失败: %w", err)
	}

	rawMessage, charsetUsed, decodedByFallback := decodeMessageBody(msg.Body)
	logrus.Infof(
		"收到消息 [Sub: %s, MsgID: %s, Charset: %s, Fallback: %t]: %s",
		subscriptionID,
		msg.MsgId,
		charsetUsed,
		decodedByFallback,
		previewText(rawMessage, 200),
	)
	logrus.Infof(
		"消息字节预览 [Sub: %s, MsgID: %s]: len=%d utf8_valid=%t hex=%s",
		subscriptionID,
		msg.MsgId,
		len(msg.Body),
		utf8.Valid(msg.Body),
		previewHex(msg.Body, 48),
	)

	// 创建消费日志
	consumeLog := models.ConsumeLog{
		SubscriptionID: subscriptionID,
		MsgID:          msg.MsgId,
		Topic:          msg.Topic,
		Tag:            msg.GetTags(),
		RawMessage:     rawMessage,
		ConsumeTime:    util.Time(time.Now()),
	}

	// 第一步：验证正则
	if sub.ValidateRegex != "" {
		matched, err := MatchTextWithPattern(rawMessage, sub.ValidateRegex)
		if err != nil {
			return fmt.Errorf("验证正则表达式错误: %w", err)
		}
		if !matched {
			// 不匹配，记录日志但不发送
			consumeLog.Matched = 0
			if err := consumeLog.Create(); err != nil {
				logrus.Errorf("创建消费日志失败: %v", err)
			}
			return nil // 不匹配不算错误，返回成功
		}
		consumeLog.Matched = 1
	} else {
		consumeLog.Matched = 1
	}

	// 第二步：提取字段
	var extractedValues map[string]string
	extractRules := subscription_rule.ParseStoredExtractRules(sub.ExtractRegex, sub.ExtractField)
	if len(extractRules) > 0 {
		values, err := subscription_rule.BuildExtractMap(rawMessage, extractRules)
		if err != nil {
			return fmt.Errorf("提取字段失败: %w", err)
		}
		extractedValues = values
	}
	consumeLog.ExtractedValues = extractedValues

	// 第三步：发送消息
	if sub.TemplateID != "" {
		err := cm.sendMessageWithTemplate(&sub, extractedValues)
		if err != nil {
			consumeLog.SendStatus = 2 // 失败
			consumeLog.SendError = err.Error()
			logrus.Errorf("发送消息失败 [Sub: %s]: %v", subscriptionID, err)
		} else {
			consumeLog.SendStatus = 1 // 成功
		}
	} else {
		consumeLog.SendStatus = 0 // 未发送
	}

	// 保存消费日志
	if err := consumeLog.Create(); err != nil {
		logrus.Errorf("创建消费日志失败: %v", err)
	}

	// 更新订阅统计
	if err := models.UpdateSubscriptionStats(subscriptionID, consumeLog.Matched, consumeLog.SendStatus); err != nil {
		logrus.Errorf("更新订阅统计失败: %v", err)
	}

	return nil
}

// sendMessageWithTemplate 使用模板发送消息
func (cm *ConsumerManager) sendMessageWithTemplate(sub *models.Subscription, variables map[string]string) error {
	// 获取模板
	template, err := models.GetTemplateByID(sub.TemplateID)
	if err != nil {
		return fmt.Errorf("获取模板失败: %w", err)
	}
	if template.Status != "enabled" {
		return fmt.Errorf("模板[%s]已禁用", template.Name)
	}

	// 构建消息内容（替换变量）
	title := ReplaceVariables(template.Name, variables)
	textContent := ReplaceVariables(template.TextTemplate, variables)
	htmlContent := ReplaceVariables(template.HTMLTemplate, variables)
	markdownContent := ReplaceVariables(template.MarkdownTemplate, variables)
	if textContent == "" && htmlContent == "" && markdownContent == "" {
		return fmt.Errorf("模板内容为空，至少需要一种内容格式")
	}

	msgService := send_message_service.SendMessageService{
		SendMode:             send_message_service.SendModeTemplate,
		TaskID:               sub.ID,
		TemplateID:           sub.TemplateID,
		PreferredContentType: normalizeTemplateContentType(sub.ConsumeMode),
		LogType:              "subscription",
		TaskType:             "subscription",
		Name:                 sub.Name,
		Title:                title,
		Text:                 textContent,
		HTML:                 htmlContent,
		MarkDown:             markdownContent,
		CallerIp:             fmt.Sprintf("[MQ Subscription] %s", sub.ID),
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": "[Subscription Send]",
		}),
	}
	msgService.Recipients = extractDynamicRecipientsFromVariables(variables)
	if len(msgService.Recipients) > 0 {
		logrus.Infof("订阅[%s]提取到动态接收者 %d 个: %v", sub.ID, len(msgService.Recipients), msgService.Recipients)
	}
	taskData, err := msgService.SendPreCheck()
	if err != nil {
		return fmt.Errorf("发送预检查失败: %w", err)
	}
	sendOutput, err := msgService.Send(taskData)
	if err != nil {
		detail := summarizeSendOutput(sendOutput)
		if detail != "" {
			return fmt.Errorf("发送失败: %w\n%s", err, detail)
		}
		return fmt.Errorf("发送失败: %w", err)
	}
	return nil
}

func extractDynamicRecipientsFromVariables(variables map[string]string) []string {
	if len(variables) == 0 {
		return nil
	}
	// 强约束：动态接收者只认 to_user 字段
	raw := strings.TrimSpace(variables["to_user"])
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	splitted := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '|' || r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
	})
	res := make([]string, 0, len(splitted))
	seen := map[string]bool{}
	for _, item := range splitted {
		v := strings.TrimSpace(item)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		res = append(res, v)
	}
	return res
}

func normalizeTemplateContentType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "html":
		return "html"
	case "markdown":
		return "markdown"
	case "text":
		return "text"
	case "push", "pull", "":
		return "text"
	default:
		return "text"
	}
}

func summarizeSendOutput(output string) string {
	if strings.TrimSpace(output) == "" {
		return ""
	}
	lines := strings.Split(output, "\n")
	selected := make([]string, 0, 12)
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.Contains(l, "失败") ||
			strings.Contains(strings.ToLower(l), "error") ||
			strings.Contains(l, "校验") ||
			strings.Contains(l, "实例") ||
			strings.Contains(l, "返回内容") {
			selected = append(selected, l)
		}
		if len(selected) >= 12 {
			break
		}
	}
	if len(selected) == 0 {
		// 兜底返回前几行，避免没有上下文
		for _, line := range lines {
			l := strings.TrimSpace(line)
			if l == "" {
				continue
			}
			selected = append(selected, l)
			if len(selected) >= 6 {
				break
			}
		}
	}
	return "发送过程诊断信息:\n- " + strings.Join(selected, "\n- ")
}

// ReplaceVariables 替换模板变量
func ReplaceVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s}", key), value)
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), value)
	}
	return result
}

// FormatTag 格式化 Tag
func FormatTag(tag string) string {
	if tag == "" {
		return "*"
	}
	tag = strings.TrimSpace(tag)
	// 多个 Tag 用 || 分隔
	tag = strings.ReplaceAll(tag, ",", "||")
	tag = strings.ReplaceAll(tag, "，", "||")
	return tag
}

// ExtractFieldsWithRegex 使用正则提取字段
func ExtractFieldsWithRegex(rawMessage, extractRegex, extractField string) (map[string]string, error) {
	if extractField == "" {
		return nil, nil
	}
	value, err := subscription_rule.ExtractValue(rawMessage, extractRegex)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		extractField: value,
	}, nil
}

func MatchTextWithPattern(text, pattern string) (bool, error) {
	return subscription_rule.MatchText(text, pattern)
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseNamesrvAddrs(raw string) []string {
	normalized := strings.TrimSpace(raw)
	normalized = strings.ReplaceAll(normalized, "，", ";")
	normalized = strings.ReplaceAll(normalized, ",", ";")
	parts := strings.Split(normalized, ";")
	addrs := make([]string, 0, len(parts))
	for _, p := range parts {
		addr := strings.TrimSpace(p)
		if addr != "" {
			addrs = append(addrs, addr)
		}
	}
	if len(addrs) == 0 && normalized != "" {
		return []string{normalized}
	}
	return addrs
}

func decodeMessageBody(body []byte) (text, charset string, fallback bool) {
	if utf8.Valid(body) {
		return string(body), "utf-8", false
	}

	decoded, _, err := transform.String(simplifiedchinese.GB18030.NewDecoder(), string(body))
	if err == nil && decoded != "" {
		return decoded, "gb18030", true
	}

	// 无法识别编码时，尽量保留可见字符
	return string(bytes.ToValidUTF8(body, []byte{})), "unknown", false
}

func previewText(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func previewHex(b []byte, max int) string {
	if max <= 0 || len(b) == 0 {
		return ""
	}
	clip := b
	if len(clip) > max {
		clip = clip[:max]
	}
	hexStr := hex.EncodeToString(clip)
	if len(b) > max {
		return hexStr + "..."
	}
	return hexStr
}
