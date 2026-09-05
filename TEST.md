# Test Results

## v1.5.12 验证
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
ok  	apigo.cc/go/log	3.301s
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
