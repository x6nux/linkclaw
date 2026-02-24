FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

FROM debian:bookworm-slim
# 安装 Docker CLI（官方静态二进制）
ARG DOCKER_VERSION=27.5.1
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/* && \
    arch=$(uname -m) && \
    case "$arch" in x86_64) arch=x86_64;; aarch64) arch=aarch64;; arm64) arch=aarch64;; esac && \
    curl -fsSL "https://download.docker.com/linux/static/stable/${arch}/docker-${DOCKER_VERSION}.tgz" | \
    tar xz --strip-components=1 -C /usr/local/bin docker/docker && \
    apt-get purge -y curl && apt-get autoremove -y
WORKDIR /app
COPY --from=builder /app/server .
COPY plugins/ /app/plugins/
EXPOSE 8080
CMD ["./server"]
