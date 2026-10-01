# ============================================
# RightHand 私人管家系统 - 服务器配置
# 部署前请修改以下配置项（搜索 TODO: 标记）
# ============================================

# 服务器配置
server:
  port: 9080
  mode: release # debug, release, test

# 数据库配置
# TODO: 请修改为实际数据库连接信息
database:
  host: 127.0.0.1
  port: 3306
  username: root
  password: root           # TODO: 请修改数据库密码
  dbname: righthand
  charset: utf8mb4
  parseTime: true
  loc: Asia/Shanghai
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

# JWT配置
# TODO: 生产环境请修改为随机安全密钥
jwt:
  secret: righthand-secret-key-2024-very-long-and-secure
  expire: 7200 # 2小时，单位秒
  issuer: righthand-auth

# 日志配置（按日期切割：logs/app-YYYY-MM-DD.log，服务器上为 /opt/righthand/logs/）
# output: both = 控制台（systemd 记入 journald，用 journalctl -u righthand 查看）+ 按日期文件
log:
  level: info # debug, info, warn, error（预留，本轮未强制生效）
  format: console # console 文本 / json 结构化（json 预留，本轮未启用）
  output: both # stdout / file / both
  file_path: ./logs/app.log # 实际写入 app-YYYY-MM-DD.log（按日期切割）
  max_size: 100 # MB（预留，本轮未启用大小切分）
  max_backups: 10 # 最多保留日志文件数（含当日）
  max_age: 30 # days（超过天数的历史文件自动清理）

# CORS配置
# 注：单机Nginx代理部署时，前后端同源，CORS不触发，无需修改此项.
# 仅当前后端分离部署（如开发环境 Vite :5173 直连后端 :9080）时才需要配置.
# TODO: 若前后分离，将下方改为实际前端地址，例如:
#   - http://10.252.20.59:5173
cors:
  allow_origins:
    - http://localhost
    - http://127.0.0.1
  allow_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allow_headers:
    - Content-Type
    - Authorization
  allow_credentials: true
  max_age: 3600
