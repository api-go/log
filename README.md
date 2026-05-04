# go/log

高性能、可插拔、支持脱敏的日志模块。

## 特性
- **零摩擦**: 自动从环境变量获取应用名、IP 等信息。
- **高性能**: 异步写入，支持对象池化与批量刷盘。
- **自动化**: 自定义日志类型只需嵌入 `BaseLog`，无需手动实现重置逻辑。
- **脱敏支持**: 内置敏感字段过滤与正则匹配脱敏。
- **多渠道**: 支持控制台、本地文件切分、Elasticsearch 批量写入。

## 安装
```bash
go get apigo.cc/go/log
```

## 基础 API
所有日志方法均支持变长额外参数，自动通过 `cast.ToMap` 转换为键值对存入 `Extra` 字段。
```go
logger.Info("用户登录", "userId", 10086, "ip", "1.2.3.4")
logger.Error("数据库连接失败", "db", "mysql", "err", err)
```

## 扩展日志 API

### 数据库日志 (DB)
自动处理耗时计算、脱敏及错误堆栈捕获。
```go
// 记录正常 SQL
logger.DB("mysql", dsn, "SELECT * FROM users WHERE id=?", []any{1}, 10.5, nil)

// 记录带错误的 SQL (自动捕获调用栈并设为 dbError 类型)
logger.DB("mysql", dsn, "SELECT...", args, usedTime, err, "k1", "v1")
```

### 任务与监控 (Task / Monitor / Statistic)
```go
// 任务执行日志 (任务名, 耗时ms, 是否成功, 消息, 额外参数...)
logger.Task("CleanCache", 150.2, true, "Success", "deleted", 100)

// 监控告警日志 (目标, 状态码, 消息, 额外参数...)
logger.Monitor("CPU", 1, "Load too high", "usage", "95%")

// 业务指标统计 (类别, 项目, 数值, 额外参数...)
logger.Statistic("Business", "OrderCount", 100, "region", "cn")
```

### 自定义日志类型
只需嵌入 `BaseLog` 即可利用对象池和自动重置功能。
```go
type MyBusinessLog struct {
    log.BaseLog
    OrderId string
    Amount  float64
}

// 使用方式
entry := log.GetEntry(reflect.TypeOf(&MyBusinessLog{})).(*MyBusinessLog)
logger.fillBase(entry, "business")
entry.OrderId = "O123"
entry.Amount = 99.8
logger.Log(entry)
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
