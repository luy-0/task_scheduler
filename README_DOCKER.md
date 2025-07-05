# Docker 部署快速参考

## 🚀 快速开始

### 使用 Makefile（推荐）

```bash
# 查看所有可用命令
make help

# 验证环境
make validate

# 构建并启动（docker-compose）
make compose-up

# 查看日志
make compose-logs

# 停止服务
make compose-down
```

### 使用 Docker Compose

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 使用 Docker 命令

```bash
# 构建镜像
docker build -t task-scheduler .

# 运行容器
docker run -d \
  --name task-scheduler \
  -v $(pwd)/configs:/app/configs:ro \
  -v $(pwd)/data:/app/plugins/auto-buy/ahr999_history \
  -e TZ=Asia/Shanghai \
  task-scheduler:latest
```

## 📁 目录结构

```
task_scheduler/
├── Dockerfile              # Docker 构建文件
├── docker-compose.yml      # Docker Compose 配置
├── Makefile               # 构建管理脚本
├── .dockerignore          # Docker 忽略文件
├── configs/               # 配置文件目录
│   ├── config.yaml
│   └── tasks/
│       └── auto-buy.yaml
├── data/                  # 数据目录（自动创建）
└── scripts/
    ├── docker-build.sh    # 构建脚本
    └── validate-dockerfile.sh # 验证脚本
```

## 🔧 配置说明

### 环境变量

- `TZ=Asia/Shanghai` - 设置时区

### 挂载目录

- `./configs:/app/configs:ro` - 配置文件（只读）
- `./data:/app/plugins/auto-buy/ahr999_history` - 数据目录

## 📋 常用命令

### Makefile 命令

| 命令 | 说明 |
|------|------|
| `make help` | 查看所有命令 |
| `make validate` | 验证环境 |
| `make build` | 构建镜像 |
| `make run` | 运行容器 |
| `make stop` | 停止容器 |
| `make logs` | 查看日志 |
| `make clean` | 清理资源 |
| `make compose-up` | 启动服务 |
| `make compose-down` | 停止服务 |
| `make compose-logs` | 查看服务日志 |
| `make shell` | 进入容器 |
| `make status` | 查看状态 |

### Docker Compose 命令

| 命令 | 说明 |
|------|------|
| `docker-compose up -d` | 后台启动 |
| `docker-compose down` | 停止服务 |
| `docker-compose logs -f` | 查看日志 |
| `docker-compose ps` | 查看状态 |
| `docker-compose restart` | 重启服务 |

## 🐛 故障排除

### 常见问题

1. **权限问题**
   ```bash
   # 创建数据目录
   mkdir -p data
   chmod 755 data
   ```

2. **配置文件问题**
   ```bash
   # 检查配置文件
   make validate
   ```

3. **容器状态**
   ```bash
   # 查看容器状态
   make status
   
   # 查看日志
   make logs
   ```

### 调试命令

```bash
# 进入容器调试
make shell

# 查看容器资源使用
docker stats task-scheduler

# 查看容器详细信息
docker inspect task-scheduler
```

## 🔄 更新部署

### 重新构建

```bash
# 使用 Makefile
make compose-rebuild

# 或使用 docker-compose
docker-compose down
docker-compose up -d --build
```

### 更新配置

1. 修改 `configs/` 目录下的配置文件
2. 重启服务：
   ```bash
   make compose-down
   make compose-up
   ```

## 📊 监控

### 健康检查

容器包含健康检查机制：
- 每 30 秒检查一次
- 超时时间 10 秒
- 重试 3 次

### 日志管理

- 日志轮转：最大 10MB，保留 3 个文件
- 日志格式：JSON 格式

## 🔒 安全特性

- 使用非 root 用户运行
- 最小化运行时镜像
- 只复制必要的文件
- 配置文件只读挂载

## 📝 生产环境

### 推荐配置

1. 使用 Docker Swarm 或 Kubernetes
2. 配置监控和告警
3. 设置日志聚合
4. 使用私有镜像仓库

### 性能优化

- 使用多阶段构建减小镜像大小
- 配置资源限制
- 使用数据卷持久化数据

---

更多详细信息请参考 [DOCKER_DEPLOY.md](DOCKER_DEPLOY.md) 