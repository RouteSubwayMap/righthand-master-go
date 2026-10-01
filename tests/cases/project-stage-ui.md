# 项目阶段前端链路（ProjectStageSection / 详情页签 / 清单列）测试用例矩阵

- 关联需求：项目管理-项目阶段——项目详情新增「项目整体计划」页签（默认激活、位于「周计划」之前，按开始时间正序展示、当前时间命中阶段行高亮并加「当前阶段」标签）；项目清单「进度」与「最近更新」之间新增「当前阶段」列（空值显示 -）；新增/编辑/删除入口由 allowedActions.manageStage 驱动
- 对应代码：righthand-frontend/src/views/project/components/ProjectStageSection.vue、src/views/project/ProjectDetail.vue、src/views/project/ProjectList.vue、src/api/project.js
- 测试文件：righthand-frontend/tests/views/project/ProjectStageSection.test.ts、tests/views/project/ProjectDetail.test.ts、tests/views/project/ProjectList.test.ts
- 更新日期：2026-09-23

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| StageSection_Render_OrderAndEmpty | 正例 | project.stages=2 条（详情接口已按开始时间正序下发）；空数组作对照 | 表格渲染 2 行且顺序与下发一致；标题显示「共 2 个阶段」；空数组显示「暂无项目阶段」空态 | 无（纯渲染） | buildProject 构造 stages + allowedActions=['manageStage'] | 无（vitest 内存态，随用例结束销毁） | - |
| StageSection_CurrentStage_HighlightAndTag | 正例 | currentStageId=命中行 id（对应「开发阶段」） | 命中行 tr 带 current-stage-row 类；行内出现「当前阶段」标签；其余行无类无标签 | 无（纯渲染） | project.stages 2 条 + currentStageId=命中 id | 无 | - |
| StageSection_NoMatch_NoHighlight | 反例 | currentStageId 缺省（null，对应无命中或终态项目） | 所有行均无 current-stage-row 类；页面无「当前阶段」标签 | 无（纯渲染） | project.stages 2 条，无 currentStageId 字段 | 无 | - |
| StageSection_ButtonGating_ByAllowedActions | 权限 | allowedActions=['manageStage'] 与 allowedActions=[] 两组 props | 含 manageStage 时显示「新增阶段」按钮与操作列；为空时按钮与操作列均不渲染 | 无（纯渲染） | 两组 project props 对照 | 无 | - |
| StageSection_Save_Success_RefreshEmitted | 正例 | 新增弹窗填写名称「联调阶段」2026-10-01~2026-10-10 提交；编辑原阶段改名提交 | createProjectStage / updateProjectStage 各调用 1 次且入参正确；弹窗关闭；emit('refresh') 各 1 次 | 无落库（mock 接口 resolve 成功） | mock 创建/更新接口成功；canManageStage=true | 无 | - |
| StageSection_NameEmpty_ValidationBlocked | 空值 | 阶段名称留空，点击保存 | 表单校验拦截（不通过 validate）；createProjectStage 0 次调用；弹窗保持打开 | 无请求发出 | mock 接口 | 无 | - |
| StageSection_DateRange_EndBeforeStart_Blocked | 边界 | 开始 2026-06-10、结束 2026-06-09，点击保存 | 校验拦截并提示「结束日期不能早于开始日期」；接口 0 次调用 | 无请求发出 | mock 接口 | 无 | - |
| StageSection_Submit_ApiError_DialogStaysOpen | 异常 | mock createProjectStage reject（接口 HTTP 500）后提交 | 接口返回 500（错误提示由拦截器统一处理）；弹窗保持打开；saving 复位 false；无 emit('refresh') | 无数据写入；刷新副作用不发生 | mock 拒绝响应 | 无 | - |
| StageSection_Delete_CancelNoRequest | 反例 | 删除确认先被取消（ElMessageBox.confirm reject 'cancel'）；再确认通过（resolve 'confirm'）作链路对照 | 取消时 deleteProjectStage 0 次调用且无 emit('refresh')；确认后调用 1 次（入参 projectId+stageId）并 emit('refresh') 1 次 | 取消无请求發出；确认后 mock 删除成功 | 阶段行 + canManageStage=true | 无 | - |
| ProjectDetail_StageTab_And_ProjectList_Column_Render | 正例 | 详情初始进入（未点击任何页签），detailData 含 stages 与 currentStageId；清单行分别含 currentStage='开发阶段' 与空值 | 详情默认激活「项目整体计划」页签且其位于「周计划」之前并渲染阶段行；清单「当前阶段」列位于「进度」与「最近更新」之间；空值行显示 - | 无（纯渲染） | mock getProjectDetail 返回含 stages；mock getProjectList 返回含 currentStage 行 | 无 | 用户反馈：点击清单「详情」默认展示「项目整体计划」页签 |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度：不适用——阶段无金额字段。
- 状态机：不适用新增——项目状态机与按钮显隐由后端 allowedActions 决定，前端仅消费；终态项目 currentStageId 为 null 见 NoMatch 用例。
- 权限与越权：适用，已覆盖——manageStage 缺省时新增按钮与操作列不渲染（ButtonGating 用例）。
- 时间边界：不适用新增——「当前阶段」判定完全复用后端下发的 currentStageId，前端不做日期比较（对齐项目管理红线）。
- 分片路由：不适用——无分片。
- 幂等与重复提交：不适用新增——弹窗保存期间 loading 保护，接口幂等由后端保证。
- 并发与乐观锁：不适用——前端单请求链路，无并发写。
- 事务回滚：不适用——前端无事务；接口失败时弹窗保持打开、不刷新（Submit_ApiError 用例）。
- 空值与可选字段：适用，已覆盖——名称空值校验拦截（NameEmpty）、描述可选、stages 空数组空态（Render_OrderAndEmpty）。
- 极值与溢出：不适用新增——名称 100 字符上限由输入框 maxlength 与后端双重约束（后端用例覆盖）。
- 中文与 Unicode：适用，已覆盖——阶段名称「开发阶段」「联调阶段」中文断言。
- 分页与排序稳定性：不适用——阶段列表内嵌于详情返回，不涉及分页；正序由后端 ORDER BY 保证（后端用例覆盖）。
- 批量与部分失败：不适用——无批量接口。
- SQL 方言与函数：不适用——前端无 SQL。

## 补充说明

- 高亮判定与清单列取值完全复用后端字段：详情 currentStageId（命中行高亮 + 标签）、清单 currentStage（表格列），前端不复制日期判定逻辑。
- 有效用例 10 条：正例 4 条（40%），反例/边界/异常/权限/空值共 6 条（60%）。
