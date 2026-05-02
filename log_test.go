package log

import (
	"testing"
)

func TestLogger(t *testing.T) {
	conf := Config{
		Name:  "test-app",
		Level: "debug",
	}
	logger := NewLogger(conf)

	// 测试 Info 日志
	logger.Info("hello", "key1", "value1")
}

func TestDesensitization(t *testing.T) {
	logger := NewLogger(Config{
		Sensitive: "phone",
	})
	
	data := map[string]any{
		"phone": "13812345678",
	}
	logger.Log(data) // 应该在输出中脱敏
}
