# 新增用户弹窗（姓名必输 / 去密码框）测试用例矩阵

- 关联需求：用户管理「新增用户」弹窗去掉密码输入框，姓名改为必输；提交不携带密码（由后端默认 88888888）
- 对应代码：righthand-frontend/src/views/system/UserManagement.vue（提交目标 /api/system/users 见 src/api/system.js createUser）
- 测试文件：righthand-frontend/tests/views/system/UserManagement.test.ts
- 更新日期：2026-09-22

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| AddDialog_NoPasswordField | 正例 | 挂载后点击「新增用户」按钮 | 弹窗标题=新增用户；#password 元素不存在；#username 与 #realName 输入框存在 | 无数据写入（loadData 仅读接口） | getUserPage 返回 records=[]、total=0；getAllRoles 返回空数组（vi.mock） | 无（内存 mock，随用例销毁） | - |
| Submit_NoPasswordInPayload | 边界 | 填 username=zhangsan、realName=张三、email、phone 后点「确定」 | createUser 调用 1 次；入参含 username=zhangsan 与 realName=张三，不含 password 键 | 弹窗关闭；loadData 重新拉取列表 1 次 | getUserPage mock + createUser mock 返回 code=200 | 无 | - |
| Submit_MissingRealName_Blocked | 反例 | 填 username=zhangsan、realName 留空后点「确定」 | createUser 0 次调用；弹窗保持打开；realName 表单项展示「请输入姓名」校验文案 | 无数据写入 | getUserPage mock | 无 | - |
| Submit_MissingUsername_Blocked | 反例 | 填 realName=张三、username 留空后点「确定」 | createUser 0 次调用；弹窗保持打开；username 表单项展示「请输入用户名」校验文案 | 无数据写入 | getUserPage mock | 无 | - |
| EditDialog_NoPasswordField | 正例 | 列表加载 2 条用户后点击第 1 行「编辑」按钮 | 弹窗标题=编辑用户；#password 不存在；#username 输入框显示 admin 且处于 disabled | 无数据写入 | getUserPage 返回 2 条记录、total=2；getUserRoles mock 返回角色数组 | 无 | - |
| Edit_Submit_NoPasswordInPayload | 边界 | 编辑态将 realName 改为 管理员2 后点「确定」 | updateUser 调用 1 次且首个入参=1；提交 payload 不含 password 键 | 弹窗关闭；loadData 重新拉取列表 1 次 | 同 EditDialog_NoPasswordField + updateUser mock 返回 code=200 | 无 | - |

## 维度适用性说明（dimension-catalog 逐项检查）

- 金额与精度 / 极值与溢出：不适用——弹窗字段为字符串，无金额与数值运算。
- 状态机：不适用——用户新增无前端状态流转，status 单选仅作为普通字段提交。
- 权限与越权：不适用——新增/编辑按钮显隐由 v-permission 与后端返回控制，本弹窗未改动权限逻辑。
- 时间边界：不适用——无日期字段参与本弹窗提交。
- 分片路由 / 批量与部分失败 / SQL 方言与函数：不适用——纯前端表单，无分片、无批量、无 SQL。
- 幂等与重复提交：不适用新增——重复用户名由后端拒绝（见 user-create-default-password.md 矩阵）；提交按钮 loading 防连点属既有行为。
- 并发与乐观锁 / 事务回滚：不适用——前端无并发写与事务边界。
- 空值与可选字段：适用，已覆盖（Submit_MissingRealName / Submit_MissingUsername 空值拦截、Submit_NoPasswordInPayload 缺省键不携带）。
- 中文与 Unicode：适用，已覆盖（姓名=张三、管理员2 及中文校验文案断言）。
- 分页与排序稳定性：不适用——弹窗不涉及分页。

## 补充说明

- 无密码框断言口径：#password 为原密码输入框的固定 id，去掉表单项后该元素在新增与编辑两种状态下均不应存在。
- 数据源契约：组件列表数据经 getUserPage（/api/system/users/page）读取 pageData.list，测试 mock 按该契约返回。
- 黑盒交互说明：组件为 script setup，测试通过 DOM（按钮点击、输入框 setValue、校验文案）驱动，不直接读写组件内部状态。
