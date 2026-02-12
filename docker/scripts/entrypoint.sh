#!/bin/bash
set -e

mkdir -p /tmp/chromium
mkdir -p /home/sandbox/workspace
mkdir -p /home/sandbox/app.supervisor.d
mkdir -p /home/sandbox/userdata

export APP_SERVICE_PORT="${APP_SERVICE_PORT:-9000}"
echo "APP_SERVICE_PORT=${APP_SERVICE_PORT}"

if [ "${ENABLE_CDP_PROXY:-true}" = "true" ]; then
    echo "CDP proxy enabled"
    export CDP_LOCATION_BLOCK='location /cdp/ {
            proxy_pass http://127.0.0.1:9222/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_read_timeout 86400s;
            proxy_send_timeout 86400s;
        }'
else
    echo "CDP proxy disabled by ENABLE_CDP_PROXY=false"
    export CDP_LOCATION_BLOCK=""
fi

NGINX_TEMPLATE="/etc/nginx/nginx.conf"
NGINX_RUNTIME="/tmp/nginx.runtime.conf"
envsubst '${APP_SERVICE_PORT} ${CDP_LOCATION_BLOCK}' < "$NGINX_TEMPLATE" > "$NGINX_RUNTIME"
echo "Generated nginx runtime config: $NGINX_RUNTIME"

SUPERVISOR_CONF="/etc/supervisor/conf.d/supervisord.conf"
RUNTIME_CONF="/tmp/supervisord.runtime.conf"

cp "$SUPERVISOR_CONF" "$RUNTIME_CONF"
sed -i "s|command=/usr/sbin/nginx -g \"daemon off;\"|command=/usr/sbin/nginx -c $NGINX_RUNTIME -g \"daemon off;\"|" "$RUNTIME_CONF"

if [ "${ENABLE_MCP:-true}" = "false" ]; then
    echo "MCP Hub disabled by ENABLE_MCP=false"
    sed -i '/\[program:mcp-hub\]/,/^$/d' "$RUNTIME_CONF"
fi

if [ "${ENABLE_BROWSER:-true}" = "false" ]; then
    echo "Browser (Chromium) disabled by ENABLE_BROWSER=false"
    sed -i '/\[program:chromium\]/,/^$/d' "$RUNTIME_CONF"
fi

if [ "${ENABLE_VNC:-true}" = "false" ]; then
    echo "VNC disabled by ENABLE_VNC=false"
    sed -i '/\[program:xvfb\]/,/^$/d' "$RUNTIME_CONF"
    sed -i '/\[program:fluxbox\]/,/^$/d' "$RUNTIME_CONF"
    sed -i '/\[program:x11vnc\]/,/^$/d' "$RUNTIME_CONF"
    sed -i '/\[program:websockify\]/,/^$/d' "$RUNTIME_CONF"
fi

SUPERVISOR_CONF_DIR="${SUPERVISOR_CONF_DIR:-/home/sandbox/app.supervisor.d}"
if [ -d "$SUPERVISOR_CONF_DIR" ]; then
    for conf in "$SUPERVISOR_CONF_DIR"/*.conf; do
        if [ -f "$conf" ]; then
            echo "Loading supervisor conf: $conf"
            echo "" >> "$RUNTIME_CONF"
            cat "$conf" >> "$RUNTIME_CONF"
        fi
    done
fi

if [ -d "/docker-entrypoint.d" ]; then
    echo "Running user init scripts from /docker-entrypoint.d..."
    for f in /docker-entrypoint.d/*; do
        if [ -f "$f" ]; then
            case "$f" in
                *.sh)
                    if [ -x "$f" ]; then
                        echo "Executing: $f"
                        "$f"
                    else
                        echo "Sourcing: $f"
                        . "$f"
                    fi
                    ;;
                *)
                    if [ -x "$f" ]; then
                        echo "Executing: $f"
                        "$f"
                    fi
                    ;;
            esac
        fi
    done
    echo "User init scripts completed."
fi

exec /usr/bin/supervisord -c "$RUNTIME_CONF"
