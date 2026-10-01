# 变更记录（CHANGELOG）

> 规范见 `docs/ai/rules.md` §12。每次改动登记：改了什么 / 为什么 / 影响范围。
> 格式参考 Keep a Changelog；条目按 feat / fix / docs / refactor / test / chore 分类。

## [Unreleased]

### fix
- Casbin 启动告警文案修正 + 补充解释描述（测试环境启动日志报 `open configs/model.conf: no such file or directory`，经核实为已知问题、功能零影响）：
  - 背景：`cmd/server/main.go` 初始化失败时输出「权限控制将不可用」，易被误读为权限体系失效；实际 `CasbinMiddleware` 未挂载到任何路由（`docs/ai/rules.md` §5 已记载），权限控制由 JWTAuth + 业务层校验 + 前端权限码（`sys_menu`/`sys_role_menu`）承担，`casbin_rule` 预置策略为预留数据
  - 报错根因（暂不修复）：`InitCasbin` 按相对路径加载 `configs/model.conf`，而发布包未收集该文件（`build-release.ps1` 打包 configs/ 仅含 config.yaml；`install.sh` 亦只同步 config.yaml），致使部署环境（systemd `WorkingDirectory=/opt/righthand`）必然缺文件；本地开发因工作目录恰为 `righthand-go/` 不报错
  - 调整：告警后补充解释行——说明当前不影响功能、给出常见原因（部署目录缺 model.conf／工作目录不正确）及启用前需补齐文件并校准策略（现 matcher 为精确匹配，与含 `*` 通配的既有策略不兼容）
  - 影响范围：仅 `cmd/server/main.go` 启动日志输出；控制流、接口与权限行为零变化
- 自动测试数据库隔离强约束：测试只允许连独立测试库，禁止触碰开发主库（用户报告：本地主库演示数据被前轮自动化测试污染，登录页展示异常）：
  - 根因：多处测试 setup 直连开发主库 `righthand`（tests 包与 project/auth 的 service/handler setup），测试写入/清理动作污染主库数据
  - 机制：`internal/database/mysql.go` 新增 fail-fast guard——测试进程（`testing.Testing()` + 二进制名后缀双判定）连接非 `_test` 结尾库名直接报错拒连；新增 `TestDBName = "righthand_test"` 常量与回归测试 `init_guard_test.go`（3 用例全绿）
  - 工具：新增 `cmd/init-test-db`（将 `sql/init-database.sql`/`init-test-data.sql` 内库目标改写为 `righthand_test` 后执行，规避脚本硬编码 `USE righthand` 误伤主库；用法 `cd righthand-go; go run ./cmd/init-test-db [-seed]`）
  - 改造：7 处触库测试 setup 统一 `cfg.Database.DBName = database.TestDBName`（tests/{route,integration}_test.go、task-service/{service/project_service_test,handler/project_handler_test}.go、internal/scheduler/project_scheduler_test.go、auth-service/{handler/auth_integration_test,service/auth_service_test}.go）
  - 修复 handler 包既有测试缺陷（原为零散 panic、从未整体跑通）：新增 `handler/main_test.go`（TestMain 统一初始化测试库）+ 补 `isAdmin` 身份上下文（16 处）、修复同 router 重复注册路由 panic、断言对齐「HTTP 200 + body.code」响应契约（contact 4 处、task 12 处）；auth `TestUserHandler_GetUser_NotFound` 对齐真实 404；`TestDictHandler_CreateDict_InvalidData` 对齐当前行为并标记待修复缺陷（CreateDict 无参数校验，空数据可入库）
  - 验证：主库 `righthand` 全程零触碰（sys_user AUTO_INCREMENT=3、UPDATE_TIME 保持 16:09:02 基线）；测试库 `righthand_test` 承载全部测试写入；受控范围复跑 internal/database、internal/scheduler、tests、task-service/handler、auth-service/{handler,service} 全绿
  - 遗留说明（已分诊，与本改造无关）：`services/task-service/service` 的 `TestTaskService_*`/`TestContactService_*` 9 用例失败——sqlite 内存库不支持模型标签 `ON UPDATE CURRENT_TIMESTAMP`（不连任何 MySQL 库），建议后续单独任务迁移到 `righthand_test`
  - 影响范围：测试代码/测试工具/文档；生产代码仅 `internal/database/mysql.go` 新增测试进程 guard 与 `TestDBName` 常量（非测试进程行为零变化）；开发/生产库数据不变

