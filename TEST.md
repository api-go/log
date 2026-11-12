## v1.5.17 验证（2026-10-08）

- `GOWORK=off go test -v ./...`、`GOWORK=off go test -bench=. ./...`、`GOWORK=off go test -race ./...` 均通过，使用已发布依赖。
- `TestFileRoutingLifecycle`：覆盖空与非空 SplitTag、父目录创建、启动前立即可读、40 个并发实例共享文件、停止排空、停止后同步写入、重启后写入及原文件名保持。
- 每条文件记录通过 `cast.UnmarshalJSON` 数组验证，并验证 `Viewable` 格式化和 `ToJSON` 对象转换；现有切分轮转与 1000 条并发文件写入测试通过。
- API 核对：当前 `ai` CLI 不支持 `tools/getExports`，改用 `GOWORK=off go doc -all` 与导出源码交叉核对，README 已补全 API。
- 文件追加继续使用现有 `os.OpenFile` 流式句柄：已发布 `go/file` 的 `Open` 仅只读、`Create` 会截断，不提供追加句柄；父目录创建和新测试读文件使用 `go/file`。

Benchmark（Apple M3 Max / arm64）：

```text
BenchmarkLoggerInfo                         1465 ns/op
BenchmarkLogger_RequestLog_Realistic         1075 ns/op    304 B/op    15 allocs/op
BenchmarkMarshalNestedSensitiveFields/disabled 697.0 ns/op 512 B/op    12 allocs/op
BenchmarkMarshalNestedSensitiveFields/normalized 1400 ns/op 696 B/op 21 allocs/op
```

异步并发基准会触发队列过载丢弃，其结果代表入队压力而非文件持久化吞吐。

# Test Results

## v1.5.16 验证
- 验证嵌套 map 与数组中的 `accessToken`、`ACCESS_TOKEN`、`access-token`、`Access_Token` 均按同一敏感字段脱敏。
- Apple M3 Max / arm64：嵌套对象不启用脱敏约 `0.70 µs/op, 512 B/op, 12 allocs/op`；接入 cast 单遍规范化脱敏约 `1.44 µs/op, 696 B/op, 21 allocs/op`。旧 cast 版本继续走安全兼容回退。
- `go test ./...` 全量通过。

## v1.5.15 验证
- 验证 `logv` 可完整处理 256 KiB 单行输入。
- 验证超过 16 MiB 的单行输入被有界拒绝，且错误包含明确的大小限制。
- 全量单元测试和基准测试均通过。

## v1.5.13 验证
- 验证 Meta 同步保留磁盘已有业务类型，并更新当前进程注册的同名类型。
- 验证本地 `logv` 格式化 `request` 日志时恢复时间、字段映射、隐藏规则和 ANSI 颜色。
- 验证 `logv` 运行后 `.log.meta.json` 仍保留 `api_call` 与 `request`。
- 全量单元测试、竞态测试和基准测试均通过。

## v1.5.10 验证
- 验证 split 文件 logger 在包初始化阶段安全创建，不依赖文件级 `init` 顺序。
- 验证文件 logger 自动创建父目录。
- 验证异步 writer 启动前的 split 日志同步写入目标文件。
- 验证代码默认配置保留未被配置文件覆盖的字段，并由配置文件覆盖已配置字段。
- 验证最终有效配置不变时 `SetDefaultConfig` 复用现有 logger。
- 全量单元测试、竞态测试和基准测试均通过。

### Benchmark（Apple M3 Max / arm64）

```text
BenchmarkLoggerInfo-16                       889536    1361 ns/op
BenchmarkLoggerAsyncConcurrent-16           4128828   286.5 ns/op
BenchmarkLogger_RequestLog_Realistic-16     1000000    1111 ns/op    975 B/op    27 allocs/op
```

## v1.5.9 验证
- 独立模块解析、全量测试与基准测试均通过。

## 单元测试报告

