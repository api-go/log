# Test Results

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
--- PASS: TestSplitTag (1.80s)
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
ok  	apigo.cc/go/log	2.246s
```

## 核心指标验证
- **初始化安全性**: `TestLoggerCore_Initialization` 确保 Logger 实例配置正确加载。
- **高并发稳定性**: `TestLoggerCore_Concurrency` 验证了在多协程竞争环境下日志写入的线程安全。
- **元数据驱动验证**: `TestMetaExtraction` 与 `TestLoadMeta` 确保 `.log.meta.json` 协议的解析与应用。
- **序列化性能**: `TestToArrayBytes` 验证了 Positional Array 格式的正确性。
- **深度脱敏能力**: `TestDeepDesensitization` 闭环验证了对复杂嵌套结构的脱敏逻辑。
- **可靠性边界**: `TestLoggerReliability` 模拟了极高压力下的日志丢弃与缓冲策略。
- **文件切分**: `TestSplitTag` 实测了基于时间滚动的文件切分能力。