### feat
- 项目详情-任务页签新增「负责人 / 任务状态」筛选（用户需求迭代）：
  - 后端：`GET /api/projects/:id/tasks` 新增可选查询参数 `userId` / `statusCode`（接口路径与响应格式不变）；`ListProjectTasks` 签名改为 `(projectID int64, page, size int, filter map[string]interface{})`，handler 改为 filter map 传参（对齐 `ListProjects` 模式）；非法 userId 忽略该条件、未知状态自然返回空列表、分页 total 为过滤后条数
  - 前端：`ProjectTaskTab.vue` 工具栏周过滤后新增负责人（数据源 `props.users`，filterable）与任务状态（`TASK_STATUS` 枚举）两个 clearable 下拉，变更即回第 1 页重新查询；切换项目时重置三个筛选
  - 测试：新增用例矩阵 `tests/cases/project-task-filter.md`（后端 8 行）+ `tests/cases/project-task-filter-ui.md`（前端 6 行），均通过 `validate-cases-matrix.ps1` 机械校验；后端新增 `TestProjectService_ListProjectTasks_Filters` 并更新 4 处旧签名调用；前端 `ProjectTaskTab.test.ts` 新增 6 用例（12/12 全绿）
  - 验证：受控范围 `npx vitest run tests/views/project/ProjectTaskTab.test.ts` 12/12 通过；浏览器真实联调通过——任务页签工具栏渲染「周过滤/负责人/状态」3 个下拉，筛选「负责人=管理员」任务 5→4 条、叠加「状态=进行中」4→1 条（截图 `ui-verify-*.png`）
  - 影响范围：项目管理-任务页签（查询参数扩展）；既有周过滤行为与响应格式不变
- 后端日志按日期切割 + 测试环境日志修复：
  - 根因：`configs/config.yaml` 的 `log.*` 配置从未接入代码、`middleware.Logger()` 为空实现（TODO）、无 logs 目录落盘 → 测试环境看不到日志文件
  - 新增 `internal/logger` 组件（零新依赖）：按 `file_path` 生成 `logs/app-YYYY-MM-DD.log`（跨日写时自动切换新文件）；保留策略按 `max_age` 天数 + `max_backups` 数量清理（当日文件必保留）；`output` 支持 `stdout/file/both`；成功后统一接管 `log.SetOutput` 与 `gin.DefaultWriter/DefaultErrorWriter`；失败降级控制台并告警
  - `middleware.Logger()` 落地：输出 `[HTTP] method path status latency clientIP`（不记录请求体/Token）；`Recovery` 补 `[PANIC]` 上下文日志
  - `cmd/server/main.go`：`gin.Default()` → `gin.New()` + 显式挂载 CORS → Recovery → Logger（修复调度触发路由未挂中间件问题）；启动时 `logger.Init`（失败降级控制台）
  - 配置：`scripts/config.yaml.tpl` 与 `righthand-go/configs/config.yaml` 统一 `log.output: both`（控制台进 journald + 按日文件）、`format: console`（json 预留）；`max_size` 大小切分预留未启用
  - 测试：矩阵 `tests/cases/app-logging.md`（8 行）通过校验；新增 `internal/logger/logger_test.go` 7 用例（file/both/跨日/保留/纯 stdout/非法路径降级/并发写不丢行）、`internal/middleware/logger_test.go`（HTTP 日志字段与敏感信息断言）
  - 验证：本机冒烟 `/health` 200 且 `logs/app-2026-09-23.log` 生成、含启动与 `[HTTP]` 日志行；测试环境查看方式 `journalctl -u righthand -f` + `tail -f /opt/righthand/logs/app-YYYY-MM-DD.log`
  - 影响范围：全部 HTTP 访问日志（按日期文件名落盘）；接口行为与响应格式不变
- 项目详情默认页签调整为「项目整体计划」（用户反馈迭代：点击清单「详情」默认展示该页签）：
  - 前端：`ProjectDetail.vue` 初始进入详情与页内切换项目时 `activeTab` 默认值由 `weeks` 改为 `stages`（页签顺序与各 action 跳转——成员/周计划/任务——逻辑不变）
  - 测试：`ProjectDetail.test.ts` 追加「默认激活『项目整体计划』页签」断言；矩阵 `tests/cases/project-stage-ui.md` 同步更新并重跑机械校验
  - 影响范围：仅前端详情页初始/重置页签值，接口与后端零改动
