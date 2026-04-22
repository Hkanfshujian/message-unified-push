# 认证配置接口说明（系统设置）

## 接口概览

- `GET /api/v1/system/auth-config`
- `POST /api/v1/system/auth-config`

用途：提供系统管理中的认证配置读取与更新能力，替代通用 `/settings/set` 的混合写入方式。

---

## 权限矩阵

| 接口 | 方法 | 权限码 | 说明 |
|---|---|---|---|
| `/api/v1/system/auth-config` | GET | `system:settings:view` | 读取认证配置 |
| `/api/v1/system/auth-config` | POST | `system:settings:edit` | 更新认证配置 |
| `/api/v1/profile/theme` | GET | `profile:settings:view` | 读取个人主题偏好 |
| `/api/v1/profile/theme` | POST | `profile:settings:edit` | 更新个人主题偏好 |
| `/api/v1/profile/password` | POST | `profile:settings:edit` | 修改个人密码 |

---

## GET 返回字段（`auth_config`）

- `register_enabled`：是否允许公开注册（`true/false`）
- `oidc_enabled`：是否启用 OIDC（`true/false`）
- `oidc_provider`：OIDC 提供商标识
- `oidc_issuer`：OIDC issuer 地址
- `oidc_client_id`：OIDC client id
- `oidc_client_secret`：OIDC client secret
- `oidc_redirect_url`：OIDC 回调地址
- `oidc_scopes`：OIDC scopes（逗号分隔）
- `oidc_auto_create_user`：首次登录是否自动建档
- `oidc_default_group_code`：默认用户组编码
- `oidc_default_role_code`：默认角色编码
- `oidc_frontend_login_path`：前端登录页路径
- `oidc_logout_url`：统一登出地址
- `oidc_conflict_mode`：冲突模式（`auto_bind/manual_review`）
- `oidc_exchange_retry`：回调 token 交换重试次数
- `oidc_alert_enabled`：告警开关
- `oidc_alert_channel`：告警通道
- `oidc_alert_webhook`：告警 webhook
- `oidc_alert_failure_threshold`：失败阈值
- `oidc_alert_window_minutes`：统计窗口分钟数
- `oidc_alert_cooldown_minutes`：告警冷却分钟数
- `oidc_last_alert_at`：最近告警时间

---

## POST 请求体

```json
{
  "data": {
    "register_enabled": "false",
    "oidc_enabled": "true",
    "oidc_conflict_mode": "manual_review",
    "oidc_exchange_retry": "2"
  }
}
```

说明：

- 支持“部分字段更新”；
- 服务端会先读取当前配置后合并再整体校验；
- 若字段不合法会返回 400 和具体校验信息。

