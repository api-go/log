# @go/log

> **Maintainer Statement:** 本项目完全由 AI 维护。任何改动均遵循代码质量与性能的最佳实践。

## 🎯 设计哲学

`@go/log` 旨在提供高性能、零摩擦的异步日志系统。其核心目标是：

*   **极致高性能**：采用 **Meta-Driven Positional Array (元数据驱动定长数组)** 架构。日志以单行 JSON 数组 (`[...]`) 形式落盘，消除 Key 冗余与装箱开销，性能提升数倍。
*   **架构解耦**：元数据外置于 `.log.meta.json`。日志包仅负责高速序列化，可视化由外部工具或 `Viewable` 接口根据元数据动态渲染。
*   **零摩擦入口**：自动识别环境上下文（应用名、IP等），无需手动构建。
*   **语义脱敏**：内置敏感信息（如手机号、密钥）的自动脱敏。
*   **高度可扩展**：支持多种写入渠道（文件切分、Elasticsearch批量传输）。

## 📦 安装

```bash
go get github.com/api-go/log
```

## 💡 快速开始

```go
import "github.com/api-go/log"

// 默认 logger (通过 log.json 或环境变量配置)
func main() {
	// 应用内置预设；log.json/log.yml、env 配置和环境变量仍可覆盖
	log.SetDefaultConfig(log.Config{
		File:     "logs/app.log",
		SplitTag: ".2006-01-02",
	})

	// 在微服务场景下动态设置应用名称
	log.SetDefaultName("my-microservice")

	log.Info("服务启动", "port", 8080)
	log.Error("数据库连接失败", "db", "mysql")

	// 创建带 traceId 的新 logger 实例
	logger := log.New("trace-xyz-123")
}
```

## ⚙️ 配置 (Configuration)

本包深度集成 `@go/config`，支持多种灵活的配置方式，优先级从高到低：

1.  **环境变量** (最高优先级)
2.  **环境特定文件** (`env.json` / `env.yml`，需增加层级 `log:`)
3.  **基础配置文件** (`log.json` / `log.yml`)

### 1. 配置文件 (`log.json`)

在项目根目录创建 `log.json` 或 `log.yml`：

```json
{
  "name": "my-cool-app",
  "level": "info",
  "file": "logs/app.log",
  "splitTag": ".2006-01-02",
  "sensitive": "phone,password,secret,token,accessToken,authorization"
}
```

### 2. 环境变量 (最高优先级)

任何配置都可以通过环境变量覆盖，变量名规则为 `LOG_` + `字段名`。

```bash
# 覆盖日志级别和输出文件
export LOG_LEVEL=debug
export LOG_FILE=console

```

### 配置项说明

*   `name`: 应用名称 (默认通过 `debug.ReadBuildInfo()` 或 `os.Args[0]` 自动识别)。
*   `level`: 日志级别 (`debug`, `info`, `warning`, `error`)。
*   `file`: 输出目标。
    *   `console`: 直接输出到控制台（默认）。
    *   `path/to/file.log`: 输出到指定文件。
    *   `es://...` 或 `ess://...`: 输出到 Elasticsearch。
*   `splitTag`: 文件切分格式，仅当 `file` 为文件路径时有效。
    *   语法遵循 Go 标准的 `time.Format` 布局，如 `".2006-01-02"` (按天切分)，`".2006-01-02-15"` (按小时切分)。
*   `truncations`: 堆栈信息截断前缀（多个以逗号分隔，默认截断 `github.com/`, `golang.org/`, `/apigo.cc/`）。
*   `sensitive`: 需要自动脱敏的字段名（多个以逗号分隔，不区分大小写，并忽略 `_`、`-`），默认处理 `phone,password,secret,token,accessToken,authorization`。该规则递归应用于嵌套对象和数组，例如 `accessToken`、`ACCESS_TOKEN` 与 `access-token` 等价。

文件输出会在 logger 初始化时自动创建父目录。文件中每行均为纯 JSON 数组，不附加标准日志时间前缀，可直接交给 `logv`。`splitTag` 为空时保持 `file` 原名；非空时文件名为 `file + "." + time.Format(splitTag)`（布局自身带点时会出现两个点，保留既有规则）。所有文件日志在异步 writer service 启动前及停止后同步刷新，运行中异步写入；停止服务会排空队列并刷新缓冲。历史文件中的时间前缀不会自动移除。

## 🛠 API 指南

### API 索引

- 配置与实例：`Config`, `Logger`, `DefaultLogger`, `NewLogger`, `New`, `SetDefaultConfig`, `SetDefaultName`；实例支持 `SetName`, `SetLevel`, `CheckLevel`, `GetTraceId`, `New`。
- 记录与填充：`Debug`, `Info`, `Warning`, `Error`, `As`；实例另支持 `Log`, `FillBase`, `FillDebug`, `FillInfo`, `FillWarning`, `FillError`, `GetCallStacks`。
- 日志模型：`LogEntry`, `BaseLog`, `DebugLog`, `InfoLog`, `WarningLog`, `ErrorLog`；模型支持 `Reset`，基础模型提供 `IsLogEntry`, `GetBaseLog`。`LevelType` 对应 `DEBUG`, `INFO`, `WARNING`, `ERROR`, `CLOSE`；`LogType*`、`LogEnv*` 与 `LogDefaultSensitive` 为类型、环境变量及默认脱敏常量。
- 模型与元数据：`RegisterType`, `GetEntry`, `WithEntry`, `PoolManager`, `MetaField`, `GetMeta`, `LoadMeta`, `SetMetaFilePath`, `Marshal`。
- 查看与转换：`Viewable`, `ToJSON`。
- 写入扩展：`Writer`（`Log`, `Run`）, `RegisterWriterMaker`, `ConsoleWriter`（`Log`, `Run`）, `FileWriter`（`Write`, `Run`, `Close`）。文件 Writer 由 `NewLogger` 配置和共享，不应复制。
- 生命周期：`WriterService.Start`, `WriterService.Stop`, `WriterService.Status`, `GetDroppedLogs`。异步队列过载时可能丢弃日志，计数可查询。
- 标准日志接入：`Logger.RedirectStdLog`, `AddStdLogFilter`, `SetStdLogOutput`。