- 新增「项目管理-项目阶段」功能（详情页「项目整体计划」页签 + 清单「当前阶段」动态列）：
  - 数据库：`init-database.sql` §4.6.6 新增 `project_stage` 表（共 21 张，`idx_project_start` 复合索引）+ 「阶段管理」按钮权限 `project:stage`（ADMIN/USER 授权）+ `project_change_log.action` COMMENT 追加 `STAGE_ADD/STAGE_EDIT/STAGE_DELETE`；新增存量库增量脚本 `righthand-go/sql/upgrade_project_stage.sql`（幂等可重复执行，含完成自检 SELECT；执行：`mysql --default-character-set=utf8mb4 -u root -p < righthand-go/sql/upgrade_project_stage.sql`，含中文禁止 PowerShell 管道；执行前先备份）
  - 后端：新增 `internal/models/project_stage.go`、DTO `StageCreateRequest/StageUpdateRequest`、阶段 CRUD 服务（校验顺序对齐 AddMember/CreateWeek：项目存在 404 → `canActAsManagerOnly` 403 → 终态 400 → 参数/重名校验；同项目未删除范围名称唯一、软删后同名可重建；变更写 STAGE_ADD/EDIT/DELETE 日志）、`projectAllowedActions` 新增 `manageStage`（仅 IN_PROGRESS/PAUSED 且 ADMIN/项目经理）；「当前阶段」由纯函数 `resolveCurrentStageItem` 统一计算（区间含两端、重叠取 `start_date` 最晚并列取 id 最大、终态不计算），详情下发 `stages`+`currentStageId`、清单批量下发 `currentStage`；新增 4 条路由 `/api/projects/:id/stages`（Casbin 既有 `/api/projects/*` 策略已覆盖）
  - 前端：`api/project.js` 新增 4 个接口；新组件 `ProjectStageSection.vue`（按开始时间正序表格、当前阶段行高亮+「当前阶段」标签、新增/编辑弹窗与删除确认、按钮显隐由 `allowedActions.manageStage` 驱动）；详情页「项目整体计划」页签位于「周计划」之前；清单「当前阶段」列位于「进度」与「最近更新」之间（空值显示 -）；`projectEnums.js` 追加 3 个变更动作标签
  - 测试：新增用例矩阵 `tests/cases/project-stage.md`（12 行，后端）+ `tests/cases/project-stage-ui.md`（10 行，前端），均通过 `validate-cases-matrix.ps1` 机械校验；后端新增 11 个服务用例（`resolveCurrentStageItem` 含两端/重叠/无命中/终态边界、重名/软删重建/越权/终态拦截）+ handler 阶段链路用例与无效 ID 400 + route 镜像 4 条；前端新增 `ProjectStageSection.test.ts` 9 用例，更新 `ProjectDetail.test.ts`（页签位于周计划之前、高亮与标签）与 `ProjectList.test.ts`（列位置与空值 -）；受控范围执行 `service/handler/route` 包与 `tests/views/project/` 全绿
  - 同步文档：`docs/ai/rules.md` §2 表清单 20→21、§3 项目管理红线追加阶段表口径（名称唯一、终态不可操作、当前阶段计算口径）
  - 影响范围：项目管理模块（表/权限/接口/前端页面）；周计划/任务/通知链路零改动；存量库需执行上述增量脚本后再部署新版本代码
