# Message Nest 对接 Casdoor OIDC 详细指南

## 1. 概述

本文档指导如何将 Message Nest 与 Casdoor OIDC 服务对接，实现单点登录（SSO）功能。

**平台信息**：
- Message Nest 域名：`https://unipush.ops.dachensky.com`
- 认证协议：OIDC (OpenID Connect)
- 授权流程：Authorization Code Flow

---

## 2. 前置条件

### 2.1 必需组件

| 组件 | 版本/要求 | 说明 |
|------|----------|------|
| Casdoor 服务 | v1.0+ | 已部署并可访问 |
| Message Nest | 最新版本 | 已完成 OIDC 渠道化改造 |
| HTTPS | 必须 | 生产环境强制要求 |

### 2.2 网络要求

- Message Nest 服务器能访问 Casdoor 的 Issuer 地址
- 用户浏览器能访问 Casdoor 登录页面
- 回调地址必须公网可访问

---

## 3. Casdoor 服务端配置

### 3.1 登录 Casdoor 管理后台

访问你的 Casdoor 管理后台（如 `https://casdoor.your-domain.com`）

### 3.2 创建应用

1. 进入 **应用管理** → **添加应用**
2. 填写应用信息：

| 字段 | 值 | 说明 |
|------|-----|------|
| 应用名称 | `Message Nest` | 显示名称 |
| 组织 | `built-in` | 或你的组织 |
| 协议 | `OIDC` | 默认 |
| 客户端ID | `unipush-oidc-client` | 自定义，后续填入 Message Nest |
| 客户端密钥 | 点击生成 | 复制保存，后续填入 Message Nest |

### 3.3 配置回调地址

在应用详情页，找到 **回调地址** 配置：

```
https://unipush.ops.dachensky.com/auth/oidc/callback
```

**注意**：
- 必须包含协议（https://）
- 必须与 Message Nest 中配置的回调地址完全一致
- 区分大小写

### 3.4 配置登出回调（可选）

```
https://unipush.ops.dachensky.com/auth/oidc/logout/callback
```

### 3.5 获取 Casdoor 配置信息

配置完成后，记录以下信息：

| 配置项 | 获取位置 | 示例值 |
|--------|---------|--------|
| Issuer | Casdoor 首页 URL | `https://casdoor.your-domain.com` |
| Client ID | 应用详情页 | `unipush-oidc-client` |
| Client Secret | 应用详情页 | `a1b2c3d4e5f6...` |

---

## 4. Message Nest 配置

### 4.1 进入认证设置

1. 登录 Message Nest 管理员账号
2. 进入 **系统管理** → **系统设置** → **认证设置**

### 4.2 基础开关配置

| 配置项 | 值 | 说明 |
|--------|-----|------|
| 启用 OIDC 登录 | `true` | 开启 OIDC 功能 |
| 允许新用户注册 | 按需 | 建议生产环境关闭，仅 OIDC 用户登录 |

### 4.3 OIDC 核心配置

| 配置项 | 值 | 说明 |
|--------|-----|------|
| OIDC 供应商 | `casdoor` | 保持默认 |
| OIDC Issuer | `https://casdoor.your-domain.com` | Casdoor 服务地址 |
| OIDC ClientID | `unipush-oidc-client` | Casdoor 中创建的客户端ID |
| OIDC ClientSecret | `a1b2c3d4e5f6...` | Casdoor 中生成的密钥 |
| OIDC 回调地址 | `https://unipush.ops.dachensky.com/auth/oidc/callback` | 必须与 Casdoor 中一致 |
| OIDC Scopes | `openid profile email` | 保持默认 |

### 4.4 登录策略配置

| 配置项 | 建议值 | 说明 |
|--------|--------|------|
| 自动建档 | `true` | 首次登录自动创建用户 |
| OIDC 新用户默认组 | 选择用户组 | 如 `default_group` |
| 回调重试次数 | `2` | 网络不稳定时可适当增加 |

### 4.5 登录页设置

| 配置项 | 建议值 |
|--------|--------|
| 登录按钮文案 | `Casdoor 登录` 或 `企业 SSO` |
| OIDC 登录图标 | 可选，建议上传 Casdoor Logo |

### 4.6 保存配置

点击 **保存** 按钮，配置将立即生效。

---

## 5. 验证测试

