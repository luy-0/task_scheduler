#!/bin/bash

# Docker 构建脚本
set -e

echo "=== 开始构建 Docker 镜像 ==="

# 设置镜像名称和标签
IMAGE_NAME="task-scheduler"
TAG="latest"

# 构建镜像
echo "构建镜像: ${IMAGE_NAME}:${TAG}"
docker build -t ${IMAGE_NAME}:${TAG} .

echo "=== 构建完成 ==="
echo "镜像名称: ${IMAGE_NAME}:${TAG}"
echo ""
echo "运行命令:"
echo "  docker run -d --name task-scheduler ${IMAGE_NAME}:${TAG}"
echo "  或者使用 docker-compose:"
echo "  docker-compose up -d" 