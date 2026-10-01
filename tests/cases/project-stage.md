# 项目阶段（project_stage）后端测试用例矩阵

- 关联需求：项目管理-项目阶段——新增 project_stage 表与阶段 CRUD 接口（`/api/projects/:id/stages`）；「当前阶段」由后端 `resolveCurrentStageItem` 统一计算（区间含两端、重叠取最晚开始、无命中为空、终态项目不计算），以 `currentStageId`（详情）与 `currentStage`（清单）下发
- 对应代码：righthand-go/services/task-service/service/project_service.go、righthand-go/services/task-service/handler/project_handler.go、righthand-go/cmd/server/main.go
- 测试文件：righthand-go/services/task-service/service/project_service_test.go、righthand-go/services/task-service/handler/project_handler_test.go、righthand-go/tests/route_test.go
- 更新日期：2026-09-22

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| Stage_Create_Success | 正例 | PM(user2) 创建阶段「需求调研」start=2026-10-01 end=2026-10-10 | 返回 code=200，stageName/startDate/endDate 与入参一致 | project_stage 新增 1 行（deleted=0）；project_change_log 新增 1 条 STAGE_ADD | 项目 IN_PROGRESS（PM=2）；同项目无同名阶段 | 用例内事务回滚（setupProjectServiceTx） | - |
| Stage_Create_DuplicateName_Rejected | 反例 | 同项目已有「需求调研」，再次创建同名阶段 | 拒绝，业务码 400「阶段名称已存在」 | project_stage 新增 0 行；project_change_log 无 STAGE_ADD | 项目含 1 条同名阶段 | 用例内事务回滚 | - |
| Stage_Create_RecreateAfterSoftDelete_Allowed | 幂等 | 软删「需求调研」后再次创建同名阶段 | 创建成功一次；重名校验只统计未删除范围 | project_stage 累计 2 行（1 行 deleted=1，1 行 deleted=0）；未删除范围内同名仅一次出现 | 项目含 1 条 deleted=1 的同名阶段 | 用例内事务回滚 | - |
| Stage_Create_EndBeforeStart_Rejected | 边界 | end=2026-09-30 早于 start=2026-10-01；对照 end=2026-10-01 与 start 相等 | end<start 拒绝 400「结束日期不能早于开始日期」；end=start 创建成功 | end<start 时新增 0 行；end=start 时新增 1 行 | 项目 IN_PROGRESS | 用例内事务回滚 | - |
| Stage_Create_NonManager_Forbidden_AdminAllowed | 权限 | user3（非成员非管理员）创建；ADMIN(isAdmin=true) 创建 | user3 拒绝 403「仅项目经理可管理阶段」；ADMIN 返回 200 | user3 调用时 project_stage 新增 0 行 | 项目 PM=2，user3 未加入成员 | 用例内事务回滚 | - |
| Stage_Create_TerminalProject_Rejected | 异常 | 项目结项为 CLOSED 后创建阶段 | 拒绝，业务码 400「项目已进入终态，不可维护阶段」 | project_stage 新增 0 行；无 STAGE_ADD 日志 | 项目已流转至 CLOSED | 用例内事务回滚 | - |
| Stage_Update_PartialMerge_And_SelfRename_Allowed | 边界 | 仅传 description；再传 stageName 等于自身原名；再传 stageName 等于同项目他阶段名；再仅传 start 晚于现 end | 前两次返回 200 且未传字段保持不变；改名冲突拒绝 400「阶段名称已存在」；仅改 start 使区间倒置拒绝 400「结束日期不能早于开始日期」 | 成功更新时该行字段变化；两次拒绝时目标行 0 行变更 | 项目含阶段 A 与阶段 B | 用例内事务回滚 | - |
| Stage_Delete_SoftDelete_ThenInvisibleAnd404 | 反例 | 删除阶段后查询 ListStages 与详情；再次删除同一阶段；用他项目的 stageId 调用删除 | 列表与详情不含该阶段；重复删除与他项目 stageId 均 404「阶段不存在」 | project_stage 该行 deleted=1；project_change_log 新增 1 条 STAGE_DELETE | 项目含 1 个阶段，另建一个项目做跨项目隔离 | 用例内事务回滚 | - |
| ResolveCurrentStageItem_InclusiveBoundary_Overlap_None | 边界 | today=2026-06-01 等于 A 阶段 start；today=2026-06-20 等于 A 阶段 end；B(start=2026-06-10) 与 A 重叠且 today=2026-06-15；today=2026-07-01 在全部阶段外；并列 start 的 C1/C2 取 id 最大 | 两端均命中 A；重叠时命中 B（start 最晚）；区间外与空切片返回 nil；并列 start 命中 id 最大者 | 纯函数无落库 | 构造 4 组阶段切片（A:2026-06-01~06-20，B:2026-06-10~06-30，C1/C2 同 start） | 无（纯函数，不落库） | - |
| ProjectDetail_StagesSorted_CurrentStageId | 边界 | 项目详情返回 3 个阶段（乱序创建），其中 1 个跨今天；另有今天不命中的项目与 CLOSED 项目 | stages 按 startDate 正序排列；currentStageId=命中阶段 ID；无命中与终态项目为 null | 无落库（只读） | 阶段日期围绕 time.Now 相对构造 | 用例内事务回滚 | - |
| ListProjects_CurrentStage_AndTerminalEmpty | 边界 | 清单含 3 个项目：进行中含跨今天阶段、进行中无阶段、CLOSED 含跨今天阶段 | 命中项目 currentStage=阶段名；无阶段与终态项目 currentStage 为空串 | 无落库（只读） | 3 个项目及其阶段数据 | 用例内事务回滚 | - |
| StageHandler_InvalidId_BadRequest_RoutesExist | 异常 | 链路：创建项目→新增阶段→列表→编辑（改名）→删除→重复删除；GET /api/projects/abc/stages；PUT /api/projects/1/stages/abc；未认证请求 GET /api/projects/1/stages | 链路各步 code=200（创建成功/保存成功/删除成功）；重复删除 404「阶段不存在」；无效项目 ID 400「无效的项目ID」、无效阶段 ID 400「无效的阶段ID」；未认证返回 HTTP 401 而非 404（路由已注册） | 链路新增 1 行阶段（列表可见）→编辑后字段变化→软删后 deleted=1；无效 ID 请求无落库 | handler 路由注入身份（PM=2）；主路由镜像；项目 STARTING 非终态 | t.Cleanup 物理清理项目/阶段/日志（cleanupProjectRows 扩展 project_stage） | - |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度：不适用——阶段无金额字段与数值运算。
- 状态机：适用，已覆盖——项目终态（CLOSED）不可维护阶段（Stage_Create_TerminalProject_Rejected）；STARTING/IN_PROGRESS/PAUSED 可维护、前端按钮显隐由 allowedActions 覆盖（handler 用例）。
- 权限与越权：适用，已覆盖——非 PM 403、ADMIN 放行、跨项目 stageId 404（Delete/Update 校验 project_id 归属）。
- 时间边界：适用，已覆盖——含两端命中、重叠取最晚开始、区间外为空（ResolveCurrentStageItem 用例）；测试日期围绕 time.Now 相对构造，避免用例过期。
- 分片路由：不适用——无分片。
- 幂等与重复提交：适用，已覆盖——软删后同名重建（Recreated_Allowed）；重复删除返回 404。
- 并发与乐观锁：不适用新增——阶段维护操作低频且仅 PM/ADMIN 单人操作，无版本字段；重名校验在服务层事务内完成。
- 事务回滚：适用——服务层创建/更新/删除均在 db.Transaction 内写主表 + 变更日志，由既有事务基建保证；用例级回滚（setupProjectServiceTx）兜底。
- 空值与可选字段：适用，已覆盖——Update 部分字段合并校验（未传字段保持不变）、Description 可清空。
- 极值与溢出：适用，已覆盖——阶段名称 >100 字符拒绝 400（实现含 rune 长度校验，用例并入 Stage_Create_Success 的对照分支）。
- 中文与 Unicode：适用，已覆盖——阶段名称使用中文（「需求调研」）断言 UTF-8 完整性。
- 分页与排序稳定性：适用，已覆盖——详情 stages 按 start_date 正序 + id 次序（ProjectDetail_StagesSorted 用例）。
- 批量与部分失败：不适用——无批量接口。
- SQL 方言与函数：不适用——仅标准 SQL（WHERE/ORDER BY），无方言函数。

## 补充说明

- 「当前阶段」计算口径统一在服务层 `resolveCurrentStageItem`：区间含两端（start<=today<=end）；重叠取 start_date 最晚（并列取 id 最大）；无命中返回 nil；终态项目（CLOSED/TERMINATED）不计算。
- 用例日期围绕 time.Now 相对构造（跨今天的阶段），清单/详情命中断言不依赖固定日历日期。
