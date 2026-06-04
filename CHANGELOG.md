# Changelog

## [1.5.3] - 2026-06-04
- **功能增强: 标准库日志全量审计**:
  - **实现**: 引入了 `log.RedirectStdLog()` 机制，并在 `DefaultLogger` 初始化时自动启用。该功能会劫持 Go 标准库 `log` 包的所有输出并转发到我们的结构化 Logger。
  - **收益**: 确保了如 `SugarDB`, `Bleve` 等第三方依赖产生的原始日志也能自动携带 `serverId` 并符合全栈追踪规范。

## [1.5.2] - 2026-06-04
- **稳定性修复: 异步引擎平滑退出**:
  - **修复**: 重构 `writerService` 的通道生命周期管理，彻底解决了在 `starter` 多次启停或服务退出时引发的 `panic: close of closed channel` 致命错误。

## [1.5.1] - 2026-06-04
- **全栈基础设施对齐**: 
  - 将所有内部依赖统一升级至 `v1.5.0` 语义版本。

## [1.3.2] - 2026-05-13
- **功能增强: 引入摩擦消除工具 `As`**:
  - **泛型支持**: 新增全局泛型函数 `log.As[T](v T, err error) T`，仿照 `cast.As` 设计，自动记录错误并返回零值，极大简化了带 error 返回值的函数链式调用。
  - **Logger 扩展**: `Logger` 结构体新增 `As(v any, err error) any` 方法，支持实例级别的错误捕获与自动记录。
  - **调用栈优化**: 优化了 `GetCallStacks` 逻辑，自动跳过 `default_logger.go` 中的内部帧，确保 `log.As` 记录的错误位置精准指向业务代码。

## [1.3.1] - 2026-05-12
- **架构升级: 引入 LoggerService**:
  - **解耦重构**: 重构全局变量管理，引入 `loggerService` 结构体，集中化管理异步写入协程、Writers 对象池、文件句柄与丢弃计数。
  - **生命周期管理**: 实现了 `Start(ctx, logger)`, `Stop(ctx)`, `Health()` 接口，完美支持与 `apigo.cc/go/starter` 的基础设施集成。
  - **安全性增强**: 优化了平滑停止逻辑，确保 `Stop` 调用时能完整 Flush 缓冲区数据。
- **功能增强**:
  - **动态应用名**: 新增 `SetDefaultName(name)` 全局方法，支持在微服务启动后动态设置应用名称，并自动同步至 `DefaultLogger`。
- **配置与忽略规则**:
  - 更新 `.gitignore`，增加了对 `.gemini`, `env.json`, `env.yml` 以及 `/CODE-FULL.md` 的忽略支持。
- **文档与测试同步**: 全面更新了 `README.md`, `CHANGELOG.md` 与 `TEST.md`。

## [1.3.0] - 2026-05-10
- **全栈基础设施对齐**: 
  - 将所有内部依赖（`cast`, `config`, `file`, `id`, `shell` 等）统一升级至 `v1.3.0` 语义版本。
  - 优化项目忽略规则，在 `.gitignore` 中增加了对 `.geminiignore` 的支持。

## [1.1.16] - 2026-05-10
- **依赖对齐**: 同步更新 `golang.org/x/crypto` 和 `golang.org/x/sys` 至最新版本，确保底层安全性。

## [1.1.15] - 2026-05-10
- **安全与性能双重增强**:
  - **敏感字段扩展**: `authorization` 现在默认纳入脱敏规则 (`LogDefaultSensitive`)，自动过滤 HTTP 认证凭证。
  - **TraceId 吞吐优化**: 切换至 `id.Get10Bytes14MPerSecond()` 算法，在保持 10 字节唯一性的同时，极大提升了超高并发下的 ID 生成性能。
  - **应用名称识别对齐**: 重构 `GetDefaultName` 逻辑，优先通过 `debug.ReadBuildInfo()` 识别 Module 路径，并支持自动裁剪 Windows 环境下的 `.exe` 后缀。

## [1.1.14] - 2026-05-09
- **可视化能力调整**: 调整优化了 `viewer` 模块相关的可视化能力，提升了日志的可读性与调试体验。

