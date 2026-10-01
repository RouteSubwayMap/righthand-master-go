# RightHand Go 后端架构设计

## 项目结构

```
righthand-go/
├── go.mod
├── go.sum
├── README.md
├── Makefile
├── docker-compose.yml
├── .env
├── cmd/                    # 应用入口
│   ├── auth/              # 认证服务
│   │   └── main.go
│   ├── memo/              # 备忘录服务
│   │   └── main.go
│   ├── task/              # 任务服务
│   │   └── main.go
│   └── gateway/           # API网关
│       └── main.go
├── internal/              # 内部包（不对外暴露）
│   ├── config/            # 配置管理
│   │   ├── config.go
│   │   └── database.go
│   ├── middleware/        # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── recovery.go
│   ├── utils/             # 工具函数
│   │   ├── jwt.go
│   │   ├── response.go
│   │   ├── validator.go
│   │   └── crypto.go
│   ├── database/          # 数据库相关
│   │   ├── mysql.go
│   │   └── redis.go
│   └── models/            # 数据模型
│       ├── user.go
│       ├── role.go
│       ├── menu.go
│       ├── memo.go
│       ├── task.go
│       └── requirement.go
├── pkg/                   # 可对外暴露的公共包
│   ├── logger/            # 日志
│   ├── cache/             # 缓存
│   └── errors/            # 错误处理
├── services/              # 业务服务层
│   ├── auth-service/      # 认证服务
│   │   ├── handler/       # HTTP处理器
│   │   ├── service/       # 业务逻辑
│   │   ├── repository/    # 数据访问层
│   │   └── dto/           # 数据传输对象
│   ├── memo-service/      # 备忘录服务
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   └── dto/
│   ├── task-service/      # 任务服务
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   └── dto/
│   └── gateway-service/   # API网关服务
│       ├── handler/
│       ├── proxy/         # 代理转发
│       └── router/        # 路由配置
├── configs/               # 配置文件
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
├── sql/                   # 数据库脚本
│   ├── righthand-all-in-one.sql  # 单库整合版初始化脚本
│   └── README.md          # 数据库脚本说明
├── scripts/               # 脚本文件
│   ├── init-db.sh
│   └── build.sh
├── tests/                 # 测试文件
│   ├── auth/
│   ├── memo/
│   ├── task/
│   └── integration/
├── docs/                  # 文档
│   └── api.md
└── deploy/                # 部署相关文件
    ├── Dockerfile.auth
    ├── Dockerfile.memo
    ├── Dockerfile.task
    └── Dockerfile.gateway
```

## 技术栈选择

### 核心框架
- **Gin**: Web框架，高性能HTTP路由
- **GORM**: ORM框架，数据库操作
- **Viper**: 配置管理
- **Zap**: 结构化日志
- **JWT-Go**: JWT认证

### 数据库
- **MySQL 8.0+**: 主数据库
- **Redis**: 缓存和会话存储

### 其他组件
- **Validator**: 参数验证
- **Swagger**: API文档
- **Testify**: 单元测试
- **Air**: 热重载开发工具

## 服务端口规划

| 服务 | 端口 | 说明 |
|------|------|------|
| Auth Service | 8081 | 认证授权服务 |
| Memo Service | 8082 | 备忘录服务 |
| Task Service | 8083 | 任务管理服务 |
| API Gateway | 8080 | API网关 |

## 核心功能模块

### 1. 认证授权服务 (auth-service)
- 用户登录/登出
- JWT Token生成与验证
- 用户管理（CRUD）
- 角色管理（CRUD）
- 菜单管理（树形结构）
- 字典管理

### 2. 备忘录服务 (memo-service)
- 备忘项管理
- 联系人管理
- 个人经历管理（教育/工作/项目/生活）
- 数据统计

### 3. 任务管理服务 (task-service)
- 任务管理（CRUD、状态流转）
- 需求管理
- 工作流管理
- 公告管理
- 消息中心
- 任务依赖与附件

### 4. API网关 (gateway-service)
- 请求路由转发
- 统一鉴权
- 跨域处理
- 请求日志
- 限流熔断

## 数据库设计

### 单库整合架构

所有模块整合到单个数据库 `righthand` 中，表名添加模块前缀以便识别：

- **AUTH 模块**: `auth_` 前缀（6个表）- 用户权限认证
- **MEMO 模块**: `memo_` 前缀（6个表）- 备忘录业务
- **TASK 模块**: `task_` 前缀（13个表）- 任务管理

**总计**: 25个表

### 初始化数据库

```bash
# 方式一：命令行执行
mysql -u root -p < sql/righthand-all-in-one.sql

# 方式二：MySQL客户端
mysql> SOURCE sql/righthand-all-in-one.sql;
```

详细说明请参考: [sql/README.md](sql/README.md)

### 默认账号

- **管理员**: admin / admin123
- **测试用户**: test / test123

## API路由规则

保持与原有Java版本相同的路由规则：
- /api/auth/** -> auth-service (8081)
- /api/system/** -> auth-service (8081)
- /api/memos/** -> memo-service (8082)
- /api/contacts/** -> memo-service (8082)
- /api/experiences/** -> memo-service (8082)
- /api/dashboard/** -> memo-service (8082)
- /api/tasks/** -> task-service (8083)
- /api/requirements/** -> task-service (8083)
- /api/announcements/** -> task-service (8083)
- /api/messages/** -> task-service (8083)

## 开发规范

### 代码规范
- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 使用golint进行代码检查
- 所有导出函数和类型必须有注释

### 测试规范
- 单元测试覆盖率不低于70%
- 集成测试覆盖核心业务流程
- 使用testify框架编写测试

### Git规范
- 分支策略：main, develop, feature/*, hotfix/*
- Commit信息遵循Conventional Commits规范
