# @go/log

> **Maintainer Statement:** 本项目完全由 AI 维护。任何改动均遵循代码质量与性能的最佳实践。

## 🎯 设计哲学

`@go/log` 旨在提供高性能、零摩擦的异步日志系统。其核心目标是：

*   **极致高性能**：采用 **Meta-Driven Positional Array (元数据驱动定长数组)** 架构。日志以单行 JSON 数组 (`[...]`) 形式落盘，消除 Key 冗余与装箱开销，性能提升数倍。
*   **架构解耦**：元数据外置于 `.log.meta.json`。日志包仅负责高速序列化，可视化由外部工具或 `Viewable` 接口根据元数据动态渲染。
*   **零摩擦入口**：自动识别环境上下文（应用名、IP等），无需手动构建。
*   **语义脱敏**：内置敏感信息（如手机号、密钥）的自动脱敏与正则过滤。
*   **高度可扩展**：支持多种写入渠道（文件切分、Elasticsearch批量传输）。

## 📦 安装

```bash
go get apigo.cc/go/log
```

## 💡 快速开始

```go
import "apigo.cc/go/log"

// 使用默认配置初始化 (或在配置中指定)
logger := log.NewLogger(log.Config{Name: "my-app", Level: "info"})

// 记录业务日志 (自动通过 cast.ToMap 处理变长参数)
logger.Info("用户登录", "userId", 10086, "ip", "1.2.3.4")
logger.Error("数据库连接失败", "db", "mysql", "err", err)
```

## 🛠 API 指南

### 核心功能

1.  **分级记录**
    * `Debug`, `Info`, `Warning`, `Error` —— 标准日志方法，支持 `message` + 变长 `extra` 参数。

2.  **通用记录 (`Log`)**
    * `Log(LogEntry)` —— 记录自定义结构的日志。注意：仅支持实现 `LogEntry` 接口的类型。

3.  **独立可视化工具 (`logv`)**
    *   **安装**：
        ```bash
        go install apigo.cc/go/log/logv@latest
        ```
    *   **使用**：`tail -f app.log | logv` `tail -f app.log | logv -json`，依赖当前目录的 `.log.meta.json` 文件。

### 自定义日志扩展

如果标准日志分级不能满足业务需求，可以轻松扩展自定义日志类型：

1.  **定义结构体**：必须嵌入 `log.BaseLog` 或 `log.ErrorLog` 等结构体以实现 `LogEntry` 接口。
2.  **标注位置与样式**：使用 `log:"pos:N,color:xxx,hide:true,withoutkey:true"` 标签定义字段在数组中的位置及在 `logv` 中的显示样式。
3.  **注册模型**：在 `init()` 中调用 `log.RegisterType("my-type", MyLog{})`。
4.  **获取与发送**：使用 `log.GetEntry[MyLog]()` 并调用 `logger.Log(entry)`。
5.  **参考示例**: log/extra.go。

```go
type BusinessLog struct {
    log.BaseLog  // 必须嵌入
    Action string `log:"pos:10,color:cyan"`
    UserId string `log:"pos:11"`
}

func init() {
    log.RegisterType("business", BusinessLog{})
}

func LogBusiness(logger *log.Logger, action, userId string) {
    entry := log.GetEntry[BusinessLog]()
    logger.FillBase(&entry.BaseLog, "business")
    entry.Action = action
    entry.UserId = userId
    logger.Log(entry) 
}
```

### 配置项 (JSON/YAML)

*   `Name`: 应用名称。
*   `Level`: 日志级别 (`debug`, `info`, `warning`, `error`)。
*   `File`: 输出目标（支持 `console` 或 `es://` 地址）。
*   `SplitTag`: 文件切分标识（仅在输出到文件时有效）。
*   `Truncations`: 堆栈信息截断前缀（多个以逗号分隔，默认截断 `github.com/`, `golang.org/`, `/apigo.cc/`）。
*   `Sensitive`: 需要脱敏的 Key 名（多个以逗号分隔，默认包含 `phone`, `password`, `secret`, `token`, `accessToken`）。


## 🧪 验证状态
测试全部通过，异步写入与性能达标。

详见：[TEST.md](./TEST.md)
