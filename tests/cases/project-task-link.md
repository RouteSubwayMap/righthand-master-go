# 项目「关联已有事项」链路（ProjectTaskTab）测试用例矩阵

- 关联需求：项目详情「任务」Tab「关联已有事项」——选择事项改为 TaskSelectDialog 弹出选择（回填只读输入框），提交后关联到周计划；新增无选择防御守卫
- 对应代码：righthand-frontend/src/views/project/components/ProjectTaskTab.vue（配合 righthand-frontend/src/components/TaskSelectDialog.vue）
- 测试文件：righthand-frontend/tests/views/project/ProjectTaskTab.test.ts
- 更新日期：2026-09-22

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| OpenLinkDialog_DefaultFirstWeek | 正例 | 项目 id=9（周 21/22），weekFilter=null，调用 openLinkDialog() | linkDialogVisible=true；linkForm.weekPlanId=21；taskId=null；taskSelectVisible=false | 组件状态如上；无数据写入 | 项目 weeks=[{id:21,启动周},{id:22,开发周}]、allowedActions=['addWeek'] | 无（vi.mock 内存态，随用例结束销毁） | - |
| SelectTask_LinkSuccess_CloseDialog | 正例 | 选择事项 id=12「修复登录bug」后调用 handleLink() | linkTaskToWeek(9,21,12) 调用 1 次；linkForm.taskId=12、taskTitle='#12 修复登录bug'；linkDialogVisible=false | 关联请求发出 1 次并成功；成功后弹窗关闭 | 项目同上；mock linkTaskToWeek resolve 成功 | 无 | - |
| Link_ApiReject_DialogStaysOpen | 异常 | mock linkTaskToWeek reject(Error('接口异常')) 后调用 handleLink() | 仍调用 1 次并以 reject 结束（无业务错误码，由请求拦截器统一提示）；linkDialogVisible 保持 true；linking 复位 false | 关联请求发出 1 次（失败）；成功副作用（关闭弹窗/刷新列表）均不发生 | 项目同上；mock 拒绝响应 | 无 | - |
| Link_NoSelection_GuardBlocks | 反例 | taskId=null、weekPlanId=21 时直接调用 handleLink() | ElMessage.warning 1 次（文案含「请先选择事项」）；linkTaskToWeek 0 次调用；linkDialogVisible 保持 true | 无请求发出；无数据写入 | 项目同上 | 无 | - |
| Reopen_ClearPreviousSelection | 反例 | 先选择 id=12（已回填），再次调用 openLinkDialog() | linkForm.taskId=null、taskTitle=''（无残留）；weekPlanId 重新按 weekFilter / 第一周取值 | 组件状态重置；无数据写入 | 项目同上 | 无 | - |
| OpenLinkDialog_WeekFilterFollows | 边界 | weekFilter=22 时调用 openLinkDialog() | linkForm.weekPlanId=22（跟随筛选周，而非默认第一周） | 无数据写入 | 项目 weeks=[{id:21},{id:22}] | 无 | - |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度 / 极值与溢出 / 分片路由 / 批量与部分失败 / SQL 方言与函数：不适用——前端链路仅发送关联请求，无金额、数值运算、分片、批量与 SQL。
- 状态机：不适用——关联 / 回填不改项目与事项状态（项目状态机与事项终态机由后端既有用例覆盖）。
- 权限与越权：不适用新增——入口显隐由 canManageTasks（项目状态 + 周权限）控制，接口鉴权在后端。
- 时间边界：不适用——归属周以外无时间语义；截止时间不参与本链路。
- 幂等与重复提交：不适用新增——重复关联同一周的幂等由后端 LinkTaskToWeek 保证（already=true 分支，已有后端用例覆盖）；前端成功后关闭弹窗，不重复提交。
- 并发与乐观锁 / 事务回滚：不适用——前端单请求链路，无并发写与事务。
- 空值与可选字段：适用，已覆盖（Link_NoSelection_GuardBlocks 空选择拦截、Reopen_ClearPreviousSelection 状态清空）。
- 中文与 Unicode：适用，已覆盖（任务标题「修复登录bug」回填断言与中文提示文案）。
- 分页与排序稳定性：不适用——本链路无分页（主表格分页由 ProjectTaskTab 既有用例覆盖）。

## 补充说明

- 本轮改造要点：原远程搜索下拉替换为 TaskSelectDialog（读 records/total 契约）；回显格式 `#<id> <title>`；新增 handleLink 防御守卫（对应 Link_NoSelection_GuardBlocks），与「关联」按钮 disabled 双保险。
- 关联成功后的列表刷新（loadTasks / refresh）依赖父组件组合行为，由浏览器联调覆盖。
