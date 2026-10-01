# AGENTS.md — 私人管家系统（RightHand）· AI 新会话定向

> 本文件是 AI 执行层入口（面向日常协作）。面向人/交付的正式文档见 `docs/`。
> 每次新会话按本文件恢复上下文，红线与契约见 `docs/ai/rules.md`。

## 项目定位
RightHand（私人管家/左右手管家）是个人与团队事务管理系统。前后端分离的单体架构：一个 Go 服务（含认证与事项两大业务域）+ 一个 Vue3 后台。

## 服务与端口
- 后端：Go + Gin 单体服务（默认 :9080），入口 `righthand-go/cmd/server/main.go`
- 前端：Vue3 后台（Vite 开发 :5173，生产 Nginx :80），代理 `/api` → :9080
- 数据库：MySQL 8.0（:3306，库 `righthand`，utf8mb4）
- 端口以 `righthand-go/configs/config.yaml` 为准

## 目录地图
- `righthand-go/cmd/server/main.go` — 服务入口（全部路由在此注册）
- `righthand-go/services/auth-service/{handler,service,repository,dto}` — 认证/用户/角色/菜单/字典
- `righthand-go/services/task-service/{handler,service,repository,dto}` — 事项/联系人/需求/批量任务/生产节点/银行参数/仪表盘
- `righthand-go/internal/{config,database,middleware,models,scheduler,utils}` — 基础设施（中间件 auth.go/casbin.go、GORM 模型、周期任务调度器、统一响应与 JWT）
- `righthand-go/sql/{init-database.sql,init-test-data.sql}` — 建库与测试数据（SSOT）
- `righthand-go/tests/` — 后端接口/集成测试
- `righthand-frontend/src/{api,views,components,composables,stores,router,utils,directives}` — 前端
- `righthand-frontend/tests/` — 前端 vitest 测试
- `scripts/` — 发布构建与部署脚本；`docs/` — 项目文档；`docs/ai/` — AI 执行层
- `righthand-parent/` — Java 版遗留工程（已被 Go 版替代），无明确要求不要改动

## 技术栈
后端：Go 1.24.8 · Gin v1.9.1 · GORM + MySQL(utf8mb4) · Casbin RBAC · JWT(HS256) · bcrypt · Viper · robfig/cron（周期任务调度）。依赖版本以 `righthand-go/go.mod` 为准。
前端：Vue 3 + Vite + Element Plus + Pinia + Vue Router + Axios + ECharts + exceljs/xlsx；测试 vitest + happy-dom。

## 构建 / 测试 / 启动
- 后端构建：`cd righthand-go; go build ./...`
- 后端测试：`cd righthand-go; go test ./...`（自动测试强制连独立测试库 `righthand_test`，先执行 `go run ./cmd/init-test-db [-seed]` 初始化；涉及 sqlite 内存库的测试需 CGO + MinGW，见 rules.md §10）
- 后端启动：`cd righthand-go; go run cmd/server/main.go`（Windows 一键：`start.bat`）
- 前端：`cd righthand-frontend; npm install; npm run dev`（:5173）/ `npm run build` / `npx vitest run`
- 初始化库：`mysql -u root -p < righthand-go/sql/init-database.sql`，开发/测试再执行 `init-test-data.sql`（生产禁止）
- 默认账号：`admin/admin123`（ADMIN）、`test/test123`（USER）
- 环境：Windows PowerShell，命令分隔用 `;`（不支持 `&&`）
- 部署（唯一链路）：Windows 执行 `scripts/build-release.ps1` 打包 → 手工上传测试服务器 → `sudo ./install.sh`（systemd + Nginx 单机，目录 `/opt/righthand`）；Docker 部署方式已封存（文件保留、不再适配），不要再走 Docker 链路

## SSOT（唯一事实来源：改这里，不要复制到别处）
- 表结构 → `righthand-go/sql/init-database.sql`
- 工程红线/契约 → `docs/ai/rules.md`（每次新会话必读）
- 接口响应格式 → `righthand-go/internal/utils/response.go`
- 路由清单 → `righthand-go/cmd/server/main.go`
- 版本号 → `VERSION`
- 设计文档（`docs/*.md`）为 V1.0 基线，细节与代码不一致时以代码为准

## 每次改动的最低要求
1. 遵守 `docs/ai/rules.md` 红线；与历史文档冲突时以 rules + 代码现状为准
2. 同步更新相关测试；后端 `go build ./...` + 相关包 `go test`，前端 `npx vitest run` 通过方视为完成
3. 变更登记到 `CHANGELOG.md`（改了什么 / 为什么 / 影响范围）

## 关键操作规范（缺一则交付不完整）
- **响应格式**：业务接口统一 `utils.Success/Error/ErrorWithCode`；`utils.Error` 固定 code=500 且只收 2 参数，禁止 `utils.Error(c, 400, "msg")` 混用
- **中文编码**：HTTP 必须 `charset=utf-8`；含中文 SQL 用 `mysql < file` 导入，禁止 PowerShell 管道；涉及中文的接口须保留 UTF-8 编码测试
- **权限三板斧**：新增页面/按钮 → `init-database.sql` 注册 `sys_menu`（一按钮一权限码）+ 角色授权 `sys_role_menu` + 前端 `hasPermission`/`v-permission` 控制显隐
- **统一事项**：任务/备忘/需求统一走 `task_item` 表与 `/api/tasks`；状态机固定 TODO→DOING→DONE/CANCELLED，终态不可再变更
- **周期性任务**：改动重复规则（recurrence_*）须同步 `internal/scheduler` 生成逻辑与前端 RecurrenceConfig 组件；子实例用 `parent_recurrence_id` 关联父级
- **软删除**：所有业务查询必须过滤 `deleted=0`；状态变更走 `PUT /{id}/status`
- **测试数据库**：自动测试只允许连独立测试库 `righthand_test`（`database.InitDB` 有 fail-fast guard，见 `docs/ai/rules.md §9`）；新增触库测试必须 `cfg.Database.DBName = database.TestDBName`，禁止改 guard 放行；初始化测试库：`cd righthand-go; go run ./cmd/init-test-db`
