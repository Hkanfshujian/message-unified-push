package subscription_service

import (
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/service/mq_consumer"
	"ops-message-unified-push/service/subscription_rule"
	"strings"

	"github.com/sirupsen/logrus"
)

// GlobalConsumerManager 全局消费者管理器实例
var GlobalConsumerManager *mq_consumer.ConsumerManager

type SubscriptionService struct {
}

type AddSubscriptionRequest struct {
	Name                string `json:"name" binding:"required,max=200"`
	SourceID            string `json:"source_id" binding:"required"`
	Topic               string `json:"topic" binding:"required,max=200"`
	Tag                 string `json:"tag" binding:"max=200"`
	GroupName           string `json:"group_name" binding:"required,max=200"`
	ValidateRegex       string `json:"validate_regex"`
	ExtractRegex        string `json:"extract_regex"`
	ExtractField        string `json:"extract_field" binding:"max=100"`
	ExtractRules        []subscription_rule.ExtractRule `json:"extract_rules"`
	TemplateID          string `json:"template_id" binding:"required"`
	TemplateContentType string `json:"template_content_type" binding:"omitempty,oneof=text html markdown"`
	ConsumeMode         string `json:"consume_mode" binding:"omitempty,oneof=text html markdown push pull"`
}

type EditSubscriptionRequest struct {
	Name                string `json:"name" binding:"required,max=200"`
	SourceID            string `json:"source_id" binding:"required"`
	Topic               string `json:"topic" binding:"required,max=200"`
	Tag                 string `json:"tag" binding:"max=200"`
	GroupName           string `json:"group_name" binding:"required,max=200"`
	ValidateRegex       string `json:"validate_regex"`
	ExtractRegex        string `json:"extract_regex"`
	ExtractField        string `json:"extract_field" binding:"max=100"`
	ExtractRules        []subscription_rule.ExtractRule `json:"extract_rules"`
	TemplateID          string `json:"template_id" binding:"required"`
	TemplateContentType string `json:"template_content_type" binding:"omitempty,oneof=text html markdown"`
	ConsumeMode         string `json:"consume_mode" binding:"omitempty,oneof=text html markdown push pull"`
}

type RegexTestRequest struct {
	Message       string `json:"message" binding:"required"`
	ValidateRegex string `json:"validate_regex"`
	ExtractRegex  string `json:"extract_regex"`
	ExtractField  string `json:"extract_field" binding:"max=100"`
	ExtractRules  []subscription_rule.ExtractRule `json:"extract_rules"`
}

type RegexTestResult struct {
	ValidateMatched bool              `json:"validate_matched"`
	ExtractedValues map[string]string `json:"extracted_values,omitempty"`
}

type SubscriptionListItem struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	SourceID            string `json:"source_id"`
	SourceName          string `json:"source_name"`
	Topic               string `json:"topic"`
	Tag                 string `json:"tag"`
	GroupName           string `json:"group_name"`
	ValidateRegex       string `json:"validate_regex"`
	ExtractRegex        string `json:"extract_regex"`
	ExtractField        string `json:"extract_field"`
	ExtractRules        []subscription_rule.ExtractRule `json:"extract_rules"`
	TemplateID          string `json:"template_id"`
	TemplateName        string `json:"template_name"`
	TemplateContentType string `json:"template_content_type"`
	Status              string `json:"status"`
	TotalConsumed       int    `json:"total_consumed"`
	TotalSent           int    `json:"total_sent"`
	TotalFailed         int    `json:"total_failed"`
	LastConsumeTime     string `json:"last_consume_time"`
	CreatedOn           string `json:"created_on"`
}

func normalizeTemplateContentType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "html":
		return "html"
	case "markdown":
		return "markdown"
	case "text":
		return "text"
	// 兼容历史值
	case "push", "pull", "":
		return "text"
	default:
		return "text"
	}
}

func normalizeTemplateContentTypeWithAlias(templateContentType, consumeMode string) string {
	if strings.TrimSpace(templateContentType) != "" {
		return normalizeTemplateContentType(templateContentType)
	}
	return normalizeTemplateContentType(consumeMode)
}

// Add 新增订阅
func (s *SubscriptionService) Add(req AddSubscriptionRequest) (*models.Subscription, error) {
	// 校验正则表达式
	if req.ValidateRegex != "" {
		if err := ValidateRegexPattern(req.ValidateRegex); err != nil {
			return nil, fmt.Errorf("验证正则表达式错误: %w", err)
		}
	}
	if req.ExtractRegex != "" {
		// legacy 校验由 NormalizeExtractRules 覆盖
	}
	extractRules, err := subscription_rule.NormalizeExtractRules(req.ExtractRules, req.ExtractRegex, req.ExtractField)
	if err != nil {
		return nil, fmt.Errorf("提取规则错误: %w", err)
	}
	extractRegexStore, extractFieldStore, err := subscription_rule.EncodeExtractRules(extractRules)
	if err != nil {
		return nil, fmt.Errorf("提取规则序列化失败: %w", err)
	}

	if err := s.validateGroupTopicUnique(req.SourceID, req.Topic, req.GroupName, ""); err != nil {
		return nil, err
	}

	// 创建订阅
	id, err := models.AddSubscription(
		req.Name,
		req.SourceID,
		req.Topic,
		req.Tag,
		req.GroupName,
		req.ValidateRegex,
		extractRegexStore,
		extractFieldStore,
		req.TemplateID,
		normalizeTemplateContentTypeWithAlias(req.TemplateContentType, req.ConsumeMode),
		"", // createdBy
	)
	if err != nil {
		return nil, fmt.Errorf("创建订阅失败: %w", err)
	}

	// 获取创建的订阅
	subscription, err := models.GetSubscriptionByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取订阅信息失败: %w", err)
	}

	return &subscription, nil
}

