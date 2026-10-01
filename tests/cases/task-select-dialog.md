# 事项选择弹窗（TaskSelectDialog）测试用例矩阵

- 关联需求：项目详情「任务」Tab「关联已有事项」弹窗中，「选择事项」由远程搜索下拉改为弹出式选择（评估结论：不重用「事项管理」页面，新增同款「筛选 + 分页」选择弹窗）
- 对应代码：righthand-frontend/src/components/TaskSelectDialog.vue（依赖 righthand-frontend/src/utils/projectEnums.js 的 TASK_ITEM_TYPE / TASK_STATUS / TASK_PRIORITY）
- 测试文件：righthand-frontend/tests/components/TaskSelectDialog.test.ts
- 更新日期：2026-09-22

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| Open_DefaultPage_LoadFirstPage | 正例 | modelValue=true 打开；getTaskList 返回 records=[事项A,事项B]、total=2 | getTaskList 调用 1 次且参数={pageNum:1,pageSize:10}；list 标题依次为「修复登录bug」「编制周报」；total=2 | 组件状态：list/total 更新；无数据写入 | 事项A(id=12，任务，高优先级，状态=待办)、事项B(id=13，备忘，低优先级，状态=已完成) | 无（vi.mock 内存态，随用例结束销毁） | - |
| Confirm_PresetRow_EmitsSelectAndCloses | 正例 | props.selectedId=12；打开后点「确定」 | emitted('select') 1 次且载荷 id=12；emitted('update:modelValue') 末次为 false（关闭） | selectedId=12 保持；无数据写入 | 事项A(id=12) | 无 | - |
| Confirm_OtherRow_EmitsThatRow | 正例 | 点选 id=13 行后点「确定」 | emitted('select') 1 次且载荷 title=「编制周报」 | selectedId=13；无数据写入 | 事项A、事项B | 无 | - |
| DblClick_Row_EmitsImmediately | 正例 | 双击 id=12 行 | 立即 emitted('select')（载荷 id=12）并关闭（update:modelValue=false） | selectedId=12；无数据写入 | 事项A(id=12) | 无 | - |
| Search_Params_EmptyKeysOmittedAndPageReset | 边界 | 预置 itemType=TASK；已有 pageNum=3；填 标题=登录、状态=待办，点「查询」 | getTaskList 末次参数={pageNum:1,pageSize:10,itemType:TASK,title:登录,statusCode=待办}，不含 priority 键（空值不携带） | 无数据写入 | 事项A（标题含「登录」） | 无 | - |
| Reset_RestorePresetTypeAndPageOne | 边界 | 改 标题=登录、类型=备忘、状态=已完成、优先级=低 后点「重置」 | filters 恢复 {title:'',itemType:TASK(预置),statusCode:'',priority:''}；getTaskList 末次参数={pageNum:1,pageSize:10,itemType:TASK} | 无数据写入 | 无 | 无 | - |
| PageChange_FetchWithNewPageNum | 边界 | pagination.pageNum 由 1 置为 2 触发加载 | getTaskList 末次参数含 pageNum=2、pageSize=10 | 无数据写入 | 事项A、事项B（total=2，第 2 页为空页边界） | 无 | - |
| PageSizeChange_BackToFirstPage | 边界 | pagination.pageSize 由 10 置为 20 触发 size-change 处理函数 | getTaskList 末次参数={pageNum:1,pageSize:20} | 无数据写入 | 事项A、事项B | 无 | - |
| Confirm_NoSelection_WarnsStaysOpen | 反例 | selectedId=null 点「确定」 | ElMessage.warning 1 次（文案含「请先选择一条事项」）；emitted('select') 0 次；未关闭 | 无数据写入 | 事项A、事项B | 无 | - |
| Confirm_StaleSelectionNotInList_Warns | 反例 | props.selectedId=999（列表无此行）点「确定」 | ElMessage.warning 1 次；emitted('select') 0 次；未关闭 | 无数据写入 | 事项A、事项B（均不含 id=999） | 无 | - |
| Load_ApiReject_ClearListAndResetLoading | 异常 | getTaskList reject(new Error('Network Error')) | list=[]、total=0、loading=false；无业务错误码（网络异常直抛），不残留旧数据 | 无数据写入 | 无（预置拒绝响应） | 无 | - |
| Closed_NoRequest | 反例 | modelValue=false 挂载 | getTaskList 0 次调用；list=[] | 无数据写入 | 无 | 无 | - |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度 / 极值与溢出：不适用——仅整数 id 与分页参数，无金额与数值运算。
- 状态机：不适用——弹窗只读查询事项，不做状态流转（事项终态机由事项模块既有用例覆盖）。
- 权限与越权：不适用新增——数据可见范围由后端 /api/tasks 控制（非 admin 仅本人），弹窗仅透传筛选参数。
- 时间边界：不适用——截止时间仅格式化展示（dayjs），无时区 / 跨天计算。
- 分片路由 / 批量与部分失败 / SQL 方言与函数：不适用——纯前端组件，无分片、无批量、无 SQL。
- 幂等与重复提交：不适用——只读选择；确认即关闭，不产生写请求。
- 并发与乐观锁 / 事务回滚：不适用——前端无并发写与事务边界。
- 空值与可选字段：适用，已覆盖（Search_Params_EmptyKeysOmittedAndPageReset 空筛选键不携带、Confirm_NoSelection 空选择、Closed_NoRequest 关闭态）。
- 中文与 Unicode：适用，已覆盖（标题「修复登录bug」「编制周报」与中文提示文案断言）。
- 分页与排序稳定性：适用，已覆盖（PageChange_FetchWithNewPageNum / PageSizeChange_BackToFirstPage 两条边界）。

## 补充说明

- 修复背景：原「关联已有事项」远程搜索读 data.list，与实际接口契约（records/total）不符导致选项恒空；本组件按 records 解析，并在 Open_DefaultPage_LoadFirstPage 对调用参数做契约断言。
- 测试策略：弹窗内容经 Teleport 挂载到 body，单测聚焦状态与事件链路；表格渲染与真实点击由浏览器联调覆盖（原因：VTU teleport stub 对 Element Plus 的 ElTeleport 误伤）。
