# Sandbox

[English](README.md) | 简体中文

AI Agent Sandbox - 基于 Golang 实现的一体化的 AI Agent 沙箱执行环境

## 快速开始

### Docker 部署

```bash
mkdir /path/to/memory
mkdir /path/to/memory/workspaces

# 设置主机路径，用于持久化存储（工作目录、提示词等）
export LOCAL_MEMORY="/path/to/memory"
# Linux 下 Docker 容器用户（UID 1000）与宿主机用户不同，需修改目录权限以避免 Permission Denied
sudo chown -R 1000:1000 "/path/to/memory"

make docker-start
```

### 访问服务

- **IndexPage**: http://localhost:8080/
- **API**: http://localhost:8080/v1/
- **VNC**: http://localhost:8080/vnc/
- **Terminal**: http://localhost:8080/terminal/
- **MCP**: http://localhost:8080/mcp
- **Health**: http://localhost:8080/health

### 本地开发（无需 Docker）

在宿主机直接跑 `sandbox-server` 和 `mcp-hub`，两者 stdout/stderr 合并到同一个日志文件并自动 tail：

```bash
make run-all      # 编译并启动两个服务，自动 tail 合并日志 /tmp/sandbox.log
make stop-all     # 停止两个服务
make tail-log     # 重新 tail 合并日志
```

## 系统架构

<img alt="cover" src="https://upload-images.jianshu.io/upload_images/12321605-f74f67020c334759.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240">


**两种接入方式**：

