FROM node:22-slim

# Docker CLI (static binary, compatible with Docker Desktop)
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates && \
    arch=$(uname -m) && \
    case "$arch" in x86_64) arch=x86_64;; aarch64) arch=aarch64;; arm64) arch=aarch64;; esac && \
    curl -fsSL "https://download.docker.com/linux/static/stable/${arch}/docker-27.5.1.tgz" | \
    tar xz --strip-components=1 -C /usr/local/bin docker/docker && \
    apt-get purge -y curl && apt-get autoremove -y && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install all dependencies (including devDeps for tsc build)
COPY package*.json ./
RUN npm ci

# Copy source and build
COPY tsconfig.json ./
COPY src/ src/
RUN npx tsc

# Prune devDependencies after build
RUN npm prune --omit=dev

# Copy container assets (agent-runner, skills, build script)
COPY container/ container/

# Runtime directories
RUN mkdir -p data groups store

CMD ["node", "dist/index.js"]
