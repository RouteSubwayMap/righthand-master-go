#!/bin/bash
# ============================================
# RightHand 私人管家系统 - 启动脚本
# ============================================
set -e

APP_HOME="/opt/righthand"
echo "=== RightHand 服务启动 ==="

# 启动后端服务
if sudo systemctl is-active --quiet righthand; then
    echo "righthand 服务已在运行中"
else
    echo "启动 righthand 后端服务..."
    sudo systemctl start righthand
    sleep 2
fi

# 检查 Nginx
if command -v nginx &> /dev/null; then
    if sudo systemctl is-active --quiet nginx; then
        echo "Nginx 已在运行中"
    else
        echo "启动 Nginx..."
        sudo systemctl start nginx
    fi
else
    echo "警告: Nginx 未安装，前端不可用"
fi

echo ""
echo "=== 服务状态 ==="
sudo systemctl status righthand --no-pager -l 2>/dev/null || echo "righthand: 状态未知"
echo ""
echo "后端API: http://localhost:9080/health"
echo "前端应用: http://localhost"
echo ""
