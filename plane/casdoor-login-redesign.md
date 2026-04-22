# Casdoor 登录功能重新设计

## 背景

现有 OIDC 实现存在以下问题：
1. **502 错误频发** - 各种类型断言、panic 等问题
2. **授权码重复使用** - 浏览器/插件导致回调被多次触发
3. **id_token 太长** - 3487 字符，无法通过 URL 传递
4. **nonce 校验失败** - Casdoor API 不返回 nonce

## 设计目标

1. **简单可靠** - 专门针对 Casdoor，不追求通用 OIDC 兼容
2. **防重放** - 确保授权码只被使用一次
3. **统一登出** - 支持 Casdoor 单点登出
4. **最小改动** - 复用现有数据库和用户体系

## 核心设计

### 1. 登录流程

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  前端   │────>│  后端   │────>│ Casdoor │────>│  后端   │
│ 点击登录│     │ /login  │     │ 授权页  │     │/callback│
└─────────┘     └─────────┘     └─────────┘     └─────────┘
                    │                               │
                    │ 1. 生成 state                 │ 4. 验证 state
                    │ 2. 存入 Redis/内存            │ 5. 换取 token
                    │ 3. 重定向到 Casdoor           │ 6. 获取用户信息
                    │                               │ 7. 创建本地会话
                    │                               │ 8. 存储 id_token
                    │                               │ 9. 重定向到前端
                    ▼                               ▼
```

### 2. 防重放机制

**方案：state 一次性使用**

```go
// 使用内存缓存存储 state（TTL 5分钟）
type StateCache struct {
    sync.RWMutex
    states map[string]*StateInfo
}

type StateInfo struct {
    Nonce     string
    CreatedAt time.Time
    Used      bool  // 标记是否已使用
}

// 验证并消费 state（原子操作）
func (c *StateCache) ConsumeState(state string) (*StateInfo, bool) {
    c.Lock()
    defer c.Unlock()
    
    info, exists := c.states[state]
    if !exists || info.Used {
        return nil, false
    }
    info.Used = true  // 标记为已使用
    delete(c.states, state)  // 立即删除
    return info, true
}
```

### 3. id_token 存储

**方案：后端存储，不经过前端**

```go
// 登录成功后，将 id_token 存入数据库或缓存
type OIDCSession struct {
    UserID     uint
    IDToken    string    // 用于统一登出
    ExpireAt   time.Time
}

// 存储位置选择：
// - 小规模：内存缓存（sync.Map）
// - 大规模：Redis
// - 持久化：数据库表
```

### 4. 统一登出

```
┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│  前端   │────>│  后端   │────>│ Casdoor │────>│  后端   │
│ 点击登出│     │ /logout │     │ 登出API │     │/callback│
└─────────┘     └─────────┘     └─────────┘     └─────────┘
                    │                               │
                    │ 1. 读取 id_token              │
                    │ 2. 构建登出 URL               │
                    │ 3. 清除本地会话               │
                    │ 4. 重定向到 Casdoor           │
                    ▼                               ▼
```

## API 设计

### 后端接口

| 接口 | 方法 | 说明 |
|-----|------|------|
| `/auth/casdoor/login` | GET | 发起 Casdoor 登录 |
| `/auth/casdoor/callback` | GET | 处理 Casdoor 回调 |
| `/auth/casdoor/logout` | POST | 发起统一登出 |
| `/auth/casdoor/logout/callback` | GET | 登出回调 |

### 配置项

**Casdoor 独立配置**，不复用 OIDC 配置。在 `settings` 表中 `section='auth_config'`：

| key | 说明 | 示例 |
|-----|------|------|
| `casdoor_enabled` | 启用/禁用 Casdoor 登录 | `true` |
| `casdoor_endpoint` | Casdoor 服务地址 | `https://sso.dachensky.com` |
| `casdoor_client_id` | OAuth2 客户端 ID | `7eb9cad71b0dc0280b7e` |
| `casdoor_client_secret` | OAuth2 客户端密钥 | `xxx` |
| `casdoor_redirect_uri` | 登录回调 URL | `https://unipush.ops.dachensky.com/auth/casdoor/callback` |
| `casdoor_auth_path` | 授权端点（可选） | `/login/oauth/authorize` |
| `casdoor_token_path` | Token 端点（可选） | `/api/login/oauth/access_token` |
| `casdoor_userinfo_path` | UserInfo 端点（可选） | `/api/get-account` |
| `casdoor_logout_path` | 登出端点（可选） | `/api/logout` |
| `casdoor_auto_create_user` | 新用户自动创建本地账号 | `true` |
| `casdoor_default_group_code` | 新用户分配的默认组 code | `normal_user` |
| `casdoor_button_text` | 登录页按钮显示文字 | `企微登录` |