- 新增「项目管理」功能（M1+M2 核心 + 任务挂周联动 + M3 自动化通知；不含 AI 周报）：
  - 数据库：`init-database.sql` 新增 5 张表（`project`/`project_weekly_plan`/`project_member`/`project_change_log`/`task_week_plan_rel`，共 20 张）+ 菜单「项目管理」目录/清单页/7 个按钮权限 + ADMIN/USER 授权 + casbin `/api/projects*`、`/api/messages*` 策略；`init-test-data.sql` 追加 2 个演示项目与周计划/任务关联/站内信
  - 后端：新增 `internal/models/project*.go` 等 5 个 GORM 模型、`services/task-service/service/{project_service,message_service}.go`、handler 3 个、DTO；路由全部 JWTAuth + 统一 `utils.*` 响应；状态机与 `CanAct` 统一权限函数、`RecalcProjectProgress` 进度只算不填、总结提交锁/解锁留痕、周计划与成员级联软删、`task_item` 零改动（多周关联表方案）；新增调度器 `internal/scheduler/project_scheduler.go`（周总结提醒/项目延期预警/周计划延期/高风险汇总，全部 01:00 执行，`GET /api/scheduler/project/trigger` 手动触发）
  - 前端：新增 `src/views/project/`（ProjectList/ProjectDetail/WeeklyPlanSection/ProjectTaskTab/ProjectFormDialog）、`src/api/project.js`、`src/utils/projectEnums.js`、`/projects` 路由、MainLayout 消息铃铛启用（未读数 + 最近消息 + 已读）；按钮显隐完全由后端 `allowedActions` 驱动
  - 测试：后端新增/更新 `project_service_test.go`、`message_service_test.go`、`project_handler_test.go`、`project_scheduler_test.go`、`tests/route_test.go`；前端新增 `tests/api/project.test.ts`、`tests/views/project/ProjectList.test.ts`、`ProjectDetail.test.ts`（37 用例全绿）；本地库重建导入 + 33 项接口冒烟通过 + 浏览器真实联调验证
  - 同步文档：`docs/ai/rules.md` §2 表清单 15→20、§3 新增「项目管理」红线（状态机/allowedActions、进度只算不填、总结提交锁、`task_week_plan_rel` 多周关联、调度器）
- 项目清单页交互对齐需求管理页（用户反馈迭代）：
  - 按钮样式：操作列由 link 文本按钮改为实心 small 按钮（详情=primary、编辑=success、状态流转=warning、删除/终止=danger），「新建项目」改为 success（对齐「新增需求」）
  - 筛选调整：「我参与的项目」复选框改为「参与人」单选下拉（filterable/clearable，与项目经理筛选框同源用户列表），可筛选任意用户参与的项目；后端查询参数由 `myProjects` 替换为 `participantId`（语义：所选用户为项目经理或项目成员）
  - 影响范围：`righthand-frontend/src/views/project/ProjectList.vue`、`src/api/project.js`、`tests/views/project/ProjectList.test.ts`（12 用例）；`righthand-go/services/task-service/{handler/project_handler.go,service/project_service.go,service/project_service_test.go}`；接口路径与响应格式不变
- 侧边栏「项目管理」菜单结构调整（用户反馈迭代）：
  - 移除原「项目管理」折叠目录层；「项目清单」菜单改名为「项目管理」，作为一级可点击菜单排列在「事项管理」正下方（图标沿用 Management，后续同级菜单顺序顺延）
  - 同步：`init-database.sql` 6.8.4 节重写 + 同级菜单 sort_order 重排（新库直接生效）；新增存量库增量脚本 `sql/adjust_project_menu.sql`（已应用于本地库）；7 项按钮权限与角色授权不变
  - 影响范围：仅菜单数据与 SQL 脚本，无代码变更；接口、路由与页面逻辑不变
- 前端测试基建补齐（原先 `@vue/test-utils`/`happy-dom` 未安装，挂载型测试无法运行）：
  - `tests/setup.ts` 全局注册 Element Plus + Pinia + `v-permission` 指令桩、补 `window.confirm/alert` 默认实现；`vite.config.ts` 测试超时放宽（真实组件挂载较慢）；修复 `Dashboard.test.ts` 的 vue-router 部分 mock
  - 影响范围：仅测试基建与既有测试文件；`npx vitest run` 可正常执行，其余历史用例遗留失败项与本次功能无关
  - 后端遗留失败说明（已分诊，均与本次功能无关）：contact/task handler 老用例 panic（测试上下文未初始化全局 DB / 缺 `isAdmin`）、`service_integration_test.go`（sqlite 不支持模型标签中的 `ON UPDATE CURRENT_TIMESTAMP`）、`internal/utils` 陈旧断言；本次新增用例（project/message service 10 个、project handler 3 个、scheduler、route 路由覆盖）全部通过
