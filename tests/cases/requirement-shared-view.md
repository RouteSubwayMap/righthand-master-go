# 需求共享视图（详情/列表/导出）测试用例矩阵

- 关联需求：需求管理 bug 修复——非本人需求点击「查看」报错。业务确认为「需求库团队共享」：所有登录用户可查看任意需求的列表与详情；编辑仅限创建人/跟进人、删除仅限创建人（本次不改动）
- 对应代码：righthand-go/services/task-service/service/requirement_service.go（GetRequirementByID / ListRequirements / ExportRequirements）
- 测试文件：righthand-go/services/task-service/service/requirement_integration_test.go
- 更新日期：2026-09-21

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| GetDetail_NonOwnerNonFollower_Success | 正例 | 查询用户2创建且follower_id为NULL的需求B（id=2） | 成功返回需求B：id=2、title与需求B一致 | 无（只读SELECT，不产生数据变更） | 需求B(user_id=2，follower_id为NULL) | 用例独立内存库，随用例结束自动销毁 | - |
| GetDetail_OwnRequirement_Success | 正例 | 查询用户1本人创建的需求A（id=1） | 成功返回需求A：id=1、title与需求A一致 | 无（只读SELECT，不产生数据变更） | 需求A(user_id=1) | 用例独立内存库，随用例结束自动销毁 | - |
| GetDetail_SoftDeleted_NotFound | 边界 | 查询已软删除的需求D（id=3，deleted=1） | 返回错误码 gorm.ErrRecordNotFound | 无（只读SELECT，不产生数据变更） | 需求D(user_id=2，deleted=1) | 用例独立内存库，随用例结束自动销毁 | - |
| GetDetail_NotExist_NotFound | 反例 | 查询不存在的需求 id=999999 | 返回错误码 gorm.ErrRecordNotFound | 无（只读SELECT，不产生数据变更） | 空库（不预置该ID的数据） | 用例独立内存库，随用例结束自动销毁 | - |
| GetDetail_InvalidID_Negative_NotFound | 异常 | 查询非法ID id=-1（负数，库中不存在） | 返回错误码 gorm.ErrRecordNotFound | 无（只读SELECT，不产生数据变更） | 空库（不预置该ID的数据） | 用例独立内存库，随用例结束自动销毁 | - |
| List_AnyUserSeesAll_Success | 权限 | 用户1查询第1页每页10条（deleted=0 可见共2条：需求A、需求B） | total=2，返回2条（含用户2创建的需求B） | 无（只读SELECT，不产生数据变更） | 需求A(user_id=1)、需求B(user_id=2) | 用例独立内存库，随用例结束自动销毁 | - |
| List_StatusFilter_AcrossUsers_Visible | 权限 | 用户1按status=DEVELOPING筛选第1页每页10条（该状态仅属于用户2的需求B） | total=1，返回1条需求B | 无（只读SELECT，不产生数据变更） | 需求B(user_id=2，status=DEVELOPING) | 用例独立内存库，随用例结束自动销毁 | - |
| Export_AnyUserExportsAll_Success | 权限 | 用户1导出（无筛选条件，deleted=0 可见共2条） | 返回2条（含用户2创建的需求B） | 无（只读SELECT，不产生数据变更） | 需求A(user_id=1)、需求B(user_id=2) | 用例独立内存库，随用例结束自动销毁 | - |
| Export_RequirementNoOfOtherUser_Found | 权限 | 用户1按需求编号REQ-B导出（REQ-B为用户2需求B的编号） | 返回1条，title与需求B一致 | 无（只读SELECT，不产生数据变更） | 需求B(user_id=2，requirement_no=REQ-B) | 用例独立内存库，随用例结束自动销毁 | - |
| Export_Deleted_Excluded | 边界 | 用户1导出（无筛选条件）；库内另有软删除的需求E（deleted=1） | 返回2条且不含需求E（仅deleted=0） | 无（只读SELECT，不产生数据变更） | 需求E(user_id=2，deleted=1) | 用例独立内存库，随用例结束自动销毁 | - |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度 / 时间边界 / 分片路由 / 极值与溢出：不适用——共享视图仅涉及 id/user_id 等值比较与 deleted 过滤，不涉及金额、时间、分片与溢出语义。
- 状态机：不适用——本次不改动状态流转；列表/导出仅对 status 做等值筛选（已由 List_StatusFilter_AcrossUsers_Visible 覆盖跨用户筛选可见）。
- 幂等与重复提交 / 并发与乐观锁冲突 / 事务回滚：不适用——详情/列表/导出均为只读查询，无写入、无锁、无事务边界。
- 中文与 Unicode：不适用新增用例——共享逻辑基于 id/deleted 条件，与字符集无关；中文数据正确性由既有 UTF-8 编码测试覆盖。
- 空值与可选字段：适用，已覆盖（follower_id 为 NULL 的需求对任意用户均可见，见 GetDetail_NonOwnerNonFollower_Success）。
- 权限与越权：适用，已覆盖（共享可见正向 7 条：详情/列表/导出跨用户可见；软删除不可见 2 条：详情与导出）。
- 分页与排序稳定性：不适用新增用例——本次未改动分页逻辑，既有 TestRequirementService_Integration_ListRequirements 的「分页测试」保持不变作回归。
- SQL 方言与函数：不适用——未使用方言函数（无 FIND_IN_SET / GROUP_CONCAT / JSON_EXTRACT 等），不命中 dialect-risk-list.md 任一条目，无需真库专项验证。

## 补充说明

- 判定表结论：验证对象为应用层过滤/共享逻辑（整数 id 比较、deleted 过滤、布尔组合），不命中方言风险清单，可用 sqlite 内存库（rules.md §9 允许，需 CGO）。
- 种子数据顺序固定（需求A、需求B、需求D 依次创建），故 id=1/2/3 可预期；测试代码直接引用创建返回的 ID 断言，不硬编码。
- 原 `TestRequirementService_Integration_DataIsolation`（断言列表按用户隔离、他人详情不可见）与本次业务确认相反，已由本矩阵对应的 `TestRequirementService_Integration_GetDetailShared` 与 `TestRequirementService_Integration_SharedView` 替代。
