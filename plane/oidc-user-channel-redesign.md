# OIDC 用户渠道化与独立管理重构方案

## 1. 需求概述

### 1.1 核心需求

1. **用户渠道区分**：用户表新增渠道字段，区分 `local`（本地）和 `oidc`（OIDC）用户
2. **默认分组功能**：认证设置中支持配置新用户的默认用户组
3. **OIDC 用户独立**：OIDC 用户不与本地用户绑定，相互独立
4. **功能精简**：基于用户独立，精简 OIDC 管理模块功能

### 1.2 现状问题

当前实现存在的问题：
- OIDC 用户通过 `auth_identities` 表与本地用户绑定（多对一关系）
- OIDC 用户必须先有本地账号，增加了复杂度
- 冲突审核、身份映射等功能过于复杂
- 无法直观区分用户来源渠道

---

## 2. 总体设计

### 2.1 核心原则

- **渠道隔离**：本地用户和 OIDC 用户在同一个表，但通过 `channel` 字段区分
- **独立认证**：OIDC 用户无需本地密码，直接通过 OIDC 认证
- **统一授权**：RBAC 权限体系对所有用户生效，与渠道无关
- **简化管理**：移除不必要的映射和冲突审核功能

### 2.2 架构对比

```
【当前架构】
OIDC 用户 → auth_identities 表 → 映射到 → auth 表（本地用户）
                    ↑
              冲突审核/身份映射管理

【新架构】
OIDC 用户 ──────────────────────────────┐
    ↓                                    │
auth 表（channel='oidc'）               │  统一 RBAC 授权
    ↓                                    │
RBAC 权限体系 ←─────────────────────────┘

本地用户
    ↓
auth 表（channel='local'）
    ↓
RBAC 权限体系
```

---

## 3. 数据库设计变更

### 3.1 用户表扩展

```sql
-- 1. 用户表新增渠道字段
ALTER TABLE message_auth ADD COLUMN channel VARCHAR(20) NOT NULL DEFAULT 'local';
ALTER TABLE message_auth ADD COLUMN external_sub VARCHAR(255) DEFAULT NULL;
ALTER TABLE message_auth ADD COLUMN external_issuer VARCHAR(255) DEFAULT NULL;
ALTER TABLE message_auth ADD COLUMN external_email VARCHAR(255) DEFAULT NULL;

-- 2. 创建索引
CREATE INDEX idx_auth_channel ON message_auth(channel);
CREATE INDEX idx_auth_external ON message_auth(external_sub, external_issuer);

-- 3. 渠道枚举约束
-- local: 本地注册/创建的用户
-- oidc: 通过 OIDC 登录自动创建的用户
```

### 3.2 数据迁移

```sql
-- 将现有 OIDC 映射用户迁移到新结构
-- 策略：将 auth_identities 中的数据合并到 auth 表

-- 步骤1：为现有 OIDC 映射用户创建独立账号
INSERT INTO message_auth (username, password, channel, external_sub, external_issuer, external_email, created_on, created_by)
SELECT 
    CONCAT('oidc_', ai.external_sub) as username,
    '' as password,  -- OIDC 用户无本地密码
    'oidc' as channel,
    ai.external_sub,
    ai.provider as external_issuer,
    ai.external_email,
    ai.created_on,
    'system'
FROM message_auth_identities ai
WHERE NOT EXISTS (
    SELECT 1 FROM message_auth a 
    WHERE a.channel = 'oidc' AND a.external_sub = ai.external_sub
);

-- 步骤2：迁移用户组/角色绑定（通过 user_id 关联）
-- 需要更新 rbac_user_role 和 rbac_user_group_member 表中的 user_id

-- 步骤3：废弃 auth_identities 表（可选，保留用于审计）
-- 或者标记为历史表
```

### 3.3 废弃/移除的表

| 表名 | 处理方式 | 说明 |
|------|----------|------|
| `message_auth_identities` | 废弃 | 不再需要身份映射 |
| `message_oidc_conflict_reviews` | 废弃 | 不再需要冲突审核 |

---

## 4. 后端改造

### 4.1 模型层变更

**models/auth.go**

```go
type Auth struct {
    ID            int    `json:"id" gorm:"autoIncrement;type:integer;primaryKey"`
    Username      string `json:"username" gorm:"type:varchar(100);default:''"`
    Password      string `json:"password" gorm:"type:varchar(100);default:''"`
    Channel       string `json:"channel" gorm:"type:varchar(20);default:'local'"` // local/oidc
    ExternalSub   string `json:"external_sub" gorm:"type:varchar(255)"`
    ExternalIssuer string `json:"external_issuer" gorm:"type:varchar(255)"`
    ExternalEmail string `json:"external_email" gorm:"type:varchar(255)"`
}

// 按渠道查询用户
func GetUserByChannelAndExternalID(channel, sub, issuer string) (*Auth, error)
func GetUsersByChannel(channel string, page, size int) ([]Auth, error)
```

### 4.2 OIDC 登录流程改造

**service/auth_service/oidc.go**