```text
=== RUN   TestLoggerCore_Initialization
--- PASS: TestLoggerCore_Initialization (0.00s)
=== RUN   TestLoggerCore_Concurrency
--- PASS: TestLoggerCore_Concurrency (0.00s)
=== RUN   TestMetaExtraction
--- PASS: TestMetaExtraction (0.00s)
=== RUN   TestWithEntry
--- PASS: TestWithEntry (0.00s)
=== RUN   TestLoggerReliability
--- PASS: TestLoggerReliability (0.01s)
=== RUN   TestToArrayBytes
    serializer_test.go:64: Raw log: ["test-app","mock_info_test",1620000000,"abc-123","Hello, World!",{"user_id":42}]
--- PASS: TestToArrayBytes (0.00s)
=== RUN   TestToArrayBytes_Desensitize
--- PASS: TestToArrayBytes_Desensitize (0.00s)
=== RUN   TestSplitTag
--- PASS: TestSplitTag (1.81s)
=== RUN   TestSensitiveDetailed
--- PASS: TestSensitiveDetailed (0.00s)
=== RUN   TestDeepDesensitization
--- PASS: TestDeepDesensitization (0.00s)
=== RUN   TestLogger
--- PASS: TestLogger (0.00s)
=== RUN   TestDesensitization
--- PASS: TestDesensitization (0.00s)
=== RUN   TestDBLog
--- PASS: TestDBLog (0.00s)
=== RUN   TestRequestLog
--- PASS: TestRequestLog (0.00s)
=== RUN   TestExtraLogs
--- PASS: TestExtraLogs (0.00s)
=== RUN   TestViewable
--- PASS: TestViewable (0.00s)
=== RUN   TestToJSON
--- PASS: TestToJSON (0.00s)
=== RUN   TestLoadMeta
--- PASS: TestLoadMeta (0.00s)
=== RUN   TestEnhancedViewable
--- PASS: TestEnhancedViewable (0.00s)
=== RUN   TestEnhancedToJSON
--- PASS: TestEnhancedToJSON (0.00s)
=== RUN   TestCallStacksViewable
--- PASS: TestCallStacksViewable (0.00s)
=== RUN   TestPrecisionViewable
--- PASS: TestPrecisionViewable (0.00s)
PASS
ok  	github.com/api-go/log	3.301s
```

## 核心指标验证
- **初始化安全性**: `TestLoggerCore_Initialization` 确保 Logger 实例配置正确加载。
- **高并发稳定性**: `TestLoggerCore_Concurrency` 验证了在多协程竞争环境下日志写入的线程安全。
- **元数据驱动验证**: `TestMetaExtraction` 与 `TestLoadMeta` 确保 `.log.meta.json` 协议的解析与应用。
- **序列化性能**: `TestToArrayBytes` 验证了 Positional Array 格式的正确性。
- **深度脱敏能力**: `TestDeepDesensitization` 闭环验证了对复杂嵌套结构的脱敏逻辑。
- **可靠性边界**: `TestLoggerReliability` 模拟了极高压力下的日志丢弃与缓冲策略。
- **文件切分**: `TestSplitTag` 实测了基于时间滚动的文件切分能力。
- **启动期文件写入**: `TestLoggerCore_SplitFileWritesBeforeWriterServiceStarts` 验证异步服务启动前日志直接写入 split 文件。
- **默认配置覆盖**: `TestLoadDefaultConfig` 与 `TestSetDefaultConfigKeepsLoggerWhenEffectiveConfigIsUnchanged` 验证代码预设、文件覆盖和无变化复用行为。

## GitHub migration validation (pending release)

Module paths and internal dependency versions are migrated. Development verification uses GOWORK=off and explicit sibling replacements; remove these replacements before release. See migration validation logs for actual results. Existing unrelated working changes are preserved and require separate review.

API inventory verification: `GOWORK=off go doc` succeeded; output retained in the migration validation directory. Source exports were preserved by the path migration.

## GitHub release verification

GOWORK=off go test -v ./... and go test -bench=. ./... passed using published GitHub dependencies, without local replacements. go mod verify passed.

## 实时压缩日志验证（2026-11-12，未发布）

- 压缩往返、嵌套脱敏、Unicode/转义、head/tail、重新打开追加、缺失/错配/损坏字典、不完整尾部恢复、follow 取消、压缩 stdin。
- 原有文件路由和生命周期测试显式验证 NoCompress 明文路径。
- API 使用 GOWORK=off go doc 与导出源码交叉核对（getExports 不在本仓库）。
- 开发验证使用 /tmp/apigo-validation-log.mod 显式替换本次 timer 修复；不依赖 go.work。

验证结果：完整测试与 race detector 通过；全量基准通过（benchtime=100ms）。同步压缩写入 BenchmarkArchiveWrite：9514 ns/op、3114 B/op、51 allocs/op（本机单次结果，不代表线上吞吐；异步路径另受队列容量约束）。

## 日志依赖联动 — 2026-11-12

开发验证使用 GOWORK=off 与 go.mod 的显式本地 replace；执行完整测试、全量基准（benchtime=100ms），API 用 go doc 与导出源码交叉核对。结果记录在 /tmp/apigo-log-integration-*.txt。
