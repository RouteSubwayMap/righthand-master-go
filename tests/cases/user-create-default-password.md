# 新增用户默认密码（CreateUser）测试用例矩阵

- 关联需求：新增用户弹窗去掉密码输入框后，新增用户初始密码一律为默认值 88888888（后端缺省兜底）；姓名必输由后端同步硬约束
- 对应代码：righthand-go/services/auth-service/{dto/user_dto.go, service/user_service.go, handler/user_handler.go}
- 测试文件：righthand-go/tests/user_create_default_test.go（真库集成：复用 route_test.go 的 setupTestRouter 与 utils.GenerateToken）
- 更新日期：2026-09-22

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| Create_NoPassword_DefaultApplied | 正例 | POST /api/system/users，body={username: it_u<运行时间戳>, realName: 集成测试用户}，不带 password 键；携带 Bearer 管理员 token | HTTP 200 且 code=200、message=创建成功；随后 POST /api/auth/login {同 username, password: 88888888} 返回 code=200 | sys_user 新增 1 行（status=1、deleted=0） | 管理员 token（userID=1）；真实 MySQL 库表 sys_user | 用例结束按 username 精确物理删除本用例创建的用户行 |
| Create_EmptyPassword_DefaultApplied | 边界 | 创建 body 显式携带 password:""（空串） | HTTP 200 且 code=200；用 88888888 登录 code=200 | sys_user 新增 1 行，password 列 bcrypt 对应 88888888 | 同 Create_NoPassword_DefaultApplied | 同上 |
| Create_ExplicitPassword_Ignored | 边界 | 创建 body 携带 password:"Custom123" | HTTP 200 且 code=200；用 88888888 登录 code=200；用 Custom123 登录 HTTP 401（传入密码未被采纳） | sys_user 新增 1 行，password 列 bcrypt 对应 88888888 | 同 Create_NoPassword_DefaultApplied | 同上 |
| Create_MissingRealName_Rejected | 反例 | 创建 body={username: it_u<运行时间戳>}，不带 realName | HTTP 400 且 code=400、message=请求参数错误；不进入写库 | sys_user 0 行新增 | 管理员 token（userID=1） | 无（无写入） |
| Create_MissingUsername_Rejected | 反例 | 创建 body={realName: 集成测试用户}，不带 username | HTTP 400 且 code=400；不进入写库 | sys_user 0 行新增 | 管理员 token（userID=1） | 无（无写入） |
| Create_DuplicateUsername_Rejected | 反例 | 第一次创建 it_u<运行时间戳> 成功；紧接着以相同 username 发起第二次创建 | 第二次 HTTP 200 且 code=500、message=用户名已存在 | sys_user 保持 1 行（首次创建的行） | 首次创建成功的用户 | 用例结束删除本用例创建的 1 个用户 |
| Create_StoredAsBcryptHash | 边界 | 创建成功后按 username 从 sys_user 读取 password 列 | 列值不等于明文 88888888；长度 60 且以 $2 开头；bcrypt.CompareHashAndPassword(列值, "88888888") 校验通过 | 无额外写入 | 本用例创建的用户 | 同上 |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度 / 极值与溢出：不适用——用户字段为字符串与状态位，无金额与数值运算；超长输入截断由 varchar 列宽约束，属既有行为。
- 状态机：不适用——新增用户无状态流转；status 缺省 1（启用）由 CreateUser 既有逻辑保证。
- 权限与越权：不适用新增——路由级 JWTAuth 已验证（未带 token 为 401 由 route_test 覆盖）；本功能未改动权限码。
- 时间边界：不适用——无日期计算；用户名时间戳后缀仅用于防重复。
- 分片路由 / 批量与部分失败 / SQL 方言与函数：不适用——单行插入、无批量、无方言函数。
- 幂等与重复提交：适用，已覆盖（Create_DuplicateUsername_Rejected 以唯一索引+用户名查重双重保障拒绝重复）。
- 并发与乐观锁 / 事务回滚：不适用——并发同名创建由 username uniqueIndex 兜底，无法稳定构造并发时序；角色分配回滚路径本次未改动。
- 空值与可选字段：适用，已覆盖（Create_NoPassword 缺键、Create_EmptyPassword 空串、Create_MissingRealName/MissingUsername 缺必填键）。
- 中文与 Unicode：适用，已覆盖（realName=集成测试用户 写入后登录链路回读）。
- 分页与排序稳定性：不适用——本接口无分页。

## 补充说明

- 默认密码实现（强制默认）：`CreateUserRequest` 不再携带 password 字段，service 层 `CreateUser` 无条件使用包级常量 DefaultPassword（88888888）加密入库，忽略任何调用方传入的密码；`ResetPassword` 同步改用该常量（SSOT 收敛）。
- 姓名必输实现：`CreateUserRequest.RealName` 增加 binding:"required"，缺失时由 ShouldBindJSON 拦截并返回 HTTP 400。
- 安全断言：Create_StoredAsBcryptHash 确保默认密码同样走 bcrypt 加密入库，不会出现明文密码列。
