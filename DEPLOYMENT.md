# RightHand 私人管家系统 - 测试环境部署指南

## 一、环境要求

| 组件 | 版本要求 |
|------|---------|
| Go | 1.24+ |
| Node.js | 20+ |
| MySQL | 8.0+ |
| npm | 10+ |

## 二、部署步骤

### 1. 数据库初始化

```bash
# 创建数据库
CREATE DATABASE IF NOT EXISTS righthand DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 执行初始化脚本
mysql -u root -p righthand < sql/init-database.sql
mysql -u root -p righthand < sql/init-test-data.sql
```

### 2. 后端部署

```bash
# 进入后端目录
cd righthand-go

# 编译
go build -o righthand-server.exe cmd/server/main.go

# 启动服务
./righthand-server.exe
```

### 3. 前端部署

```bash
# 进入前端目录
cd righthand-frontend

# 安装依赖
npm install

# 开发模式启动
npm run dev

# 或生产构建
npm run build
npm run preview -- --host 0.0.0.0 --port 5173
```

### 4. 一键启动

```bash
# 直接运行启动脚本
cd righthand-go
start.bat
```

## 三、服务访问

| 服务 | 地址 |
|------|------|
| 前端应用 | http://localhost:5173 |
| 后端API | http://localhost:9080 |
| 健康检查 | http://localhost:9080/health |

## 四、默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 超级管理员 |
| test | test123 | 普通用户 |

## 五、集成验收测试清单

### 5.1 用户认证模块

| 测试项 | 预期结果 |
|--------|---------|
| 登录页面显示 | ✅ 正常显示登录表单 |
| 正确账号登录 | ✅ 登录成功，跳转到仪表盘 |
| 错误密码登录 | ✅ 提示登录失败 |
| 空密码登录 | ✅ 提示密码不能为空 |
| 退出登录 | ✅ 返回登录页面 |

### 5.2 仪表盘模块

| 测试项 | 预期结果 |
|--------|---------|
| 今日代办统计 | ✅ 显示截止日期为今天的待办数量 |
| 所有代办统计 | ✅ 显示所有待办事项数量 |
| 进行中统计 | ✅ 显示进行中事项数量 |
| 已逾期统计 | ✅ 显示已逾期事项数量 |
| 今日代办列表 | ✅ 分页展示今日待办事项 |
| 所有代办列表 | ✅ 分页展示所有待办事项 |
| 进行中列表 | ✅ 分页展示进行中事项 |
| 已逾期列表 | ✅ 分页展示已逾期事项 |

### 5.3 事项管理模块

| 测试项 | 预期结果 |
|--------|---------|
| 事项列表展示 | ✅ 分页展示所有事项 |
| 新增一次性事项 | ✅ 创建成功，显示在列表中 |
| 新增周期性事项 | ✅ 创建成功，按规则生成实例 |
| 编辑事项 | ✅ 修改后保存成功 |
| 删除事项 | ✅ 删除后从列表消失 |
| 状态变更 | ✅ TODO → DOING → DONE |
| 优先级设置 | ✅ HIGH/MEDIUM/LOW 正常显示 |
| 标签管理 | ✅ 标签正常显示和筛选 |

### 5.4 联系人管理模块

| 测试项 | 预期结果 |
|--------|---------|
| 联系人列表 | ✅ 分页展示所有联系人 |
| 新增联系人 | ✅ 创建成功 |
| 编辑联系人 | ✅ 修改后保存成功 |
| 删除联系人 | ✅ 删除后从列表消失 |
| 分组筛选 | ✅ 按分组过滤显示 |
| 搜索功能 | ✅ 按姓名/手机号搜索 |

### 5.5 需求管理模块

| 测试项 | 预期结果 |
|--------|---------|
| 需求列表 | ✅ 分页展示所有需求 |
| 新增需求 | ✅ 创建成功 |
| 编辑需求 | ✅ 修改后保存成功 |
| 删除需求 | ✅ 删除后从列表消失 |
| 状态流转 | ✅ PENDING_REVIEW → ACCEPTED → IN_PROGRESS → COMPLETED |
| 优先级设置 | ✅ HIGH/MEDIUM/LOW 正常显示 |

### 5.6 系统管理模块

