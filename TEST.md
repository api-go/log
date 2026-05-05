# 日志性能测试报告

## 测试环境
- 操作系统: darwin
- 架构: amd64
- CPU: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

## 基准测试结果 (v1.1.7)

| 测试用例 | 迭代次数 | 耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkLogger_RequestLog_Realistic` | 607,719 | 2,029 | 296 | 14 |
| `BenchmarkLoggerInfo` | 383,230 | 2,979 | - | - |
| `BenchmarkLoggerAsyncConcurrent` | 1,230,997 | 1,059 | - | - |

## 版本对比评估

| 版本 | 机制 | 存储格式 | 可视化 | 性能 (Async) |
| :--- | :--- | :--- | :--- | :--- |
| **v1.0.3** | Map 序列化 | JSON Object | 内置 | ~8,773 ns/op |
| **v1.1.4** | Meta-Driven Array | JSON Array | 独立工具/Meta | ~7,080 ns/op |
| **v1.1.6** | BaseLog Pointer Opt | JSON Array | 独立工具/Meta | ~7,445 ns/op |
| **v1.1.7** | Dead Code Removal | **JSON Array** | 独立工具/Meta | **~1,059 ns/op** |

## 总结
- **性能质变**: v1.1.7 通过 **移除冗余逻辑与死代码**，异步并发性能得到了跨越式提升（由 7445ns 降至 **1059ns**）。
- **脱敏加固**: 实现了全类型字段脱敏，并支持 `CamelCase` 与 `snake_case` 的自动对齐。
- **功能验证**: 闭环验证了 `SplitTag` 动态切分能力，确保在大规模日志滚动场景下的稳定性。

