#!/bin/bash
set -e

# 启动后端
cd /app/backend && ./server &

# 启动前端 (Next.js standalone)
cd /app/frontend && node server.js &

# 启动 Caddy (前台运行，保持容器存活)
exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
