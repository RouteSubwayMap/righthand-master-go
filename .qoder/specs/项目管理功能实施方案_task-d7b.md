# 项目管理功能实施方案（M1+M2 + 任务挂周 + M3 自动化）

## 摘要
- 范围：项目清单页、项目详情页（周计划时间轴 / 周总结 / 自动进度 / 成员管理 / 变更历史 / 任务 Tab）、`task_item` 挂周、M3 每日自动化通知（站内信）；**不含 AI 周报**。
- 遵循 `docs/ai/rules.md` 红线：直接改 `sql/init-database.sql`（不新增 add_*.sql）、`deleted TINYINT` 软删除、`create_time/update_time/create_user/update_user` 审计字段、统一 `utils.Success/Error/ErrorWithCode` 响应、权限三件套。

## 一、对 test.md 设计稿的修正（决策已定）
1. 字段规范对齐：`created_at/updated_at/deleted_at` → `create_time/update_time/create_user/update_user` + `deleted TINYINT DEFAULT 0`
2. `risk_level` 用字符串枚举 HIGH/MEDIUM/LOW（复用现有 `risk_level` 字典，与 requirement 同口径，废弃设计稿的 TINYINT 1/2/3）
3. `delivery_status` 去掉「已终止」（与 status=TERMINATED 重复）：NORMAL / AT_RISK / CANNOT_FINISH
4. 项目状态字符串枚举：STARTING / IN_PROGRESS / PAUSED / TERMINATED(终态) / CLOSED(终态)；状态→操作映射为后端唯一权威（常量 + 接口返回 `allowedActions`），前端只渲染文案，不复制状态逻辑
5. 进度公式归一：`progress_pct = 已完成周数 ÷ 非跳过周计划数 × 100`（SKIPPED 不计入分子分母；0 周计划时 = 0），仅由服务端 `RecalcProjectProgress(projectID)` 计算（冗余字段 `week_count/done_week_count` 同步），周计划增/删/状态变更与总结解锁时触发，禁止任何手动写 progress
6. 周计划「已延期」不落库：查询时后端按 `end_date < today AND status IN (NOT_STARTED, IN_PROGRESS)` 计算返回 `overdue` 布尔，前端显示「已延期」标签；落库标记仅由 M3 job 处理项目级
7. `task_change_log` 不存在 → 新建轻量 `project_change_log`（仅记项目级关键变更：创建/状态流转/风险变更/成员变动/结项终止/总结解锁）
8. `project_key` 不用 DB 唯一约束（软删除后 key 会误占用），改服务层在 `deleted=0` 范围内校验唯一
9. 删除项目仅允许 STARTING 状态（防误删），级联软删周计划 + 成员（事务内），已提交总结的周禁止改/删计划
10. 总结提交即锁（`summary_at` 非空即锁定）；PM/ADMIN 可解锁（`summary_at` 置 NULL，原 summary 保留可编辑，写 change_log 留痕）

## 二、数据库（改 `righthand-go/sql/init-database.sql`）

新增 4 张表（均 utf8mb4、软删除、审计字段、索引按最左前缀）：
- `project`：project_name、project_key、description、plan_desc、manager_id、start_date、end_date、status(默认 STARTING)、risk_level(默认 MEDIUM)、delivery_status(默认 NORMAL)、risk_note、progress_pct、week_count、done_week_count、paused_at、closed_at + 索引 idx_manager / idx_status_risk / idx_date_range
- `project_weekly_plan`：project_id、week_no(项目内自增)、title、plan_desc、start_date、end_date、owner_id、status(NOT_STARTED/IN_PROGRESS/DONE/SKIPPED)、remark、summary(JSON TEXT)、summary_by、summary_at、sort_order、deleted + 索引(project_id,sort_order) / idx_owner_status / idx_date；week_no 由服务端按项目内 max+1 生成，不用 YEARWEEK
- `project_member`：project_id、user_id、member_role(PM/MEMBER)、deleted + 索引(project_id,user_id)；服务层去重
- `project_change_log`：project_id、action、detail、operator_id、create_time + 索引(project_id,create_time)

修改：
- `task_item` 加 `week_plan_id BIGINT DEFAULT NULL COMMENT '归属周计划（NULL=未排期）'` + `INDEX idx_week_plan (week_plan_id)`（同步 `internal/models/task_item.go`）
- sys_menu：新增目录「项目管理」(menu_type=0, path=/projects, icon=Management, sort_order=8, product_code=auth) → 子菜单「项目清单」(menu_type=1, path=/projects, component=project/ProjectList) + 7 个按钮权限：`project:add` / `project:edit` / `project:delete` / `project:status` / `project:member` / `project:week` / `project:summary`；按现有格式给 ADMIN(role_id=1) 与 USER(role_id=2) 授权 sys_role_menu
- casbin_rule：`/api/projects*`、`/api/messages*` 的 admin/user 策略（随 7.x 区域追加）
- `init-test-data.sql`：追加 2 个演示项目（1 进行中 + 1 启动）与若干周计划，便于联调
- 现有开发库升级方式：重新执行 `mysql < init-database.sql` + `init-test-data.sql`（DROP 重建）