**默认值：**
- `casdoor_auth_path`: `/login/oauth/authorize`
- `casdoor_token_path`: `/api/login/oauth/access_token`
- `casdoor_userinfo_path`: `/api/get-account`
- `casdoor_logout_path`: `/api/logout`
- `casdoor_auto_create_user`: `true`
- `casdoor_default_group_code`: 组 ID=2

### 快速初始化 SQL

```sql
INSERT INTO settings (section, `key`, value) VALUES
('auth_config', 'casdoor_enabled', 'true'),
('auth_config', 'casdoor_endpoint', 'https://sso.dachensky.com'),
('auth_config', 'casdoor_client_id', 'your_client_id'),
('auth_config', 'casdoor_client_secret', 'your_client_secret'),
('auth_config', 'casdoor_redirect_uri', 'https://unipush.ops.dachensky.com/auth/casdoor/callback'),
('auth_config', 'casdoor_button_text', '企微登录');
```

## 数据库设计

### 复用现有表

- `auth` - 用户表，使用 `channel='casdoor'` 和 `external_sub` 字段
- `auth_identity` - 身份绑定表，`provider='casdoor'`
- `login_log` - 登录日志

### id_token 存储

**方案：内存缓存（当前实现）**

```go
// IDTokenStore 使用 sync.RWMutex + map 实现
type IDTokenStore struct {
    mu     sync.RWMutex
    tokens map[uint]string // userID -> idToken
}
```

**特点：**
- 简单轻量，无外部依赖
- 重启后 id_token 丢失，但不影响登录（只影响统一登出）
- 如需持久化，可改用 Redis 或数据库表

## 前端设计

### 登录页

```vue
<template>
  <Button @click="casdoorLogin" :loading="loading">
    <img src="/casdoor-icon.svg" />
    企微登录
  </Button>
</template>

<script setup>
const casdoorLogin = () => {
  // 直接跳转，不需要 AJAX
  window.location.href = '/auth/casdoor/login'
}
</script>
```

### 回调处理

后端直接重定向到前端页面，带上本地 token：
```
/login?token=xxx&login_type=casdoor
```

前端只需要：
1. 从 URL 获取 token
2. 存储到 localStorage
3. 跳转到首页
4. **立即清除 URL 参数**

## 实现状态

### 第一阶段：核心登录 ✅

1. [x] 新建 `service/casdoor_service/` 目录
2. [x] 实现 state 缓存机制 (`state_cache.go`)
3. [x] 实现 `/auth/casdoor/login` 接口
4. [x] 实现 `/auth/casdoor/callback` 接口
5. [x] 实现用户映射/自动创建
6. [x] 前端登录按钮对接

### 第二阶段：统一登出 ✅

1. [x] 实现 id_token 存储 (`IDTokenStore`)
2. [x] 实现 `/api/v1/auth/casdoor/logout` 接口
3. [x] 实现 `/auth/casdoor/logout/callback` 接口
4. [x] 前端登出对接

### 第三阶段：管理功能 ✅

1. [x] 配置独立化（不复用 OIDC 配置）
2. [x] 配置页面对接（AuthSettings.vue）
3. [ ] 会话管理（可选）
4. [ ] 审计日志（可选）

## 新增文件清单

| 文件 | 说明 |
|------|------|
| `service/casdoor_service/config.go` | 配置加载，独立 `casdoor_*` 配置项 |
| `service/casdoor_service/state_cache.go` | 内存 state 缓存，防重放 |
| `service/casdoor_service/casdoor.go` | 核心服务：Token 交换、用户信息、用户映射 |
| `routers/api/auth_casdoor.go` | API 路由处理 |

## 与现有 OIDC 的关系

**完全独立，不复用任何 OIDC 代码和配置**

| 对比项 | Casdoor | OIDC（原有） |
|--------|---------|------------|
| 配置前缀 | `casdoor_*` | `oidc_*` |
| 服务文件 | `casdoor_service/` | `auth_service/oidc.go` |
| 路由前缀 | `/auth/casdoor/*` | `/auth/oidc/*` |
| 用户标识 | `channel='casdoor'` | `channel='oidc'` |

**无任何交集，可安全并存或独立移除**

## 风险评估

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 内存缓存丢失 | state 失效 | 设置合理 TTL，失败可重试 |
| Casdoor 不可用 | 无法登录 | 保留本地账号登录 |
| Token 泄露 | 安全风险 | HTTPS + HttpOnly Cookie |

## 时间估计

- 第一阶段：2-3 小时
- 第二阶段：1-2 小时
- 第三阶段：1 小时

**总计：4-6 小时**