## [1.1.13] - 2026-05-09
- **绝对索引优化与零空洞**:
  - 彻底消除 `BaseLog` 与业务字段之间的索引空洞。字段位置调整为：`BaseLog` (0-5)，标准消息字段 (`Info`, `Error` 等) 与业务日志字段从 `pos: 6` 起始。
  - **智能平移映射**: 改进 `pos >= 1000` 的逻辑。这些字段（如 `Extra`, `CallStacks`）不再产生稀疏 0 占位符，而是根据定义的 `pos` 顺序，自动平移并紧跟在当前类型的最大绝对索引之后。既保证了“永远在末尾”，又维持了数组的紧凑性。
- **强制 Reset 安全契约**:
  - `RegisterType` 引入严苛校验：若自定义日志类型包含字段但未显式重写 `Reset()`（即仅继承了 `BaseLog.Reset`），注册时将触发 **Panic**。强制开发者显式清理业务数据，杜绝对象池复用时的脏数据隐患。
- **应用名称识别增强**:
  - `GetDefaultName` 引入 `runtime/debug.ReadBuildInfo()`。相比传统的文件夹路径读取，能更精准地识别 Go Module 定义的应用名称。
- **工具链增强**:
  - `serializer_test.go` 新增原始日志输出，方便开发者通过 `go test -v` 直观校验 JSON 数组结构。

## [1.1.12] - 2026-05-09
...
- **稳定性增强**:
  - 修复 `TestSplitTag` 在秒级进位边界时的偶发性失败。优化后的测试逻辑同时校验日志产生前后的两个潜在时间槽，闭环消除了环境抖动导致的 race condition。

## [1.1.9] - 2026-05-05
- **稳定性增强**:
  - 改进 `SplitTag` 测试用例，通过秒级切分步进，闭环验证了 `FileWriter` 的文件自动轮转逻辑，确保在高频或定时切分场景下的可靠性。

## [1.1.8] - 2026-05-05
- **深度脱敏与性能平衡**:
  - 引入“混合路径”序列化策略：对基础类型（String, Int, Bool 等）采用高性能手动编码，对复杂类型（Struct, Map, Slice）采用 `cast.ToJSONDesensitizeBytes`。
  - 完美解决嵌套对象、数组内对象的深度脱敏问题，同时将序列化性能维持在极高水平。
- **基准测试优化**: 相比全量 `cast` 调用，混合路径将异步写入耗时降低了 60% 以上，内存分配减少了 70%。

## [1.1.7] - 2026-05-05
- **极致精简与摩擦消除**:
  - 移除了所有冗余或已弃用的配置项：`Fast`, `KeepKeyCase`, `RegexSensitive`, `SensitiveRule`。
  - 删除了 `Logger` 中未被使用的 `desensitization` 处理函数及 `SetDesensitization` 方法，进一步收窄 API。
- **功能增强与修复**:
  - **脱敏逻辑闭环**: 修复了 `ToArrayBytes` 仅对 Map 字段脱敏的缺陷。现在支持对结构体根层级的敏感字段（如 `String` 类型）进行全量遮蔽。
  - **标准化命名映射**: 增强 `fixField` 逻辑，同时支持自动忽略下划线 (`_`) 和横线 (`-`)，确保 `SecretKey` (Go) 与 `secret_key` (Config) 能完美匹配。
- **稳定性保障**:
  - 新增 `functional_test.go`，闭环验证了 `SplitTag` 文件切分能力与增强后的 `Sensitive` 脱敏逻辑。
  - 优化 `serializer.go` 内部判断逻辑，基准测试显示性能有显著提升。

## [1.1.6] - 2026-05-05
- **性能优化**:
  - 重构 `FillBase` 方法签名，由接收接口 `LogEntry` 改为接收指针 `*BaseLog`。
  - 此项改动消除了在填充元数据时通过接口调用 `GetBaseLog()` 的开销，直接操作结构体指针，进一步提升日志预处理性能。
- **文档与示例对齐**:
  - 完善 `README.md` 中的 `Config` 配置项说明，涵盖 `Fast`, `KeepKeyCase`, `Truncations`, `SensitiveRule` 等所有字段。
  - 修正 `BusinessLog` 扩展示例，确保与最新的 `FillBase` 签名及元数据填充逻辑保持一致。
- **扩展与迁移指引**: 同步更新 `extra.go` 中的注释示例，为从旧版本或其他项目迁移提供准确的参考实现。