## 三、后端实现（`righthand-go/`，遵循 Handler → Service → Model 现状风格）

新增文件：
- `internal/models/project.go`、`project_weekly_plan.go`、`project_member.go`、`project_change_log.go`（GORM 模型，实现 TableName()，JSON tag camelCase）
- `internal/scheduler/project_scheduler.go`：独立调度器，3 个 cron Job + `ManualTrigger()`；在 `cmd/server/main.go` 启动；暴露 `GET /api/scheduler/project/trigger`（仿现有 /api/scheduler/trigger）
- `services/task-service/dto/project_dto.go`：创建/更新/状态流转/风险变更/周计划/总结/成员 请求 DTO
- `services/task-service/service/project_service.go`：核心服务
  - 状态机常量：`allowedTransitions`（STARTING→IN_PROGRESS；IN_PROGRESS→PAUSED/CLOSED/TERMINATED；PAUSED→IN_PROGRESS/TERMINATED；终态不可再变）+ 操作语义（开始/暂停记 paused_at/恢复/结项/终止必填 reason 写入 risk_note 并记 closed_at）
  - `CanAct(projectID, userID, isAdmin, action)` 统一权限函数（ADMIN 放行；项目查看所有人；编辑/删除/流转=manager 或 create_user；周计划/成员=PM 或 ADMIN；总结=该周 owner 或 PM）
  - `RecalcProjectProgress`、周计划 CRUD（week_no 自增、总结锁校验）、总结提交/解锁、成员增删（去重）、变更日志写入、级联软删（事务）
  - 列表查询一次聚合经理姓名（JOIN sys_user）+ week_count/done_week_count，避免 N+1；支持筛选：状态多选/风险/经理/日期范围/我参与；默认按 end_date 升序
- `services/task-service/service/message_service.go`：`CreateMessage`（内部调用，写 sys_message，type=PROJECT，business_id=项目ID）+ 查询/未读数/标记已读
- `services/task-service/handler/project_handler.go`（项目+成员）、`handler/project_week_handler.go`（周计划+总结）、`handler/message_handler.go`

路由注册（`cmd/server/main.go`，全部挂 JWTAuth，响应统一 utils.*）：
- `GET/POST /api/projects`、`GET/PUT/DELETE /api/projects/:id`、`PUT /api/projects/:id/status`、`PUT /api/projects/:id/risk`
- `GET/POST /api/projects/:id/members`、`DELETE /api/projects/:id/members/:memberId`
- `POST /api/projects/:id/weeks`、`PUT/DELETE /api/projects/:id/weeks/:weekId`、`PUT .../weeks/:weekId/status`、`POST .../weeks/:weekId/summary`、`PUT .../weeks/:weekId/summary/unlock`
- `GET /api/projects/:id/tasks`（团队可见，聚合本项目周计划下的任务，支持 weekPlanId 过滤，仿 ListTasksByRequirementID 模式）、`GET/PUT /api/projects/:id/logs` 合并进详情接口返回
- `GET /api/messages`、`GET /api/messages/unread-count`、`PUT /api/messages/:id/read`、`PUT /api/messages/read-all`
- `PUT /api/tasks` 体系扩展：ListTasks filter 支持 `weekPlanId`；创建/更新事项接受 weekPlanId
- 详情接口返回 `allowedActions`（按状态+当前用户计算），列表接口返回每行 `allowedActions` 供按钮渲染

M3 调度器 Job（确定性 SQL，幂等策略：仅状态变更时推一次通知）：
1. 周总结提醒：cron 每周五 18:00；`status=IN_PROGRESS AND 今天落在某周 [start,end] 内 AND 该周 summary_at IS NULL` → 站内信给 owner
2. 延期预警：每日 09:00；项目级 `end_date < 今天 AND status NOT IN (CLOSED,TERMINATED) AND delivery_status != AT_RISK` → 置 AT_RISK + 通知 manager + 记 change_log；周级 `end_date < 今天 AND status IN (NOT_STARTED,IN_PROGRESS)` → 通知 owner 与 manager（不落库）
3. 高风险汇总：每日 09:00；`risk_level=HIGH OR delivery_status=CANNOT_FINISH`（未终结）汇总一条 → 发给所有 ADMIN 角色用户