### 核心功能

1.  **应用默认配置**
    * `SetDefaultConfig(Config)` —— 设置应用内置预设，然后加载配置文件与环境变量覆盖；最终有效配置未变化时复用当前默认 logger。

2.  **分级记录**
    * `Debug`, `Info`, `Warning`, `Error` —— 标准日志方法，支持 `message` + 变长 `extra` 参数。

3.  **摩擦消除 (`As`)**
    * `As(v, err)` —— 仿照 `cast.As`，忽略错误并返回零值，但会自动将错误记录到日志中。支持全局调用 (`log.As`) 或实例调用 (`logger.As`)。
    * **优势**: 在类型转换或快速赋值场景下，无需繁琐的 `if err != nil` 判断，同时确保异常被记录。

4.  **通用记录 (`Log`)**
    * `Log(LogEntry)` —— 记录自定义结构的日志。

5.  **独立可视化工具 (`logv`)**
    *   **安装**: `go install github.com/api-go/log/logv@latest`
    *   **使用**: `tail -f app.log | logv` 或 `tail -f app.log | logv -json`。
    *   `logv` 从当前目录加载 `.log.meta.json`；基础 Meta 会与应用已有业务 Meta 合并，不会覆盖业务日志类型。
    *   单条日志最大支持 16 MiB；超过限制时会返回非零退出码并报告读取错误。

### 自定义日志扩展 (规范)

为保证高性能与内存安全，扩展自定义日志类型必须遵循以下规范：

1.  **定义结构体**
    *   必须嵌入 `log.BaseLog` (或其子类，如 `log.ErrorLog`)。
    *   **索引 (`pos`) 规范**:
        *   `0`-`6` 由 `BaseLog`。
        *   业务字段从 `6` 开始紧凑递增编号 (`pos:6`, `pos:7`, ...)，如果删除了某个字段请留空 pos 以实现向前兼容。
		*   如果继承自 `ErrorLog` 等，则业务字段应从 `7` 开始（查询父类最大值 + 1）。
        *   `Extra` 固定使用 `pos:1000`，`CallStacks` 固定使用 `pos:1001` (它们会被自动平移到数组末尾)。

2.  **实现 `Reset()` 方法 (强制)**
    *   **必须**重写 `Reset()` 方法以初始化/清空数据避免对象池复用时产生脏数据。
    *   `Reset()` 方法中必须首先调用父级的 `Reset()` (如 `l.BaseLog.Reset()`)。
    *   **安全保障**: 若未重写 `Reset`，`RegisterType` 将在启动时 **Panic**，以防止对象池复用时产生脏数据。
	*   **建议**: map / slice 类型建第一次初始化一个容量，之后使用 clear() 方法清空数据避免内存重复分配。

3.  **注册模型**
    *   在 `init()` 中调用 `log.RegisterType("my-type", MyLog{})` 完成注册。

#### 示例: `DBErrorLog`

```go
package main

import "github.com/api-go/log"

// 1. 定义结构体 (字段从 pos:6 开始)
type DBErrorLog struct {
	log.ErrorLog // 嵌入 ErrorLog，自动获得 Error 和 CallStacks 字段
	DB       string  `log:"pos:7,color:blue"`
	SQL      string  `log:"pos:8"`
	Args     []any   `log:"pos:9"`
	UsedTime float32 `log:"pos:10,color:cyan"`
}

// 2. 实现 Reset() 方法 (强制)
func (l *DBErrorLog) Reset() {
	l.ErrorLog.Reset() // 必须先调用父级 Reset
	l.DB = ""
	l.SQL = ""
	if l.Args == nil {
		l.Args = make([]any, 0, 10)
	} else {
		clear(l.Args) // 清空内容
		l.Args = l.Args[:0] // 清空长度
	}
	l.UsedTime = 0
}

// 3. 注册
func init() {
	log.RegisterType("dbError", &DBErrorLog{})
}

// 4. 使用示例
func LogDBError(logger *log.Logger, db, sql string, args []any, err error, usedTime float32) {
	entry := log.GetEntry[DBErrorLog]()

	// 自动填充基础字段和 ErrorLog 字段
	logger.FillError(&entry.ErrorLog, err.Error())
	
	// 填充自定义字段
	entry.DB = db
	entry.SQL = sql
	entry.Args = append(entry.Args, args...)
	entry.UsedTime = usedTime

	logger.Log(entry)
}
```


## 🧪 验证状态
测试全部通过，异步写入与性能达标。

详见：[TEST.md](./TEST.md)