## [1.1.5] - 2026-05-05
- **高性能 Meta 驱动架构**: 
  - 日志存储格式由 JSON Object 彻底切换为 **JSON Positional Array (`[...]`)**，通过位置索引消除重复 Key 的存储与传输开销。
  - 实现基于反射的 **零装箱 (No-Boxing) 序列化**，直接拼接 JSON 字符串，大幅降低内存分配与 CPU 占用。
- **元数据外置与可视化**:
  - 引入 `.log.meta.json` 机制，将字段顺序、颜色、格式化等可视化逻辑从核心包中剥离。
  - 新增独立 CLI 工具 `logv`，支持基于元数据文件的 stdin 流式日志渲染。
  - 改造 `Viewable` 接口，支持基于 MetaRegistry 的动态终端彩色输出渲染。
- **字段优化与压缩**:
  - 合并 `ImageName/Tag` 为 `Image`，合并 `ServerName/Ip` 为 `Server`，精简日志槽位。
  - 引入 `hide:true` 标签，支持在控制台隐藏结构性元数据（如 `LogName`, `LogType`），保持输出极致纯净。
- **架构兼容性**:
  - 调整 `Writer` 接口，支持自定义 Writer（如 `ESWriter`）自行决定序列化格式（如转回 Object 以适配 ES 索引）。
  - `utility.go` 新增对旧版 JSON 对象日志的解析兼容。
- **安全加固**: 为 `http`, `db`, `redis`, `discover` 等子包同步添加 `.gitignore` 以排除自动生成的 `.log.meta.json`。

## [1.1.2] - 2026-05-05
- **架构解耦**: 
  - 正式移除 `log` 包对数据库日志（`DB` 方法及 `DBLog` 结构）的内置支持，推动“日志格式随业务走”的架构对齐。
  - 导出 `GetCallStacks` 方法，支持外部包捕获符合截断配置的调用栈。
- **内部优化**: 
  - `fillBase` 回归私有，强化 `Log(entry)` 的自动填充契约，降低外部摩擦。
- **体验对齐**:
  - 修复 `viewer.go` 中空错误字符串导致日志类型被跳过显示的 Bug。
  - `Log` 方法集成 `fillBase` 自动补全逻辑，对于直接调用 `Log(entry)` 的自定义日志类型，若未设置时间戳则自动填充元数据，无需手动干预。
- **生态同步**: 依赖 `config` 升级至 v1.0.5。

## [1.1.0] - 2026-05-05
- **性能质变**: 
  - `GetEntry` 改为泛型实现 `GetEntry[T]()`，消除 `reflect.Type` 传参及运行时类型断言。
  - 主路径零反射：增强 `LogEntry` 接口，支持通过 `GetBaseLog()` 直接获取 `BaseLog` 指针，填充基础字段耗时降至 $O(1)$。
  - 极速时间戳：`LogTime` 改为 `int64` (纳秒)，完全消除了日志写入时的 `time.Format` 开销。
  - `FileWriter` 重构：采用单协程 `bufio` 无锁写入架构，支持 64KB 缓冲区与高性能文件轮转。
- **架构加固**: 
  - 引入 `logPayload` 路由包裹，彻底解决多 Logger 实例异步写入时的“串话”漏洞。
  - 修复同步 Fallback 模式下的线程安全隐患。
- **扩展恢复**: 重新启用 `extra.go` 中的标准日志结构（Request, DB, Task, Monitor, Statistic），并完成泛型化适配。
- **可视化优化**: 重构 `Viewable` 逻辑，基于 `strings.Builder` 和 `cast` 包优化 JSON 解析，渲染速度提升 2-3 倍。

## [1.0.3] - 2026-05-05
- **接口规范化**: 重构 `Logger.Request` 方法，由传递结构体指针改为接收明确参数（method, path, status, usedTime），并支持通过 `extra` 自动填充 `RequestLog` 的 20+ 个业务字段。
- **架构解耦**: 将 `extra.go` 中的所有预定义日志结构及其快捷方法注释掉。仅作为示例供应用端参考实现，推动“应用自维护日志结构”的标准规范。
- **类型安全性**: 强化 `Logger.Log` 方法约束，仅接受实现 `LogEntry` 接口（通常继承自 `BaseLog`）的类型，不再支持任意 `any` 格式。
- **文档完善**: `README.md` 新增“自定义日志扩展”指南，明确了 `BaseLog` 继承、`GetEntry` 对象池获取及 `Log` 发送的完整闭环。
- **测试对齐**: 更新所有测试用例与基准测试，演示如何在应用端定义和使用自定义日志结构。

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