// Edit 编辑订阅
func (s *SubscriptionService) Edit(id string, req EditSubscriptionRequest) error {
	// 校验正则表达式
	if req.ValidateRegex != "" {
		if err := ValidateRegexPattern(req.ValidateRegex); err != nil {
			return fmt.Errorf("验证正则表达式错误: %w", err)
		}
	}
	if req.ExtractRegex != "" {
		// legacy 校验由 NormalizeExtractRules 覆盖
	}
	extractRules, err := subscription_rule.NormalizeExtractRules(req.ExtractRules, req.ExtractRegex, req.ExtractField)
	if err != nil {
		return fmt.Errorf("提取规则错误: %w", err)
	}
	extractRegexStore, extractFieldStore, err := subscription_rule.EncodeExtractRules(extractRules)
	if err != nil {
		return fmt.Errorf("提取规则序列化失败: %w", err)
	}

	// 检查订阅是否存在
	sub, err := models.GetSubscriptionByID(id)
	if err != nil {
		return fmt.Errorf("订阅不存在: %w", err)
	}

	// 如果订阅正在运行，不允许编辑
	if sub.Status == "running" {
		return fmt.Errorf("订阅正在运行中，请先停止后再编辑")
	}

	if err := s.validateGroupTopicUnique(req.SourceID, req.Topic, req.GroupName, id); err != nil {
		return err
	}

	// 更新订阅
	data := map[string]interface{}{
		"name":           req.Name,
		"source_id":      req.SourceID,
		"topic":          req.Topic,
		"tag":            req.Tag,
		"group_name":     req.GroupName,
		"validate_regex": req.ValidateRegex,
		"extract_regex":  extractRegexStore,
		"extract_field":  extractFieldStore,
		"template_id":    req.TemplateID,
		"consume_mode":   normalizeTemplateContentTypeWithAlias(req.TemplateContentType, req.ConsumeMode),
	}

	return models.UpdateSubscription(id, data)
}

// Delete 删除订阅
func (s *SubscriptionService) Delete(id string) error {
	sub := models.Subscription{}
	if err := sub.GetByUUID(id); err != nil {
		return err
	}

	// 运行中的订阅不允许删除
	if sub.Status == "running" {
		return fmt.Errorf("订阅正在运行中，请先停止后再删除")
	}

	return models.DeleteSubscription(id)
}

// GetByID 根据 ID 获取订阅
func (s *SubscriptionService) GetByID(id string) (*models.Subscription, error) {
	subscription, err := models.GetSubscriptionByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %w", err)
	}
	return &subscription, nil
}

func (s *SubscriptionService) TestRegex(req RegexTestRequest) (*RegexTestResult, error) {
	result := &RegexTestResult{
		ValidateMatched: true,
	}

	if req.ValidateRegex != "" {
		matched, err := MatchTextWithPattern(req.Message, req.ValidateRegex)
		if err != nil {
			return nil, fmt.Errorf("验证正则表达式错误: %w", err)
		}
		result.ValidateMatched = matched
	}

	extractRules, err := subscription_rule.NormalizeExtractRules(req.ExtractRules, req.ExtractRegex, req.ExtractField)
	if err != nil {
		return nil, fmt.Errorf("提取规则错误: %w", err)
	}
	if len(extractRules) > 0 {
		values, err := subscription_rule.BuildExtractMap(req.Message, extractRules)
		if err != nil {
			return nil, fmt.Errorf("提取正则表达式错误: %w", err)
		}
		result.ExtractedValues = values
	}

	return result, nil
}

func (s *SubscriptionService) validateGroupTopicUnique(sourceID, topic, groupName, excludeID string) error {
	allInSource, err := models.GetSubscriptions(0, 0, "", sourceID, "", "")
	if err != nil {
		return fmt.Errorf("校验订阅唯一性失败: %w", err)
	}

	targetTopic := strings.TrimSpace(topic)
	targetGroup := strings.TrimSpace(groupName)
	for _, sub := range allInSource {
		if excludeID != "" && sub.ID == excludeID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(sub.Topic), targetTopic) &&
			strings.EqualFold(strings.TrimSpace(sub.GroupName), targetGroup) {
			return fmt.Errorf("同一数据源下 Consumer Group [%s] 已订阅 Topic [%s]，请更换 Group 或 Topic", groupName, topic)
		}
	}
	return nil
}

