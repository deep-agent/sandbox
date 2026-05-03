# 环境变量说明

本文档梳理 `sandbox` 项目中所有的环境变量,并按 **程序依赖**、**Docker 依赖**、**共享变量** 三类进行说明。共计 **23 个** 唯一变量。

---

## 一、程序依赖(Go 代码直接读取)

这些变量由 Go 源码通过 `os.Getenv` 读取,即使脱离 Docker 直接运行二进制也需要关心。

### 1.1 主程序(`internal/config/config.go`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SANDBOX_SRV_PORT` | `8000` | 主 HTTP 服务端口 |
| `MCP_HUB_PORT` | `8001` | MCP Hub 服务端口 |

### 1.2 鉴权中间件(`internal/api/middleware/`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `JWT_SECRET` | 空 | JWT 签名密钥(HS256/384/512 HMAC)。为空时 JWT 鉴权自动关闭 |
| `JWT_AUTH_REQUIRED` | `false` | 设为 `true` 时,若 `JWT_SECRET` 未配置则拒绝所有请求(防止裸奔) |


## 二、Docker 依赖(仅构建/容器编排使用)

这些变量不会被 Go 代码读取,仅在 Dockerfile 构建阶段、docker-compose 编排或 entrypoint 脚本中生效。

### 2.1 构建期参数(`docker/Dockerfile` 的 `ARG`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `TARGETOS` | 由 buildx 自动注入 | 目标操作系统(多架构构建用) |
| `TARGETARCH` | 由 buildx 自动注入 | 目标 CPU 架构 |
| `HTTP_PROXY` | 空 | 构建时下载依赖使用的 HTTP 代理 |
| `HTTPS_PROXY` | 空 | 构建时下载依赖使用的 HTTPS 代理 |
| `NO_PROXY` | `localhost,127.0.0.1` | 不走代理的主机列表 |
| `NODE_VERSION` | `24.13.0` | 镜像内 Node.js 版本 |
| `DEBIAN_FRONTEND` | `noninteractive` | apt 静默安装模式 |

### 2.2 容器运行期(`docker/Dockerfile` 的 `ENV`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `TZ` | `Asia/Shanghai` | 容器时区 |
| `HOME` | `/home/sandbox` | 沙箱用户家目录 |
| `PATH` | `/home/sandbox/.local/bin:${PATH}` | 可执行文件搜索路径 |
| `DISPLAY` | `:99` | X11 显示编号(Xvfb 使用) |

### 2.3 Compose 编排(`docker-compose.yaml`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HOST_PORT` | `8080` | 宿主机映射到容器 8080 的端口 |
| `LOCAL_MEMORY` | `./docker/volumes` | 宿主机挂载基础目录(workspaces/agent/bin) |
| `LOCAL_SUPERVISOR_CONF` | `./docker/volumes/app.supervisor.d` | Supervisor 配置目录(宿主机侧) |
| `LOCAL_USERDATA` | `./docker/volumes/userdata` | 用户数据目录(宿主机侧) |
| `LOCAL_INIT_SCRIPTS` | `./docker/volumes/init.d` | 容器启动初始化脚本目录 |

### 2.4 Entrypoint 脚本(`docker/scripts/entrypoint.sh`)

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SUPERVISOR_CONF_DIR` | `/home/sandbox/app.supervisor.d` | 容器内 Supervisor 配置目录 |
| `APP_SERVICE_PORT` | `9000` | 用户 app 服务端口(由 nginx 反代) |

---

## 三、共享变量(程序和 Docker 都会读)

这三个开关由 `docker-compose.yaml` 注入,`entrypoint.sh` 根据值拼装 supervisor 子进程配置,决定容器内是否启动对应子服务:

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `ENABLE_MCP` | `true` | 是否启动 MCP Hub |
| `ENABLE_BROWSER` | `true` | 是否启动 Chromium |
| `ENABLE_VNC` | `true` | 是否启动 Xvfb + x11vnc + websockify(浏览器访问 noVNC) |

---

## 四、快速参考:典型部署场景

### 4.1 本地直接跑 Go 二进制(不用 Docker)

只需关心 **1.1 / 1.2** 两节:

```bash
export SANDBOX_SRV_PORT=8000
export JWT_SECRET=your-shared-secret   # 需要鉴权时
./sandbox
```

### 4.2 Docker Compose 部署

复制 `.env.example` 为 `.env`,通常只需修改:

- `HOST_PORT`:对外端口
- `JWT_SECRET` / `JWT_AUTH_REQUIRED`:开启鉴权
- `ENABLE_MCP` / `ENABLE_BROWSER` / `ENABLE_VNC`:按需关闭以节省资源

### 4.3 构建镜像到私有仓库

通过 Makefile 构建时,可注入:

```bash
HTTP_PROXY=http://proxy:8080 HTTPS_PROXY=http://proxy:8080 make build
```

---

## 五、已知问题与建议

1. **`LOCAL_MEMORY` 是宿主机挂载路径**（compose 使用），对应 `${LOCAL_MEMORY}/workspaces` 目录，通过 `docker-compose.yaml` 的 volume 映射到容器内 `/home/sandbox/workspaces`。
2. **JWT 默认关闭**：`.env.example` 中 `JWT_SECRET` / `JWT_AUTH_REQUIRED` 均留空，部署方自行决定是否开启。强制开启方式：设置 `JWT_SECRET=<密钥>` + `JWT_AUTH_REQUIRED=true`
3. **示例程序变量已隔离**:`ARK_API_KEY` / `ARK_MODEL` / `MCP_URL` 仅在 `examples/web/` 下使用,配置模板位于 `examples/web/.env.example`,不要混入根目录 `.env`
4. **部分内部端口是硬编码的,环境变量改不动**:以下端口在配置文件中直接写死,修改 env 不生效,需要同步改配置文件后重建镜像:

   | 端口 | 用途 | 硬编码位置 |
   |------|------|-----------|
   | `8000` | Sandbox Server(`SANDBOX_SRV_PORT` 默认值) | `docker/nginx/nginx.conf`(多处 `proxy_pass`) |
   | `8001` | MCP Hub(`MCP_HUB_PORT` 默认值) | `docker/nginx/nginx.conf:46` |
   | `5900` | x11vnc 监听端口 | `docker/scripts/supervisord.conf:26` (`-rfbport 5900`) |
   | `6080` | websockify / noVNC 代理端口 | `docker/nginx/nginx.conf:65` + `supervisord.conf:34` |
   | `9222` | Chromium 远程调试端口 | `docker/scripts/supervisord.conf:42` (`--remote-debugging-port=9222`) |

   对外只暴露 `HOST_PORT`(默认 `8080`,由 nginx 统一反代),上述内部端口理论上不需要改。如果确实要改,记得三个地方一起动:`nginx.conf` / `supervisord.conf` / Go 默认值(如有)。
