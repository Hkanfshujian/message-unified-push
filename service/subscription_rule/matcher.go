package subscription_rule

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidatePattern 校验规则语法（仅 DSL）
func ValidatePattern(pattern string) error {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return nil
	}
	if !isDSLPattern(p) {
		return fmt.Errorf("仅支持 DSL 规则，请使用 dsl: 前缀")
	}
	_, err := evaluateDSLValidate("", stripDSLPrefix(p))
	return err
}

// ValidateExtractPattern 校验提取规则语法（仅 DSL）
func ValidateExtractPattern(pattern string) error {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return nil
	}
	if !isDSLPattern(p) {
		return fmt.Errorf("仅支持 DSL 规则，请使用 dsl: 前缀")
	}
	_, err := evaluateDSLExtract("", stripDSLPrefix(p))
	return err
}

// MatchText 验证规则匹配
func MatchText(rawMessage, pattern string) (bool, error) {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return true, nil
	}
	if !isDSLPattern(p) {
		return false, fmt.Errorf("仅支持 DSL 规则，请使用 dsl: 前缀")
	}
	return evaluateDSLValidate(rawMessage, stripDSLPrefix(p))
}

// ExtractValue 提取规则匹配并返回单值
func ExtractValue(rawMessage, pattern string) (string, error) {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return rawMessage, nil
	}
	if !isDSLPattern(p) {
		return "", fmt.Errorf("仅支持 DSL 规则，请使用 dsl: 前缀")
	}
	return evaluateDSLExtract(rawMessage, stripDSLPrefix(p))
}

func isDSLPattern(pattern string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(pattern)), "dsl:")
}

func stripDSLPrefix(pattern string) string {
	raw := strings.TrimSpace(pattern)
	if len(raw) >= 4 && strings.EqualFold(raw[:4], "dsl:") {
		return strings.TrimSpace(raw[4:])
	}
	return raw
}

func evaluateDSLValidate(rawMessage, expr string) (bool, error) {
	msg := newDSLMessage(rawMessage)
	return evalBoolExpr(msg, strings.TrimSpace(expr))
}

func evaluateDSLExtract(rawMessage, expr string) (string, error) {
	msg := newDSLMessage(rawMessage)
	return evalValueExpr(msg, strings.TrimSpace(expr))
}

type dslMessage struct {
	raw      string
	jsonData map[string]interface{}
}

func newDSLMessage(raw string) dslMessage {
	msg := dslMessage{raw: raw}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err == nil {
		msg.jsonData = obj
	}
	return msg
}

func evalBoolExpr(msg dslMessage, expr string) (bool, error) {
	e := strings.TrimSpace(expr)
	if e == "" {
		return false, nil
	}
	e = trimWrappedParens(e)
	if parts, ok := splitTopLevel(e, "||"); ok {
		for _, part := range parts {
			b, err := evalBoolExpr(msg, part)
			if err != nil {
				return false, err
			}
			if b {
				return true, nil
			}
		}
		return false, nil
	}
	if parts, ok := splitTopLevel(e, "&&"); ok {
		for _, part := range parts {
			b, err := evalBoolExpr(msg, part)
			if err != nil {
				return false, err
			}
			if !b {
				return false, nil
			}
		}
		return true, nil
	}
	if strings.HasPrefix(strings.TrimSpace(e), "!") {
		b, err := evalBoolExpr(msg, strings.TrimSpace(e[1:]))
		if err != nil {
			return false, err
		}
		return !b, nil
	}
	return evalPredicate(msg, e)
}

