# 私人管家工程契约（rules.md）

> AI 每次新会话必读。本文件优先级高于任何历史文档；与传统文档（`docs/*.md`）冲突时，以本文件 + 代码现状为准。
> 定位见 `AGENTS.md`。红线随项目演进更新，变更需同步更新本文件并在 `CHANGELOG.md` 登记。

## 0. SSOT（唯一事实来源）
- 表结构 → `righthand-go/sql/init-database.sql`（唯一权威；测试数据 → `init-test-data.sql`，生产禁用）
- 接口响应格式 → `righthand-go/internal/utils/response.go`
- 路由清单 → `righthand-go/cmd/server/main.go`
- 版本号 → `VERSION`
- 设计文档涉及表结构/接口细节 → 引用代码与 SSOT，不复制粘贴；细节冲突时以代码为准

## 1. 架构与分层
- 单体服务：所有业务模块运行在单一进程；入口 `righthand-go/cmd/server/main.go`，全部路由在此统一注册，不在 handler 包内私挂
- 业务模块放 `services/{auth-service,task-service}`；跨模块公共设施放 `internal/`
- 分层固定：Handler → Service → Repository → Model；Handler 只做参数解析/校验与响应，不写业务逻辑；调用只允许相邻层
- 中间件：`internal/middleware/auth.go`（JWTAuth/CORS/Logger/Recovery）、`casbin.go`（InitCasbin/CasbinMiddleware）
- 日志：应用日志由 `internal/logger` 统一按日期切割输出 `logs/app-YYYY-MM-DD.log`（跨日自动切换）；`log.output` 支持 `stdout/file/both`（默认 both：控制台 + 按日文件）；HTTP 访问日志由 `middleware.Logger()` 输出，禁止记录请求体/Token/密码
- 全局中间件顺序固定：CORS → Recovery → Logger；所有业务路由组必须挂 `JWTAuth`
- 依赖版本以 `righthand-go/go.mod` 为准（Go 1.24.8 锁定）；升级依赖需单独说明理由

## 2. 数据库与建模
- 表结构变更：直接改 `righthand-go/sql/init-database.sql`；不再新增 `add_*.sql`/`migration_*.sql` 独立脚本（历史增量脚本仅作记录）
- 含中文的 SQL 文件：读写必须 UTF-8；导入用 `mysql < file` 或 `SOURCE`，禁止 PowerShell 管道（防乱码）
- 字符集一律 utf8mb4；连接串必须含 `charset=utf8mb4&parseTime=true&loc=Local`
- 表命名小写+下划线、不加业务前缀（如 `task_item`，非 `task_xxx_item`）；避免复数
- 字段规范：主键 `id` BIGINT 自增；外键 `{table}_id`；时间 `create_time`/`update_time`（DATETIME）；审计 `create_user`/`update_user`；软删除 `deleted` TINYINT 默认 0
- 业务查询必须过滤 `deleted=0`；外键与高频查询字段建索引，复合索引遵循最左前缀
- GORM 模型必须实现 `TableName()`；JSON tag 用 camelCase；事务边界在 Service 层
- 现有 21 张表：`sys_user`/`sys_role`/`sys_user_role`/`sys_menu`/`sys_role_menu`/`sys_dict`/`sys_message`、`contact`、`task_item`、`requirement`、`task_requirement_rel`、`bank_param`、`batch_task`、`prod_node`、`casbin_rule`、`project`、`project_weekly_plan`、`project_member`、`project_change_log`、`project_stage`、`task_week_plan_rel`
- Casbin 策略表 `casbin_rule` 的预置策略随 init-database.sql 维护，策略变更须同步脚本

## 3. 统一事项与状态机
- 任务/备忘/需求在 `task_item` 统一建模，用 `item_type` 区分：TASK/MEMO/REQUIREMENT
- 状态机固定：TODO → DOING → DONE/CANCELLED；DONE/CANCELLED 为终态不可再变更；禁止新增状态或跳级流转
- 优先级 HIGH/MEDIUM/LOW；进度 0-100
- API 统一 `/api/tasks`；禁止恢复 `/api/memo/*`、`/api/task/*` 旧路径
- 需求（`requirement` 表）独立建模：状态 PENDING_CLARIFY(待澄清)/PENDING_SCHEDULING(待排期)/DEVELOPING(开发中)/TESTING(测试中)/PENDING_RELEASE(待投产)/COMPLETED(已完成)；与事项的关联走 `task_requirement_rel`

