# Agent

`agent` 是部署在被监控服务器上的探针进程，负责向 Mashiro backend 周期性上报数据。

## 主要职责

- 采集 CPU、内存、磁盘、运行时间
- 采集累计上下行流量与实时速率
- 获取公网 IPv4 / IPv6、系统名称
- 拉取延迟任务并执行 `TCP`、`ICMP`、`HTTP` 探测
- 将系统数据和延迟结果回传到 backend

## 启动方式

```bash
cd agent
export MASHIRO_SERVER_URL="http://localhost:8080/api/agent/report"
export MASHIRO_AGENT_ID="server-auth-token"
go run .
```

## 关键环境变量

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `MASHIRO_SERVER_URL` | 否 | `http://localhost:8080/api/agent/report` | Agent 上报地址 |
| `MASHIRO_AGENT_ID` | 否 | `default-agent-1` | 服务器的认证 token |

## 与 backend 的交互

- 常规上报：`POST /api/agent/report`
- 拉取延迟任务：`GET /api/agent/latency/tasks`
- 回传延迟结果：`POST /api/agent/latency/results`

## 部署建议

- 推荐通过后台生成的一键部署命令安装
- Agent 应部署在真实被监控主机上，而不是和 backend / frontend 一起放进同一个部署 compose
- 如果放进容器内运行，默认采集到的是容器视角而不是宿主机视角
