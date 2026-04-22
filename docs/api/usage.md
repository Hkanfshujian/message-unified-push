# API 使用说明

ops-message-unified-push 提供模板化的消息推送 API 接口。

## 接口地址

```
POST /api/v2/message/send
```

## 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | 模板令牌，在消息模板中获取 |
| title | string | 是 | 消息标题 |
| placeholders | object | 否 | 模板占位符键值对 |
| recipients | array | 否 | 动态接收者列表（支持的渠道有效） |

## 请求示例

```json
{
    "token": "your_template_token",
    "title": "订单通知",
    "placeholders": {
        "order_id": "20241206001",
        "status": "已发货"
    }
}
```

## 响应格式

### 成功响应

```json
{
    "code": 200,
    "msg": "success",
    "data": {
        "status": "sent"
    }
}
```

### 失败响应

```json
{
    "code": 400,
    "msg": "error message",
    "data": null
}
```

## 获取 Token

1. 登录 ops-message-unified-push 管理后台
2. 进入"消息模板"页面
3. 创建模板并配置推送渠道实例
4. 保存后获得模板 Token
   - 获得唯一的推送令牌（Token）

3. **调用API发送消息**
   - 使用获得的 Token
   - 发送标题和内容
   - 消息会自动推送到配置的所有渠道

## 注意事项

::: warning 重要
- Token 是唯一的，请妥善保管
- 消息内容支持纯文本和Markdown格式（取决于推送渠道）
- 建议使用异步方式调用API，避免阻塞主流程
:::

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 404 | Token不存在 |
| 500 | 服务器内部错误 |

## 下一步

查看各语言的 [调用示例](/api/examples)。
