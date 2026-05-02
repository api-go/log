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

## 快速开始
```go
import "apigo.cc/go/log"

func main() {
    // 使用默认 Logger
    log.Info("server started", "port", 8080)
    
    // 创建带 traceId 的子 Logger
    logger := log.New("unique-trace-id")
    logger.Info("request processed")
    
    // 错误日志带堆栈
    logger.Error("database failed", "db", "mysql")
}
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
