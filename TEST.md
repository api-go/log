# 日志性能测试报告

## 测试环境
- 操作系统: darwin
- 架构: amd64
- CPU: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

## 基准测试结果 (v1.1.8)

| 测试用例 | 迭代次数 | 耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkLogger_RequestLog_Realistic` | 540,817 | 2,256 | 561 | 17 |
| `BenchmarkLoggerInfo` | 341,980 | 3,164 | - | - |
| `BenchmarkLoggerAsyncConcurrent` | 1,289,307 | 914 | - | - |

## 版本对比评估

| 版本 | 机制 | 存储格式 | 可视化 | 性能 (Async) |
| :--- | :--- | :--- | :--- | :--- |
| **v1.0.3** | Map 序列化 | JSON Object | 内置 | ~8,773 ns/op |
| **v1.1.4** | Meta-Driven Array | JSON Array | 独立工具/Meta | ~7,080 ns/op |
| **v1.1.6** | BaseLog Pointer Opt | JSON Array | 独立工具/Meta | ~7,445 ns/op |
| **v1.1.7** | Dead Code Removal | JSON Array | 独立工具/Meta | ~1,059 ns/op |
| **v1.1.8** | **Hybrid Deep Masking** | **JSON Array** | 独立工具/Meta | **~914 ns/op** |

## 总结
- **性能质变**: v1.1.8 通过 **混合序列化路径**，在实现深度脱敏的同时，将异步并发性能提升到了巅峰（**914ns**），真正做到了安全与性能兼得。
- **深度安全**: 借助 `cast.ToJSONDesensitizeBytes`，现在可以自动穿透并屏蔽嵌套在 Map、Struct 或 Slice 内部的任何敏感信息。
- **摩擦消除**: 统一了基础类型与复杂类型的处理逻辑，代码更易维护。