func evalPredicate(msg dslMessage, expr string) (bool, error) {
	name, args, ok, err := parseCall(expr)
	if err != nil {
		return false, err
	}
	if !ok {
		// 兼容 true/false 字面量
		v := strings.ToLower(strings.TrimSpace(expr))
		if v == "true" {
			return true, nil
		}
		if v == "false" {
			return false, nil
		}
		return false, fmt.Errorf("DSL 布尔表达式非法: %s", expr)
	}
	switch strings.ToLower(name) {
	case "contains":
		if len(args) == 1 {
			sub, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			return strings.Contains(msg.raw, sub), nil
		}
		if len(args) == 2 {
			a, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			b, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			return strings.Contains(a, b), nil
		}
	case "equals", "eq":
		if len(args) == 2 {
			a, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			b, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			return a == b, nil
		}
	case "startswith":
		if len(args) == 2 {
			a, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			b, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			return strings.HasPrefix(a, b), nil
		}
	case "endswith":
		if len(args) == 2 {
			a, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			b, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			return strings.HasSuffix(a, b), nil
		}
	case "exists":
		if len(args) == 1 {
			v, ok := resolvePath(msg, strings.TrimSpace(args[0]))
			return ok && strings.TrimSpace(toString(v)) != "", nil
		}
	case "regex":
		if len(args) == 1 {
			pat, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			re, err := regexp.Compile(pat)
			if err != nil {
				return false, err
			}
			return re.MatchString(msg.raw), nil
		}
		if len(args) == 2 {
			target, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			pat, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			re, err := regexp.Compile(pat)
			if err != nil {
				return false, err
			}
			return re.MatchString(target), nil
		}
	case "in":
		if len(args) >= 2 {
			left, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			for _, rightArg := range args[1:] {
				right, err := evalValueExpr(msg, rightArg)
				if err != nil {
					return false, err
				}
				if left == right {
					return true, nil
				}
			}
			return false, nil
		}
	case "empty":
		if len(args) == 1 {
			v, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			return strings.TrimSpace(v) == "", nil
		}
	case "notempty":
		if len(args) == 1 {
			v, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			return strings.TrimSpace(v) != "", nil
		}
	case "gt", "gte", "lt", "lte":
		if len(args) == 2 {
			left, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			right, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			lv, rv, err := parseTwoFloat(left, right)
			if err != nil {
				return false, err
			}
			switch strings.ToLower(name) {
			case "gt":
				return lv > rv, nil
			case "gte":
				return lv >= rv, nil
			case "lt":
				return lv < rv, nil
			default:
				return lv <= rv, nil
			}
		}
	case "between":
		if len(args) == 3 {
			val, err := evalValueExpr(msg, args[0])
			if err != nil {
				return false, err
			}
			minV, err := evalValueExpr(msg, args[1])
			if err != nil {
				return false, err
			}
			maxV, err := evalValueExpr(msg, args[2])
			if err != nil {
				return false, err
			}
			v, minF, err := parseTwoFloat(val, minV)
			if err != nil {
				return false, err
			}
			maxF, err := parseFloat(maxV)
			if err != nil {
				return false, err
			}
			return v >= minF && v <= maxF, nil
		}
	}
	return false, fmt.Errorf("DSL 谓词不支持或参数错误: %s", expr)
}

