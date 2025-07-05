# Docker 部署说明

## 概述

本项目提供了完整的 Docker 容器化部署方案，包括：
- 多阶段构建的 Dockerfile
- docker-compose 配置
- 构建和部署脚本

## 快速开始

### 1. 使用 docker-compose（推荐）

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 2. 使用 Docker 命令

```bash
# 构建镜像
./scripts/docker-build.sh

# 运行容器
docker run -d \
  --name task-scheduler \
  -v $(pwd)/configs:/app/configs:ro \
  -v $(pwd)/data:/app/plugins/auto-buy/ahr999_history \
  -e TZ=Asia/Shanghai \
  task-scheduler:latest

# 查看日志
docker logs -f task-scheduler

# 停止容器
docker stop task-scheduler
docker rm task-scheduler
```

## 配置说明

### 挂载目录

- `./configs:/app/configs:ro` - 配置文件目录（只读）
- `./data:/app/plugins/auto-buy/ahr999_history` - AHR999 历史数据目录

### 环境变量

- `TZ=Asia/Shanghai` - 设置时区为上海

## 目录结构

```
task_scheduler/
├── Dockerfile              # Docker 构建文件
├── docker-compose.yml      # Docker Compose 配置
├── .dockerignore          # Docker 忽略文件
├── configs/               # 配置文件目录
│   ├── config.yaml
│   └── tasks/
│       └── auto-buy.yaml
├── data/                  # 数据目录（容器运行时创建）
└── scripts/
    └── docker-build.sh    # 构建脚本
```

## 构建优化

### 多阶段构建

Dockerfile 使用多阶段构建：
1. **构建阶段**：使用 `golang:1.21-alpine` 编译应用
2. **运行阶段**：使用 `alpine:latest` 运行应用

### 安全特性

- 使用非 root 用户运行
- 最小化运行时镜像
- 只复制必要的文件

### 健康检查

容器包含健康检查机制：
- 每 30 秒检查一次
- 超时时间 10 秒
- 重试 3 次

## 故障排除

### 查看容器状态

```bash
# 查看容器状态
docker ps -a

# 查看容器日志
docker logs task-scheduler

# 进入容器调试
docker exec -it task-scheduler sh
```

### 常见问题

1. **配置文件权限问题**
   ```bash
   # 确保配置文件可读
   chmod 644 configs/*.yaml
   ```

2. **数据目录权限问题**
   ```bash
   # 创建数据目录并设置权限
   mkdir -p data
   chmod 755 data
   ```

3. **时区问题**
   ```bash
   # 确保容器时区正确
   docker exec task-scheduler date
   ```

## 生产环境部署

### 1. 使用 Docker Swarm

```bash
# 初始化 Swarm
docker swarm init

# 部署服务
docker stack deploy -c docker-compose.yml task-scheduler
```

### 2. 使用 Kubernetes

创建 Kubernetes 部署文件：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: task-scheduler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: task-scheduler
  template:
    metadata:
      labels:
        app: task-scheduler
    spec:
      containers:
      - name: task-scheduler
        image: task-scheduler:latest
        env:
        - name: TZ
          value: "Asia/Shanghai"
        volumeMounts:
        - name: configs
          mountPath: /app/configs
          readOnly: true
        - name: data
          mountPath: /app/plugins/auto-buy/ahr999_history
      volumes:
      - name: configs
        hostPath:
          path: /path/to/configs
      - name: data
        persistentVolumeClaim:
          claimName: task-scheduler-data
```

## 监控和日志

### 日志管理

Docker Compose 配置了日志轮转：
- 最大文件大小：10MB
- 保留文件数：3个

### 监控建议

1. 使用 Prometheus + Grafana 监控容器指标
2. 配置日志聚合系统（如 ELK Stack）
3. 设置告警机制

## 更新部署

### 更新镜像

```bash
# 重新构建镜像
./scripts/docker-build.sh

# 重启服务
docker-compose down
docker-compose up -d
```

### 滚动更新

```bash
# 使用 docker-compose 滚动更新
docker-compose up -d --no-deps --build task-scheduler
``` 