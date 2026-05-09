# Frontend

`frontend` 是 Mashiro 的 Web 界面，基于 React、Vite、TypeScript 和 Tailwind CSS。

## 主要职责

- 展示服务器仪表盘、流量、速率、在线状态和延迟结果
- 提供 grid / list 两种展示方式
- 提供管理后台：服务器管理、延迟任务管理、系统设置和登录

## 本地开发

```bash
cd frontend
pnpm install
pnpm dev
```

## 构建

```bash
cd frontend
pnpm build
```

## 关键环境变量

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `VITE_API_BASE_URL` | 否 | `/api` | 前端请求 API 的基础路径 |

## 开发代理

Vite 开发环境会将 `/api` 代理到 `http://localhost:8080`，配置见 [vite.config.ts](file:///g:/demo/test/Mashiro/frontend/vite.config.ts)。

## Docker 镜像

- Dockerfile: [Dockerfile](file:///g:/demo/test/Mashiro/frontend/Dockerfile)
- Nginx 配置: [nginx.conf](file:///g:/demo/test/Mashiro/docker/nginx.conf)
- 容器内由 Nginx 提供静态文件，并将 `/api` 反向代理到 `backend:8080`
