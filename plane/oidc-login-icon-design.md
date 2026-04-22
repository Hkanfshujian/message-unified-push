# OIDC 登录按钮图标设计方案（认证设置）

## 1. 目标

- 在「系统管理 → 系统设置 → 认证设置」中支持配置 OIDC 登录按钮图标。
- 支持两种存储后端：
  - 本地存储（默认）
  - S3 兼容对象存储（后续切换）
- 支持上传后自动裁剪，保证按钮图标统一视觉尺寸。

## 2. 配置模型（沿用 `auth_config`）

新增键：

- `oidc_login_icon_enabled`：`true/false`，是否启用图标
- `oidc_login_icon_url`：最终可访问 URL
- `oidc_login_icon_storage`：`local/s3`
- `oidc_login_icon_size`：目标尺寸，默认 `24`
- `oidc_login_icon_updated_at`：更新时间

按钮文案沿用已接入键：

- `oidc_login_button_text`

## 3. 存储设计

### 3.1 本地存储（默认）

- 物理目录：`./data/uploads/auth-icons/YYYY/MM/`
- 文件名：`oidc-login-{timestamp}-{rand}.png`
- 对外访问：
  - 新增静态路由前缀：`/public/uploads/*`
  - 生成 URL：`{apiUrl}{pathPrefix}/public/uploads/auth-icons/...`

优点：部署简单，不依赖外部云资源。

### 3.2 S3 存储（可切换）

- Bucket：由系统存储配置统一管理
- Key：`auth-icons/oidc-login/{yyyy}/{mm}/{uuid}.png`
- URL：
  - 私有桶：返回后端代理下载地址（推荐）
  - 公有桶：返回对象直链（可选）

建议在未来「系统设置-存储配置」提供全局存储驱动，认证图标仅消费该能力。

## 4. 上传与裁剪流程

1. 前端选择图片（jpg/png/webp，<= 2MB）。
2. 打开裁剪框，固定 1:1 比例。
3. 前端导出 128x128 PNG（基础裁剪）；后端再统一压缩为 24x24/32x32 资源。
4. 后端二次校验 MIME、尺寸、体积。
5. 后端保存文件到 local/s3，返回 `url`。
6. 前端回填 `oidc_login_icon_url` 并保存认证配置。

说明：前端裁剪提升交互，后端二次处理保证一致性与安全。

## 5. API 设计

### 5.1 上传图标

- `POST /api/v1/system/auth-config/oidc-icon/upload`
- 权限：`system:settings:edit`
- `multipart/form-data` 字段：`file`

返回：

```json
{
  "code": 200,
  "msg": "上传成功",
  "data": {
    "url": "/public/uploads/auth-icons/2026/03/oidc-login-xxx.png",
    "width": 24,
    "height": 24
  }
}
```

### 5.2 删除图标（可选）

- `POST /api/v1/system/auth-config/oidc-icon/delete`
- 权限：`system:settings:edit`

效果：清空 `oidc_login_icon_url`，并按策略删除物理文件/对象。

## 6. 前端交互（认证设置 + 登录页）

### 6.1 认证设置页

- 新增：
  - 图标启用开关
  - 上传按钮
  - 预览区
  - 重置按钮
  - 提示文案：推荐透明背景 PNG
- 保存时写入 `auth_config`。

### 6.2 登录页

- 若 `oidc_enabled=true` 且 `oidc_login_icon_enabled=true` 且有 `oidc_login_icon_url`：
  - 按钮左侧显示图标（24x24）
- 加载失败回退纯文本按钮，不阻断登录。

## 7. 安全与治理

- 严格校验扩展名与 MIME，拒绝 SVG（防脚本注入）。
- 限制最大上传体积（2MB）。
- 文件统一转码为 PNG，去除元数据。
- 防覆盖：随机文件名。
- 审计：记录操作人、IP、时间、旧值/新值。

## 8. 演进建议（你提到的本地 + S3）

- 第一步：先落地本地存储（最快可用）。
- 第二步：抽象 `StorageProvider` 接口（Local/S3）。
- 第三步：系统设置新增“存储配置”页，统一给头像、LOGO、认证图标复用。

## 9. 交付拆分（建议）

- Phase A（1 次迭代）：
  - 上传接口（local）
  - 认证设置上传与预览
  - 登录页展示图标
- Phase B：
  - 裁剪体验增强（缩放、拖拽）
  - 删除/替换历史资源清理策略
- Phase C：
  - S3 接入与存储配置中心

