package subscription_rule

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ExtractRule struct {
	Field string `json:"field"`
	Regex string `json:"regex"`
}

// ParseStoredExtractRules 解析存储层规则（兼容旧字段）
func ParseStoredExtractRules(extractRegex, extractField string) []ExtractRule {
	regexRaw := strings.TrimSpace(extractRegex)
	fieldRaw := strings.TrimSpace(extractField)
	if regexRaw == "" && fieldRaw == "" {
		return nil
	}

	// 新格式：extract_regex 存 JSON 数组
	if strings.HasPrefix(regexRaw, "[") {
		var rules []ExtractRule
		if err := json.Unmarshal([]byte(regexRaw), &rules); err == nil {
			out := make([]ExtractRule, 0, len(rules))
			for _, r := range rules {
				field := strings.TrimSpace(r.Field)
				regex := strings.TrimSpace(r.Regex)
				if field == "" {
					continue
				}
				out = append(out, ExtractRule{
					Field: field,
					Regex: regex,
				})
			}
			return out
		}
	}

	// 旧格式：单组规则
	if fieldRaw == "" {
		return nil
	}
	return []ExtractRule{
		{
			Field: fieldRaw,
			Regex: regexRaw,
		},
	}
}

// NormalizeExtractRules 规范化输入规则并做语法校验
func NormalizeExtractRules(rules []ExtractRule, legacyRegex, legacyField string) ([]ExtractRule, error) {
	base := rules
	if len(base) == 0 {
		base = ParseStoredExtractRules(legacyRegex, legacyField)
	}
	out := make([]ExtractRule, 0, len(base))
	seen := map[string]bool{}
	for idx, r := range base {
		field := strings.TrimSpace(r.Field)
		regex := strings.TrimSpace(r.Regex)
		if field == "" && regex == "" {
			continue
		}
		if field == "" {
			return nil, fmt.Errorf("第%d组提取规则缺少字段名", idx+1)
		}
		if seen[field] {
			return nil, fmt.Errorf("提取字段名重复: %s", field)
		}
		seen[field] = true
		if regex != "" {
			if err := ValidateExtractPattern(regex); err != nil {
				return nil, fmt.Errorf("第%d组提取规则语法错误: %w", idx+1, err)
			}
		}
		out = append(out, ExtractRule{
			Field: field,
			Regex: regex,
		})
	}
	return out, nil
}

func EncodeExtractRules(rules []ExtractRule) (extractRegex string, extractField string, err error) {
	if len(rules) == 0 {
		return "", "", nil
	}
	b, err := json.Marshal(rules)
	if err != nil {
		return "", "", err
	}
	// extract_field 仅作兼容，保留首字段
	return string(b), rules[0].Field, nil
}

func BuildExtractMap(rawMessage string, rules []ExtractRule) (map[string]string, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	result := make(map[string]string, len(rules))
	for _, rule := range rules {
		value, err := ExtractValue(rawMessage, rule.Regex)
		if err != nil {
			return nil, fmt.Errorf("提取字段[%s]失败: %w", rule.Field, err)
		}
		result[rule.Field] = value
	}
	return result, nil
}

