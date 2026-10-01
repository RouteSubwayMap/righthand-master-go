@echo off
chcp 65001 >nul
echo ==================== RightHand 私人管家系统 - 测试环境部署脚本 ====================
echo.

:: 检查是否存在可执行文件
if not exist "righthand-server.exe" (
    echo 错误: 未找到 righthand-server.exe，请先运行 build.bat
    pause
    exit /b 1
)

echo 1. 检查数据库连接配置...
echo    - 数据库主机: 127.0.0.1
echo    - 数据库端口: 3306
echo    - 数据库名称: righthand
echo    - 用户名: root
echo.

echo 2. 请确保已执行数据库初始化脚本:
echo    - sql/init-database.sql
echo    - sql/init-test-data.sql
echo.

echo 3. 启动后端服务...
start "RightHand Server" righthand-server.exe

echo 4. 服务已启动，请等待几秒后启动前端...
timeout /t 3 /nobreak >nul

echo 5. 启动前端开发服务器...
cd ..\righthand-frontend
start "RightHand Frontend" npm run dev

echo.
echo ==================== 部署完成 ====================
echo.
echo 后端服务: http://localhost:9080
echo 前端应用: http://localhost:5173
echo.
echo 默认登录账号:
echo   用户名: admin
echo   密码: admin123
echo.
pause