# Backend

`backend` 是 Mashiro 的 API 服务，基于 Gin、GORM 和 SQLite。

## 主要职责

- 提供登录、修改密码、服务器管理、系统配置等管理接口
- 提供首页仪表盘和服务器列表的公开接口
- 接收 Agent 的系统数据上报和延迟结果上报
- 管理延迟任务，并按服务器维度下发给 Agent
- 生成 Agent 一键部署脚本和源码包

## 启动方式

```bash
cd backend
export MASHIRO_JWT_SECRET="replace-me"
export MASHIRO_ADMIN_PASSWORD="replace-me"
go run .
```

## 关键环境变量

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `MASHIRO_PORT` | 否 | `8080` | backend 监听端口 |
| `MASHIRO_DB_PATH` | 否 | `mashiro.db` | SQLite 数据库文件路径 |
| `MASHIRO_JWT_SECRET` | 是 | 无 | JWT 签名密钥，未设置会直接启动失败 |
| `MASHIRO_ADMIN_USERNAME` | 否 | `admin` | 首次初始化管理员用户名 |
| `MASHIRO_ADMIN_PASSWORD` | 首次启动必填 | 无 | 首次初始化管理员密码 |

## 数据持久化

- 默认数据库文件位于当前工作目录下的 `mashiro.db`
- Docker 部署时建议挂载到 `/data/mashiro.db`

## 重要接口分组

- `/api/login`
- `/api/dashboard/*`
- `/api/admin/*`
- `/api/agent/report`
- `/api/agent/latency/tasks`
- `/api/agent/latency/results`
- `/api/agent/install.sh`
- `/api/agent/install.ps1`
- `/api/agent/package.zip`

## Docker 镜像

- Dockerfile: [Dockerfile](file:///g:/demo/test/Mashiro/backend/Dockerfile)
- 默认暴露端口：`8080`
- 运行时必须提供 `MASHIRO_JWT_SECRET`
