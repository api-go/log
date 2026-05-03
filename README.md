# go/log

高性能、可插拔、支持脱敏的日志模块。

## 特性
- **零摩擦**: 自动从环境变量获取应用名、IP 等信息。
- **高性能**: 异步写入，支持批量刷盘。
- **脱敏支持**: 内置敏感字段过滤与正则匹配脱敏。
- **多渠道**: 支持控制台、本地文件切分、Elasticsearch 批量写入。
- **现代化**: 深度集成 `apigo.cc/go` 基础库。

## 安装
```bash
go get apigo.cc/go/log
```

## 标准化日志 API

除了基础的 `Debug`, `Info`, `Warning`, `Error` 外，`go/log` 还提供了一系列针对特定场景优化的标准化日志 API：

### 数据库日志 (DB)
自动处理耗时计算、脱敏及错误堆栈捕获。
```go
// 记录正常 SQL
logger.DB("mysql", dsn, "SELECT * FROM users WHERE id=?", []any{1}, 10.5)

// 记录带错误的 SQL (自动捕获调用栈并设为 dbError 类型)
logger.DB("mysql", dsn, "SELECT...", args, usedTime, "table not found")
```

### 请求日志 (Request)
针对高性能 HTTP 服务设计的结构化日志。
```go
req := &log.RequestLog{
    Method: "GET",
    Path: "/api/user",
    // ... 填充其他字段
}
logger.Request(req)
```

### 任务与监控 (Task / Monitor / Statistic)
```go
// 任务执行日志
logger.Task("CleanCache", 150.2, true, "Success")

// 监控告警日志
logger.Monitor("CPU", 1, "Load too high")

// 业务指标统计
logger.Statistic("Business", "OrderCount", 100)
```

## 配置项 (JSON/YAML)
可以在配置文件中的 `log` 节点进行配置：
- `Name`: 应用名称（默认自动获取）
- `Level`: 日志级别 (debug, info, warning, error)
- `File`: 输出目标 (console, ./app.log, es://user:pass@host:9200/group)
- `SplitTag`: 文件切分格式 (如 20060102)
- `Sensitive`: 敏感字段列表
- `RegexSensitive`: 脱敏正则

## 脱敏规则
默认规则为 `12:4*4, 11:3*4, 7:2*2, 3:1*1, 2:1*0`。
格式为 `长度阈值:左保留*右保留`。