- **MCP Hub**：内置 MCP 协议服务，提供标准化的 Tool 接口（bash、file、web 等），可直接对接 Claude、OpenAI 等支持 MCP 的 AI 模型；结合 [Eino](https://github.com/cloudwego/eino) 框架，几行代码即可实现完整的 ReAct Agent。

- **Sandbox SDK**：提供 Go SDK，通过编程方式调用 Bash、FileSystem、Grep、Web、JSONL 等服务。既可封装为 Tool 供 AI Agent 调用，也可直接用于产品功能开发（如文件浏览器、Git 管理、LSP 服务等）。

## 项目结构

```
sandbox/
├── cmd/
│   ├── sandbox-server/main.go    # 主服务入口
│   └── mcp-hub/main.go           # MCP Hub 入口
├── internal/
│   ├── api/
│   │   ├── router.go             # 路由配置
│   │   ├── middleware/           # JWT 鉴权、Logger、Context
│   │   └── handlers/             # Bash/File/Grep/JSONL/Terminal/Web/WS/Sandbox/Swagger
│   ├── services/
│   │   ├── bash/                 # Bash 命令执行
│   │   ├── filesystem/           # 文件系统操作 (manager, glob, grep, read, write, replacer, operations)
│   │   ├── jsonl/                # JSONL 追加 / 读取 / 计数
│   │   ├── terminal/             # 网页终端 (PTY)
│   │   └── web/                  # Web 服务 (fetch, search)
│   ├── mcp/
│   │   ├── registry.go           # MCP Tool 注册
│   │   ├── server.go             # MCP 协议服务
│   │   └── tools/                # MCP Tools 实现
│   └── config/config.go
├── types/
│   ├── consts/                   # 常量定义 (env, headers)
│   └── model/                    # 共享数据模型 (bash, file, grep, web, jsonl, response)
├── pkg/
│   ├── ctxutil/                  # Context 工具 (工作目录、会话)
│   ├── logger/                   # 纯文本日志封装
│   └── safe/                     # 安全工具函数
├── docker/
│   ├── Dockerfile
│   ├── nginx/nginx.conf          # Nginx 网关配置
│   ├── scripts/
│   │   ├── supervisord.conf      # 进程管理
│   │   └── entrypoint.sh
│   └── volumes/                  # 挂载卷示例
├── docs/
│   └── ENV.md                    # 环境变量说明
├── web/
│   ├── index.html                # Web 首页
│   └── terminal/index.html       # 网页终端前端 (xterm.js)
├── sdk/
│   └── go/                       # Go SDK (bash, file, grep, web, jsonl)
├── examples/
│   ├── filesystem/               # 文件系统示例
│   └── web/                      # Web 示例
├── docker-compose.yaml
├── Makefile
└── go.mod
```

## 核心服务

| 服务 | 端口 | 描述 |
|------|------|------|
| Nginx | 8080 | 统一入口，反向代理 |
| Sandbox Server | 8000 | HTTP API + WebSocket API |
| MCP Hub | 8001 | MCP 协议服务 |
| noVNC | 6080 | VNC Web 客户端 |
| VNC Server | 5900 | VNC 服务 |
| Chromium CDP | 9222 | Chrome DevTools Protocol（内部使用，供 agent-browser 调用） |

## 进程管理 (Supervisord)

容器内通过 Supervisord 管理以下 8 个进程，按优先级顺序启动：

| # | 程序名 | 命令 | 优先级 | 作用 |
|---|--------|------|--------|------|
| 1 | xvfb | `Xvfb :99 -screen 0 1280x1024x24` | 100 | 虚拟帧缓冲 X 服务器，创建虚拟显示器 `:99` |
| 2 | fluxbox | `fluxbox` | 200 | 轻量级窗口管理器 |
| 3 | x11vnc | `x11vnc -display :99 -forever -shared -rfbport 5900 -nopw` | 300 | VNC 服务器，共享虚拟显示器到 5900 端口 |
| 4 | websockify | `websockify --web=/usr/share/novnc 6080 localhost:5900` | 400 | WebSocket 代理，支持 noVNC 网页访问 |
| 5 | chromium | `chromium-browser --no-sandbox --remote-debugging-port=9222 ...` | 500 | Chromium 浏览器，启用 CDP 远程调试 |
| 6 | sandbox-server | `sandbox-server` | 600 | 沙箱主服务 (Bash/File/Grep/JSONL/Web API) |
| 7 | mcp-hub | `mcp-hub` | 700 | MCP Hub 服务 (MCP 协议 + Tool 接口) |
| 8 | nginx | `nginx -g "daemon off;"` | 800 | Nginx 反向代理网关 |

**启动顺序**: Xvfb → Fluxbox → x11vnc → websockify → Chromium → sandbox-server → mcp-hub → Nginx

**架构说明**:
- 程序 1-4 构成 **VNC 远程桌面** 栈，让用户可以通过浏览器访问虚拟桌面
- 程序 5 是 **Chromium 浏览器**，通过 9222 CDP 端口供内部（例如外置 agent-browser）调用
- 程序 6-7 是 **业务服务**，提供 API 和 MCP 协议
- 程序 8 是 **网关**，统一对外暴露服务

## API 端点

### 系统

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/docs` | GET | Swagger UI 文档 |
| `/v1/openapi.json` | GET | OpenAPI 规范 |
| `/v1/sandbox` | GET | 获取沙箱环境信息 |

### Bash

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/bash/exec` | POST | 执行 Bash 命令 |
| `/v1/bash/exec/stream` | POST | 流式执行 Bash 命令 |

### 文件系统

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/file/read` | POST | 读取文件 |
| `/v1/file/write` | POST | 写入文件 |
| `/v1/file/append` | POST | 追加写入文件 |
| `/v1/file/list` | POST | 列出目录 |
| `/v1/file/delete` | POST | 删除文件 |
| `/v1/file/move` | POST | 移动文件 |
| `/v1/file/copy` | POST | 复制文件 |
| `/v1/file/mkdir` | POST | 创建目录 |
| `/v1/file/exists` | GET | 检查文件是否存在 |
| `/v1/file/stat` | POST | 获取文件元信息 |
| `/v1/file/upload` | POST | 上传文件 (multipart) |
| `/v1/file/download` | GET | 下载文件 (浏览器可内联预览) |
| `/v1/grep/search` | POST | Grep 搜索文件内容 |

### JSONL

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/jsonl/count` | POST | 统计 JSONL 文件行数 |
| `/v1/jsonl/read` | POST | 按行读取 JSONL 文件 |
| `/v1/jsonl/append` | POST | 向 JSONL 文件追加一行或多行 |

### Web

| 端点 | 方法 | 描述 |
|------|------|------|
| `/v1/web/fetch` | POST | 抓取网页内容 |
| `/v1/web/search` | POST | 网页搜索 |

### WebSocket

| 端点 | 方法 | 描述 |
|------|------|------|
| `/mcp` | WebSocket | MCP 协议端点 |
| `/websockify` | WebSocket | VNC (noVNC) 远程桌面代理 |
| `/terminal/` | GET | 网页终端页面 |
| `/v1/terminal/ws` | WebSocket | 终端 WebSocket 连接 |
| `/v1/ws` | WebSocket | 通用 WebSocket 接口 |

## Go SDK 使用示例

```go
package main

import (
    "fmt"
    sandbox "github.com/deep-agent/sandbox/sdk/go"
    "github.com/deep-agent/sandbox/types/model"
)

func main() {
    client := sandbox.NewClient("http://localhost:8080")

    // 获取沙箱信息
    ctx, _ := client.GetContext()
    fmt.Printf("Workspace: %s\n", ctx.Workspace)

    // 执行 Bash 命令
    result, _ := client.BashExec(&model.BashExecRequest{
        Command: "ls -la",
    })
    fmt.Println(result.Output)

    // 读取文件
    content, _ := client.FileRead(&model.FileReadRequest{
        File: "/home/sandbox/.bashrc",
    })
    fmt.Println(content.Content)

    // 检查文件是否存在
    exists, _ := client.FileExists("/home/sandbox/.bashrc")
    fmt.Printf("File exists: %v\n", exists.Exists)

    // 向 JSONL 文件追加一行
    _ = client.JSONLAppendLine(&model.JSONLAppendRequest{
        File:       "/home/sandbox/workspaces/events.jsonl",
        JSONString: []string{`{"event":"hello"}`},
    })
}
```

## MCP Tools

MCP Hub 提供以下工具：

### 文件系统

| Tool | 描述 |
|------|------|
| `Bash` | 执行 Bash 命令 |
| `Glob` | 文件 glob 匹配 |
| `Grep` | 文件内容搜索 |
| `Read` | 读取文件内容 |
| `Write` | 写入文件内容 |
| `Edit` | 编辑文件 (搜索替换) |

### Web

| Tool | 描述 |
|------|------|
| `WebFetch` | 抓取网页内容 |
| `WebSearch` | 网页搜索 |

## 环境变量

### 容器运行时配置

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `SANDBOX_SRV_PORT` | 8000 | Sandbox Server 端口 |
| `MCP_HUB_PORT` | 8001 | MCP Hub 端口 |
| `SUPERVISOR_CONF_DIR` | /home/sandbox/app.supervisor.d | Supervisord 配置目录 |
| `APP_SERVICE_PORT` | 9000 | 用户 HTTP 服务端口（通过 `/app/` 访问） |
| `JWT_SECRET` | - | JWT HMAC 共享密钥 (可选) |
| `JWT_AUTH_REQUIRED` | false | 强制要求鉴权 (可选) |
| `ENABLE_MCP` | true | 是否启用 MCP Hub 进程 |
| `ENABLE_BROWSER` | true | 是否启用 Chromium 进程 |
| `ENABLE_VNC` | true | 是否启用 VNC/noVNC 进程 |
| `TZ` | Asia/Shanghai | 时区 |

### JWT 鉴权 (可选)

设置 `JWT_SECRET` 后，会对 `/v1` 路由启用 `Authorization: Bearer <token>` 校验（HS256/384/512）。

如果设置 `JWT_AUTH_REQUIRED=true`，即使未配置 `JWT_SECRET`，也会拒绝所有请求（返回 401），防止意外暴露未鉴权的服务。

示例：
```bash
export JWT_SECRET="your-secret-key"
export JWT_AUTH_REQUIRED="true"            # 可选，强制鉴权
```

### Docker Compose 变量

这些变量由宿主机侧的 `docker-compose` 消费。

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `HOST_PORT` | 8080 | 宿主机映射到容器 8080 的端口 |
| `LOCAL_MEMORY` | ./docker/volumes | 持久化存储基础路径；其 `workspaces` 子目录挂载到 `/home/sandbox/workspaces` |
| `LOCAL_SUPERVISOR_CONF` | ./docker/volumes/app.supervisor.d | 挂载到 `/home/sandbox/app.supervisor.d` 的宿主机路径 |
| `LOCAL_USERDATA` | ./docker/volumes/userdata | 挂载到 `/home/sandbox/userdata` 的宿主机路径 |
| `LOCAL_INIT_SCRIPTS` | ./docker/volumes/init.d | 挂载到 `/docker-entrypoint.d` 的宿主机路径 |

### Docker 构建参数 (可选)

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `HTTP_PROXY` | - | `docker build` 代理 |
| `HTTPS_PROXY` | - | `docker build` 代理 |
| `NO_PROXY` | localhost,127.0.0.1 | `docker build` 不走代理列表 |

## 扩展能力

Sandbox 支持用户注入自定义程序和脚本，提供以下扩展机制：

### 目录结构

```
docker/volumes/
├── workspaces/         # 用户工作目录，挂载到 /home/sandbox/workspaces
├── app.supervisor.d/   # Supervisord 配置目录，挂载到 /home/sandbox/app.supervisor.d
├── userdata/           # 用户数据目录，挂载到 /home/sandbox/userdata (脚本、二进制文件)
└── init.d/             # 初始化脚本目录，挂载到 /docker-entrypoint.d
```

### 1. Supervisord 配置 (app.supervisor.d)

将 Supervisord 配置文件放入 `docker/volumes/app.supervisor.d/` 目录，用于注册后台服务。

**自动注册到 Supervisord**：在 app.supervisor.d 目录下创建 `.conf` 文件：

```
docker/volumes/app.supervisor.d/
└── my-service.conf         # supervisord 配置
```

`my-service.conf` 示例：
```ini
[program:my-service]
command=/home/sandbox/userdata/my-service --port=9000
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
priority=900
```

### 2. 用户数据 (UserData)

将用户脚本或二进制文件放入 `docker/volumes/userdata/` 目录，容器启动后可在 `/home/sandbox/userdata/` 访问。

```
docker/volumes/userdata/
├── my-service              # 可执行程序
├── script.sh               # Shell 脚本
└── data.json               # 数据文件
```

### 3. 初始化脚本 (Init Scripts)

将脚本放入 `docker/volumes/init.d/` 目录，容器启动时会自动按字母顺序执行：

```
docker/volumes/init.d/
├── 01-setup-env.sh         # 环境配置
├── 02-install-deps.sh      # 安装依赖
└── 03-start-services.sh    # 启动服务
```

脚本示例：
```bash
#!/bin/bash
echo "Setting up environment..."
pip install requests numpy
npm install -g typescript
```

**执行规则**：
- `.sh` 文件：可执行则执行，否则 source
- 其他文件：可执行则执行

### 4. 用户 HTTP 服务代理

如果用户程序是 HTTP 服务，可以通过 `/app/` 路径访问。Nginx 会自动将请求转发到 `APP_SERVICE_PORT` 端口（默认 9000）。

```
用户请求: http://localhost:8080/app/api/hello
    ↓
Nginx 转发: http://127.0.0.1:9000/api/hello
    ↓
用户服务处理
```

**使用方式**：
1. 用户服务监听 `APP_SERVICE_PORT` 端口
2. 通过 `/app/` 路径访问

**自定义端口**：
```bash
APP_SERVICE_PORT=3000 docker-compose up
```

### 5. 基于 Sandbox 镜像扩展

用户可以基于 Sandbox 镜像创建自定义镜像：

```dockerfile
FROM your-registry/sandbox:latest

USER root
RUN apt-get update && apt-get install -y your-package

COPY --chown=sandbox:sandbox my-binary /home/sandbox/userdata/
COPY --chown=sandbox:sandbox my-init.sh /docker-entrypoint.d/

USER sandbox
```

### 环境变量配置

通过 docker-compose 或环境变量自定义挂载路径：

```bash
LOCAL_MEMORY=/path/to/memory \
LOCAL_SUPERVISOR_CONF=/path/to/app.supervisor.d \
LOCAL_USERDATA=/path/to/userdata \
LOCAL_INIT_SCRIPTS=/path/to/init.d \
APP_SERVICE_PORT=3000 \
docker-compose up
```

## License

Apache License 2.0
