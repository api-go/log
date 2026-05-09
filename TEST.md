# 日志性能测试报告

## 测试环境
- 操作系统: darwin
- 架构: amd64
- CPU: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

## 基准测试结果 (v1.1.10)

| 测试用例 | 迭代次数 | 耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkLogger_RequestLog_Realistic` | 344,300 | 3,338 | 1,331 | 19 |
| `BenchmarkLoggerInfo` | 291,952 | 4,083 | - | - |
| `BenchmarkLoggerAsyncConcurrent` | 784,453 | 1,466 | - | - |

## 版本对比评估

| 版本 | 机制 | 存储格式 | 可视化 | 性能 (Async) |
| :--- | :--- | :--- | :--- | :--- |
| **v1.0.3** | Map 序列化 | JSON Object | 内置 | ~8,773 ns/op |
| **v1.1.7** | Dead Code Removal | JSON Array | 独立工具/Meta | ~1,059 ns/op |
| **v1.1.10** | Stability & Infrastructure | JSON Array | 独立工具/Meta | ~919 ns/op |
| **v1.1.11** | **Absolute Indexing (Schema)** | **Fixed Array** | **LogType Opt** | **~1,466 ns/op** |

## 总结
- **Schema 兼容性**: v1.1.11 实现了 `pos` 绝对索引。虽然因数组稀疏化（填充 0）导致序列化开销略微增加（~1.4µs），但换取了极强的 Schema 稳定性，适配各类数仓接入。
- **可观测性**: 引入 `droppedLogs` 监控，解决了高并发场景下日志丢弃“黑盒”的问题。
- **鲁棒性**: 替换为 UDP 拨号法获取 IP，消除了在 K8s 等复杂网络环境下的识别摩擦。


