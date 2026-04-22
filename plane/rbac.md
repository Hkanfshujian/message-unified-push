# Message Nest RBAC 鉴权体系落地方案（开发计划）

## 1. 背景与目标

当前系统已具备登录与 JWT 认证能力，但缺少完整的 RBAC（Role-Based Access Control）授权体系，主要表现为：

- 登录后“是否有权限访问某接口/页面”缺少统一判断标准
- 权限粒度无法覆盖“菜单可见性 + 按钮操作 + 接口调用”
- 缺少角色与权限的配置管理能力（后台不可维护）

本方案目标：

1. 建立标准 RBAC 模型：**用户-用户组-角色-权限**；
2. 完成后端接口级鉴权（强约束）；
3. 完成前端菜单/按钮级鉴权（体验层）；
4. 支持平滑迁移，兼容当前单管理员模式；
5. 可扩展到数据权限（按部门、按资源归属）场景。

---

## 2. 现状评估（基于当前代码）

已具备基础：

- JWT 登录认证与中间件机制
- Gin 路由分层、v1/v2 API 结构清晰
- GORM 模型能力完整，便于扩展角色/权限表

当前缺口：

- 数据模型中无 `role`、`permission`、`user_role` 等核心实体
- 路由无统一权限元数据（无法声明“这个接口需要什么权限”）
- 中间件只有“是否登录”校验，没有“是否有权限”校验
- 前端路由与页面操作按钮无统一权限码控制

---

## 3. RBAC 模型设计

### 3.1 核心实体

1. `sys_user`（已有 auth 表可演进）
2. `sys_user_group`（用户组）
3. `sys_role`
4. `sys_permission`
5. `sys_user_role`（用户-角色，多对多）
6. `sys_group_role`（用户组-角色，多对多）
7. `sys_user_group_member`（用户-用户组，多对多）
8. `sys_role_permission`（角色-权限，多对多）

### 3.2 权限粒度

权限码规范：

`{模块}:{资源}:{动作}`

示例：

- `system:user:view`
- `system:user:create`
- `message:template:view`
- `message:template:edit`
- `message:sendways:test`
- `message:sendways:delete`

### 3.3 权限类型

- `menu`：菜单权限（前端导航可见）
- `action`：页面操作权限（按钮级）
- `api`：接口权限（后端强校验）

建议 `sys_permission` 增加字段：

- `code`（唯一权限码）
- `name`
- `type`（menu/action/api）
- `method`（GET/POST/...，api 类型使用）
- `path`（接口路径，api 类型使用）
- `parent_id`（菜单树支持）
- `sort`
- `status`

### 3.4 生效权限规则（建议）

用户最终权限 = 用户直绑角色权限 ∪ 用户所在用户组角色权限

优先级建议：

1. `super_admin` 角色直接放行；
2. 角色取并集，不做“拒绝优先”；
3. 后续如需“显式拒绝”，新增 deny 规则表再扩展。

---

## 4. 数据库设计与迁移计划

## 4.1 新增表

1. `message_sys_role`
2. `message_sys_permission`
3. `message_sys_user_group`
4. `message_sys_user_role`
5. `message_sys_group_role`
6. `message_sys_user_group_member`
7. `message_sys_role_permission`

## 4.2 字段约束建议

- 唯一索引：
  - `sys_role.code`
  - `sys_permission.code`
  - `sys_user_group.code`
  - `sys_user_role(user_id, role_id)`
  - `sys_group_role(group_id, role_id)`
  - `sys_user_group_member(user_id, group_id)`
  - `sys_role_permission(role_id, permission_id)`
- 外键（可选，若保持轻量可仅逻辑约束）
- 通用审计字段：`created_on/created_by/modified_on/modified_by`

## 4.3 迁移策略

1. 首次迁移仅建表，不改变现有登录流程；
2. 初始化一个 `super_admin` 角色与全量权限；
3. 将现有管理员账号自动绑定 `super_admin`；
4. 鉴权中间件先“记录日志 + 灰度开关”，后强制执行。

---

## 5. 后端改造计划（Gin + GORM）

## 5.1 模型层（models）

新增模型文件：