```go
// HandleCallback 改造后的回调处理
func (s *OIDCAuthService) HandleCallback(ctx context.Context, code, nonce string) (string, *Auth, error) {
    // ... 获取 OIDC token 和 profile ...
    
    // 1. 查找是否已存在该 OIDC 用户
    user, err := models.GetUserByChannelAndExternalID("oidc", profile.Sub, s.Config.Provider)
    if err == nil && user != nil {
        // 已存在，直接登录
        token, err := util.GenerateToken(user.Username, "", expDays) // OIDC 用户无密码
        return token, user, err
    }
    
    // 2. 不存在，自动创建（如果开启自动建档）
    if !s.Config.AutoCreateUser {
        return "", nil, fmt.Errorf("用户未注册，请联系管理员")
    }
    
    // 3. 创建新的 OIDC 用户
    username := generateOIDCUsername(profile)
    user = &models.Auth{
        Username:       username,
        Password:       "",  // OIDC 用户无本地密码
        Channel:        "oidc",
        ExternalSub:    profile.Sub,
        ExternalIssuer: s.Config.Provider,
        ExternalEmail:  profile.Email,
    }
    
    if err := models.CreateUser(user); err != nil {
        return "", nil, err
    }
    
    // 4. 绑定默认用户组
    if s.Config.DefaultGroupCode != "" {
        bindDefaultGroup(user.ID, s.Config.DefaultGroupCode)
    }
    
    // 5. 生成 token
    token, err := util.GenerateToken(user.Username, "", expDays)
    return token, user, err
}
```

### 4.3 本地登录改造

**models/auth.go**

```go
// CheckAuth 改造：区分本地用户和 OIDC 用户
func CheckAuth(username, password string) (*Auth, error) {
    var auth Auth
    err := db.Where("username = ?", username).First(&auth).Error
    if err != nil {
        return nil, err
    }
    
    // OIDC 用户不允许本地密码登录
    if auth.Channel == "oidc" {
        return nil, errors.New("OIDC 用户请使用 OIDC 登录")
    }
    
    // 本地用户校验密码
    encryptedPwd := util.EncodeMD5(password)
    if auth.Password != encryptedPwd {
        return nil, errors.New("密码错误")
    }
    
    return &auth, nil
}
```

### 4.4 认证设置扩展

**新增配置项**

```go
type AuthConfig struct {
    // ... 现有配置 ...
    
    // 新增：默认用户组配置
    DefaultGroupCode string `json:"oidc_default_group_code"` // 也用于本地新用户
}
```

配置存储：
- `auth_config.oidc_default_group_code`: OIDC 新用户默认组
- `auth_config.local_default_group_code`: 本地新用户默认组（可选，可与上面共用）

### 4.5 用户管理改造

**service/rbac_service/manage_service.go**

```go
// UserManageService 添加渠道字段
type UserManageService struct {
    ID       int
    Username string
    Password string
    Channel  string // 新增
    // ...
}

// Add 创建用户时指定渠道
func (s *UserManageService) Add() error {
    // ...
    auth := &models.Auth{
        Username: s.Username,
        Password: util.EncodeMD5(s.Password),
        Channel:  s.Channel, // local/oidc
    }
    // ...
    
    // 绑定默认用户组
    defaultGroup := getDefaultGroupForChannel(s.Channel)
    if defaultGroup != "" {
        bindDefaultGroup(user.ID, defaultGroup)
    }
}
```

---

## 5. 前端改造

### 5.1 用户管理列表

**用户列表新增渠道列**

```vue
<Table>
  <TableHead>用户名</TableHead>
  <TableHead>渠道</TableHead>  <!-- 新增 -->
  <TableHead>用户组</TableHead>
  <TableHead>角色</TableHead>
  <TableHead>操作</TableHead>
</Table>

<TableCell>
  <Badge :variant="item.channel === 'oidc' ? 'blue' : 'default'">
    {{ item.channel === 'oidc' ? 'OIDC' : '本地' }}
  </Badge>
</TableCell>
```

### 5.2 认证设置页面

**登录策略板块新增默认用户组配置**

```vue
<div class="space-y-4">
  <div class="text-sm font-semibold">登录策略</div>
  
  <!-- 现有：身份冲突处理 -->
  <div>...</div>
  
  <!-- 新增：默认用户组配置 -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div class="space-y-2">
      <label>本地新用户默认组</label>
      <select v-model="config.local_default_group_code">
        <option v-for="g in userGroups" :key="g.code" :value="g.code">
          {{ g.name }}
        </option>
      </select>
    </div>
    <div class="space-y-2">
      <label>OIDC 新用户默认组</label>
      <select v-model="config.oidc_default_group_code">
        <option v-for="g in userGroups" :key="g.code" :value="g.code">
          {{ g.name }}
        </option>
      </select>
    </div>
  </div>
</div>
```

### 5.3 登录页提示

**OIDC 用户尝试本地登录时提示**

