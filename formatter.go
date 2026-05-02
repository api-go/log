package log

import (
	"apigo.cc/go/cast"
)

// Formatter 日志格式化接口
type Formatter interface {
	Format(data any, sensitiveKeys []string) ([]byte, error)
}

// JSONFormatter 默认的 JSON 格式化器
type JSONFormatter struct{}

func (f *JSONFormatter) Format(data any, sensitiveKeys []string) ([]byte, error) {
	if len(sensitiveKeys) > 0 {
		return cast.ToJSONDesensitizeBytes(data, sensitiveKeys)
	}
	return cast.ToJSONBytes(data)
}

// TextFormatter 文本格式化器 (示例)
type TextFormatter struct{}

func (f *TextFormatter) Format(data any, sensitiveKeys []string) ([]byte, error) {
	// 简单的文本格式化实现
	return []byte(cast.String(data)), nil
}
