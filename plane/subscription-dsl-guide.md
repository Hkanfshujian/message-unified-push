# 订阅规则 DSL 使用说明

本文档用于说明“订阅消息”中的 `验证正则` 与 `提取正则` 在 `dsl:` 模式下的语法。

## 一、入口规则

- `验证正则`：仅支持 `dsl:` 业务 DSL
- `提取正则`：仅支持 `dsl:` 提取 DSL

## 二、验证 DSL（`dsl:`）

### 2.1 布尔运算

- `&&` 与
- `||` 或
- `!` 非
- 支持括号

### 2.2 谓词函数

- `contains(a, b)`：包含判断
- `equals(a, b)` / `eq(a, b)`：相等判断
- `startsWith(a, b)`：前缀判断
- `endsWith(a, b)`：后缀判断
- `exists(path)`：字段存在且非空
- `regex(a, pattern)`：字符串正则匹配
- `regex(pattern)`：对整条原文匹配
- `in(a, b, c, ...)`：集合包含
- `gt(a, b)` / `gte(a, b)`：数值比较（大于/大于等于）
- `lt(a, b)` / `lte(a, b)`：数值比较（小于/小于等于）
- `between(v, min, max)`：区间比较（含边界）
- `empty(v)` / `notEmpty(v)`：空值判断

### 2.3 字段路径

- JSON 路径：`$.department`、`$.user.name`
- 原文：`raw` 或 `$`

### 2.4 示例

```text
dsl:contains($.department, "研发部")
```

```text
dsl:contains($.department, "研发部") && exists($.name)
```

```text
dsl:regex($.department, ".*研发部.*") && in($.level, "P5", "P6", "P7")
```

## 三、提取 DSL（`dsl:`）

提取 DSL 最终返回一个字符串，写入“提取字段名”对应变量。

### 3.1 核心函数

- `pick(path)` / `json(path)`：提取 JSON 字段值
- `lower(v)`：小写
- `upper(v)`：大写
- `trim(v)`：去空白
- `replace(v, old, new)`：替换
- `concat(a, b, ...)`：拼接
- `split(v, sep, idx)`：分割取第 idx 项
- `regex(v, pattern[, group])`：正则提取
- `regexAll(v, pattern[, group][, sep])`：提取所有匹配并拼接（第三参也可直接传 sep）
- `findIdsByValue(path, target[, idKey][, sep])`：按 JSON 结构递归查找 `value=target` 的对象并返回 id 列表
- `arrayLen(path)`：获取 JSON 数组长度
- `valuesByKey(path, key[, sep])`：递归提取指定 key 的所有值
- `findByField(path, matchKey, matchValue, returnKey[, sep])`：按字段过滤并返回目标字段
- `findByFieldRaw(path, matchKey, matchValue[, sep])`：按字段过滤并返回匹配对象 JSON
- `default(v, fallback)`：为空时回退
- `coalesce(a, b, c, ...)`：返回第一个非空值
- `if(cond, a, b)`：条件表达式
- `len(v)`：字符串长度
- `substr(v, start[, length])`：字符串截取
- `toInt(v)` / `toFloat(v)`：数值转换
- `add(a, b)` / `sub(a, b)` / `mul(a, b)` / `div(a, b)`：四则运算

### 3.2 示例

```text
dsl:pick($.name)
```

```text
dsl:if(contains($.department, "研发部"), pick($.name), "")
```

```text
dsl:default(trim(pick($.name)), "unknown")
```

```text
dsl:concat("user:", lower(pick($.name)))
```

```text
dsl:if(gt(toFloat($.cost), 100), "HIGH", "NORMAL")
```

```text
dsl:coalesce(pick($.name), pick($.nickname), "unknown")
```

```text
dsl:regexAll(raw, "(?s)\\{[^{}]*\"id\"\\s*:\\s*(\\d+)[^{}]*\"value\"\\s*:\\s*\"target\"[^{}]*\\}", 1, "|")
```

```text
dsl:regexAll(raw, "(?s)\\{[^{}]*\"id\"\\s*:\\s*(\\d+)[^{}]*\"value\"\\s*:\\s*\"target\"[^{}]*\\}", "|")
```

```text
dsl:findIdsByValue($, "target", "id", "|")
```

```text
dsl:arrayLen($.children)
```

```text
dsl:valuesByKey($, "id", "|")
```

```text
dsl:findByField($, "value", "target", "id", "|")
```

```text
dsl:findByFieldRaw($, "value", "target", "|")
```

## 四、完整业务示例

目标：当 `department` 包含“研发部”时，提取 `name` 给模板变量 `MESSAGE_TEST`。

- 验证正则：

```text
dsl:contains($.department, "研发部")
```

- 提取正则：

```text
dsl:pick($.name)
```

- 提取字段名：

```text
MESSAGE_TEST
```

- 模板变量：

```text
{{MESSAGE_TEST}}
```

## 五、发送命令样例

```bash
sh mqadmin sendMessage -n "127.0.0.1:9876" -t "hukanfa-test-20250205" -p '{"department":"平台研发部","name":"kanfa.hu","text":"这是一条中文测试消息"}'
```

## 六、注意事项

- `提取字段名` 为空时，不会写入模板变量。
- 规则引擎不执行系统命令，仅做 DSL 匹配与提取。
- 中文建议 UTF-8 发送，避免上游把内容发成 `????`。
