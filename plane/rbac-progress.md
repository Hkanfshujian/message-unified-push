# RBAC 开发进度跟踪

## 总体说明

- 主计划文档：`plane/rbac.md`
- 当前目标：推进迭代六（设置域重构）

## 阶段状态

### 迭代一：数据层与基础能力

- [x] 新增 RBAC 核心模型（角色/权限/用户组/关系表）
- [x] 接入数据库自动迁移
- [x] 初始化 `super_admin` 角色
- [x] 初始化默认权限种子数据并绑定到 `super_admin`
- [x] 初始化 `default_group` 用户组并绑定 `super_admin`
- [x] 将 `admin` 自动绑定到 `super_admin` 和 `default_group`
- [x] 提供“当前用户权限聚合”服务
- [x] 提供 API：`GET /api/v1/rbac/me/permissions`

### 迭代二：后端接口强鉴权

- [x] 新增权限中间件 `RequirePermission`
- [x] 新增权限中间件 `RequireAnyPermission`
- [x] 核心路由挂载权限码（模板/渠道/定时消息/系统设置/统计/日志）
- [x] 增加鉴权失败审计日志
- [x] 增加灰度开关（`rbac.authz_mode`：monitor/enforce/disable）

### 迭代三：前端可见性控制

- [x] Pinia 存储权限码
- [x] 登录后拉取并缓存 `me/permissions`
- [x] 路由守卫权限校验
- [x] 菜单权限过滤
- [x] 按钮级 `v-permission` 指令
- [x] 在模板/渠道/定时消息关键操作接入按钮级权限控制

### 迭代四：管理后台与完善

- [x] 角色管理页面（列表/新增/编辑/删除/授权）
- [x] 用户组管理页面（列表/新增/编辑/删除/成员管理）
- [x] 权限管理页面（列表/新增/编辑）
- [x] 授权关系管理页面
- [x] 后端管理接口（角色/用户组/权限/授权关系）

### 迭代五：外部认证接入（Casdoor OIDC）

- [x] OIDC 登录与回调
- [x] 外部身份映射本地用户
- [x] 与本地 RBAC 联动
- [x] 单点登出与会话策略
- [x] 统一认证配置项初始化（`auth_config`）
- [x] OIDC 失败审计日志与告警指标接口
- [x] OIDC 身份映射管理页面（查看与解绑）
- [x] 单点登出回调链路与前端状态恢复
- [x] OIDC 回调异常分类与可重试策略
- [x] OIDC 身份映射冲突治理（人工审核流）
- [x] OIDC 审计看板与告警通知通道配置

### 迭代六：设置域重构（系统设置下沉 + 个人设置拆分）

- [x] 输出设置域重构规划文档
- [x] 系统设置下沉至系统管理菜单（`/system/settings`）
- [x] 个人设置独立路由与页面（`/profile/settings`）
- [x] 系统设置与个人设置功能拆分完成（密码迁移到个人设置）
- [x] 用户个性化主题色支持（全局默认 + 用户覆盖）
- [x] 兼容路由 `/settings` 重定向与回归验证
- [x] 认证配置子页接入系统设置（注册开关/OIDC策略）
- [x] 补齐个人设置权限模型并接入菜单/路由控制
- [x] 认证配置拆分为独立后端接口（`/api/v1/system/auth-config`）
- [x] 系统设置与个人设置子模块异步懒加载拆分
- [x] 新增 RBAC 权限聚合单测（个人设置内建权限）
- [x] 新增前端自动化测试（主题偏好与 `/settings` 兼容重定向）
- [x] 系统设置拆分为独立路由子页面（`/system/settings/*`）
- [x] 个人设置拆分为独立路由子页面（`/profile/settings/*`）
- [x] 认证配置接口文档与字段级权限矩阵说明
- [x] 前端 CI 测试命令与执行规范文档
- [x] 系统/个人设置子路由配置与守卫辅助单测
- [x] 前端 CI Workflow 文件落库（`.github/workflows/web-ci.yml`）

## 当前 API 草案

- `GET /api/v1/rbac/me/permissions`
- `GET /api/v1/rbac/roles`
- `POST /api/v1/rbac/roles`
- `POST /api/v1/rbac/roles/edit`
- `POST /api/v1/rbac/roles/delete`
- `GET /api/v1/rbac/roles/permissions`
- `POST /api/v1/rbac/roles/assign-permissions`
- `GET /api/v1/rbac/groups`
- `POST /api/v1/rbac/groups`
- `POST /api/v1/rbac/groups/edit`
- `POST /api/v1/rbac/groups/delete`
- `GET /api/v1/rbac/groups/roles`
- `GET /api/v1/rbac/groups/members`
- `POST /api/v1/rbac/groups/assign-roles`
- `POST /api/v1/rbac/groups/assign-members`
- `GET /api/v1/rbac/permissions`
- `POST /api/v1/rbac/permissions`
- `POST /api/v1/rbac/permissions/edit`
- `GET /api/v1/rbac/users`
- `GET /api/v1/rbac/users/role-ids`
- `GET /api/v1/rbac/users/group-ids`
- `POST /api/v1/rbac/users/assign-roles`
- `POST /api/v1/rbac/users/assign-groups`
- `GET /auth/oidc/login`
- `GET /auth/oidc/callback`
- `POST /auth/oidc/logout`
- `GET /auth/oidc/logout/callback`
- `GET /api/v1/oidc/metrics`
- `GET /api/v1/oidc/audits`
- `GET /api/v1/oidc/conflicts`
- `POST /api/v1/oidc/conflicts/approve`
- `POST /api/v1/oidc/conflicts/reject`
- `GET /api/v1/oidc/identities`
- `POST /api/v1/oidc/identities/unbind`
- `GET /api/v1/oidc/alert-config`
- `POST /api/v1/oidc/alert-config`
- `GET /api/v1/profile/theme`
- `POST /api/v1/profile/theme`
- `POST /api/v1/profile/password`
- `GET /api/v1/system/auth-config`
- `POST /api/v1/system/auth-config`

返回字段：

- `user_id`
- `username`
- `roles`
- `groups`
- `permissions`
- `is_super_admin`

## 下一个开发任务

1. 在 CI 中接入后端测试链路（`go test ./... -vet=off`）；
2. 增加系统/个人设置页面级 E2E 冒烟测试；
3. 评估修复 `go test ./...` 的历史 vet 问题并恢复全量通过。

规划文档：

- `plane/settings-refactor-plan.md`
- `plane/auth-config-api.md`
- `plane/ci-test-spec.md`
