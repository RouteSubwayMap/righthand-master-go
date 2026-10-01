# 跨多周周计划显示适配

## 背景与目标
- 现状：`week_no` 为项目内序号（每条 +1），起止日期可自由跨多周（如 2026-06-01 ~ 2026-07-03 跨 5 周），但所有位置仍显示"第N周"，与实际起止时间不符。
- 目标：跨度 ≤7 天显示「第N周」（现状不变）；>7 天显示「第N周 - 第M周」并在卡片突出日期范围；所有展示位置统一走同一标签逻辑；表结构不改，任务侧（deadline、关联周逻辑）不动。

## 核心规则
- span（跨周数）= ceil((end - start + 1) / 7)，最小 1；例：6/1 ~ 7/3 = 33 天 → span 5。
- 标签：span=1 → `第N周`；span>1 → `第N周 - 第(N+span-1)周`（例：`第1周 - 第5周`）。
- 编号占用制：新建周期 `week_no` = 现存所有周计划的最大"结束编号"（week_no + span - 1）+ 1，无记录为 1。旧数据（每条 ≤7 天）行为与现状完全一致。
- 边界：编辑历史周期跨度变化不回溯重排后续编号；删除中间周期允许编号存在空洞（展示始终以日期为准，可接受）。

## 后端变更（righthand-go）
1. `internal/models/project_weekly_plan.go`：新增 `Span() int` 与 `Label() string` 方法（models 包，供 service 与 scheduler 共用），实现上述规则。
2. `services/task-service/service/project_service.go`：
   - `buildWeekView`（L304 返回体）：新增 `weekLabel`、`weekSpan` 字段。
   - `taskWeekView`（L1342）：weeks 元素新增 `weekLabel`、`weekSpan`。
   - `CreateWeek`（L1046-1050）：week_no 生成改为占用制（遍历该项目全部周计划取 max 结束编号 +1）。
   - 硬编码"第%d周"文案改用 `Label()`：变更日志 WEEK_ADD/WEEK_EDIT/WEEK_DELETE/WEEK_STATUS/SUMMARY_SUBMIT/SUMMARY_UNLOCK、错误文案"第%d周总结已锁定，不可移动任务"（L1666）。
3. `internal/scheduler/project_scheduler.go`（L92）：周总结提醒消息改用 `Label()`。

## 前端变更（righthand-frontend）
4. `src/utils/projectEnums.js`：新增展示 helper `weekLabelOf(week)`（优先 `week.weekLabel`，回退 `第${week.weekNo}周`，空值安全），符合该文件"仅展示映射"的定位。
5. `src/views/project/components/WeeklyPlanSection.vue`：
   - L27 卡片标题 → `weekLabelOf(week)`；`weekSpan>1` 时增收 `跨N周` el-tag；week-meta 日期范围在跨周时加主色/加粗样式（突出日期）。
   - L128 编辑弹窗标题、L162 总结弹窗标题、L343 删除确认、L410 解锁确认文案 → `weekLabelOf`。
6. `src/views/project/components/ProjectTaskTab.vue`：
   - L11 周过滤下拉、L117 新建任务归属周、L159 关联事项归属周下拉 → `weekLabelOf(w)`。
   - L37 所属周 tag → `w.weekLabel || 第${w.weekNo}周`。
   - L73 移出、L87 移入、L398 移动成功提示 → 对应 week 的 label。

## 测试
7. 后端 `services/task-service/service/project_service_test.go` 新增用例：
   - `Span/Label` 边界：7 天→`第N周`；8 天→span 2；33 天→`第1周 - 第5周`。
   - 编号占用：建 33 天周期（week_no=1）→ 再建 7 天 → week_no=6；再建 8 天 → 下一条编号 = 结束编号+1。
   - `buildWeekView`/`taskWeekView` 返回 `weekLabel/weekSpan` 断言。
   - 既有 7 天系列用例（week_no 1/2/3、文案断言"第3周总结已锁定"）必须保持通过、无需改动——验证向后兼容。
8. `internal/scheduler/project_scheduler_test.go` 保持通过（断言消息含"尚未提交周总结"不变，补一条单周 label 断言）。
9. 前端 `tests/views/project/ProjectDetail.test.ts`：mock 周数据补 `weekLabel/weekSpan`，新增跨周渲染断言（`第1周 - 第5周`、`跨5周` tag）。
10. `tests/views/project/ProjectTaskTab.test.ts`：所属周 tag、周过滤下拉 label 断言更新。
11. 根 `CHANGELOG.md` 按"改了什么/为什么/影响范围"登记。

## 验证命令
- `cd righthand-go; go build ./...; go test ./services/task-service/... ./internal/scheduler/...`
- `cd righthand-frontend; npx vitest run tests/views/project`

## 验收标准
- 新增 2026-06-01 ~ 2026-07-03 周期：卡片显示 `第1周 - 第5周` + `跨5周` tag + 日期范围突出；再建 2026-07-04 ~ 2026-07-10 → 显示 `第6周`（衔接不重叠）。
- 周过滤下拉、任务所属周 tag、编辑/总结/删除/解锁弹窗、调度提醒消息、变更日志文案全部一致。
- 既有单周（≤7 天）数据展示与行为零变化。

## Assumptions
- 编号占用制是「第N周 - 第M周」自洽的前提（否则相邻周期编号重叠）；旧数据行为不变。
- 不新增"必须 7 天"的起止校验，允许任意跨度。
- 任务侧 deadline/归属周联动本次不处理。