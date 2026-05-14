package log

import (
	"apigo.cc/go/config"
)

var DefaultLogger *Logger

func init() {
	RegisterWriterMaker("es", newESWriter)
	RegisterWriterMaker("ess", newESWriter)

	var conf Config
	_ = config.Load(&conf, "log")
	DefaultLogger = NewLogger(conf)
}

// New 创建带有 traceId 的 Logger 副本
func New(traceId string) *Logger {
	return DefaultLogger.New(traceId)
}

// SetDefaultName 设置全局默认应用名称，并同步更新 DefaultLogger
func SetDefaultName(name string) {
	if name == "" {
		return
	}
	globalDefaultName = name
	if DefaultLogger != nil {
		DefaultLogger.SetName(name)
	}
}

// As 仿照 cast.As，使用 DefaultLogger 记录错误并返回零值 (消除摩擦)
func As[T any](v T, err error) T {
	if err != nil {
		if DefaultLogger != nil {
			DefaultLogger.Error(err.Error())
		}
		var zero T
		return zero
	}
	return v
}

// Debug 记录一条调试级别日志
func Debug(message string, extra ...any) {
	DefaultLogger.Debug(message, extra...)
}

// Info 记录一条信息级别日志
func Info(message string, extra ...any) {
	DefaultLogger.Info(message, extra...)
}

// Warning 记录一条警告级别日志
func Warning(message string, extra ...any) {
	DefaultLogger.Warning(message, extra...)
}

// Error 记录一条错误级别日志
func Error(message string, extra ...any) {
	DefaultLogger.Error(message, extra...)
}
