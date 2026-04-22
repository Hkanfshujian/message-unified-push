# API 调用示例

本页提供 `ops-message-unified-push` 的常见调用方式。  
新接入统一使用 **V2 模板接口**。

## 调用前准备

### 1. 服务地址

- 本地开发：`http://127.0.0.1:8000`
- 生产环境：`https://your-domain`

### 2. Token 说明

- **V2 接口**：使用“消息模板”中生成的模板 Token

### 3. 环境变量（可选）

```bash
export BASE_URL="http://127.0.0.1:8000"
export TOKEN="your_token"
```

PowerShell:

```powershell
$env:BASE_URL = "http://127.0.0.1:8000"
$env:TOKEN = "your_token"
```

## V2 模板发送（推荐）

接口地址：`POST /api/v2/message/send`

### cURL

```bash
curl -X POST "$BASE_URL/api/v2/message/send" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "'$TOKEN'",
    "title": "服务告警",
    "placeholders": {
      "service": "order-api",
      "status": "degraded",
      "time": "2026-04-14 12:00:00"
    }
  }'
```

### Python (requests)

```python
import requests

base_url = "http://127.0.0.1:8000"
token = "your_token"

payload = {
    "token": token,
    "title": "服务告警",
    "placeholders": {
        "service": "order-api",
        "status": "degraded",
        "time": "2026-04-14 12:00:00"
    }
}

resp = requests.post(f"{base_url}/api/v2/message/send", json=payload, timeout=10)
print(resp.status_code, resp.text)
```

### Node.js (fetch)

```javascript
async function main() {
  const baseUrl = 'http://127.0.0.1:8000'
  const token = 'your_token'

  const payload = {
    token,
    title: '服务告警',
    placeholders: {
      service: 'order-api',
      status: 'degraded',
      time: '2026-04-14 12:00:00'
    }
  }

  const resp = await fetch(`${baseUrl}/api/v2/message/send`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })

  console.log(resp.status, await resp.text())
}

main().catch(console.error)
```

### Go

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	jsonBody := []byte(`{
  "token": "your_token",
  "title": "服务告警",
  "placeholders": {
    "service": "order-api",
    "status": "degraded",
    "time": "2026-04-14 12:00:00"
  }
}`)

	req, _ := http.NewRequest("POST", "http://127.0.0.1:8000/api/v2/message/send", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(body))
}
```

### Java

```java
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class OpsMessageUnifiedPushV2Example {
    public static void main(String[] args) throws Exception {
        String payload = """
            {
              "token": "your_token",
              "title": "服务告警",
              "placeholders": {
                "service": "order-api",
                "status": "degraded",
                "time": "2026-04-14 12:00:00"
              }
            }
            """;

        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create("http://127.0.0.1:8000/api/v2/message/send"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(payload))
                .build();

        HttpResponse<String> resp = HttpClient.newHttpClient().send(req, HttpResponse.BodyHandlers.ofString());
        System.out.println(resp.statusCode() + " " + resp.body());
    }
}
```

## 常见错误排查

| 现象 | 常见原因 | 处理建议 |
|------|----------|----------|
| `401` 未授权 | Token 无效或已禁用 | 在管理后台重新生成/启用 Token |
| `404` Token 不存在 | 请求的 Token 不属于当前环境 | 检查环境变量与服务地址是否一致 |
| `400` 参数错误 | 缺少必填字段或 JSON 格式错误 | 对照参数文档检查请求体 |
| 请求超时 | 服务不可达或网络问题 | 检查网关/Nginx/防火墙配置 |

## 下一步

- 查看 [API 使用说明](/api/usage) 了解参数与响应定义
- 查看 [V2 API 文档](/api/v2) 了解模板发送的完整能力
- 如需迁移历史任务，请参考仓库中的 V1 历史文档（默认不再作为主文档入口）