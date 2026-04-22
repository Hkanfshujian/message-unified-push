# 设置域重构规划（系统设置下沉 + 个人设置拆分 + 主题个性化）

## 1. 背景与目标

当前“设置”入口位于一级菜单，且一个页面混合了系统级配置与个人级偏好，存在以下问题：

- 权限边界不清：普通用户与管理员可见内容耦合在同一路径；
- 信息架构不清：站点配置、日志清理、个人密码、登录日志混在一起；
- 个性化能力不足：主题颜色仅有全局配置，缺少用户粒度覆盖；
- 迭代成本高：后续新增设置项难以判断归属（系统/个人）。

本次重构目标：

1. 将“系统设置”下沉到“系统管理”域；
2. 拆分“系统设置”与“个人设置”两套页面与接口；
3. 建立用户个性化主题（颜色）能力，支持“全局默认 + 用户覆盖”；
4. 提供平滑迁移路径，保证线上无中断切换。

---

## 2. 现状梳理

## 2.1 前端现状

- 当前页面：`/settings`，组件 [Settings.vue](file:///f:/hukanfa/myself/myhub/ops-message-unified-push/web/src/components/pages/settings/Settings.vue)
- 当前侧边分组（混合）：重置密码、数据清理、登录日志、站点设置、加解密工具、站点关于  
  组件 [SettingsSidebar.vue](file:///f:/hukanfa/myself/myhub/ops-message-unified-push/web/src/components/pages/settings/SettingsSidebar.vue)

混合项建议归类：

- **个人设置**：重置密码、我的主题偏好；
- **系统设置**：站点设置、数据清理、登录日志、认证开关、OIDC、告警通道；
- **工具页**：加解密工具（可保留在系统管理或迁到“开发工具”分组）。

## 2.2 后端现状

- 设置存储以 `message_settings(section,key,value)` 为主，偏全局；
- 缺少用户级偏好存储模型；
- 注册开关已接入 `auth_config.register_enabled`。

---

## 3. 目标信息架构

## 3.1 菜单与路由

1. 系统管理（管理员）
   - 角色管理
   - 用户组管理
   - 权限管理
   - 用户管理
   - 授权关系
   - 身份映射
   - 系统设置（新）

2. 个人中心（登录用户）
   - 个人设置（新）
   - 可含：修改密码、主题偏好、会话信息

建议路由：

- `/system/settings`：系统设置
- `/profile/settings`：个人设置

保留兼容路由：

- `/settings` 临时重定向到 `/profile/settings`（普通用户）或 `/system/settings`（管理员）

## 3.2 权限模型

新增权限码建议：

- `system:settings:view`（已有）
- `system:settings:edit`（已有）
- `profile:settings:view`（新增）
- `profile:settings:edit`（新增）

---

## 4. 功能拆分方案

## 4.1 系统设置（管理员）

保留/迁入：

- 站点设置（标题、logo、默认主题色、cookie 过期）
- 日志清理策略
- 认证配置（含注册开关、OIDC、冲突策略、告警参数）
- 登录日志查看
- 关于信息（可选）

## 4.2 个人设置（用户）

迁入：

- 修改密码
- 主题偏好（颜色、深浅模式）

后续可扩展：

- 通知偏好（站内/邮件）
- 首页默认落点

## 4.3 主题个性化（重点）

策略采用“系统默认 + 用户覆盖”：

1. 系统默认色来源：`site_config.theme_color`
2. 用户覆盖来源：`user_preferences.theme_color`
3. 前端加载顺序：用户偏好 > 系统默认 > 内置默认

建议颜色集（先固定枚举）：

- `blue` / `green` / `purple` / `orange` / `red` / `zinc`

---

## 5. 数据与接口设计

## 5.1 数据模型

新增表建议：`message_user_preferences`

- `id`
- `user_id`（唯一索引）
- `theme_color`
- `theme_mode`（light/dark/system）
- `created_on/created_by/modified_on/modified_by`

说明：

- 不建议把用户偏好继续塞入全局 settings，避免“按用户查询”复杂化；
- 后续个人设置扩展时更易演进。

## 5.2 接口草案

系统设置：

- `GET /api/v1/system/settings`
- `POST /api/v1/system/settings`

个人设置：

- `GET /api/v1/profile/settings`
- `POST /api/v1/profile/settings`
- `POST /api/v1/profile/password`

主题偏好：

- `GET /api/v1/profile/theme`
- `POST /api/v1/profile/theme`

---

## 6. 实施阶段

## Phase 1：信息架构与路由拆分

- 新增 `/system/settings`、`/profile/settings`
- Sidebar 调整：系统设置移入系统管理
- `/settings` 做兼容重定向

## Phase 2：前后端接口拆分

- 现有 Settings 组件拆为 `SystemSettings` 与 `ProfileSettings`
- 后端按权限拆接口，避免一个接口承载双语义

## Phase 3：主题个性化

- 建立 `user_preferences`
- 登录后加载用户主题并落地到前端 store/localStorage
- 提供颜色选择 UI

## Phase 4：迁移与清理

- 清理旧 `/settings` 直连依赖
- 更新文档与测试用例

---

## 7. 风险与规避

- 风险：旧页面入口失效  
  规避：保留 `/settings` 兼容重定向至少一个版本周期；

- 风险：权限配置漏配导致页面不可见  
  规避：seed 增加新权限并自动绑定 `super_admin`；

- 风险：主题变量冲突  
  规避：统一主题变量定义，限制用户可选颜色集合。

---

## 8. 验收标准

1. 菜单中“系统设置”位于“系统管理”分组内；
2. 普通用户仅可访问“个人设置”；
3. 管理员可访问“系统设置”并配置注册开关/OIDC等；
4. 用户可设置个人主题色，刷新后保持；
5. `/settings` 兼容跳转生效，无 404。