### 项目管理（`project` 领域）
- 状态机：项目 STARTING→IN_PROGRESS→PAUSED/CLOSED/TERMINATED（CLOSED/TERMINATED 为终态不可再变）；周计划 NOT_STARTED→IN_PROGRESS→DONE/SKIPPED；可用操作由后端按状态+权限计算 `allowedActions` 返回，前端只渲染不复制状态逻辑；删除项目仅限 STARTING（软删+级联事务）
- **进度只算不填**：`progress_pct = 已完成周数 ÷ 非跳过周数 × 100`（SKIPPED 不计入分子分母，0 周计划时=0），仅由服务端 `RecalcProjectProgress` 在周计划增删/状态变更/总结解锁时触发，禁止任何手动写 progress
- **总结提交即锁定**：`summary_at` 非空即锁，该周计划禁止改/删；仅 PM/ADMIN 可解锁（置空 summary_at，原内容保留可编辑，写 `project_change_log` 留痕）
- 跨周任务走多周关联表 `task_week_plan_rel`（物理删除、唯一键幂等），`task_item` 零改动，禁止给 `task_item` 加项目/周字段
- M3 调度器 `internal/scheduler/project_scheduler.go` 全部 Job 凌晨 01:00 执行（cron `0 0 1 * * *` / 周五 `0 0 1 * * 5`），通知走站内信 `sys_message`（type=PROJECT）；调试用 `GET /api/scheduler/project/trigger`
- 项目阶段（`project_stage`）：同一项目未删除范围内阶段名称唯一（服务层校验，软删后同名可重建）；新增/编辑/删除与成员/周计划卡片同一口径（`allowedActions.manageStage` 驱动前端显隐，后端校验顺序：项目存在 404 → ADMIN/项目经理 403 → 终态项目 400），变更写 `project_change_log`（STAGE_ADD/STAGE_EDIT/STAGE_DELETE）；「当前阶段」由服务端纯函数 `resolveCurrentStageItem` 统一计算（`start<=今天<=end` 含两端，重叠取 `start_date` 最晚、并列取 id 最大，终态项目不计算），详情经 `currentStageId`、清单经 `currentStage` 下发，前端禁止复制日期判定逻辑

## 4. 周期性任务
- 重复类型：NONE/DAILY/WEEKLY/MONTHLY/YEARLY；结束条件 NEVER/DATE/COUNT；规则存 `recurrence_rule`（JSON）
- 调度器 `internal/scheduler/recurrence_scheduler.go` 每日检查并生成实例；新实例继承父级属性并用 `parent_recurrence_id` 关联
- 改动生成/规则逻辑必须同步四处：模型、调度器、前端 `RecurrenceConfig` 组件、相关测试
- 调试可用 `GET /api/scheduler/trigger` 手动触发（仅测试用途）

## 5. 认证与权限
- 认证：JWT（HS256），前端存 `righthand_token`，请求头 `Authorization: Bearer <token>`；过期时间以 config.yaml 为准
- 密码 bcrypt 加密存储（DefaultCost）；任何场景禁止明文存储/传输密码
- 权限数据源：`sys_user_role`（用户-角色）+ `sys_role_menu`（角色-菜单）；菜单类型 1 目录 / 2 菜单 / 3 按钮
- 前端：`utils/permission.js` 的 `hasPermission`（ADMIN 直接放行，其余按权限集合匹配）+ `v-permission` 指令控制显隐；权限码格式 `模块:资源:操作`
- 后端：业务路由统一 `JWTAuth` 认证；Casbin 已集成（Enforcer + `casbin_rule` 表），`CasbinMiddleware` 已实现但当前未挂载到路由——如启用须同步补齐策略并做全量回归
- 新增页面/按钮必须三件套齐全：init-database.sql 注册 `sys_menu`（一按钮一权限码）+ 对应角色授权 + 前端显隐控制
- 用户数据隔离：普通用户只能访问自己的事项/联系人数据；管理员可查看全部
- 需求为团队共享视图：所有登录用户可查看任意需求的列表与详情；编辑仅限需求创建人或跟进人（管理员不限），删除仅限创建人（管理员不限）

