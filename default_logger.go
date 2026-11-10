package log

import (
	"github.com/api-go/config"
)

var (
	defaultConfig Config
	DefaultLogger *Logger
)

func loadDefaultConfig() Config {
	conf := defaultConfig
	_ = config.Load(&conf, "log")
	return conf
}

func init() {
	RegisterWriterMaker("es", newESWriter)
	RegisterWriterMaker("ess", newESWriter)

	setDefaultLogger(loadDefaultConfig())
}

func setDefaultLogger(conf Config) {
	DefaultLogger = NewLogger(conf)
	DefaultLogger.RedirectStdLog()
}

// SetDefaultConfig sets application defaults, then applies file and environment
// overrides before rebuilding DefaultLogger.
func SetDefaultConfig(conf Config) {
	defaultConfig = conf
	finalConfig := normalizeConfig(loadDefaultConfig())
	if DefaultLogger != nil && DefaultLogger.config == finalConfig {
		return
	}
	setDefaultLogger(finalConfig)
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
