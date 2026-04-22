# Message Nest 外部认证接入方案（Casdoor + OIDC）

## 1. 目标与边界

本方案用于将“认证”能力外置到 Casdoor，授权继续使用本项目本地 RBAC。

- 认证：Casdoor（OIDC）
- 授权：Message Nest 本地 RBAC（用户组/角色/权限）
- 原则：认证外置、授权内置

---

## 2. 协议选型

推荐协议：

- OIDC（OpenID Connect）
- OAuth2 Authorization Code Flow
- PKCE（建议启用）

不建议优先选用 SAML/CAS，除非必须兼容历史系统。

---

## 3. 登录流程设计

1. 前端点击登录，跳转后端 `/auth/oidc/login`
2. 后端重定向到 Casdoor 授权端点
3. 用户在 Casdoor 完成认证后回调 `/auth/oidc/callback`
4. 后端用授权码换 token（id_token / access_token）
5. 后端校验 id_token（issuer/audience/exp/nonce）
6. 提取用户标识（sub/email），做本地用户映射
7. 生成本系统 JWT（兼容现有鉴权链路）
8. 前端后续仍使用本系统 JWT 调用业务接口

---

## 4. 用户映射策略

建议映射键优先级：

1. `sub`（外部唯一ID）
2. `email`

建议新增字段：

- `auth_source`（local/casdoor）
- `external_sub`
- `external_issuer`

首次登录策略（可配置）：

- 自动创建本地用户（默认）
- 默认加入 `default_group`
- 默认分配最小权限角色

---

## 5. 与 RBAC 的衔接方式

采用“弱耦合”策略：

- Casdoor 负责“你是谁”
- 本地 RBAC 负责“你能做什么”

可选增强：

- 将 Casdoor claim（groups/roles）映射到本地用户组/角色
- 支持“首次映射 + 手工覆盖”

---

## 6. 后端改造清单（Gin）

新增配置项（`conf/app.ini`）：

- `OIDCEnabled`
- `OIDCIssuer`
- `OIDCClientID`
- `OIDCClientSecret`
- `OIDCRedirectURL`
- `OIDCScopes`（`openid profile email`）
- `OIDCLogoutURL`（可选）

新增模块建议：

- `service/auth_service/oidc_service.go`
- `middleware/oidc_state.go`（state/nonce 安全校验）
- `routers/api/auth_oidc.go`

新增接口建议：

- `GET /auth/oidc/login`
- `GET /auth/oidc/callback`
- `POST /auth/oidc/logout`

---

## 7. 前端改造清单（Vue3）

- 登录页增加“Casdoor 登录”入口
- 回调后处理本系统 JWT 落库
- 401 跳转到统一登录入口
- 登出时调用后端登出接口，再清理本地状态

---

## 8. 安全要求

- 必须校验 `state`、`nonce`
- 必须校验 `iss`、`aud`、`exp`
- 使用 HTTPS
- Token 不写日志
- 本系统 JWT 设置合理有效期与刷新策略

---

## 9. 分阶段实施计划

### 阶段 A：最小可用登录

- 打通 Casdoor 登录回调
- 完成本地用户映射
- 签发本系统 JWT

### 阶段 B：与 RBAC 联动

- 默认用户组策略
- claim 到用户组/角色映射

### 阶段 C：完善体验与安全

- 单点登出
- 会话续期与失效策略
- 审计日志与告警

---

## 10. 验收标准

1. 可通过 Casdoor 成功登录并访问业务接口
2. 本地 RBAC 权限控制仍完全生效
3. 未授权用户返回 403
4. 登录回调全过程具备安全校验
5. 登出后本地与统一认证会话状态一致