- `models/rbac_role.go`
- `models/rbac_permission.go`
- `models/rbac_user_group.go`
- `models/rbac_user_role.go`
- `models/rbac_group_role.go`
- `models/rbac_user_group_member.go`
- `models/rbac_role_permission.go`

新增查询能力：

- 按用户查询角色列表
- 按用户查询用户组列表
- 按角色查询权限码集合
- 按用户聚合权限码集合（缓存友好）

## 5.2 服务层（service）

新增服务：

- `service/rbac_service/role_service.go`
- `service/rbac_service/group_service.go`
- `service/rbac_service/permission_service.go`
- `service/rbac_service/authz_service.go`

核心方法：

- `GetUserPermissionCodes(userID string) ([]string, error)`
- `UserHasPermission(userID, code string) bool`
- `AssignRolesToUser(userID string, roleIDs []string)`
- `AssignGroupsToUser(userID string, groupIDs []string)`
- `AssignRolesToGroup(groupID string, roleIDs []string)`
- `AssignPermissionsToRole(roleID string, permissionIDs []string)`

## 5.3 中间件层（middleware）

新增中间件：

1. `RequirePermission(code string)`：接口级权限拦截；
2. `RequireAnyPermission(codes ...string)`：支持任意一个权限即可；
3. 与现有 JWT 中间件串联：先鉴权身份，再鉴权权限。

返回码建议：

- `401` 未登录/Token 失效
- `403` 已登录但无权限

## 5.4 路由层（routers）

在路由注册处为关键接口声明权限码，例如：

- 模板管理：`message:template:view/edit/delete`
- 渠道管理：`message:sendways:view/add/edit/delete/test`
- 统计查询：`message:statistics:view`
- 系统设置：`system:settings:view/edit`

做法建议：

- 保持 v1/v2 路由结构不变；
- 在注册路由时增加权限中间件；
- 对公共接口（登录、健康检查）保留免鉴权。

---

## 6. 前端改造计划（Vue3 + Pinia）

## 6.1 权限数据流

登录后拉取当前用户权限码列表，存入 Pinia：

- `stores/auth.ts` 中维护：
  - `roles: string[]`
  - `permissions: string[]`

## 6.2 路由与菜单控制

- 在路由 meta 中声明 `requiredPermissions`
- 路由守卫（`router.beforeEach`）做页面级拦截
- 菜单渲染时按权限过滤

## 6.3 按钮级控制

提供统一能力：

1. 指令：`v-permission="'message:template:edit'"`
2. 组合式函数：`usePermission().can('xxx')`

应用范围：

- 模板管理页“新增/编辑/删除/实例”等按钮
- 渠道管理页“测试/删除”等操作
- 设置页敏感配置入口

## 6.4 接口兜底

即使前端已隐藏按钮，后端仍必须做强校验（防抓包绕过）。

---

## 7. API 规划（RBAC 管理端）

建议新增管理接口：

- 角色管理
  - `GET /api/v1/rbac/roles`
  - `POST /api/v1/rbac/roles`
  - `POST /api/v1/rbac/roles/edit`
  - `POST /api/v1/rbac/roles/delete`
- 权限管理
  - `GET /api/v1/rbac/permissions`
  - `POST /api/v1/rbac/permissions`
  - `POST /api/v1/rbac/permissions/edit`
- 授权关系
  - `POST /api/v1/rbac/user/assign-roles`
  - `POST /api/v1/rbac/user/assign-groups`
  - `POST /api/v1/rbac/groups/assign-roles`
  - `POST /api/v1/rbac/role/assign-permissions`
- 用户组管理
  - `GET /api/v1/rbac/groups`
  - `POST /api/v1/rbac/groups`
  - `POST /api/v1/rbac/groups/edit`
  - `POST /api/v1/rbac/groups/delete`
- 当前用户权限
  - `GET /api/v1/rbac/me/permissions`

---

## 8. 缓存与性能方案

## 8.1 建议缓存对象

- `user_permissions:{userID}` -> `[]permissionCode`
- `role_permissions:{roleID}` -> `[]permissionCode`
- `group_roles:{groupID}` -> `[]roleID`

## 8.2 失效策略