## 6. 后端接口与响应
- 统一响应 `{code, message, data}`；分页 `{total, pageNum, pageSize, list}`（以 `utils/response.go` 为准）
- 响应函数：`Success` / `SuccessWithMessage` / `Error(c,msg)`（固定 code=500，只收 2 参数）/ `ErrorWithCode(c,code,msg)` / `NotFound` / `Unauthorized` / `Forbidden` / `BadRequest`；业务接口禁止自拼 `gin.H` 响应（仅 `/health`、调度触发等运维端点例外）
- 禁止 `utils.Error(c, 400, "msg")` 式错误调用；自定义错误码必须用 `ErrorWithCode`
- RESTful 约定：GET 列表/详情、POST 创建、PUT 更新、DELETE 删除；状态变更走 `PUT /{id}/status`
- 错误不得吞掉：`if err != nil` 必须处理并返回明确信息；日志禁止记录密码/Token；错误日志含上下文
- `/health` 健康检查语义不可改动（部署验证依赖）
- 批量导入/导出接口（requirement/prod-node/bank-param/batch-task 的 import/export）须同步维护模板与校验逻辑

## 7. 前端规范
- Vue3 Composition API（`<script setup>`）；组件 PascalCase；Props 明确类型与默认值；样式 scoped SCSS
- API 调用统一走 `src/api/*` 模块，禁止组件内直接用 axios；axios 封装 `utils/request.js`（自动附带 token、401 清 token 跳登录、路由切换取消请求）
- 新增页面：`views/` 建组件 + `router/index.js` 注册（懒加载）+ 菜单权限码与后端一致
- 路由守卫负责：token 校验、菜单加载、`requiredPermission` 校验；新增受控页面须配置 meta
- 状态管理用 Pinia（`stores/user`、`stores/dict`）；字典数据优先走 dict store 缓存
- 图标/组件用 Element Plus；构建 Vite，别名 `@` → `src`；开发代理 `/api` → :9080
- 大数据量列表必须分页；请求防抖/节流

## 8. 中文与 UTF-8 编码（重要）
- 数据库/表/字段一律 utf8mb4；HTTP 请求与响应 `Content-Type: application/json; charset=utf-8`
- PowerShell 调 API 创建中文数据必须显式 UTF-8 转换（`[System.Text.Encoding]::UTF8` 转字节再转字符串），禁止直接管道传 JSON
- 含中文的 SQL 用 `mysql < file` 导入或 `SOURCE`，禁止 `Get-Content file | mysql`
- 涉及中文字符的功能必须保留 UTF-8 编码测试（`internal/utils/utf8_encoding_test.go` 等），禁止删除
- 乱码排查顺序：① 验证 API 返回字节（中文每字符 3 字节）→ ② 查库 `HEX()` → ③ 查库/表字符集
- GORM/Gin/Axios 原生处理 UTF-8，无需特殊转码；问题根因通常在 PowerShell/导入方式

## 9. 测试与交付
- 顺序固定：编译通过 → 单元测试 → 集成测试；未通过不得声称完成
- **自动测试数据库强约束（红线）**：自动测试凡涉及数据库，只允许连接独立测试库 `righthand_test`，严禁连接开发/生产主库 `righthand`
  - 机制：`internal/database.InitDB` 内置 fail-fast guard——测试进程连接非 `_test` 结尾库名时直接报错拒连；`init_guard_test.go` 为其回归测试，禁止移除、绕过或放行
  - 触库测试 setup 必须显式切换：`cfg.Database.DBName = database.TestDBName`；包内多个测试共用全局 DB 时用 `TestMain` 统一初始化（参考 `services/task-service/handler/main_test.go`）
  - 测试库初始化/重建：`cd righthand-go; go run ./cmd/init-test-db`（`-seed` 一并导入演示业务数据）；该工具自动把 SQL 脚本内的库目标改写为 `righthand_test`（脚本含硬编码 `USE righthand`，禁止直接 source 到测试库）