#### 用户管理
| 测试项 | 预期结果 |
|--------|---------|
| 用户列表 | ✅ 分页展示所有用户 |
| 新增用户 | ✅ 创建成功 |
| 编辑用户 | ✅ 修改后保存成功 |
| 删除用户 | ✅ 删除后从列表消失 |
| 分配角色 | ✅ 角色分配成功 |

#### 角色管理
| 测试项 | 预期结果 |
|--------|---------|
| 角色列表 | ✅ 展示所有角色 |
| 新增角色 | ✅ 创建成功 |
| 编辑角色 | ✅ 修改后保存成功 |
| 删除角色 | ✅ 删除后从列表消失 |
| 分配菜单 | ✅ 菜单权限分配成功 |

#### 菜单管理
| 测试项 | 预期结果 |
|--------|---------|
| 菜单树 | ✅ 树形展示菜单结构 |
| 新增菜单 | ✅ 创建成功（支持目录/菜单/按钮） |
| 编辑菜单 | ✅ 修改后保存成功 |
| 删除菜单 | ✅ 删除后从树消失 |

#### 字典管理
| 测试项 | 预期结果 |
|--------|---------|
| 字典列表 | ✅ 按类型分组展示 |
| 新增字典项 | ✅ 创建成功 |
| 编辑字典项 | ✅ 修改后保存成功 |
| 删除字典项 | ✅ 删除后从列表消失 |

### 5.7 接口测试

| 接口 | 方法 | 状态码 |
|------|------|--------|
| /api/auth/login | POST | 200 |
| /api/auth/info | GET | 200 |
| /api/auth/menus | GET | 200 |
| /api/dashboard/stats | GET | 200 |
| /api/tasks | GET | 200 |
| /api/tasks | POST | 201 |
| /api/contacts | GET | 200 |
| /api/contacts | POST | 201 |
| /api/requirements | GET | 200 |
| /api/requirements | POST | 201 |
| /api/system/users | GET | 200 |
| /api/system/users | POST | 201 |
| /api/system/roles | GET | 200 |
| /api/system/roles | POST | 201 |
| /api/system/menus | GET | 200 |
| /api/system/dicts | GET | 200 |

## 六、配置说明

### 配置文件位置
```
righthand-go/configs/config.yaml
```

### 主要配置项

```yaml
server:
  port: 9080                    # 后端服务端口
  mode: debug                   # 运行模式: debug/release/test

database:
  host: 127.0.0.1              # 数据库主机
  port: 3306                    # 数据库端口
  username: root                # 数据库用户名
  password: ""                  # 数据库密码
  dbname: righthand             # 数据库名称

jwt:
  secret: your-secret-key       # JWT密钥
  expire: 7200                  # token过期时间(秒)

cors:
  allow_origins:                # 允许的前端域名
    - http://localhost:5173
```

## 七、目录结构

```
righthand/
├── righthand-go/              # 后端服务
│   ├── cmd/
│   │   └── server/
│   │       └── main.go       # 启动入口
│   ├── services/              # 业务服务
│   │   ├── auth-service/      # 认证服务
│   │   └── task-service/      # 事项服务
│   ├── internal/              # 内部模块
│   ├── configs/               # 配置文件
│   ├── sql/                   # SQL脚本
│   └── righthand-server.exe   # 可执行文件
└── righthand-frontend/        # 前端应用
    ├── src/
    │   ├── views/             # 页面组件
    │   ├── components/        # 通用组件
    │   ├── api/               # API封装
    │   └── stores/            # 状态管理
    ├── dist/                  # 构建产物
    └── package.json           # 依赖配置
```

## 八、问题排查

### 端口占用
```bash
# 检查端口占用
netstat -ano | findstr "9080"
netstat -ano | findstr "5173"

# 杀死占用进程
taskkill /F /PID <进程ID>
```

### 数据库连接失败
1. 检查 MySQL 服务是否启动
2. 检查数据库配置文件中的用户名和密码
3. 确认数据库 `righthand` 已创建
4. 执行初始化脚本

### 前端跨域问题
检查 `config.yaml` 中的 CORS 配置是否包含前端地址。

## 九、测试命令

```bash
# 后端测试
cd righthand-go
go test ./services/... -v

# 前端测试
cd righthand-frontend
npm run test:run

# 前端构建
npm run build
```

---

**部署完成** - 系统已准备就绪，可进行集成验收测试！