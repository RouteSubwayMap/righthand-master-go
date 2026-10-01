#!/bin/bash
# ============================================
# RightHand 私人管家系统 - 一键安装脚本
# 使用方法: sudo ./install.sh
# 执行前请先修改本脚本中的配置项
# ============================================
set -e

# ============================================
# 配置项（部署前根据实际环境修改）
# ============================================

# 数据库初始化配置
AUTO_INIT_DB=false            # 是否自动执行 SQL 初始化（true/false）
DB_HOST="127.0.0.1"          # 数据库主机
DB_PORT="3306"               # 数据库端口
DB_USER="root"               # 数据库用户
DB_PASS=root                   # 数据库密码（空密码留空）
DB_NAME="righthand"          # 数据库名称

# ============================================
# 以下为安装逻辑，一般无需修改
# ============================================

APP_HOME="/opt/righthand"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "========================================"
echo "  RightHand 私人管家系统 - 一键安装"
echo "========================================"
echo ""

# 检查 root 权限
if [ "$(id -u)" -ne 0 ]; then
    echo "错误: 请使用 root 权限执行 (sudo ./install.sh)"
    exit 1
fi

# 0. 赋权：解压后的文件可能没有执行权限
echo "[0/7] 赋予执行权限..."
chmod +x "$SCRIPT_DIR"/bin/righthand-server 2>/dev/null || true
chmod +x "$SCRIPT_DIR"/*.sh 2>/dev/null || true
echo "      执行权限已设置"

# 0.5 停止原有服务（避免 "Text file busy" 错误）
if systemctl is-active --quiet righthand 2>/dev/null; then
    echo "      检测到 righthand 服务正在运行，正在停止..."
    systemctl stop righthand 2>/dev/null || true
    sleep 1
fi

# 兜底：杀掉残留的 righthand-server 进程
if pgrep -f righthand-server &> /dev/null; then
    echo "      检测到残留进程，正在清理..."
    pkill -f righthand-server 2>/dev/null || true
    sleep 1
fi

# 1. 创建目录结构
echo "[1/7] 创建目录结构..."
mkdir -p "$APP_HOME/bin"
mkdir -p "$APP_HOME/web"
mkdir -p "$APP_HOME/configs"
mkdir -p "$APP_HOME/logs"
echo "      目录创建完成: $APP_HOME"

# 2. 复制文件
echo "[2/7] 复制应用文件..."

# 后端可执行文件
if [ -f "$SCRIPT_DIR/bin/righthand-server" ]; then
    cp "$SCRIPT_DIR/bin/righthand-server" "$APP_HOME/bin/"
    chmod +x "$APP_HOME/bin/righthand-server"
    echo "      后端可执行文件已复制"
else
    echo "      警告: 未找到 bin/righthand-server"
fi

# 前端静态文件
if [ -d "$SCRIPT_DIR/web" ]; then
    rm -rf "$APP_HOME/web"/*
    cp -r "$SCRIPT_DIR/web"/* "$APP_HOME/web/"
    echo "      前端文件已复制"
else
    echo "      警告: 未找到 web/ 目录"
fi

# 配置文件
if [ -f "$SCRIPT_DIR/configs/config.yaml" ]; then
    if [ -f "$APP_HOME/configs/config.yaml" ]; then
        echo "      配置文件已存在，备份为 config.yaml.bak"
        cp "$APP_HOME/configs/config.yaml" "$APP_HOME/configs/config.yaml.bak"
    fi
    cp "$SCRIPT_DIR/configs/config.yaml" "$APP_HOME/configs/config.yaml"
    echo "      配置文件已复制"
else
    echo "      警告: 未找到 configs/config.yaml"
fi

echo "      文件复制完成"

# 3. 数据库初始化（可选）
echo "[3/6] 数据库初始化..."
if [ "$AUTO_INIT_DB" = "true" ]; then
    if ! command -v mysql &> /dev/null; then
        echo "      错误: 未找到 mysql 客户端，请先安装"
        exit 1
    fi
    echo "      正在连接数据库 ${DB_HOST}:${DB_PORT}/${DB_NAME} ..."

    if [ -f "$SCRIPT_DIR/sql/init-database.sql" ]; then
        echo "      执行 SQL 初始化脚本..."
        if [ -n "$DB_PASS" ]; then
            mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" -p"${DB_PASS}" < "$SCRIPT_DIR/sql/init-database.sql"
        else
            mysql -h"${DB_HOST}" -P"${DB_PORT}" -u"${DB_USER}" < "$SCRIPT_DIR/sql/init-database.sql"
        fi
        echo "      数据库初始化完成"
    else
        echo "      警告: 未找到 sql/init-database.sql"
    fi
else
    echo "      AUTO_INIT_DB=false，跳过数据库初始化"
    echo "      请手动执行: mysql -h${DB_HOST} -u${DB_USER} -p < sql/init-database.sql"
fi

# 4. 安装并配置 Nginx
echo "[4/6] 配置 Nginx..."

# 自动安装 Nginx（如未安装）
if ! command -v nginx &> /dev/null; then
    echo "      Nginx 未安装，正在自动安装..."
    if command -v apt &> /dev/null; then
        apt update -qq && apt install -y -qq nginx
    elif command -v yum &> /dev/null; then
        yum install -y -q nginx
    elif command -v dnf &> /dev/null; then
        dnf install -y -q nginx
    else
        echo "      错误: 无法识别包管理器，请手动安装 Nginx"
        exit 1
    fi
    echo "      Nginx 安装完成"
fi

# 复制配置
if [ -f "$SCRIPT_DIR/nginx/righthand.conf" ]; then
    cp "$SCRIPT_DIR/nginx/righthand.conf" /etc/nginx/conf.d/righthand.conf
    echo "      Nginx 配置已复制到 /etc/nginx/conf.d/"

    # 清理可能冲突的默认站点
    if [ -f /etc/nginx/sites-enabled/default ]; then
        rm -f /etc/nginx/sites-enabled/default
        echo "      已移除默认站点配置"
    fi

    # 测试并重载
    if nginx -t &> /dev/null; then
        systemctl enable nginx 2>/dev/null || true
        systemctl restart nginx 2>/dev/null || nginx -s reload
        echo "      Nginx 已启动"
    else
        echo "      警告: Nginx 配置测试失败，请检查"
        nginx -t
    fi
else
    echo "      警告: 未找到 nginx/righthand.conf"
fi

# 5. 配置 systemd 服务
echo "[5/6] 注册 systemd 服务..."
if [ -f "$SCRIPT_DIR/systemd/righthand.service" ]; then
    cp "$SCRIPT_DIR/systemd/righthand.service" /etc/systemd/system/righthand.service
    systemctl daemon-reload
    systemctl enable righthand
    systemctl start righthand
    sleep 2

    if systemctl is-active --quiet righthand; then
        echo "      righthand 服务已启动"
    else
        echo "      警告: righthand 服务启动失败，请检查日志"
        systemctl status righthand --no-pager -l 2>/dev/null || true
    fi
else
    echo "      警告: 未找到 systemd/righthand.service"
fi

# 6. 完成
echo "[6/7] 安装完成"
echo ""
echo "========================================"
echo "  RightHand 私人管家系统 - 安装完成"
echo "========================================"
echo ""
echo "  安装目录: $APP_HOME"
echo "  配置文件: $APP_HOME/configs/config.yaml"
echo "  日志目录: $APP_HOME/logs/"
echo ""
echo "  服务管理:"
echo "    启动: systemctl start righthand"
echo "    停止: systemctl stop righthand"
echo "    状态: systemctl status righthand"
echo "    重启: systemctl restart righthand"
echo ""
echo "  访问地址:"
echo "    前端应用: http://localhost"
echo "    后端API:  http://localhost:9080"
echo "    健康检查: http://localhost:9080/health"
echo ""
echo "  默认账号:"
echo "    用户名: admin"
echo "    密码:   admin123"
echo ""
