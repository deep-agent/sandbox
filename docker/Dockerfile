FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/sandbox-server ./cmd/sandbox-server
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o bin/mcp-hub ./cmd/mcp-hub

FROM debian:bookworm-slim

ARG TARGETARCH
ARG NODE_VERSION=24.13.0

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

RUN apt-get update && apt-get install -y \
    wget \
    curl \
    gnupg \
    ca-certificates \
    supervisor \
    nginx \
    gettext-base \
    xvfb \
    x11vnc \
    fluxbox \
    novnc \
    websockify \
    chromium \
    fonts-liberation \
    fonts-noto-cjk \
    python3 \
    python-is-python3 \
    python3-pip \
    git \
    ripgrep \
    xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
	arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
	case "${arch}" in \
		amd64) node_arch="x64" ;; \
		arm64) node_arch="arm64" ;; \
		*) echo "unsupported architecture=${arch}"; exit 1 ;; \
	esac; \
	curl -fsSLO "https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-linux-${node_arch}.tar.xz"; \
	tar -xJf "node-v${NODE_VERSION}-linux-${node_arch}.tar.xz" -C /usr/local --strip-components=1; \
	rm "node-v${NODE_VERSION}-linux-${node_arch}.tar.xz"; \
	node -v; \
	npm -v

RUN useradd -m -s /bin/bash sandbox && \
    mkdir -p /home/sandbox/workspace && \
    mkdir -p /home/sandbox/app.supervisor.d && \
    mkdir -p /home/sandbox/userdata && \
    mkdir -p /docker-entrypoint.d && \
    chown -R sandbox:sandbox /home/sandbox /docker-entrypoint.d

RUN mkdir -p /var/lib/nginx/body /var/lib/nginx/proxy /var/lib/nginx/fastcgi /var/lib/nginx/uwsgi /var/lib/nginx/scgi /var/www/terminal && \
    chown -R sandbox:sandbox /var/lib/nginx /var/log/nginx /var/run /var/www && \
    rm -f /etc/nginx/sites-enabled/default

ENV HOME=/home/sandbox
ENV WORKSPACE=/home/sandbox/workspace
ENV SUPERVISOR_CONF_DIR=/home/sandbox/app.supervisor.d
ENV USERDATA_DIR=/home/sandbox/userdata
ENV DISPLAY=:99
ENV VNC_SERVER_PORT=5900
ENV WEBSOCKET_PROXY_PORT=6080
ENV SANDBOX_SRV_PORT=8000
ENV MCP_HUB_PORT=8001
ENV BROWSER_REMOTE_DEBUGGING_PORT=9222
ENV PUBLIC_PORT=8080
ENV APP_SERVICE_PORT=9000

COPY --chown=sandbox:sandbox docker/scripts/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY --chown=sandbox:sandbox docker/nginx/nginx.conf /etc/nginx/nginx.conf
COPY --chown=sandbox:sandbox docker/scripts/entrypoint.sh /entrypoint.sh
COPY --chown=sandbox:sandbox web/index.html /var/www/index.html
COPY --chown=sandbox:sandbox web/terminal/ /var/www/terminal/

COPY --from=builder --chown=sandbox:sandbox /app/bin/sandbox-server /usr/local/bin/sandbox-server
COPY --from=builder --chown=sandbox:sandbox /app/bin/mcp-hub /usr/local/bin/mcp-hub

RUN chmod +x /entrypoint.sh /usr/local/bin/sandbox-server /usr/local/bin/mcp-hub

EXPOSE 8080

USER sandbox
WORKDIR /home/sandbox

ENTRYPOINT ["/entrypoint.sh"]
