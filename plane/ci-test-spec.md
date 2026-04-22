# 前端 CI 测试命令与执行规范

## 命令规范

前端在 CI 中固定执行：

1. `npm ci`
2. `npm run test:unit`
3. `npm run build`

本地开发可用：

1. `npm install`
2. `npm run test:unit`
3. `npm run build`

---

## 通过标准

- `test:unit` 全部通过；
- `build` 必须成功；
- 单测文件放置在 `web/src/**/*.test.ts`；
- 新增路由守卫、主题逻辑、关键工具函数时必须补对应单测。

---

## 当前覆盖点

- `/settings` 兼容重定向判定测试：
  - `web/src/router/guard-utils.test.ts`
- 主题应用逻辑测试：
  - `web/src/util/theme.test.ts`

---

## 推荐接入（GitHub Actions 示例）

```yaml
name: web-ci
on:
  pull_request:
    paths:
      - 'web/**'
jobs:
  test-build:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: web
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: web/package-lock.json
      - run: npm ci
      - run: npm run test:unit
      - run: npm run build
```

