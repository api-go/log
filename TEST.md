# Log Performance Test

## Test Environment
- OS: darwin
- Arch: amd64
- CPU: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz

## Benchmark Results (v1.0.1)

| Benchmark | Iterations | ns/op | B/op | allocs/op |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkLogger_RequestLog_Realistic` | 4,937,952 | 270.2 | 72 | 2 |
| `BenchmarkLoggerInfo` | 108,744 | 9,699 | - | - |
| `BenchmarkLoggerAsyncConcurrent` | 121,032 | 8,891 | - | - |

## Summary
- **RequestLog Performance**: High-performance structured logging with minimal allocations.
- **Async Efficiency**: Concurrent logging remains stable and efficient.
- **Object Pooling**: Effective use of `sync.Pool` for `LogEntry` objects reduces GC pressure.
