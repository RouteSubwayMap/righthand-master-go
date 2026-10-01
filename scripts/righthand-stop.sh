#!/bin/bash
# ============================================
# RightHand 私人管家系统 - 停止脚本
# ============================================
set -e

echo "=== RightHand 服务停止 ==="

# 停止后端服务
if sudo systemctl is-active --quiet righthand; then
    echo "停止 righthand 后端服务..."
    sudo systemctl stop righthand
    echo "righthand 服务已停止"
else
    echo "righthand 服务未在运行"
fi

echo ""
