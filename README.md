# Mashiro

Mashiro 是一个轻量级自托管服务器监控项目，包含 `backend`、`frontend` 和 `agent` 三个部分：

- `backend`：Gin + GORM + SQLite，负责管理端 API、公开仪表盘 API、Agent 上报、延迟任务和一键部署脚本。
- `frontend`：React + Vite，负责首页仪表盘和管理后台。
- `agent`：Go 编写的轻量探针，负责系统采集、网络采集、延迟任务执行和结果回传。

## 仓库结构

```text
.
├─ backend/                # Go API 服务
├─ frontend/               # React 前端
├─ agent/                  # Go 探针
├─ docker-compose.yml      # 部署编排
├─ .env.example            # 部署环境变量示例
└─ .github/workflows/      # GitHub Actions
```

## 本地开发

### 1. 启动 backend

```bash
cd backend
export MASHIRO_JWT_SECRET="replace-me"
export MASHIRO_ADMIN_PASSWORD="replace-me"
go run .
```

默认监听 `:8080`。

### 2. 启动 frontend

```bash
cd frontend
pnpm install
pnpm dev
```

开发模式下，Vite 会将 `/api` 代理到 `http://localhost:8080`。

### 3. 启动 agent

```bash
cd agent
export MASHIRO_SERVER_URL="http://localhost:8080/api/agent/report"
export MASHIRO_AGENT_ID="your-server-token"
go run .
```

## 组件文档

- [backend 文档](file:///g:/demo/test/Mashiro/backend/README.md)
- [frontend 文档](file:///g:/demo/test/Mashiro/frontend/README.md)
- [agent 文档](file:///g:/demo/test/Mashiro/agent/README.md)

## Docker 部署

### 镜像发布

仓库包含 GitHub Actions 工作流 [docker-release.yml](file:///g:/demo/test/Mashiro/.github/workflows/docker-release.yml)：

- 触发条件：push 任意 Git tag
- 推送目标：`ghcr.io/<owner>/<repo>-backend:<tag>` 和 `ghcr.io/<owner>/<repo>-frontend:<tag>`
- 额外标签：同一次构建会附带一个短 SHA 标签

### docker-compose

1. 复制环境变量模板：

```bash
cp .env.example .env
```

2. 修改 `.env` 中的镜像名、版本、JWT 密钥和初始化管理员密码。

3. 启动：

```bash
docker compose up -d
```

4. 访问：

- 前端：`http://<your-host>/`
- 后端 API：由前端容器通过 Nginx 反向代理到 `backend:8080`

## 初始化说明

- 第一次启动时，后端会根据 `MASHIRO_ADMIN_USERNAME` 和 `MASHIRO_ADMIN_PASSWORD` 初始化管理员账号。
- 如果数据库里已经存在用户，后续重启不会重新覆盖账号。
- Agent 不包含在 `docker-compose.yml` 中，应该部署到被监控的目标主机上。
