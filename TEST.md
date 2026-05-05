# 日志性能测试报告

## 测试环境
- 操作系统: darwin
- 架构: amd64
- CPU: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

## 基准测试结果 (v1.1.4)

| 测试用例 | 迭代次数 | 耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkLogger_RequestLog_Realistic` | 510,711 | 2,122 | 292 | 5 |
| `BenchmarkLoggerInfo` | 144,194 | 9,547 | - | - |
| `BenchmarkLoggerAsyncConcurrent` | 159,004 | 7,080 | - | - |

## 版本对比评估

| 版本 | 机制 | 存储格式 | 可视化 | 性能 (Async) |
| :--- | :--- | :--- | :--- | :--- |
| **v1.0.3** | Map 序列化 | JSON Object | 内置 | ~8,773 ns/op |
| **v1.1.4** | Meta-Driven Array | **JSON Array** | 独立工具/Meta | **~7,080 ns/op** |

## 总结
- **性能质变**: v1.1.4 通过 **Meta-Driven Positional Array** 架构，在异步并发场景下性能提升了约 20%。
- **存储优化**: 采用数组格式彻底消除了日志中重复 Key 的存储开销，极大地降低了磁盘占用与 ES 索引压力。
- **架构解耦**: 核心包不再感知具体的字段名称，通过外置的 `.log.meta.json` 实现极致的灵活扩展。
- **内存效率**: 通过零装箱 (No-Boxing) 直接字符串拼接技术，保持了极低的内存分配。
- **独立工具**: 配合 `logv` CLI 工具，实现了“落盘高性能数组，查看友好彩色文本”的完美闭环。
