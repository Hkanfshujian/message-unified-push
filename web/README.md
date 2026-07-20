# Ops Message Unified Push Web

消息统一推送中台前端控制台，基于 Vue 3、TypeScript、Vite、Element Plus、UnoCSS、Pinia、ECharts 和 WangEditor 构建。

## 运行环境

- Node.js：`>=20.19.0`
- npm：随 Node.js 安装版本即可
- Sass：使用 Dart Sass `sass`，禁止使用 `node-sass`

安装依赖时会自动执行 Node.js 版本检查；版本不满足时会终止安装并提示升级。

## 常用命令

```sh
npm install
npm run dev
npm run typecheck
npm run test:unit
npm run check:deps
npm run build
npm run preview
```

## 依赖合规检查

`npm run check:deps` 会校验前端依赖是否满足重构约束，包括 Vue、TypeScript、Vite、Element Plus、UnoCSS、Dart Sass、Pinia、ECharts、@antv/g2、@wangeditor/editor 的版本声明，并确保 `node-sass` 不存在。

## 开发说明

- 开发服务使用 `npm run dev`，Vite 会启用热更新，常规前端修改不需要重启后端。
- 生产构建使用 `npm run build`，会先执行 TypeScript 类型检查，再执行 Vite 构建。
- 生产预览使用 `npm run preview`，用于验证静态资源、路由和部署路径行为。
