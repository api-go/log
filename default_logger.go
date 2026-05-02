package log

import (
	"apigo.cc/go/config"
)

var DefaultLogger *Logger

func init() {
	RegisterWriterMaker("es", NewESWriter)
	RegisterWriterMaker("ess", NewESWriter)

	var conf Config
	_ = config.Load("log", &conf)
	DefaultLogger = NewLogger(conf)
}

// New 创建带有 traceId 的 Logger 副本
func New(traceId string) *Logger {
	return DefaultLogger.New(traceId)
}