- 用户角色变更：清理对应用户缓存
- 用户组成员变更：清理对应用户缓存
- 用户组角色变更：清理该组下所有用户缓存
- 角色权限变更：清理该角色下所有用户缓存
- 权限元数据变更：可全量清理权限缓存

可先使用内存缓存，后续支持 Redis（项目已有 Redis 依赖可演进）。

---

## 9. 安全设计要点

1. 任何业务接口都以**后端权限校验为准**；
2. 权限码不可由前端自由传入决定；
3. 严格区分 `401` 与 `403`，便于审计与排障；
4. 记录鉴权失败日志（用户ID、接口、权限码、来源IP）；
5. 高危操作可叠加二次确认（如删除渠道、删除模板）。

---

## 10. 分阶段实施计划（建议 5 个迭代）

## 迭代一：数据层与基础能力

- 新增 RBAC 数据表与模型
- 初始化超级管理员角色与权限种子数据
- 提供“查询当前用户权限”接口
- 用户组与组成员/组角色关系最小实现

交付物：

- 可运行迁移
- 可查询权限码列表

## 迭代二：后端接口强鉴权

- 增加权限中间件
- 给核心业务路由声明权限码
- 支持灰度开关（仅记录/强拦截）

交付物：

- 关键接口受控
- 鉴权日志可追踪

## 迭代三：前端可见性控制

- Pinia 权限状态管理
- 菜单过滤与路由守卫
- 按钮级 `v-permission` 指令

交付物：

- UI 层权限一致性
- 用户只看到有权限的功能入口

## 迭代四：管理后台与完善

- 角色管理、权限管理、授权页面
- 用户组管理、组成员管理、组角色授权页面
- 批量授权能力
- 自动化测试与文档补全

交付物：

- 完整可配置 RBAC 管理能力

## 迭代五：外部认证接入（Casdoor OIDC）

- 接入 Casdoor 作为统一认证中心（OIDC Authorization Code + PKCE）
- 完成本地用户映射（sub/email）与首次登录自动建档策略
- 完成外部身份与本地用户组/角色绑定策略
- 保持“认证外置、授权本地”（RBAC 仍由本系统执行）
- 增加登录回调、登出回调与会话过期处理

交付物：

- 可用的 Casdoor 单点登录能力
- 与本地 RBAC 联动的身份映射机制
- 外部认证接入实施文档（见 `oidc-casdoor.md`）

---

## 11. 测试计划

### 11.1 后端测试

- 中间件单测：401/403 场景
- 权限聚合单测：用户-角色-权限链路
- 路由集成测试：有权/无权访问行为

### 11.2 前端测试

- 路由守卫行为测试
- 指令 `v-permission` 显隐测试
- 菜单渲染测试

### 11.3 回归测试

- 登录、模板发送、渠道管理、定时任务等主流程回归
- 超级管理员兼容性验证

---

## 12. 风险与回滚策略

主要风险：

- 权限码与接口映射遗漏导致误拦截
- 旧账号无角色绑定导致全量 403

控制策略：

1. 默认赋予管理员 `super_admin`
2. 鉴权开关支持快速降级到“仅记录不拦截”
3. 分模块灰度启用（先模板管理，再渠道管理，再系统设置）

---

## 13. 与当前项目的集成建议（落地路径）

建议优先落在以下目录：

- `models/`：新增 RBAC 相关模型
- `service/rbac_service/`：权限聚合与授权服务
- `middleware/`：新增权限中间件
- `routers/`：给路由挂载权限码
- `web/src/stores/`：权限状态管理
- `web/src/directives/`：`v-permission` 指令

兼容当前现状：

- 不改变现有业务 API 路径结构
- 在现有 JWT 认证基础上增量扩展授权
- 先补齐核心模块权限，再全量覆盖

---

## 14. 里程碑验收标准（DoD）

1. 已登录用户访问无权限接口返回 403；
2. 菜单与按钮显隐与后端权限一致；
3. 超级管理员可访问全部功能；
4. 普通角色仅可访问授权模块；
5. 关键业务链路（模板、渠道、定时任务）回归通过；
6. 有完整 RBAC 操作审计日志。

---

## 15. 后续增强方向

- 数据权限（按组织、项目、资源所有者）
- 权限表达式（ABAC 混合模型）
- 多租户隔离
- 审计报表与异常告警联动
