# Changelog

## [1.0.2] - 2026-05-04
- **设计优化**: 引入 `ResetLogEntry` 自动化重置机制，基于反射和缓存实现日志对象字段的自动初始化与清空（Map/Slice 默认容量 8）。
- **接口精简**: 简化 `LogEntry` 接口为标记接口，移除了冗余的 `Base()` 和 `Reset()` 手动实现。
- **扩展性增强**: `Task`, `Monitor`, `Statistic`, `DB` 等快捷方法全面支持变长 `extra ...any` 参数，并集成 `cast.ToMap` 自动转换。
- **构建修复**: 修复了 `convert` 模块对 `cast` 新 API 的兼容性问题。

## [1.0.1] - 2026-05-04
- **结构增强**: `DBLog` 结构体新增 `Error` 和 `CallStacks` 字段，提升数据库错误诊断效率。
- **DB 方法重构**: `Logger.DB` 方法支持可选错误参数，自动处理 `dbError` 类型并记录调用栈。
- **扩展日志支持**: 新增 `TaskLog`, `MonitorLog`, `StatisticLog` 标准结构及其 `Logger` 快捷方法，置于 `extra.go`。
- **RequestLog 封装**: `Logger` 新增 `Request` 方法，简化请求日志记录流程。
- **调用栈优化**: 优化 `getCallStacks` 逻辑，确保能正确捕获业务代码和测试代码的调用位置，同时过滤掉日志库内部帧。

## [1.0.0] - 2026-05-02
- **初始版本**: 由 `ssgo/log` 迁移并基于 `apigo.cc/go` 标准重构。
- **高性能引擎**: 引入 `LogEntry` 池化与 `sync.Pool` 复用，支持零分配日志对象。
- **异步写入**: 实现基于 Channel 的非阻塞异步写入引擎，将 IO 压力从主路径完全剥离。
- **批量刷盘**: `FileWriter` 引入 `bufio.Writer` 缓冲，`ESWriter` 优化 Bulk 请求构造，大幅提升 IO 吞吐量。
- **灵活格式**: 引入 `Formatter` 接口，支持 JSON 与文本等自定义格式。
- **增强视图**: 内置 `ConsoleWriter` 与增强型 `Viewer`，支持 `RequestLog` 高亮显示与状态码变色。
- **字段规范**: 统一字段 Key 为小写，确保跨平台解析一致性。
- **安全性**: 集成高性能字段脱敏能力，并提供幂等停机与 Panic 恢复机制。