- 「关联已有事项」选择弹窗链路完善（用户反馈迭代）：
  - 前端：新增通用组件 `src/components/TaskSelectDialog.vue`（标题/类型/状态/优先级筛选 + 分页单选，打开才加载、关闭态零请求）；`ProjectTaskTab.vue` 关联弹窗接入该组件（只读回显 `#id 标题`），提交按钮 disabled 与 `handleLink` 守卫双保险拦截空参请求
  - 契约修复：选择弹窗按后端分页结构读取 `data.records`/`data.total`，修复首次打开列表恒为空
  - 测试：新增用例矩阵 `tests/cases/task-select-dialog.md`（12 行）与 `tests/cases/project-task-link.md`（6 行），均通过 `validate-cases-matrix.ps1` 机械校验（正例占比 33.3%）；`TaskSelectDialog.test.ts` 10→12 用例（新增 `PageSizeChange_BackToFirstPage`、`Closed_NoRequest`）、`ProjectTaskTab.test.ts` 6 用例（`Link_NoSelection_GuardBlocks` 守卫反例），用例命名与矩阵场景 1:1；4 个相关测试文件 `testTimeout` 统一放宽至 90s（环境变慢致 ProjectList 4 用例 30s 超时，非逻辑回归）
  - 验证：受控范围执行 4 个测试文件 `vitest run` 44 用例全绿（262s）；浏览器真实联调「关联已有事项 → 选择弹窗筛选 → 选择回填 → 关联成功」（列表 3→4 条），并实测第 1 周「周总结已提交锁定」被业务规则正确拦截（截图 `verify_pm_link_dialog.png`、`verify_pm_link_success.png`）
- 用户管理「新增用户」弹窗改造（用户反馈：姓名必输、去掉密码框、初始密码统一默认值）：
  - 前端 `UserManagement.vue`：删除「密码」表单项（新增/编辑弹窗均无密码框）；`formData`/`handleEdit`/`resetForm` 移除 password；rules 新增姓名必填（提示「请输入姓名」）
  - 后端 `auth-service`：新增 `DefaultPassword = "88888888"` 常量（SSOT，CreateUser 与 ResetPassword 共用）；`CreateUser` 忽略调用方传入密码、统一用默认密码 bcrypt 加密；`CreateUserRequest` 移除 `Password` 字段、`RealName` 添加 `binding:"required"`（缺失返回 HTTP 400）
  - 测试：后端新增 `tests/user_create_default_test.go`（矩阵 `tests/cases/user-create-default-password.md`，7 场景含显式密码被忽略、缺姓名 400、bcrypt 落库断言）7/7 通过；前端重写 `tests/views/system/UserManagement.test.ts`（矩阵 `tests/cases/user-add-dialog-validation.md`，6 场景）6/6 通过
  - 测试基建修复 `vite.config.ts`：vitest `server.deps.inline` 内联 element-plus/async-validator，修复 async-validator 无 exports 字段致 Node ESM 解析出 `{default: fn}` 包装对象、`new AsyncValidator()` 抛 TypeError 被 ElForm 静默吞掉、表单必填校验恒通过的问题（必填拦截类断言此前为假阳性）
  - 验证：`vite build` 通过（1729 模块，43.5s）；浏览器真实联调 8 步全过——弹窗无密码框、空提交出现「请输入姓名」/「请输入用户名」且弹窗不关闭、创建成功、新用户以 `88888888` 登录成功（截图 `verify_user_dialog_empty.png`、`verify_user_dialog_validation.png`、`verify_user_dialog_filled.png`、`verify_user_list_created.png`、`verify_user_login_default_pwd.png`）
  - 基线说明：前端全量 227 用例中 30 个失败，经「临时移除 inline 配置」基线对照实验确认为前轮遗留（同批文件对照下同样失败，本轮零新增回归）；对照后联调测试用户已删除，库内用户恢复为 admin/wujp/test
  - 影响范围：`righthand-frontend/{src/views/system/UserManagement.vue,vite.config.ts,tests/views/system/UserManagement.test.ts}`、`righthand-go/services/auth-service/{dto/user_dto.go,handler/user_handler.go,service/user_service.go}`、`righthand-go/tests/user_create_default_test.go`；接口路径与响应格式不变（请求体传入 password 时被忽略，不报错）
