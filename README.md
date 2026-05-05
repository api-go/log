# @go/log

> **Maintainer Statement:** 本项目完全由 AI 维护。任何改动均遵循代码质量与性能的最佳实践。

## 🎯 设计哲学

`@go/log` 旨在提供高性能、零摩擦的异步日志系统。其核心目标是：

*   **零摩擦入口**：自动识别环境上下文（应用名、IP等），无需手动构建。
*   **极致高性能**：异步写入架构，支持对象池复用，大幅降低内存分配。
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
    * `Log(LogEntry)` —— 记录自定义结构的日志。注意：仅支持实现 `LogEntry` 接口的类型（即嵌入了 `BaseLog` 的结构体）。

3.  **专业日志扩展**
    * **请求日志 (`Request`)**: 记录 HTTP 请求，包含方法、路径、状态码、耗时等。
    * **数据库日志 (`DB`)**: 自动计算耗时、捕获调用栈并支持脱敏。
    * **监控与统计 (`Monitor`, `Statistic`)**: 用于应用指标监控。
    * **任务执行 (`Task`)**: 用于任务耗时与状态记录。

### 自定义日志扩展

如果标准日志分级不能满足业务需求，可以轻松扩展自定义日志类型：

1.  **定义结构体**：必须嵌入 `log.BaseLog` 以自动获得基础字段和池化能力。
2.  **获取对象**：使用 `log.GetEntry[MyLog]()` 从对象池获取，避免频繁分配内存。
3.  **业务逻辑**：仅需关注业务相关的字段，`BaseLog` 中的字段（时间、TraceId、服务器信息等）由框架自动填充，map/slice等字段框架会自动初始化好（避免对象重复创建）直接使用即可。
4.  **发送日志**：调用 `logger.Log(entry)`。

```go
type BusinessLog struct {
    log.BaseLog  // 必须嵌入
    Action string
    UserId string
}

func LogBusiness(logger *log.Logger, action, userId string) {
    entry := log.GetEntry[BusinessLog]()
    entry.Action = action
    entry.UserId = userId
    logger.Log(entry) // 框架会自动填充 BaseLog 并异步写入后回收对象
}
```

### 配置项 (JSON/YAML)

*   `Name`: 应用名称。
*   `Level`: 日志级别 (`debug`, `info`, `warning`, `error`)。
*   `File`: 输出目标（支持 `console` 或 `es://` 地址）。
*   `Sensitive`, `RegexSensitive`: 脱敏配置。

## 🧪 验证状态
测试全部通过，异步写入与性能达标。

详见：[TEST.md](./TEST.md)
