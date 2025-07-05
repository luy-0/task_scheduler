# 多阶段构建 Dockerfile
# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task_scheduler .

# 第二阶段：运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/task_scheduler .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p /app/plugins/auto-buy/ahr999_history && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口（如果需要的话）
# EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ps aux | grep task_scheduler || exit 1

# 启动命令
CMD ["./task_scheduler"] 