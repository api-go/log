# 日志性能测试报告

## 测试环境
- 操作系统: darwin
- 架构: amd64
- CPU: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

## 基准测试结果 (v1.1.10)

| 测试用例 | 迭代次数 | 耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkLogger_RequestLog_Realistic` | 544,791 | 2,230 | 561 | 17 |
| `BenchmarkLoggerInfo` | 368,821 | 3,042 | - | - |
| `BenchmarkLoggerAsyncConcurrent` | 1,216,018 | 919 | - | - |

## 版本对比评估

| 版本 | 机制 | 存储格式 | 可视化 | 性能 (Async) |
| :--- | :--- | :--- | :--- | :--- |
| **v1.0.3** | Map 序列化 | JSON Object | 内置 | ~8,773 ns/op |
| **v1.1.4** | Meta-Driven Array | JSON Array | 独立工具/Meta | ~7,080 ns/op |
| **v1.1.6** | BaseLog Pointer Opt | JSON Array | 独立工具/Meta | ~7,445 ns/op |
| **v1.1.7** | Dead Code Removal | JSON Array | 独立工具/Meta | ~1,059 ns/op |
| **v1.1.8** | Hybrid Deep Masking | JSON Array | 独立工具/Meta | ~914 ns/op |
| **v1.1.10** | **Stability & Infrastructure** | **JSON Array** | 独立工具/Meta | **~919 ns/op** |

## 总结
- **稳定性增强**: v1.1.10 修复了 `SplitTag` 轮转测试在秒级边界时的 race condition，测试用例更加健壮。
- **基础设施对齐**: 核心元数据解析与测试用例全面对齐 `go/cast` 与 `go/file` 基础设施，消除了原生 `strconv` 与 `os` 操作的摩擦。
- **性能维持**: 在增强稳定性的同时，异步并发性能稳定维持在 **~919ns**，保持了极高的吞吐能力。