- 后端：testify 断言；测试文件 `{module}_test.go`，函数 `Test{FunctionName}`；覆盖率目标 > 70%
- sqlite 内存库（需 CGO）可用于不依赖 MySQL 特有语法/方言的纯逻辑单测；凡涉及 SQL 语义、方言、约束、迁移的测试必须连测试库（反面案例：`service_integration_test.go` 的 sqlite 迁移因模型标签 `ON UPDATE CURRENT_TIMESTAMP` 失败）
- 前端：vitest + happy-dom（`tests/` 目录）；新增组件/工具函数须补测试
- 改动必须同步更新相关测试；测试失败必须分析并修复，禁止跳过、注释、删除失败用例
- 后端验收：`go build ./...` + 相关包 `go test ./...`；前端验收：`npx vitest run`
- 测试数据含中文字符并验证正确性；边界场景（空串、emoji、超长文本）须覆盖

## 10. 环境与命令
- Windows PowerShell：命令分隔用 `;`，不用 `&&`
- Go 版本严格执行 1.24.8（go.mod 已锁定），不得降级或模糊写"1.24"
- GOPROXY=https://goproxy.cn,direct；GOSUMDB=sum.golang.google.cn；官方源失败先用国内镜像
- CGO/sqlite 测试环境：`$env:CGO_ENABLED="1"` + `$env:Path="C:\msys64\mingw64\bin;"+$env:Path`
- 端口：后端 9080、前端 5173（dev）/80（prod）、MySQL 3306；以 `configs/config.yaml` 为准
- 数据库连接池：MaxOpenConns=100 / MaxIdleConns=10 / ConnMaxLifetime=3600

## 11. 部署与发布
- **唯一发布链路（红线）**：Windows 执行 `scripts/build-release.ps1` 打包（读 VERSION + commit 交叉编译 Linux amd64，用 `scripts/config.yaml.tpl` 生成包内 `configs/config.yaml`）→ 手工上传测试服务器 → `sudo ./install.sh`（systemd + Nginx 单机部署，目录 `/opt/righthand`）；安装后验证 `/`、`/health`、`/api/auth/login`
- **Docker 部署方式已封存**：相关文件（Dockerfile、docker-compose*.yml、deploy*.ps1、deploy.sh）保留不删，不再维护、不再适配
- 日志查看（`output: both`）：`journalctl -u righthand -f`（控制台）+ `tail -f /opt/righthand/logs/app-YYYY-MM-DD.log`（按日期文件）
- 生产必改项：`database.password`、`jwt.secret`、`cors.allow_origins`；`server.mode=release`
- 部署/排障优先参考 `docs/私人管家-部署指南.md`（Nginx 404、CRLF 换行、端口冲突等已收录）

## 12. 文档维护治理（冷热分离）
- **热文档**（每次改动实时更新）：代码 + 测试、`init-database.sql`、`rules.md`（红线）、`CHANGELOG.md`（登记）
- **冷文档**（里程碑/交付节点批量更新）：`docs/` 下设计文档（需求说明书、整体架构设计、详细方案设计、部署指南、全局开发规范）
- **SSOT 单一归属**：每类事实只有一个权威源，其余只引用不复制——表结构→`init-database.sql`；红线→`rules.md`；变更历史→`CHANGELOG.md`
- **事件驱动**：只更新被本次改动真正影响的文档，不做无关全量刷新
- **每次改动顺序（热）**：① 改代码 + 同步测试 → ② 涉表结构改 `init-database.sql` → ③ 涉红线改 `rules.md` → ④ `CHANGELOG.md` 登记；`go build` + 相关测试通过为完成
- **冲突处理**：`rules.md` + 代码现状 > `docs/` 历史文档
- **命名约定**：核心文档沿用"私人管家-"前缀；AI 执行层文件放 `docs/ai/`（`README.md`、`rules.md` 保留英文名）

## 13. AI 行为约束
- 优先编辑现有文件（SearchReplace）；不创建不必要的文件；不删除用户未要求的文件
- 不破坏现有功能；不随意重构无关代码；保持与现有代码风格一致
- 改动前先读相关代码与涉及的红线；重大操作（删数据、改表结构、改公共契约）先询问用户确认
- 中文回复，简洁直接，给出可执行的下一步；禁止输出 CODE_EDIT_BLOCK 等中间产物
- 测试未完成不得声称功能正常；不确定的结论必须验证后再给
