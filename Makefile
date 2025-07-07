# Makefile for Task Scheduler Docker Management

# 变量定义
IMAGE_NAME = task-scheduler
TAG = latest
CONTAINER_NAME = task-scheduler

# 默认目标
.PHONY: help
help:
	@echo "可用的命令:"
	@echo "  make env-setup  - 设置环境变量文件"
	@echo "  make build      - 构建 Docker 镜像"
	@echo "  make run        - 运行容器"
	@echo "  make stop       - 停止容器"
	@echo "  make logs       - 查看容器日志"
	@echo "  make clean      - 清理容器和镜像"
	@echo "  make validate   - 验证 Dockerfile"
	@echo "  make compose-up - 使用 docker-compose 启动"
	@echo "  make compose-down - 使用 docker-compose 停止"

# 构建镜像
.PHONY: build
build:
	@echo "构建 Docker 镜像..."
	docker build -t $(IMAGE_NAME):$(TAG) .
	@echo "构建完成: $(IMAGE_NAME):$(TAG)"

# 运行容器
.PHONY: run
run:
	@echo "运行容器..."
	docker run -d \
		--name $(CONTAINER_NAME) \
		--restart unless-stopped \
		-v $(PWD)/configs:/app/configs:ro \
		-v $(PWD)/data:/app/plugins/auto-buy/ahr999_history \
		-e TZ=Asia/Shanghai \
		$(IMAGE_NAME):$(TAG)
	@echo "容器已启动: $(CONTAINER_NAME)"

# 停止容器
.PHONY: stop
stop:
	@echo "停止容器..."
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true
	@echo "容器已停止并删除"

# 查看日志
.PHONY: logs
logs:
	@echo "查看容器日志..."
	docker logs -f $(CONTAINER_NAME)

# 清理
.PHONY: clean
clean:
	@echo "清理容器和镜像..."
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true
	docker rmi $(IMAGE_NAME):$(TAG) || true
	@echo "清理完成"

# 设置环境变量文件
.PHONY: env-setup
env-setup:
	@echo "设置环境变量文件..."
	@./scripts/setup-env.sh

# 验证 Dockerfile
.PHONY: validate
validate:
	@echo "验证 Dockerfile..."
	./scripts/validate-dockerfile.sh

# 使用 docker-compose 启动
.PHONY: compose-up
compose-up:
	@echo "使用 docker-compose 启动服务..."
	docker-compose up -d
	@echo "服务已启动"

# 使用 docker-compose 停止
.PHONY: compose-down
compose-down:
	@echo "使用 docker-compose 停止服务..."
	docker-compose down
	@echo "服务已停止"

# 查看 docker-compose 日志
.PHONY: compose-logs
compose-logs:
	@echo "查看 docker-compose 日志..."
	docker-compose logs -f

# 重新构建并启动
.PHONY: rebuild
rebuild: clean build run
	@echo "重新构建并启动完成"

# 重新构建并启动 (docker-compose)
.PHONY: compose-rebuild
compose-rebuild:
	@echo "重新构建并启动 (docker-compose)..."
	docker-compose down
	docker-compose up -d --build
	@echo "重新构建并启动完成"

# 进入容器
.PHONY: shell
shell:
	@echo "进入容器..."
	docker exec -it $(CONTAINER_NAME) sh

# 检查容器状态
.PHONY: status
status:
	@echo "容器状态:"
	docker ps -a --filter name=$(CONTAINER_NAME)

# 检查镜像
.PHONY: images
images:
	@echo "Docker 镜像:"
	docker images | grep $(IMAGE_NAME) 