# Changelog

## [1.0.0] - 2026-05-02
- **初始版本**: 由 `ssgo/log` 迁移并基于 `apigo.cc/go` 标准重构。
- **高性能引擎**: 引入 `LogEntry` 池化与 `sync.Pool` 复用，支持零分配日志对象。
- **异步写入**: 实现基于 Channel 的非阻塞异步写入引擎，将 IO 压力从主路径完全剥离。
- **批量刷盘**: `FileWriter` 引入 `bufio.Writer` 缓冲，`ESWriter` 优化 Bulk 请求构造，大幅提升 IO 吞吐量。
- **灵活格式**: 引入 `Formatter` 接口，支持 JSON 与文本等自定义格式。
- **增强视图**: 内置 `ConsoleWriter` 与增强型 `Viewer`，支持 `RequestLog` 高亮显示与状态码变色。
- **字段规范**: 统一字段 Key 为小写，确保跨平台解析一致性。
- **安全性**: 集成高性能字段脱敏能力，并提供幂等停机与 Panic 恢复机制。
