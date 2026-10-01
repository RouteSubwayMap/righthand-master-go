# 项目任务页签筛选交互（前端） 测试用例矩阵

- 关联需求：项目管理-详情-任务页签新增「负责人」「任务状态」筛选（前端交互）
- 对应代码：`righthand-frontend/src/views/project/components/ProjectTaskTab.vue`、`righthand-frontend/tests/views/project/ProjectTaskTab.test.ts`
- 更新日期：2026-09-23
- 维度说明：
  - 已覆盖：交互联动（筛选变更回第 1 页、三条件组合、清空不带参）、状态重置（切换项目）、空值与可选字段（清空筛选）。
  - 不适用：金额/并发/事务回滚/分片/方言（纯前端，接口以 mock 拦截）；权限（页面入口权限既有路由守卫覆盖，本次仅工具栏筛选）。

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| Filter_User_TriggersSearchFromFirstPage | 正例 | 先翻到第 2 页，再设置 userFilter=2 并触发查询 | getProjectTasks 以 (9, pageNum=1, pageSize=10, userId=2) 调用，pageNum 回到第 1 页 | 无（mock 拦截，不写库） | 项目 id=9；props.users=[{id:2},{id:3}] | 无（内存 mock 态，随用例销毁） | - |
| Filter_Status_AppliesStatusCode | 正例 | 设置 statusFilter=DONE 并触发查询 | getProjectTasks 参数含 statusCode=DONE 且 pageNum=1 | 无（mock 拦截，不写库） | 同上 | 无（内存 mock 态，随用例销毁） | - |
| Filter_Combined_WeekUserStatus | 边界 | 同时设置 weekFilter=22、userFilter=2、statusFilter=DONE 并查询 | getProjectTasks 参数同时含 weekPlanId=22、userId=2、statusCode=DONE，共 3 个过滤条件 | 无（mock 拦截，不写库） | 周选项含 id=21、22；props.users=[{id:2},{id:3}] | 无（内存 mock 态，随用例销毁） | - |
| Filter_Cleared_OmitsFilterParams | 反例 | 三个筛选均置空后触发查询 | getProjectTasks 参数仅含 pageNum/pageSize，不含 weekPlanId/userId/statusCode | 无（mock 拦截，不写库） | 同上 | 无（内存 mock 态，随用例销毁） | - |
| Filter_Reset_OnProjectSwitch | 反例 | 三个筛选有值后，props.project.id 由 9 切换为 10 | 三个筛选重置为 null，pageNum=1，并以 (10, 无过滤参数) 重新查询 | 无（mock 拦截，不写库） | 两个项目对象（id=9、id=10） | 无（内存 mock 态，随用例销毁） | - |
| Filter_Toolbar_ThreeSelectsRendered | 边界 | 挂载组件读取工具栏 | toolbar-left 内渲染 3 个下拉（周过滤、负责人、状态），负责人/状态下拉分别显示占位文案「全部负责人」「全部状态」 | 无（只读渲染） | 同上 | 无（内存 mock 态，随用例销毁） | - |
