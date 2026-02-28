# Sandbox

English | [简体中文](README_CN.md)

AI Agent Sandbox - An sandbox execution environment for AI Agents, built with Golang

## Architecture

<img alt="cover" src="https://upload-images.jianshu.io/upload_images/12321605-f74f67020c334759.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240">

**Two Integration Methods**:

- **MCP Hub**: Built-in MCP protocol service providing standardized Tool interfaces (bash, file, browser, etc.). Directly integrates with AI models supporting MCP like Claude and OpenAI. Combined with the [Eino](https://github.com/cloudwego/eino) framework, you can implement a complete ReAct Agent with just a few lines of code.

- **Sandbox SDK**: Provides a Go SDK for programmatic access to Bash, FileSystem, Browser, and other services. Can be wrapped as Tools for AI Agent calls or used directly for product development (file browser, Git management, LSP services, etc.).

## Project Structure

```
sandbox/
├── cmd/
│   ├── sandbox-server/main.go    # Main server entry
│   └── mcp-hub/main.go           # MCP Hub entry
├── internal/
│   ├── api/
│   │   ├── router.go             # Router configuration
│   │   ├── middleware/           # JWT auth, Logger
│   │   └── handlers/             # Bash/File/Browser/Terminal/Web API
│   ├── services/
│   │   ├── bash/                 # Bash command execution
│   │   ├── filesystem/           # Filesystem operations (manager, glob, grep, read, write, replacer, operations)
│   │   ├── browser/              # Browser control (CDP)
│   │   ├── terminal/             # Web terminal (PTY)
│   │   └── web/                  # Web services (fetch, search)
│   ├── mcp/
│   │   ├── registry.go           # MCP Tool registration
│   │   ├── server.go             # MCP protocol server
│   │   └── tools/                # MCP Tools implementation
│   └── config/config.go
├── types/
│   ├── consts/                   # Constants (env, headers)
│   └── model/                    # Shared data models (bash, browser, file, grep, web, response)
├── pkg/
│   ├── ctxutil/                  # Context utilities (workspace path, session)
│   └── safe/                     # Safety utility functions
├── docker/
│   ├── Dockerfile
│   ├── nginx/nginx.conf          # Nginx gateway config
│   ├── scripts/
│   │   ├── supervisord.conf      # Process management
│   │   └── entrypoint.sh
│   └── volumes/                  # Volume mount examples
├── docs/
│   ├── tools.json                # MCP Tools documentation
│   └── web_tools.json            # Web Tools documentation
├── web/
│   ├── index.html                # Web homepage
│   └── terminal/index.html       # Web terminal frontend (xterm.js)
├── sdk/
│   └── go/                       # Go SDK (sandbox, interface, bash, file, browser, grep)
├── examples/
│   ├── cdp/                      # CDP examples
│   ├── filesystem/               # Filesystem examples
│   └── web/                      # Web examples
├── docker-compose.yaml
├── Makefile
└── go.mod
```

## Core Services

| Service | Port | Description |
|---------|------|-------------|
| Nginx | 8080 | Unified entry, reverse proxy |
| Sandbox Server | 8000 | HTTP API + WebSocket API |
| MCP Hub | 8001 | MCP protocol service |
| noVNC | 6080 | VNC Web client |
| VNC Server | 5900 | VNC service |
| Chromium CDP | 9222 | Chrome DevTools Protocol |

## Process Management (Supervisord)

The container manages 8 processes via Supervisord, started in priority order:

| # | Program | Command | Priority | Purpose |
|---|---------|---------|----------|---------|
| 1 | xvfb | `Xvfb :99 -screen 0 1280x1024x24` | 100 | Virtual framebuffer X server, creates virtual display `:99` |
| 2 | fluxbox | `fluxbox` | 200 | Lightweight window manager |
| 3 | x11vnc | `x11vnc -display :99 -forever -shared -rfbport 5900 -nopw` | 300 | VNC server, shares virtual display to port 5900 |
| 4 | websockify | `websockify --web=/usr/share/novnc 6080 localhost:5900` | 400 | WebSocket proxy, enables noVNC web access |
| 5 | chromium | `chromium-browser --no-sandbox --remote-debugging-port=9222 ...` | 500 | Chromium browser with CDP remote debugging |
| 6 | sandbox-server | `sandbox-server` | 600 | Sandbox main service (Bash/File/Browser API) |
| 7 | mcp-hub | `mcp-hub` | 700 | MCP Hub service (MCP protocol + Tool interfaces) |
| 8 | nginx | `nginx -g "daemon off;"` | 800 | Nginx reverse proxy gateway |

**Startup Order**: Xvfb → Fluxbox → x11vnc → websockify → Chromium → sandbox-server → mcp-hub → Nginx

**Architecture Notes**:
- Programs 1-4 form the **VNC Remote Desktop** stack, allowing browser access to virtual desktop
- Program 5 is the **Chromium Browser**, supporting CDP protocol for automation
- Programs 6-7 are **Business Services**, providing API and MCP protocol
- Program 8 is the **Gateway**, unified external service exposure

## Quick Start

### Docker Deployment

```bash
# Build and push multi-platform image
make docker-build

# Start with docker-compose
docker-compose up -d

# Rebuild and start
make docker-rebuild

# Reload (stop and restart)
make docker-reload
```

### Access Services

- **IndexPage**: http://localhost:8080/
- **API**: http://localhost:8080/v1/
- **VNC**: http://localhost:8080/vnc/
- **Terminal**: http://localhost:8080/terminal/
- **MCP**: http://localhost:8080/mcp
- **Health**: http://localhost:8080/health

## API Endpoints

### System

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/docs` | GET | Swagger UI documentation |
| `/v1/openapi.json` | GET | OpenAPI specification |
| `/v1/sandbox` | GET | Get sandbox environment info |

### Bash

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/bash/exec` | POST | Execute Bash command |
| `/v1/bash/exec/stream` | POST | Stream execute Bash command |

### Filesystem

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/file/read` | POST | Read file |
| `/v1/file/write` | POST | Write file |
| `/v1/file/list` | POST | List directory |
| `/v1/file/delete` | POST | Delete file |
| `/v1/file/move` | POST | Move file |
| `/v1/file/copy` | POST | Copy file |
| `/v1/file/mkdir` | POST | Create directory |
| `/v1/file/exists` | GET | Check if file exists |
| `/v1/grep/search` | POST | Grep search file content |

### Browser

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/browser/info` | GET | Get browser info (CDP URL) |
| `/v1/browser/navigate` | POST | Navigate to URL |
| `/v1/browser/screenshot` | POST | Browser screenshot |
| `/v1/browser/click` | POST | Click element |
| `/v1/browser/type` | POST | Type text |
| `/v1/browser/evaluate` | POST | Execute JavaScript |
| `/v1/browser/url` | GET | Get current URL |
| `/v1/browser/title` | GET | Get page title |
| `/v1/browser/scroll` | POST | Scroll page |
| `/v1/browser/html` | POST | Get element HTML |
| `/v1/browser/wait` | POST | Wait for element visible |
| `/v1/browser/page` | GET | Get page info |
| `/v1/browser/pdf` | POST | Export PDF |

### Web

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/web/fetch` | POST | Fetch web content |
| `/v1/web/search` | POST | Web search |

### WebSocket

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/mcp` | WebSocket | MCP protocol endpoint |
| `/vnc/` | WebSocket | VNC remote desktop |
| `/terminal/` | GET | Web terminal |
| `/v1/terminal/ws` | WebSocket | Terminal WebSocket connection |
| `/v1/ws` | WebSocket | General WebSocket interface |

## Go SDK Usage Example

```go
package main

import (
    "fmt"
    sandbox "github.com/deep-agent/sandbox/sdk/go"
    "github.com/deep-agent/sandbox/types/model"
)

func main() {
    client := sandbox.NewClient("http://localhost:8080")

    // Get sandbox info
    ctx, _ := client.GetContext()
    fmt.Printf("Workspace: %s\n", ctx.Workspace)

    // Execute Bash command
    result, _ := client.BashExec(&model.BashExecRequest{
        Command: "ls -la",
    })
    fmt.Println(result.Output)

    // Read file
    content, _ := client.FileRead(&model.FileReadRequest{
        File: "/home/sandbox/.bashrc",
    })
    fmt.Println(content.Content)

    // Browser screenshot
    screenshot, _ := client.BrowserScreenshot(&model.BrowserScreenshotRequest{
        Full: true,
    })
    fmt.Println("Screenshot (base64):", screenshot.Screenshot[:100])

    // Get browser info
    info, _ := client.BrowserGetInfo()
    fmt.Printf("CDP URL: %s\n", info.CDPURL)

    // Check if file exists
    exists, _ := client.FileExists("/home/sandbox/.bashrc")
    fmt.Printf("File exists: %v\n", exists.Exists)
}
```

## MCP Tools

MCP Hub provides the following tools:

### Filesystem

| Tool | Description |
|------|-------------|
| `bash` | Execute Bash command |
| `glob` | File glob matching |
| `grep` | File content search |
| `read` | Read file content |
| `write` | Write file content |
| `edit` | Edit file (search & replace) |

### Browser

| Tool | Description |
|------|-------------|
| `browser_navigate` | Navigate to URL |
| `browser_screenshot` | Browser screenshot |
| `browser_click` | Click element |
| `browser_type` | Type text |
| `browser_get_url` | Get current URL |
| `browser_get_title` | Get page title |
| `browser_get_html` | Get element HTML |
| `browser_evaluate` | Execute JavaScript |
| `browser_scroll` | Scroll page |
| `browser_wait_visible` | Wait for element visible |
| `browser_get_page_info` | Get page info |
| `browser_pdf` | Export PDF |

### Web

| Tool | Description |
|------|-------------|
| `web_fetch` | Fetch web content |
| `web_search` | Web search |

## Environment Variables

### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SANDBOX_SRV_PORT` | 8000 | Sandbox Server port |
| `MCP_HUB_PORT` | 8001 | MCP Hub port |
| `BROWSER_REMOTE_DEBUGGING_PORT` | 9222 | Chrome CDP port |
| `VNC_SERVER_PORT` | 5900 | VNC service port |
| `WEBSOCKET_PROXY_PORT` | 6080 | WebSocket proxy port (noVNC) |
| `WORKSPACE` | $HOME | Working directory |
| `JWT_SECRET` | - | JWT HMAC shared secret (optional) |
| `JWT_AUTH_REQUIRED` | false | Enforce authentication (optional) |
| `TZ` | Asia/Shanghai | Timezone |

### JWT Authentication (Optional)

Setting `JWT_SECRET` enables `Authorization: Bearer <token>` validation (HS256/384/512) for `/v1` routes.

If `JWT_AUTH_REQUIRED=true` is set, all requests will be rejected (returning 401) even if `JWT_SECRET` is not configured, preventing accidental exposure of unauthenticated services.

Example:
```bash
export JWT_SECRET="your-secret-key"
export JWT_AUTH_REQUIRED="true"            # Optional, enforce auth
```

### Docker Extended Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SUPERVISOR_CONF_DIR` | /home/sandbox/app.supervisor.d | Supervisord config directory |
| `USERDATA_DIR` | /home/sandbox/userdata | User data directory (scripts, binaries) |
| `APP_SERVICE_PORT` | 9000 | User HTTP service port |

## Extensibility

Sandbox supports user-injected custom programs and scripts through the following extension mechanisms:

### Directory Structure

```
docker/volumes/
├── workspace/          # User workspace, mounted to /home/sandbox/workspace
├── app.supervisor.d/   # Supervisord config dir, mounted to /home/sandbox/app.supervisor.d
├── userdata/           # User data dir, mounted to /home/sandbox/userdata (scripts, binaries)
└── init.d/             # Init scripts dir, mounted to /docker-entrypoint.d
```

### 1. Supervisord Configuration (app.supervisor.d)

Place Supervisord configuration files in `docker/volumes/app.supervisor.d/` directory to register background services.

**Auto-register with Supervisord**: Create `.conf` files in the app.supervisor.d directory:

```
docker/volumes/app.supervisor.d/
└── my-service.conf         # supervisord config
```

`my-service.conf` example:
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

### 2. User Data (UserData)

Place user scripts or binaries in `docker/volumes/userdata/` directory, accessible at `/home/sandbox/userdata/` after container startup.

```
docker/volumes/userdata/
├── my-service              # Executable program
├── script.sh               # Shell script
└── data.json               # Data file
```

### 3. Init Scripts

Place scripts in `docker/volumes/init.d/` directory, automatically executed in alphabetical order at container startup:

```
docker/volumes/init.d/
├── 01-setup-env.sh         # Environment setup
├── 02-install-deps.sh      # Install dependencies
└── 03-start-services.sh    # Start services
```

Script example:
```bash
#!/bin/bash
echo "Setting up environment..."
pip install requests numpy
npm install -g typescript
```

**Execution Rules**:
- `.sh` files: Execute if executable, otherwise source
- Other files: Execute if executable

### 4. User HTTP Service Proxy

If your user program is an HTTP service, access it via the `/app/` path. Nginx automatically forwards requests to the `APP_SERVICE_PORT` (default 9000).

```
User request: http://localhost:8080/app/api/hello
    ↓
Nginx forwards: http://127.0.0.1:9000/api/hello
    ↓
User service handles
```

**Usage**:
1. User service listens on `APP_SERVICE_PORT`
2. Access via `/app/` path

**Custom Port**:
```bash
APP_SERVICE_PORT=3000 docker-compose up
```

### 5. Extend from Sandbox Image

Users can create custom images based on the Sandbox image:

```dockerfile
FROM your-registry/sandbox:latest

USER root
RUN apt-get update && apt-get install -y your-package

COPY --chown=sandbox:sandbox my-binary /home/sandbox/userdata/
COPY --chown=sandbox:sandbox my-init.sh /docker-entrypoint.d/

USER sandbox
```

### Environment Variable Configuration

Customize mount paths via docker-compose or environment variables:

```bash
LOCAL_WORKSPACE=/path/to/workspace \
LOCAL_SUPERVISOR_CONF=/path/to/app.supervisor.d \
LOCAL_USERDATA=/path/to/userdata \
LOCAL_INIT_SCRIPTS=/path/to/init.d \
APP_SERVICE_PORT=3000 \
docker-compose up
```

## License

Apache License 2.0