```vue
<!-- 登录表单 -->
<form @submit="handleLogin">
  <Input v-model="form.username" placeholder="用户名" />
  <Input v-model="form.password" type="password" placeholder="密码" />
  <Button type="submit">登录</Button>
</form>

<!-- 错误提示 -->
<div v-if="errorMsg === 'OIDC_USER_USE_OIDC_LOGIN'">
  该账号为 OIDC 用户，请使用下方 OIDC 登录按钮
</div>

<!-- OIDC 登录按钮 -->
<Button variant="outline" @click="oidcLogin">
  <img src="/oidc-icon.svg" /> OIDC 登录
</Button>
```

---

## 6. OIDC 管理模块精简

### 6.1 功能调整

| 功能模块 | 当前状态 | 新状态 | 说明 |
|----------|----------|--------|------|
| **告警监控** | 保留 | 保留 | 监控 OIDC 登录失败和告警 |
| **冲突审核** | 保留 | **移除** | 用户独立后不再需要 |
| **身份映射** | 保留 | **移除** | 用户独立后不再需要 |
| **审计日志** | 保留 | 保留 | 记录 OIDC 登录事件 |

### 6.2 精简后的 OIDC 管理

```
系统管理 - OIDC 管理
├── 告警监控（原告警配置+指标卡片合并）
│   ├── 实时指标卡片
│   └── 告警配置（Switch + 阈值设置）
└── 审计日志
    └── OIDC 登录/登出/失败记录
```

### 6.3 后端 API 调整

**移除的接口**

- `GET /api/v1/oidc/identities` - 身份映射列表
- `POST /api/v1/oidc/identities/unbind` - 解绑身份
- `GET /api/v1/oidc/conflicts` - 冲突审核列表
- `POST /api/v1/oidc/conflicts/approve` - 通过审核
- `POST /api/v1/oidc/conflicts/reject` - 驳回审核

**保留的接口**

- `GET /api/v1/oidc/metrics` - 指标统计
- `GET /api/v1/oidc/alert-config` - 获取告警配置
- `POST /api/v1/oidc/alert-config` - 更新告警配置
- `GET /api/v1/oidc/audits` - 审计日志列表

---

## 7. RBAC 适配

### 7.1 权限设计

所有 RBAC 权限对用户渠道透明：

```
用户（local/oidc）
    ↓
用户组（与渠道无关）
    ↓
角色（与渠道无关）
    ↓
权限（与渠道无关）
```

### 7.2 用户组绑定

- 本地新用户 → 绑定 `local_default_group_code` 配置的组
- OIDC 新用户 → 绑定 `oidc_default_group_code` 配置的组
- 管理员可手动调整任何用户的用户组

---

## 8. 实施计划

### 阶段一：数据库迁移（1天）

1. 用户表添加 `channel`, `external_sub`, `external_issuer`, `external_email` 字段
2. 数据迁移：将 `auth_identities` 数据合并到 `auth` 表
3. 更新 `rbac_user_role` 和 `rbac_user_group_member` 的 user_id 关联
4. 废弃旧表（或保留为历史）

### 阶段二：后端改造（2天）

1. 改造 `models/auth.go` - 新增渠道相关字段和方法
2. 改造 `service/auth_service/oidc.go` - 简化登录流程，移除映射逻辑
3. 改造 `routers/api/auth.go` - 区分本地和 OIDC 用户登录
4. 改造认证设置 - 添加默认用户组配置
5. 移除 OIDC 管理相关接口（冲突审核、身份映射）

### 阶段三：前端改造（2天）

1. 用户管理列表添加渠道列和筛选
2. 认证设置页面添加默认用户组配置
3. 登录页添加 OIDC 用户提示
4. 精简 OIDC 管理页面（移除冲突审核、身份映射）

### 阶段四：测试验证（1天）

1. 本地用户注册/登录测试
2. OIDC 用户首次登录测试（自动建档）
3. OIDC 用户重复登录测试
4. 用户组自动绑定测试
5. RBAC 权限验证（本地和 OIDC 用户）

---

## 9. 回滚策略

### 数据库回滚

```sql
-- 如果出现问题，恢复 auth_identities 表
-- 1. 从备份恢复 auth_identities 数据
-- 2. 删除 auth 表新增字段
-- 3. 恢复旧代码版本
```

### 代码回滚

- 保留原 OIDC 服务代码分支
- 保留原 auth_identities 模型
- 通过 feature flag 控制新旧逻辑切换

---

## 10. 验收标准

1. ✅ 用户表有 `channel` 字段，正确区分 local/oidc
2. ✅ 本地用户只能通过账号密码登录
3. ✅ OIDC 用户只能通过 OIDC 登录
4. ✅ 新用户自动绑定配置的默认用户组
5. ✅ 用户管理列表显示渠道标识
6. ✅ OIDC 管理页面只保留告警监控和审计日志
7. ✅ RBAC 权限对两种用户都生效
8. ✅ 审计日志正确记录 OIDC 登录事件

---

## 11. 后续优化方向

1. **渠道扩展**：支持更多渠道（LDAP、SAML、企业微信等）
2. **渠道图标**：不同渠道显示不同图标
3. **渠道统计**：按渠道统计用户数量和活跃度
4. **渠道策略**：不同渠道可配置不同的默认角色和权限