## 四、前端实现（`righthand-frontend/`）

新增目录 `src/views/project/`（功能页面统一归入「项目管理」目录）：
- `ProjectList.vue`：筛选器（状态多选/风险/经理/日期范围/我参与）、默认 end_date 升序可切换、列（名称、起止+工期天数、经理、状态标签、风险色标、进度条、最近更新）、动态操作按钮（渲染后端 allowedActions：开始/暂停/恢复/结项/终止/编辑/删除）、新建/编辑弹窗（名称+key+起止+经理+初始状态）
- `ProjectDetail.vue`：头部信息卡（名称/描述/计划/起止/经理/状态/风险内联修改+原因）、进度概览（X/Y 周已完成 + 延期天数）、Tabs：周计划时间轴 / 任务 / 成员 / 变更历史
- `components/WeeklyPlanSection.vue`：纵向时间轴卡片（起止、负责人、状态标签、已延期标记）、新增/编辑/删除周计划、状态流转、总结弹窗（完成/下周/风险/所需支持 四项结构化表单，提交锁死展示 summary_by/at，PM/ADMIN 解锁按钮）
- `components/ProjectTaskTab.vue`：项目任务列表（按周分组/过滤）、新增任务弹窗（预置 weekPlanId）、移动任务到其他周、跳转 TaskDetail
- `components/ProjectFormDialog.vue`：项目新建/编辑、成员管理弹窗
- `src/api/project.js`（全部项目接口）、扩展 `src/api/message.js`
- `src/utils/projectEnums.js`：状态/风险/交付状态/周状态 → 中文标签与色值映射（仅展示用，操作能力由后端 allowedActions 决定）
- `src/router/index.js`：`/projects` 父路由（MainLayout，requiresAuth）+ ''（列表）+ ':id/detail'（详情），懒加载
- `src/layouts/MainLayout.vue`：启用顶栏消息铃铛（未读数 badge + 下拉近 10 条 + 标记已读/全部已读，复用现有被注释的 UI 骨架）

## 五、测试计划

后端（`go build ./...` + 相关包 go test；sqlite 内存库单测需 CGO 环境）：
- `service/project_service_test.go`、`handler/project_handler_test.go`、`service/message_service_test.go`、更新 `tests/route_test.go` 路由覆盖
- 用例矩阵（正例占比 ≤40%，覆盖反例/边界/权限/幂等/UTF-8 中文）：
  1. 进度自动计算：5 周 3 完成=60；手动 PUT progress 被拒绝（反例）
  2. 非法状态跳转（STARTING→CLOSED 400）、终态再流转 400
  3. 非 PM 改/删周计划 403；跨项目 weekId 404
  4. 已提交总结后改/删计划 400；解锁后再改 200（边界）
  5. 终止必填原因（空 400）；结项记 closed_at（正例）
  6. 级联软删：删项目后周计划/成员 deleted=1（事务）
  7. project_key 重复 400（反例）；成员重复添加 幂等
  8. 总结重复提交 400（幂等）；summary JSON 结构正确
  9. 列表筛选（状态多选/我参与）与默认排序正确（正例）
  10. job：过期项目标 AT_RISK 仅一次 + 通知落库 + change_log；高风险汇总发给 ADMIN
- 前端 vitest：`tests/views/project/ProjectList.test.ts`、`ProjectDetail.test.ts`、`tests/api/project.test.ts`（列表渲染、按钮按 allowedActions 显隐、总结锁定态、周计划渲染）

验收命令：`cd righthand-go; go build ./...; go test ./services/task-service/... ./internal/...`；`cd righthand-frontend; npx vitest run; npm run build`

## 六、文档与收尾
- 更新 `docs/ai/rules.md`：表清单 15→19 张；新增「项目管理」红线段（状态机、进度只算不填、总结提交锁、SKIPPED 不计分母）
- `CHANGELOG.md` 登记（改了什么/为什么/影响范围）
- 开发库重新导入 init-database.sql 后冒烟：admin/admin123 登录 → 菜单出现「项目管理」→ 建项目 → 加周计划 → 提总结 → 进度自动变化 → 触发 `/api/scheduler/project/trigger` 验证通知

## 七、假设与边界
- 不做 AI 周报（本期）；周计划「已延期」为展示态；`delivery_status=CANNOT_FINISH` 允许编辑项目时手动置位，job 不覆盖
- 「我参与的项目」= manager_id 或 project_member 成员（ADMIN 看全部）；项目详情团队共享可见（同需求视图模式）
- 项目删除仅 STARTING；TEST/DONE 状态周计划不可删
