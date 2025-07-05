#!/bin/bash

# Dockerfile 验证脚本
set -e

echo "=== 验证 Dockerfile 语法 ==="

# 检查 Dockerfile 是否存在
if [ ! -f "Dockerfile" ]; then
    echo "错误: Dockerfile 不存在"
    exit 1
fi

# 检查必要的文件是否存在
echo "检查必要文件..."

required_files=(
    "main.go"
    "go.mod"
    "go.sum"
    "configs/config.yaml"
    "configs/tasks/auto-buy.yaml"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "警告: $file 不存在"
    else
        echo "✓ $file"
    fi
done

# 检查 .dockerignore 是否存在
if [ ! -f ".dockerignore" ]; then
    echo "警告: .dockerignore 不存在"
else
    echo "✓ .dockerignore"
fi

# 检查 docker-compose.yml 是否存在
if [ ! -f "docker-compose.yml" ]; then
    echo "警告: docker-compose.yml 不存在"
else
    echo "✓ docker-compose.yml"
fi

echo ""
echo "=== 验证完成 ==="
echo ""
echo "如果所有文件都存在，可以运行以下命令构建镜像："
echo "  docker build -t task-scheduler ."
echo "  或者使用脚本："
echo "  ./scripts/docker-build.sh" 