### 5.1 测试步骤

1. **退出登录**
   - 点击右上角头像 → 退出登录

2. **访问登录页**
   - 访问 `https://unipush.ops.dachensky.com/#/login`

3. **检查 OIDC 登录按钮**
   - 应显示配置的按钮文案（如 "Casdoor 登录"）
   - 如果配置了图标，应显示图标

4. **点击 OIDC 登录**
   - 应跳转至 Casdoor 登录页面

5. **Casdoor 登录**
   - 输入 Casdoor 用户名密码
   - 首次登录可能需要授权应用访问权限

6. **返回 Message Nest**
   - 登录成功后自动跳转回 Message Nest
   - 应进入系统首页

### 5.2 验证用户创建

1. 进入 **系统管理** → **用户管理**
2. 检查新登录的用户是否在列表中
3. 查看用户详情，确认：
   - 渠道为 `OIDC`
   - 已绑定默认用户组
   - 外部标识正确

---

## 6. 常见问题排查

### 6.1 登录按钮不显示

**现象**：登录页没有 OIDC 登录按钮

**排查**：
1. 检查 **启用 OIDC 登录** 开关是否为 `true`
2. 检查 Issuer、ClientID、ClientSecret 是否已填写
3. 检查浏览器控制台是否有错误

### 6.2 跳转后显示配置错误

**现象**：点击登录后显示 "OIDC配置校验失败"

**排查**：
1. 检查 Issuer 地址是否正确（需包含 https://）
2. 检查 Message Nest 服务器是否能访问 Casdoor
3. 检查 Casdoor 服务是否正常运行

### 6.3 Casdoor 提示回调地址不匹配

**现象**：Casdoor 页面显示 "redirect_uri mismatch"

**排查**：
1. 检查 Casdoor 应用中的回调地址
2. 检查 Message Nest 中的回调地址
3. 确保两者完全一致（包括协议、域名、路径）

### 6.4 登录后提示用户未注册

**现象**：返回 Message Nest 后提示用户不存在

**排查**：
1. 检查 **自动建档** 开关是否为 `true`
2. 检查 Casdoor 用户是否有 email 字段
3. 检查 Message Nest 日志获取详细错误

### 6.5 权限不正确

**现象**：登录后无法访问某些功能

**排查**：
1. 检查 **OIDC 新用户默认组** 是否配置正确
2. 检查用户组是否绑定了正确的角色
3. 手动为用户分配角色测试

---

## 7. 高级配置

### 7.1 统一登出（Single Logout）

如需实现统一登出，在 Casdoor 和 Message Nest 中配置：

**Casdoor 应用配置**：
```
登出回调地址: https://unipush.ops.dachensky.com/auth/oidc/logout/callback
```

**Message Nest 配置**：
```
统一登出地址: https://casdoor.your-domain.com/api/logout
```

### 7.2 用户属性映射

Message Nest 支持获取以下 OIDC 属性：

| OIDC Claim | Message Nest 字段 | 说明 |
|------------|------------------|------|
| `sub` | external_sub | 用户唯一标识 |
| `email` | external_email | 邮箱地址 |
| `name` | - | 显示名称（暂未使用）|
| `preferred_username` | username | 用户名生成依据 |

### 7.3 安全增强

建议在生产环境启用以下安全选项：

1. **PKCE**（已默认启用）
2. **State 校验**（已默认启用）
3. **Nonce 校验**（已默认启用）
4. **HTTPS 强制**（生产环境必需）

---

## 8. 回滚方案

如需禁用 OIDC 登录：

1. 进入 **系统管理** → **系统设置** → **认证设置**
2. 将 **启用 OIDC 登录** 设为 `false`
3. 保存配置
4. 登录页将立即恢复为本地账号密码登录

**注意**：已创建的 OIDC 用户仍可正常使用，但无法通过 OIDC 登录（需使用本地密码，如已设置）

---

## 9. 技术支持

如遇到问题，请提供以下信息：

1. Message Nest 版本号
2. Casdoor 版本号
3. 浏览器控制台错误日志
4. Message Nest 服务端日志
5. 配置截图（敏感信息打码）

---

## 10. 更新记录

| 版本 | 日期 | 更新内容 |
|------|------|---------|
| 1.0 | 2026-03-31 | 初始版本，基于 OIDC 渠道化改造 |
