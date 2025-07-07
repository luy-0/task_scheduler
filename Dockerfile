# 阶段1：带缓存的构建阶段
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG BUILD_VERSION=1.0.0
ARG TARGETOS=linux
ARG TARGETARCH

RUN apk add --no-cache --virtual .build-deps \
    git \
    ca-certificates \
    tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download -x && \
    go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w -X main.Version=${BUILD_VERSION}" \
    -o /app/task_scheduler .

# 阶段2：最小化运行时镜像
FROM alpine:latest
WORKDIR /app

# 元数据
LABEL maintainer="your-email@example.com"
LABEL org.opencontainers.image.version="${BUILD_VERSION}"

# 系统配置
RUN apk --no-cache add \
    ca-certificates \
    tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    mkdir -p /app/configs /app/plugins/auto-buy/ahr999_history && \
    chown -R appuser:appgroup /app

# 精确复制构建产物
COPY --chown=appuser:appgroup --from=builder /app/task_scheduler /app/
COPY --from=builder /app/configs /app/configs/

# 环境配置
ENV TZ=Asia/Shanghai \
    BUILD_VERSION=${BUILD_VERSION}

USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ps -o comm | grep -q '^task_scheduler$' || exit 1

ENTRYPOINT ["/app/task_scheduler"]
