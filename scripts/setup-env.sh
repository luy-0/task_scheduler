#!/bin/bash

# 设置环境变量脚本
# 用于创建和配置 .env 文件

set -e

echo "=== 环境变量设置脚本 ==="
echo

# 检查是否已存在 .env 文件
if [ -f .env ]; then
    echo "⚠️  .env 文件已存在"
    read -p "是否要覆盖现有文件？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "操作已取消"
        exit 0
    fi
fi

echo "📝 创建 .env 文件..."

# 创建 .env 文件
cat > .env << 'EOF'
# Binance API Configuration
# 请将下面的占位符替换为您的实际 API 密钥
BINANCE_API_KEY=your_binance_api_key_here
BINANCE_SECRET_KEY=your_binance_secret_key_here

# 其他环境变量（可选）
# TZ=Asia/Shanghai
# LOG_LEVEL=info
EOF

echo "✅ .env 文件已创建"
echo
echo "🔧 下一步操作："
echo "1. 编辑 .env 文件，填入您的实际 API 密钥"
echo "2. 运行 'make validate' 验证环境"
echo "3. 运行 'make compose-up' 启动服务"
echo
echo "📋 .env 文件内容："
echo "----------------------------------------"
cat .env
echo "----------------------------------------"
echo
echo "⚠️  安全提醒："
echo "- 确保 .env 文件不会被提交到版本控制"
echo "- 定期轮换您的 API 密钥"
echo "- 在生产环境中使用安全的密钥管理服务" 