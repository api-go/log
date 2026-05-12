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
