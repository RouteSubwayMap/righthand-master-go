# 项目任务列表筛选（负责人/状态） 测试用例矩阵

- 关联需求：项目管理-详情-任务页签新增「负责人」「任务状态」筛选
- 对应代码：`righthand-go/services/task-service/service/project_service.go` `ListProjectTasks`、`righthand-go/services/task-service/handler/project_week_handler.go`
- 更新日期：2026-09-23
- 状态取值说明：待办（状态机首态）、DOING（进行中）、DONE（已完成）
- 维度说明：
  - 已覆盖：分页与排序稳定性（过滤后 total 与越界页）、空值与可选字段（非法 userId、未知状态码）。
  - 不适用：金额精度/分片路由/幂等/事务回滚（纯只读查询）；方言（无方言函数，过滤条件均为等值比较）；权限（沿用既有项目可见性校验，本次不新增规则）；时间边界（截止时间排序既有用例已覆盖）。

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| ListProjectTasks_FilterByAssignee_OnlyAssigneeRows | 正例 | userId=3 | list 仅含 T1（title=任务T1），total=1 | 无（只读查询，不写库） | 项目 P2：周 W1、W2；T1（user3，DOING，挂 W1+W2）、T2（user2，待办，挂 W1）、T3（user2，DONE，挂 W2） | 事务回滚（t.Cleanup 回滚，开发库零副作用） | - |
| ListProjectTasks_FilterByStatusCode_OnlyDoingRows | 正例 | statusCode=DOING | list 仅含 T1，total=1 | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
| ListProjectTasks_CombinedFilters_Intersection | 边界 | weekPlanId=W2、userId=2、statusCode=DONE | list 为三条件交集且仅含 T3，total=1（T2 不在 W2 被排除） | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
| ListProjectTasks_FilteredPagination_FirstPageUnfinishedFirst | 边界 | userId=2、pageNum=1、pageSize=1 | total=2、list 长度=1 且元素为 T2（过滤后未完成优先排序） | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
| ListProjectTasks_InvalidUserId_Ignored | 反例 | userId=abc（非数字） | 忽略该条件，返回全部 3 条，total=3 | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
| ListProjectTasks_UnknownStatusCode_EmptyList | 反例 | statusCode=FOO | list 长度=0，total=0 | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
| ListProjectTasks_FilteredPagination_BeyondLastPageEmpty | 边界 | userId=2、pageNum=3、pageSize=1 | total=2、list 长度=0（越界空页，total 仍为过滤后条数） | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
| ListProjectTasks_NoMatchedUser_EmptyList | 反例 | userId=999（不存在的用户） | list 长度=0，total=0 | 无（只读查询，不写库） | 同上 | 事务回滚（t.Cleanup 回滚） | - |
