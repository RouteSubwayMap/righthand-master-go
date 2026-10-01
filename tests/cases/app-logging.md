# 后端日志按日期切割 测试用例矩阵

- 关联需求：后端日志按日期生成文件 + 测试环境（单机 systemd 部署）日志可查
- 对应代码：`righthand-go/internal/logger/logger.go`、`righthand-go/internal/middleware/auth.go`（Logger）、`righthand-go/cmd/server/main.go`
- 更新日期：2026-09-23
- 维度说明：
  - 已覆盖：时间边界（跨日切换、保留天数）、并发（多协程写入）、异常与降级（路径非法）、安全（HTTP 日志不含敏感信息）、空值（output 缺省回退 stdout）。
  - 不适用：金额/分片/事务回滚/方言（日志为本地文件 IO，无数据库参与）；幂等（写入语义非幂等，重复写入追加属预期行为，已由行数断言覆盖）。

| 场景 | 类型 | 输入 | 预期响应 | 预期副作用 | 依赖数据 | 数据清理 | 关联需求 |
|---|---|---|---|---|---|---|---|
| Logger_FileOutput_WritesDatedFile | 正例 | output=file，now=2026-09-23，写入行「启动日志」 | 生成 app-2026-09-23.log，内容与写入行完全一致 | 新增当日日志文件 1 个 | t.TempDir 空目录；maxAge=30、maxBackups=10 | t.TempDir 自动清理 | - |
| Logger_BothOutput_ConsoleAndFile | 正例 | output=both，now=2026-09-23，注入 stdout 缓冲，写入 1 行 | 缓冲含该行且文件 app-2026-09-23.log 含该行，两处各 1 行 | 控制台缓冲追加 1 行；文件新增 1 行 | t.TempDir 空目录；stdout 以 bytes.Buffer 注入 | t.TempDir 自动清理 | - |
| Logger_CrossDay_SwitchesToNewFile | 边界 | now 先固定 2026-09-23 写 1 行，再切 2026-09-24 写 1 行 | 两个文件各含 1 行；09-23 文件大小不再变化（写入已切换） | 新增第 2 个文件；第 1 个文件不再追加 | t.TempDir 空目录 | t.TempDir 自动清理 | - |
| Logger_Retention_AgeThenBackupCount | 边界 | now=2026-09-23、maxAge=30、maxBackups=2，预置旧文件 app-2026-08-23.log（31 天前）、app-2026-08-25.log、app-2026-08-26.log、app-2026-08-27.log，触发写入 | 删除 08-23（超龄）与 08-25、08-26（超量），保留 app-2026-08-27.log 与当日文件共 2 个（等于 maxBackups） | 删除 3 个历史文件；当日文件保留 | 预置 4 个旧文件（文件名日期用于判定），文件内容为占位文本 | t.TempDir 自动清理 | - |
| Logger_StdoutOutput_NoFileCreated | 反例 | output=stdout，注入 stdout 缓冲，写入 1 行 | 缓冲含该行，目录内文件数=0 | 无文件创建 | t.TempDir 空目录 | t.TempDir 自动清理 | - |
| Logger_InvalidFilePath_InitFails | 反例 | file_path 的父路径被普通文件占据，output=file | 初始化返回非空 error 且不创建日志文件 | 无文件创建；调用方可降级控制台输出 | 临时目录内预置普通文件 fake-dir | t.TempDir 自动清理 | - |
| Logger_ConcurrentWrites_AllLinesIntact | 并发 | 50 个协程各写 1 行（不同内容），output=file | 文件行数与写入次数一致（50=50），每行内容完整无交错 | 仅 1 个当日文件；无半行或丢失行 | t.TempDir 空目录 | t.TempDir 自动清理 | - |
| Logger_MiddlewareHttpLine_NoSensitiveBody | 反例 | 经 Logger 中间件的 POST /api/auth/login（请求体 password=secret-password-123，响应 200） | 日志恰 1 行：含 POST、/api/auth/login、200、耗时与客户端 IP；不含 secret-password-123，也不含完整请求体 | 仅追加 1 行访问日志；无业务数据变更 | gin.New 测试引擎 + Logger 中间件 + httptest 请求 | 恢复 log 默认输出目标；t.TempDir 自动清理 | - |