- 新增数据库增量升级脚本 `righthand-go/sql/upgrade_project_management.sql`（存量库升级：测试环境先执行脚本、再部署最新版本代码）：
  - 覆盖内容：`sys_message` 新增 `deleted` 列；新建 5 张项目模块表（`project`/`project_weekly_plan`/`project_member`/`project_change_log`/`task_week_plan_rel`）；「项目管理」一级菜单 + 7 个按钮权限 + ADMIN/USER 角色授权 + 同级菜单 `sort_order` 对齐；casbin `/api/projects*`、`/api/messages*` 共 18 条策略；附可选演示数据（默认注释，按需解锁）
  - 幂等设计：`sys_message.deleted` 列存在性判断（information_schema + 动态 SQL）；建表 `IF NOT EXISTS`；菜单/按钮防重守卫；角色授权 `INSERT IGNORE`；casbin 目标范围 DELETE + INSERT「确保存在」语义
  - 验证（本地 MySQL 8.0 真库双库对照）：HEAD 基线库执行增量脚本 vs 最新 `init-database.sql` 全量初始化，8 个维度（表清单/列定义/索引/一级菜单/全部菜单/项目菜单/角色授权/casbin）逐行一致；幂等复测二次执行无错、无重复数据、对比仍全绿
  - 实操注意：Windows mysql 客户端默认字符集为 gbk，导入含中文脚本必须加 `--default-character-set=utf8mb4`（否则中文字节按 GBK 错位解析，触发 `Data too long` 等异常）；用户管理改造（姓名必填/默认密码 88888888）为应用层改动，无 SQL 变更
  - 影响范围：仅新增 SQL 脚本文件，无代码变更

### fix
- 修复 `tests/route_test.go` `TestInvalidUserID` 既有失败（期望 400 实得 401）：
  - 根因：测试路由在业务路由组前置挂载 `JWTAuth`，请求未携带 Token 被认证中间件拦截返回 401，无法触达「无效 userId 返回 400」的目标断言路径
  - 修复：用例携带 `utils.GenerateToken` 生成的合法 Bearer Token（保留原 400 断言语义不变）
  - 影响范围：仅测试文件，无生产代码改动
- 修复需求管理「非本人需求点击查看报错」：
  - 根因：列表/导出接口无用户过滤（全团队可见），而详情接口 `GetRequirementByID` 按创建人/跟进人隔离，非本人需求详情返回 404，造成「列表可见、点击查看报错」
  - 修复：明确需求为团队共享视图，`GetRequirementByID` 移除隔离条件（签名简化为 `(id int64)`，所有登录用户可查看任意需求详情，软删除不可见）；列表/导出维持全量可见；编辑/删除限制不变
  - 同步调整：`docs/ai/rules.md` §5 数据隔离条款（需求改为共享视图）；`requirement_integration_test.go` 以共享视图断言替代原 `DataIsolation` 用例（新增 `GetDetailShared` / `SharedView`，矩阵 `tests/cases/requirement-shared-view.md`），并补齐 13 处过时签名调用（ListRequirements/ExportRequirements 参数缺失）
  - 附带修复：`UpdateStatus` 完成态取时间改用注入库 `s.getDB()`（原全局 `database.GetDB()` 在测试环境未初始化会 panic）；测试断言 `retrieved.Deleted` 改为解引用比较（`Deleted` 为 `*int`）
  - 影响范围：`righthand-go/services/task-service/{service,handler}` 需求模块；接口路径与响应格式不变，前端无需改动

### docs
- 全局约束更新：部署链路修正 + 日志组件约定：
  - `docs/ai/rules.md`：§1 补日志组件约定（`internal/logger` 按日切割、`output: stdout/file/both`）；§11 改写为真实发布链路红线（Windows `scripts/build-release.ps1` 打包 → 手工上传 → `sudo ./install.sh`，systemd + Nginx 单机）并固化「Docker 部署已封存、不再维护与适配」
  - `AGENTS.md`：补唯一部署链路与 Docker 封存说明，避免后续会话再走 Docker 链路
  - `docs/私人管家-部署指南.md`：§4.3 目录树日志名改 `app-YYYY-MM-DD.log`、§4.8 日志查看命令、§4.10 模板对比表 `log.output`、§六新增「查看日志」排查小节
  - 影响范围：仅文档，无代码变更
- 建立 AI 执行层全局约束体系（参考 IPMS 积分项目约束方式，整合 `docs/` 五大设计文档）：
  - 新增 `/AGENTS.md`：AI 新会话定向入口（项目定位、服务端口、目录地图、技术栈、构建测试、SSOT、关键操作规范）
  - 新增 `docs/ai/rules.md`：工程契约红线 14 类（架构分层、数据库、状态机、周期任务、权限、接口响应、前端、UTF-8 编码、测试交付、环境、部署、文档治理、AI 行为）
  - 新增 `docs/ai/README.md`：AI 执行层索引（读写时机、阅读顺序、引用关系、冷热分离维护约定）
  - 新增本文件 `CHANGELOG.md`：变更登记机制
- 影响范围：仅文档，无代码变更；`docs/私人管家-全局开发规范.md` 保留为历史文档（冷），与 `rules.md` 冲突时以 `rules.md` + 代码为准