// GetAll 获取订阅列表
func (s *SubscriptionService) GetAll(name, status, sourceID string, page, pageSize int) ([]SubscriptionListItem, int64, error) {
	subscriptions, err := models.GetSubscriptions(page, pageSize, name, sourceID, status, "")
	if err != nil {
		return nil, 0, fmt.Errorf("获取订阅列表失败: %w", err)
	}

	total, err := models.GetSubscriptionsTotal(name, sourceID, status, "")
	if err != nil {
		return nil, 0, fmt.Errorf("获取订阅总数失败: %w", err)
	}

	sourceNameMap := map[string]string{}
	templateNameMap := map[string]string{}
	result := make([]SubscriptionListItem, 0, len(subscriptions))
	for _, sub := range subscriptions {
		if _, ok := sourceNameMap[sub.SourceID]; !ok && sub.SourceID != "" {
			source, sourceErr := models.GetMQSourceByID(sub.SourceID)
			if sourceErr == nil {
				sourceNameMap[sub.SourceID] = source.Name
			}
		}
		if _, ok := templateNameMap[sub.TemplateID]; !ok && sub.TemplateID != "" {
			template, tplErr := models.GetTemplateByID(sub.TemplateID)
			if tplErr == nil && template != nil {
				templateNameMap[sub.TemplateID] = template.Name
			}
		}

		lastConsumeTime := ""
		if t := sub.LastConsumeTime.String(); t != "" && !strings.HasPrefix(t, "0001-01-01") {
			lastConsumeTime = t
		}

		result = append(result, SubscriptionListItem{
			ID:                  sub.ID,
			Name:                sub.Name,
			SourceID:            sub.SourceID,
			SourceName:          sourceNameMap[sub.SourceID],
			Topic:               sub.Topic,
			Tag:                 sub.Tag,
			GroupName:           sub.GroupName,
			ValidateRegex:       sub.ValidateRegex,
			ExtractRegex:        sub.ExtractRegex,
			ExtractField:        sub.ExtractField,
			ExtractRules:        subscription_rule.ParseStoredExtractRules(sub.ExtractRegex, sub.ExtractField),
			TemplateID:          sub.TemplateID,
			TemplateName:        templateNameMap[sub.TemplateID],
			TemplateContentType: normalizeTemplateContentType(sub.ConsumeMode),
			Status:              sub.Status,
			TotalConsumed:       sub.TotalConsumed,
			TotalSent:           sub.TotalSent,
			TotalFailed:         sub.TotalFailed,
			LastConsumeTime:     lastConsumeTime,
			CreatedOn:           sub.CreatedAt.String(),
		})
	}

	return result, total, nil
}

// Count 获取订阅总数
func (s *SubscriptionService) Count(name, status, sourceID string) (int64, error) {
	return models.GetSubscriptionsTotal(name, sourceID, status, "")
}

// Start 启动订阅
func (s *SubscriptionService) Start(id string) error {
	sub := models.Subscription{}
	if err := sub.GetByUUID(id); err != nil {
		return err
	}

	if sub.Status == "running" {
		return fmt.Errorf("订阅已在运行中")
	}

	// 更新状态
	if err := sub.UpdateStatus("running"); err != nil {
		return fmt.Errorf("更新订阅状态失败: %w", err)
	}

	// 启动消费者
	if GlobalConsumerManager != nil {
		go func() {
			if err := GlobalConsumerManager.StartSubscription(id); err != nil {
				// 启动失败，回滚状态
				logrus.Errorf("启动订阅 %s 失败: %v，回滚状态", id, err)
				sub.UpdateStatus("stopped")
			} else {
				logrus.Infof("订阅 %s 消费者启动成功", id)
			}
		}()
	}

	return nil
}

// Stop 停止订阅
func (s *SubscriptionService) Stop(id string) error {
	sub := models.Subscription{}
	if err := sub.GetByUUID(id); err != nil {
		return err
	}

	if sub.Status == "stopped" {
		return nil
	}

	// 停止消费者
	if GlobalConsumerManager != nil {
		if err := GlobalConsumerManager.StopSubscription(id); err != nil {
			// 订阅状态是 running 但进程内没有消费者实例时，允许幂等停止
			if !strings.Contains(err.Error(), "未在运行") {
				return err
			}
			logrus.Warnf("订阅 %s 停止时未找到消费者实例，按幂等停止处理", id)
		}
	}

	// 更新状态
	return sub.UpdateStatus("stopped")
}

// ValidateRegexPattern 验证正则表达式
func ValidateRegexPattern(pattern string) error {
	return subscription_rule.ValidatePattern(pattern)
}

func ValidateExtractPattern(pattern string) error {
	return subscription_rule.ValidateExtractPattern(pattern)
}

func MatchTextWithPattern(text, pattern string) (bool, error) {
	return subscription_rule.MatchText(text, pattern)
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

// FormatTagList 格式化 Tag 列表
func FormatTagList(tag string) string {
	if tag == "" {
		return "*"
	}
	// 多个 Tag 用 || 分隔
	tag = strings.TrimSpace(tag)
	tag = strings.ReplaceAll(tag, ",", "||")
	tag = strings.ReplaceAll(tag, "，", "||")
	return tag
}