func evalValueExpr(msg dslMessage, expr string) (string, error) {
	e := strings.TrimSpace(expr)
	if e == "" {
		return "", nil
	}
	if quoted, ok := tryUnquote(e); ok {
		return quoted, nil
	}
	if strings.EqualFold(e, "raw") || strings.EqualFold(e, "$") {
		return msg.raw, nil
	}
	if looksLikePath(e) {
		v, _ := resolvePath(msg, e)
		return toString(v), nil
	}

	name, args, ok, err := parseCall(e)
	if err != nil {
		return "", err
	}
	if !ok {
		// 非函数时按字面量处理
		return e, nil
	}

	switch strings.ToLower(name) {
	case "pick", "json":
		if len(args) != 1 {
			return "", fmt.Errorf("%s 需要 1 个参数", name)
		}
		v, _ := resolvePath(msg, strings.TrimSpace(args[0]))
		return toString(v), nil
	case "lower":
		if len(args) != 1 {
			return "", fmt.Errorf("lower 需要 1 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		return strings.ToLower(v), nil
	case "upper":
		if len(args) != 1 {
			return "", fmt.Errorf("upper 需要 1 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		return strings.ToUpper(v), nil
	case "trim":
		if len(args) != 1 {
			return "", fmt.Errorf("trim 需要 1 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(v), nil
	case "replace":
		if len(args) != 3 {
			return "", fmt.Errorf("replace 需要 3 个参数")
		}
		src, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		oldVal, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		newVal, err := evalValueExpr(msg, args[2])
		if err != nil {
			return "", err
		}
		return strings.ReplaceAll(src, oldVal, newVal), nil
	case "concat":
		var parts []string
		for _, arg := range args {
			v, err := evalValueExpr(msg, arg)
			if err != nil {
				return "", err
			}
			parts = append(parts, v)
		}
		return strings.Join(parts, ""), nil
	case "split":
		if len(args) < 2 || len(args) > 3 {
			return "", fmt.Errorf("split 需要 2 或 3 个参数")
		}
		src, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		sep, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		parts := strings.Split(src, sep)
		if len(args) == 2 {
			return strings.Join(parts, ","), nil
		}
		idxStr, err := evalValueExpr(msg, args[2])
		if err != nil {
			return "", err
		}
		idx, err := strconv.Atoi(strings.TrimSpace(idxStr))
		if err != nil {
			return "", fmt.Errorf("split 索引非法: %v", err)
		}
		if idx < 0 || idx >= len(parts) {
			return "", nil
		}
		return parts[idx], nil
	case "regex":
		// regex(target, pattern [, groupIndex])
		if len(args) < 2 || len(args) > 3 {
			return "", fmt.Errorf("regex 提取需要 2 或 3 个参数")
		}
		target, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		pat, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return "", err
		}
		matches := re.FindStringSubmatch(target)
		if matches == nil {
			return "", nil
		}
		groupIdx := 1
		if len(args) == 3 {
			idxStr, err := evalValueExpr(msg, args[2])
			if err != nil {
				return "", err
			}
			v, err := strconv.Atoi(strings.TrimSpace(idxStr))
			if err != nil {
				return "", fmt.Errorf("regex group 索引非法: %v", err)
			}
			groupIdx = v
		}
		if groupIdx < 0 || groupIdx >= len(matches) {
			return "", nil
		}
		return matches[groupIdx], nil
	case "regexall":
		// regexAll(target, pattern [, groupIndex] [, sep])
		if len(args) < 2 || len(args) > 4 {
			return "", fmt.Errorf("regexAll 提取需要 2 到 4 个参数")
		}
		target, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		pat, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return "", err
		}
		groupIdx := 1
		sep := "|"
		if len(args) >= 3 {
			third, err := evalValueExpr(msg, args[2])
			if err != nil {
				return "", err
			}
			third = strings.TrimSpace(third)
			if third != "" {
				// 兼容写法：第三个参数既可传 group，也可直接传 sep
				if v, convErr := strconv.Atoi(third); convErr == nil {
					groupIdx = v
				} else {
					sep = third
				}
			}
		}
		if len(args) == 4 {
			sep, err = evalValueExpr(msg, args[3])
			if err != nil {
				return "", err
			}
			if sep == "" {
				sep = "|"
			}
		}
		all := re.FindAllStringSubmatch(target, -1)
		if len(all) == 0 {
			return "", nil
		}
		values := make([]string, 0, len(all))
		for _, m := range all {
			if groupIdx < 0 || groupIdx >= len(m) {
				continue
			}
			values = append(values, m[groupIdx])
		}
		if len(values) == 0 {
			return "", nil
		}
		return strings.Join(values, sep), nil
	case "findidsbyvalue":
		// findIdsByValue(path, targetValue [, idKey] [, sep])
		if len(args) < 2 || len(args) > 4 {
			return "", fmt.Errorf("findIdsByValue 需要 2 到 4 个参数")
		}
		path := strings.TrimSpace(args[0])
		targetValue, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		idKey := "id"
		if len(args) >= 3 {
			idKey, err = evalValueExpr(msg, args[2])
			if err != nil {
				return "", err
			}
			idKey = strings.TrimSpace(idKey)
			if idKey == "" {
				idKey = "id"
			}
		}
		sep := "|"
		if len(args) == 4 {
			sep, err = evalValueExpr(msg, args[3])
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(sep) == "" {
				sep = "|"
			}
		}

		var root interface{}
		if path == "" || path == "$" || strings.EqualFold(path, "raw") {
			root = msg.jsonData
		} else {
			var ok bool
			root, ok = resolvePath(msg, path)
			if !ok {
				return "", nil
			}
		}
		if root == nil {
			return "", nil
		}
		values := findIDsByValue(root, targetValue, idKey)
		if len(values) == 0 {
			return "", nil
		}
		return strings.Join(values, sep), nil
	case "arraylen":
		// arrayLen(path)
		if len(args) != 1 {
			return "", fmt.Errorf("arrayLen 需要 1 个参数")
		}
		node, ok := resolveNodeByPath(msg, strings.TrimSpace(args[0]))
		if !ok || node == nil {
			return "0", nil
		}
		if arr, ok := node.([]interface{}); ok {
			return strconv.Itoa(len(arr)), nil
		}
		return "0", nil
	case "valuesbykey":
		// valuesByKey(path, key [, sep])
		if len(args) < 2 || len(args) > 3 {
			return "", fmt.Errorf("valuesByKey 需要 2 或 3 个参数")
		}
		node, ok := resolveNodeByPath(msg, strings.TrimSpace(args[0]))
		if !ok || node == nil {
			return "", nil
		}
		key, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return "", nil
		}
		sep := "|"
		if len(args) == 3 {
			sep, err = evalValueExpr(msg, args[2])
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(sep) == "" {
				sep = "|"
			}
		}
		values := collectValuesByKey(node, key)
		if len(values) == 0 {
			return "", nil
		}
		return strings.Join(values, sep), nil
	case "findbyfield":
		// findByField(path, matchKey, matchValue, returnKey [, sep])
		if len(args) < 4 || len(args) > 5 {
			return "", fmt.Errorf("findByField 需要 4 或 5 个参数")
		}
		node, ok := resolveNodeByPath(msg, strings.TrimSpace(args[0]))
		if !ok || node == nil {
			return "", nil
		}
		matchKey, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		matchValue, err := evalValueExpr(msg, args[2])
		if err != nil {
			return "", err
		}
		returnKey, err := evalValueExpr(msg, args[3])
		if err != nil {
			return "", err
		}
		matchKey = strings.TrimSpace(matchKey)
		returnKey = strings.TrimSpace(returnKey)
		if matchKey == "" || returnKey == "" {
			return "", nil
		}
		sep := "|"
		if len(args) == 5 {
			sep, err = evalValueExpr(msg, args[4])
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(sep) == "" {
				sep = "|"
			}
		}
		values := collectValuesByField(node, matchKey, matchValue, returnKey)
		if len(values) == 0 {
			return "", nil
		}
		return strings.Join(values, sep), nil
	case "findbyfieldraw":
		// findByFieldRaw(path, matchKey, matchValue [, sep])
		if len(args) < 3 || len(args) > 4 {
			return "", fmt.Errorf("findByFieldRaw 需要 3 或 4 个参数")
		}
		node, ok := resolveNodeByPath(msg, strings.TrimSpace(args[0]))
		if !ok || node == nil {
			return "", nil
		}
		matchKey, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		matchValue, err := evalValueExpr(msg, args[2])
		if err != nil {
			return "", err
		}
		matchKey = strings.TrimSpace(matchKey)
		if matchKey == "" {
			return "", nil
		}
		sep := "|"
		if len(args) == 4 {
			sep, err = evalValueExpr(msg, args[3])
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(sep) == "" {
				sep = "|"
			}
		}
		values := collectObjectsByField(node, matchKey, matchValue)
		if len(values) == 0 {
			return "", nil
		}
		return strings.Join(values, sep), nil
	case "default":
		if len(args) != 2 {
			return "", fmt.Errorf("default 需要 2 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(v) != "" {
			return v, nil
		}
		return evalValueExpr(msg, args[1])
	case "coalesce":
		for _, arg := range args {
			v, err := evalValueExpr(msg, arg)
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(v) != "" {
				return v, nil
			}
		}
		return "", nil
	case "len":
		if len(args) != 1 {
			return "", fmt.Errorf("len 需要 1 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		return strconv.Itoa(len(v)), nil
	case "substr":
		if len(args) < 2 || len(args) > 3 {
			return "", fmt.Errorf("substr 需要 2 或 3 个参数")
		}
		src, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		startStr, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		start, err := strconv.Atoi(strings.TrimSpace(startStr))
		if err != nil {
			return "", fmt.Errorf("substr 起始索引非法: %v", err)
		}
		if start < 0 {
			start = 0
		}
		if start >= len(src) {
			return "", nil
		}
		if len(args) == 2 {
			return src[start:], nil
		}
		lengthStr, err := evalValueExpr(msg, args[2])
		if err != nil {
			return "", err
		}
		l, err := strconv.Atoi(strings.TrimSpace(lengthStr))
		if err != nil {
			return "", fmt.Errorf("substr 长度非法: %v", err)
		}
		if l <= 0 {
			return "", nil
		}
		end := start + l
		if end > len(src) {
			end = len(src)
		}
		return src[start:end], nil
	case "toint":
		if len(args) != 1 {
			return "", fmt.Errorf("toInt 需要 1 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		f, err := parseFloat(v)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(int64(f), 10), nil
	case "tofloat":
		if len(args) != 1 {
			return "", fmt.Errorf("toFloat 需要 1 个参数")
		}
		v, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		f, err := parseFloat(v)
		if err != nil {
			return "", err
		}
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	case "add", "sub", "mul", "div":
		if len(args) != 2 {
			return "", fmt.Errorf("%s 需要 2 个参数", name)
		}
		aStr, err := evalValueExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		bStr, err := evalValueExpr(msg, args[1])
		if err != nil {
			return "", err
		}
		a, b, err := parseTwoFloat(aStr, bStr)
		if err != nil {
			return "", err
		}
		var result float64
		switch strings.ToLower(name) {
		case "add":
			result = a + b
		case "sub":
			result = a - b
		case "mul":
			result = a * b
		default:
			if b == 0 {
				return "", fmt.Errorf("div 除数不能为 0")
			}
			result = a / b
		}
		return strconv.FormatFloat(result, 'f', -1, 64), nil
	case "if":
		if len(args) != 3 {
			return "", fmt.Errorf("if 需要 3 个参数")
		}
		cond, err := evalBoolExpr(msg, args[0])
		if err != nil {
			return "", err
		}
		if cond {
			return evalValueExpr(msg, args[1])
		}
		return evalValueExpr(msg, args[2])
	}

	return "", fmt.Errorf("DSL 函数不支持: %s", name)
}

func findIDsByValue(root interface{}, targetValue, idKey string) []string {
	return collectValuesByField(root, "value", targetValue, idKey)
}

func resolveNodeByPath(msg dslMessage, path string) (interface{}, bool) {
	p := strings.TrimSpace(path)
	if p == "" || p == "$" || strings.EqualFold(p, "raw") {
		if msg.jsonData != nil {
			return msg.jsonData, true
		}
		return nil, false
	}
	return resolvePath(msg, p)
}

func collectValuesByKey(root interface{}, key string) []string {
	results := make([]string, 0)
	seen := map[string]bool{}
	var walk func(node interface{})
	walk = func(node interface{}) {
		switch vv := node.(type) {
		case map[string]interface{}:
			if val, ok := vv[key]; ok {
				s := strings.TrimSpace(toString(val))
				if s != "" && !seen[s] {
					seen[s] = true
					results = append(results, s)
				}
			}
			for _, child := range vv {
				walk(child)
			}
		case []interface{}:
			for _, child := range vv {
				walk(child)
			}
		}
	}
	walk(root)
	return results
}

func collectValuesByField(root interface{}, matchKey, matchValue, returnKey string) []string {
	results := make([]string, 0)
	seen := map[string]bool{}
	var walk func(node interface{})
	walk = func(node interface{}) {
		switch vv := node.(type) {
		case map[string]interface{}:
			if val, ok := vv[matchKey]; ok && toString(val) == matchValue {
				if returnVal, ok := vv[returnKey]; ok {
					s := strings.TrimSpace(toString(returnVal))
					if s != "" && !seen[s] {
						seen[s] = true
						results = append(results, s)
					}
				}
			}
			for _, child := range vv {
				walk(child)
			}
		case []interface{}:
			for _, child := range vv {
				walk(child)
			}
		}
	}
	walk(root)
	return results
}

func collectObjectsByField(root interface{}, matchKey, matchValue string) []string {
	results := make([]string, 0)
	seen := map[string]bool{}
	var walk func(node interface{})
	walk = func(node interface{}) {
		switch vv := node.(type) {
		case map[string]interface{}:
			if val, ok := vv[matchKey]; ok && toString(val) == matchValue {
				if b, err := json.Marshal(vv); err == nil {
					s := string(b)
					if s != "" && !seen[s] {
						seen[s] = true
						results = append(results, s)
					}
				}
			}
			for _, child := range vv {
				walk(child)
			}
		case []interface{}:
			for _, child := range vv {
				walk(child)
			}
		}
	}
	walk(root)
	return results
}

func parseFloat(v string) (float64, error) {
	s := strings.TrimSpace(v)
	if s == "" {
		return 0, fmt.Errorf("数值参数为空")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("数值参数非法: %s", v)
	}
	return f, nil
}

func parseTwoFloat(a, b string) (float64, float64, error) {
	av, err := parseFloat(a)
	if err != nil {
		return 0, 0, err
	}
	bv, err := parseFloat(b)
	if err != nil {
		return 0, 0, err
	}
	return av, bv, nil
}

func parseCall(expr string) (name string, args []string, ok bool, err error) {
	e := strings.TrimSpace(expr)
	if e == "" {
		return "", nil, false, nil
	}
	i := strings.IndexByte(e, '(')
	if i <= 0 || !strings.HasSuffix(e, ")") {
		return "", nil, false, nil
	}
	name = strings.TrimSpace(e[:i])
	if name == "" {
		return "", nil, false, fmt.Errorf("DSL 函数名为空")
	}
	inside := strings.TrimSpace(e[i+1 : len(e)-1])
	args, err = splitArgs(inside)
	if err != nil {
		return "", nil, false, err
	}
	return name, args, true, nil
}

func splitArgs(s string) ([]string, error) {
	if strings.TrimSpace(s) == "" {
		return []string{}, nil
	}
	var res []string
	start := 0
	depth := 0
	inQuote := byte(0)
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			inQuote = ch
			continue
		}
		if ch == '(' {
			depth++
			continue
		}
		if ch == ')' {
			if depth == 0 {
				return nil, fmt.Errorf("DSL 参数括号不匹配")
			}
			depth--
			continue
		}
		if ch == ',' && depth == 0 {
			res = append(res, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	if inQuote != 0 || depth != 0 {
		return nil, fmt.Errorf("DSL 参数语法不完整")
	}
	res = append(res, strings.TrimSpace(s[start:]))
	return res, nil
}

func splitTopLevel(s, op string) ([]string, bool) {
	var parts []string
	start := 0
	depth := 0
	inQuote := byte(0)
	escaped := false
	found := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		if ch == '(' {
			depth++
			continue
		}
		if ch == ')' {
			if depth > 0 {
				depth--
			}
			continue
		}
		if depth == 0 && i+len(op) <= len(s) && s[i:i+len(op)] == op {
			part := strings.TrimSpace(s[start:i])
			parts = append(parts, part)
			start = i + len(op)
			i += len(op) - 1
			found = true
		}
	}
	if !found {
		return nil, false
	}
	parts = append(parts, strings.TrimSpace(s[start:]))
	return parts, true
}

func trimWrappedParens(s string) string {
	cur := strings.TrimSpace(s)
	for {
		if len(cur) < 2 || cur[0] != '(' || cur[len(cur)-1] != ')' {
			return cur
		}
		if !isOuterParens(cur) {
			return cur
		}
		cur = strings.TrimSpace(cur[1 : len(cur)-1])
	}
}

func isOuterParens(s string) bool {
	depth := 0
	inQuote := byte(0)
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inQuote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 && i != len(s)-1 {
				return false
			}
		}
	}
	return depth == 0
}

func tryUnquote(s string) (string, bool) {
	if len(s) < 2 {
		return "", false
	}
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		v, err := strconv.Unquote(`"` + strings.ReplaceAll(s[1:len(s)-1], `"`, `\"`) + `"`)
		if err != nil {
			// 兜底：不做转义处理
			return s[1 : len(s)-1], true
		}
		return v, true
	}
	return "", false
}

func looksLikePath(expr string) bool {
	e := strings.TrimSpace(expr)
	if strings.HasPrefix(e, "$.") {
		return true
	}
	if e == "$" || strings.EqualFold(e, "raw") {
		return true
	}
	// 兼容直接 key 形式：department / name
	for _, r := range e {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.') {
			return false
		}
	}
	return e != ""
}

func resolvePath(msg dslMessage, path string) (interface{}, bool) {
	p := strings.TrimSpace(path)
	if p == "$" || strings.EqualFold(p, "raw") {
		return msg.raw, true
	}
	p = strings.TrimPrefix(p, "$.")
	if p == "" {
		return nil, false
	}
	if msg.jsonData == nil {
		return nil, false
	}
	cur := interface{}(msg.jsonData)
	for _, part := range strings.Split(p, ".") {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		v, ok := m[part]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func toString(v interface{}) string {
	switch vv := v.(type) {
	case nil:
		return ""
	case string:
		return vv
	case float64:
		if vv == float64(int64(vv)) {
			return strconv.FormatInt(int64(vv), 10)
		}
		return strconv.FormatFloat(vv, 'f', -1, 64)
	case bool:
		if vv {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(vv)
		return string(b)
	}
}
